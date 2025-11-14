package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/toxictoast/toxictoastgo/shared/kafka"
	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/repository"
	"toxictoast/services/warcraft-service/pkg/blizzard"
)

type CharacterUseCase struct {
	repo              repository.CharacterRepository
	detailsRepo       repository.CharacterDetailsRepository
	equipmentRepo     repository.CharacterEquipmentRepository
	statsRepo         repository.CharacterStatsRepository
	raceRepo          repository.RaceRepository
	classRepo         repository.ClassRepository
	factionRepo       repository.FactionRepository
	guildRepo         repository.GuildRepository
	blizzardClient    *blizzard.Client
	kafkaProducer     *kafka.Producer
}

func NewCharacterUseCase(
	repo repository.CharacterRepository,
	detailsRepo repository.CharacterDetailsRepository,
	equipmentRepo repository.CharacterEquipmentRepository,
	statsRepo repository.CharacterStatsRepository,
	raceRepo repository.RaceRepository,
	classRepo repository.ClassRepository,
	factionRepo repository.FactionRepository,
	guildRepo repository.GuildRepository,
	blizzardClient *blizzard.Client,
	kafkaProducer *kafka.Producer,
) *CharacterUseCase {
	return &CharacterUseCase{
		repo:              repo,
		detailsRepo:       detailsRepo,
		equipmentRepo:     equipmentRepo,
		statsRepo:         statsRepo,
		raceRepo:          raceRepo,
		classRepo:         classRepo,
		factionRepo:       factionRepo,
		guildRepo:         guildRepo,
		blizzardClient:    blizzardClient,
		kafkaProducer:     kafkaProducer,
	}
}

func (uc *CharacterUseCase) CreateCharacter(ctx context.Context, name, realm, region string) (*domain.Character, error) {
	// Check if character already exists
	existing, _ := uc.repo.FindByNameRealmRegion(ctx, name, realm, region)
	if existing != nil {
		return nil, errors.New("character already exists")
	}

	// Fetch from Blizzard API
	profile, err := uc.blizzardClient.GetCharacter(ctx, name, realm, region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch character from Blizzard API: %w", err)
	}

	// Create Character
	character := &domain.Character{
		ID:        uuid.New().String(),
		Name:      profile.Name,
		Realm:     profile.Realm,
		Region:    profile.Region,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	character, err = uc.repo.Create(ctx, character)
	if err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	// Create CharacterDetails with reference data
	if err := uc.createOrUpdateCharacterDetails(ctx, character.ID, profile); err != nil {
		// Rollback character creation
		_ = uc.repo.Delete(ctx, character.ID)
		return nil, fmt.Errorf("failed to create character details: %w", err)
	}

	// Publish character created event
	if uc.kafkaProducer != nil {
		event := kafka.WarcraftCharacterCreatedEvent{
			CharacterID: character.ID,
			Name:        character.Name,
			Realm:       character.Realm,
			Region:      character.Region,
			CreatedAt:   character.CreatedAt,
		}
		if err := uc.kafkaProducer.PublishWarcraftCharacterCreated("warcraft.character.created", event); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: Failed to publish character created event: %v\n", err)
		}
	}

	return character, nil
}

func (uc *CharacterUseCase) GetCharacter(ctx context.Context, id string) (*domain.Character, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *CharacterUseCase) ListCharacters(ctx context.Context, page, pageSize int, region, realm, faction *string) ([]*domain.Character, int, error) {
	filters := make(map[string]interface{})
	if region != nil {
		filters["region"] = *region
	}
	if realm != nil {
		filters["realm"] = *realm
	}
	if faction != nil {
		filters["faction"] = *faction
	}

	return uc.repo.List(ctx, page, pageSize, filters)
}

func (uc *CharacterUseCase) UpdateCharacter(ctx context.Context, id string, guildID *string) (*domain.Character, error) {
	character, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update character details with guild_id
	if guildID != nil {
		details, err := uc.detailsRepo.FindByCharacterID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to find character details: %w", err)
		}

		details.GuildID = guildID
		details.UpdatedAt = time.Now()

		_, err = uc.detailsRepo.Update(ctx, details)
		if err != nil {
			return nil, fmt.Errorf("failed to update character details: %w", err)
		}
	}

	character.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, character)
}

func (uc *CharacterUseCase) DeleteCharacter(ctx context.Context, id string) error {
	// Get character for event
	character, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete related data first
	_ = uc.detailsRepo.Delete(ctx, id)
	_ = uc.equipmentRepo.Delete(ctx, id)
	_ = uc.statsRepo.Delete(ctx, id)

	// Delete character (soft delete)
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish character deleted event
	if uc.kafkaProducer != nil {
		event := kafka.WarcraftCharacterDeletedEvent{
			CharacterID: character.ID,
			Name:        character.Name,
			Realm:       character.Realm,
			Region:      character.Region,
			DeletedAt:   time.Now(),
		}
		if err := uc.kafkaProducer.PublishWarcraftCharacterDeleted("warcraft.character.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish character deleted event: %v\n", err)
		}
	}

	return nil
}

func (uc *CharacterUseCase) RefreshCharacter(ctx context.Context, id string) (*domain.Character, error) {
	character, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Fetch fresh data from Blizzard API
	profile, err := uc.blizzardClient.GetCharacter(ctx, character.Name, character.Realm, character.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh character from Blizzard API: %w", err)
	}

	// Update character details
	if err := uc.createOrUpdateCharacterDetails(ctx, character.ID, profile); err != nil {
		return nil, fmt.Errorf("failed to update character details: %w", err)
	}

	character.UpdatedAt = time.Now()
	character, err = uc.repo.Update(ctx, character)
	if err != nil {
		return nil, err
	}

	// Publish character synced event
	if uc.kafkaProducer != nil {
		event := kafka.WarcraftCharacterSyncedEvent{
			CharacterID:       character.ID,
			Name:              character.Name,
			Realm:             character.Realm,
			Region:            character.Region,
			Level:             profile.Level,
			ItemLevel:         profile.ItemLevel,
			ClassName:         profile.ClassName,
			RaceName:          profile.RaceName,
			FactionName:       profile.FactionType,
			AchievementPoints: profile.AchievementPoints,
			SyncedAt:          time.Now(),
		}
		if err := uc.kafkaProducer.PublishWarcraftCharacterSynced("warcraft.character.synced", event); err != nil {
			fmt.Printf("Warning: Failed to publish character synced event: %v\n", err)
		}
	}

	return character, nil
}

func (uc *CharacterUseCase) GetCharacterEquipment(ctx context.Context, characterID string) (*domain.CharacterEquipment, error) {
	character, err := uc.repo.FindByID(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// Fetch equipment from Blizzard API
	equipment, err := uc.blizzardClient.GetCharacterEquipment(ctx, character.Name, character.Realm, character.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch equipment from Blizzard API: %w", err)
	}

	// Set character ID and save
	equipment.ID = uuid.New().String()
	equipment.CharacterID = characterID
	equipment.CreatedAt = time.Now()
	equipment.UpdatedAt = time.Now()

	// CreateOrUpdate equipment in database
	equipment, err = uc.equipmentRepo.CreateOrUpdate(ctx, equipment)
	if err != nil {
		return nil, err
	}

	// Publish equipment updated event
	if uc.kafkaProducer != nil {
		event := kafka.WarcraftCharacterEquipmentUpdatedEvent{
			CharacterID: characterID,
			Name:        character.Name,
			UpdatedAt:   equipment.UpdatedAt,
		}
		if err := uc.kafkaProducer.PublishWarcraftCharacterEquipmentUpdated("warcraft.character.equipment.updated", event); err != nil {
			fmt.Printf("Warning: Failed to publish equipment updated event: %v\n", err)
		}
	}

	return equipment, nil
}

func (uc *CharacterUseCase) GetCharacterStats(ctx context.Context, characterID string) (*domain.CharacterStats, error) {
	character, err := uc.repo.FindByID(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// Fetch stats from Blizzard API
	stats, err := uc.blizzardClient.GetCharacterStats(ctx, character.Name, character.Realm, character.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats from Blizzard API: %w", err)
	}

	// Set character ID and save
	stats.ID = uuid.New().String()
	stats.CharacterID = characterID
	stats.CreatedAt = time.Now()
	stats.UpdatedAt = time.Now()

	// CreateOrUpdate stats in database
	stats, err = uc.statsRepo.CreateOrUpdate(ctx, stats)
	if err != nil {
		return nil, err
	}

	// Publish stats updated event
	if uc.kafkaProducer != nil {
		event := kafka.WarcraftCharacterStatsUpdatedEvent{
			CharacterID: characterID,
			Name:        character.Name,
			UpdatedAt:   stats.UpdatedAt,
		}
		if err := uc.kafkaProducer.PublishWarcraftCharacterStatsUpdated("warcraft.character.stats.updated", event); err != nil {
			fmt.Printf("Warning: Failed to publish stats updated event: %v\n", err)
		}
	}

	return stats, nil
}

// Helper function to create or update character details with reference data
func (uc *CharacterUseCase) createOrUpdateCharacterDetails(ctx context.Context, characterID string, profile *blizzard.CharacterProfile) error {
	// Get or create Faction
	factionKey := strings.ToLower(profile.FactionType)
	faction, err := uc.factionRepo.FindByKey(ctx, factionKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to find faction: %w", err)
	}
	if faction == nil {
		faction = &domain.Faction{
			ID:        uuid.New().String(),
			Key:       factionKey,
			Name:      strings.Title(profile.FactionType),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		faction, err = uc.factionRepo.Create(ctx, faction)
		if err != nil {
			return fmt.Errorf("failed to create faction: %w", err)
		}
	}

	// Get or create Race
	raceKey := strings.ToLower(strings.ReplaceAll(profile.RaceName, " ", "-"))
	race, err := uc.raceRepo.FindByKey(ctx, raceKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to find race: %w", err)
	}
	if race == nil {
		race = &domain.Race{
			ID:        uuid.New().String(),
			Key:       raceKey,
			Name:      profile.RaceName,
			FactionID: faction.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		race, err = uc.raceRepo.Create(ctx, race)
		if err != nil {
			return fmt.Errorf("failed to create race: %w", err)
		}
	}

	// Get or create Class
	classKey := strings.ToLower(strings.ReplaceAll(profile.ClassName, " ", "-"))
	class, err := uc.classRepo.FindByKey(ctx, classKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to find class: %w", err)
	}
	if class == nil {
		class = &domain.Class{
			ID:        uuid.New().String(),
			Key:       classKey,
			Name:      profile.ClassName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		class, err = uc.classRepo.Create(ctx, class)
		if err != nil {
			return fmt.Errorf("failed to create class: %w", err)
		}
	}

	// Find guild if character has one
	var guildID *string
	if profile.GuildName != nil && profile.GuildRealm != nil {
		guild, err := uc.guildRepo.FindByNameRealmRegion(ctx, *profile.GuildName, *profile.GuildRealm, profile.Region)
		if err == nil && guild != nil {
			guildID = &guild.ID
		}
	}

	// Check if details already exist
	now := time.Now()
	existing, err := uc.detailsRepo.FindByCharacterID(ctx, characterID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing details: %w", err)
	}

	if existing != nil {
		// Update existing details
		existing.DisplayName = profile.DisplayName
		existing.DisplayRealm = profile.DisplayRealm
		existing.Level = profile.Level
		existing.ItemLevel = profile.ItemLevel
		existing.ClassID = class.ID
		existing.RaceID = race.ID
		existing.FactionID = faction.ID
		existing.GuildID = guildID
		existing.ThumbnailURL = profile.ThumbnailURL
		existing.AchievementPoints = profile.AchievementPoints
		existing.LastSyncedAt = &now
		existing.UpdatedAt = now

		_, err = uc.detailsRepo.Update(ctx, existing)
		return err
	}

	// Create new details
	details := &domain.CharacterDetails{
		ID:                uuid.New().String(),
		CharacterID:       characterID,
		DisplayName:       profile.DisplayName,
		DisplayRealm:      profile.DisplayRealm,
		Level:             profile.Level,
		ItemLevel:         profile.ItemLevel,
		ClassID:           class.ID,
		RaceID:            race.ID,
		FactionID:         faction.ID,
		GuildID:           guildID,
		ThumbnailURL:      profile.ThumbnailURL,
		AchievementPoints: profile.AchievementPoints,
		LastSyncedAt:      &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	_, err = uc.detailsRepo.Create(ctx, details)
	return err
}
