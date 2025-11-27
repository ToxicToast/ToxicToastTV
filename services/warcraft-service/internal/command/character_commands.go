package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/repository"
	"toxictoast/services/warcraft-service/pkg/blizzard"
)

// ============================================================================
// Commands
// ============================================================================

// CreateCharacterCommand creates a new character by fetching data from Blizzard API
type CreateCharacterCommand struct {
	cqrs.BaseCommand
	Name   string `json:"name"`
	Realm  string `json:"realm"`
	Region string `json:"region"`
}

func (c *CreateCharacterCommand) CommandName() string {
	return "create_character"
}

func (c *CreateCharacterCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Realm == "" {
		return errors.New("realm is required")
	}
	if c.Region == "" {
		return errors.New("region is required")
	}
	return nil
}

// UpdateCharacterCommand updates a character's guild assignment
type UpdateCharacterCommand struct {
	cqrs.BaseCommand
	GuildID *string `json:"guild_id"`
}

func (c *UpdateCharacterCommand) CommandName() string {
	return "update_character"
}

func (c *UpdateCharacterCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("character_id is required")
	}
	return nil
}

// DeleteCharacterCommand soft-deletes a character and related data
type DeleteCharacterCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteCharacterCommand) CommandName() string {
	return "delete_character"
}

func (c *DeleteCharacterCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("character_id is required")
	}
	return nil
}

// RefreshCharacterCommand refreshes character data from Blizzard API
type RefreshCharacterCommand struct {
	cqrs.BaseCommand
}

func (c *RefreshCharacterCommand) CommandName() string {
	return "refresh_character"
}

func (c *RefreshCharacterCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("character_id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

// CreateCharacterHandler handles character creation
type CreateCharacterHandler struct {
	repo          repository.CharacterRepository
	detailsRepo   repository.CharacterDetailsRepository
	raceRepo      repository.RaceRepository
	classRepo     repository.ClassRepository
	factionRepo   repository.FactionRepository
	guildRepo     repository.GuildRepository
	blizzardClient *blizzard.Client
	kafkaProducer *kafka.Producer
}

func NewCreateCharacterHandler(
	repo repository.CharacterRepository,
	detailsRepo repository.CharacterDetailsRepository,
	raceRepo repository.RaceRepository,
	classRepo repository.ClassRepository,
	factionRepo repository.FactionRepository,
	guildRepo repository.GuildRepository,
	blizzardClient *blizzard.Client,
	kafkaProducer *kafka.Producer,
) *CreateCharacterHandler {
	return &CreateCharacterHandler{
		repo:          repo,
		detailsRepo:   detailsRepo,
		raceRepo:      raceRepo,
		classRepo:     classRepo,
		factionRepo:   factionRepo,
		guildRepo:     guildRepo,
		blizzardClient: blizzardClient,
		kafkaProducer: kafkaProducer,
	}
}

func (h *CreateCharacterHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateCharacterCommand)

	// Check if character already exists
	existing, _ := h.repo.FindByNameRealmRegion(ctx, createCmd.Name, createCmd.Realm, createCmd.Region)
	if existing != nil {
		return errors.New("character already exists")
	}

	// Fetch from Blizzard API
	profile, err := h.blizzardClient.GetCharacter(ctx, createCmd.Name, createCmd.Realm, createCmd.Region)
	if err != nil {
		return fmt.Errorf("failed to fetch character from Blizzard API: %w", err)
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

	character, err = h.repo.Create(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	// Set AggregateID for the command
	createCmd.AggregateID = character.ID

	// Create CharacterDetails with reference data
	if err := h.createOrUpdateCharacterDetails(ctx, character.ID, profile); err != nil {
		// Rollback character creation
		_ = h.repo.Delete(ctx, character.ID)
		return fmt.Errorf("failed to create character details: %w", err)
	}

	// Publish character created event
	if h.kafkaProducer != nil {
		event := kafka.WarcraftCharacterCreatedEvent{
			CharacterID: character.ID,
			Name:        character.Name,
			Realm:       character.Realm,
			Region:      character.Region,
			CreatedAt:   character.CreatedAt,
		}
		if err := h.kafkaProducer.PublishWarcraftCharacterCreated("warcraft.character.created", event); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: Failed to publish character created event: %v\n", err)
		}
	}

	return nil
}

// Helper function to create or update character details with reference data
func (h *CreateCharacterHandler) createOrUpdateCharacterDetails(ctx context.Context, characterID string, profile *blizzard.CharacterProfile) error {
	// Get or create Faction
	factionKey := strings.ToLower(profile.FactionType)
	faction, err := h.factionRepo.FindByKey(ctx, factionKey)
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
		faction, err = h.factionRepo.Create(ctx, faction)
		if err != nil {
			return fmt.Errorf("failed to create faction: %w", err)
		}
	}

	// Get or create Race
	raceKey := strings.ToLower(strings.ReplaceAll(profile.RaceName, " ", "-"))
	race, err := h.raceRepo.FindByKey(ctx, raceKey)
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
		race, err = h.raceRepo.Create(ctx, race)
		if err != nil {
			return fmt.Errorf("failed to create race: %w", err)
		}
	}

	// Get or create Class
	classKey := strings.ToLower(strings.ReplaceAll(profile.ClassName, " ", "-"))
	class, err := h.classRepo.FindByKey(ctx, classKey)
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
		class, err = h.classRepo.Create(ctx, class)
		if err != nil {
			return fmt.Errorf("failed to create class: %w", err)
		}
	}

	// Find guild if character has one
	var guildID *string
	if profile.GuildName != nil && profile.GuildRealm != nil {
		guild, err := h.guildRepo.FindByNameRealmRegion(ctx, *profile.GuildName, *profile.GuildRealm, profile.Region)
		if err == nil && guild != nil {
			guildID = &guild.ID
		}
	}

	// Check if details already exist
	now := time.Now()
	existing, err := h.detailsRepo.FindByCharacterID(ctx, characterID)
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

		_, err = h.detailsRepo.Update(ctx, existing)
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

	_, err = h.detailsRepo.Create(ctx, details)
	return err
}

// UpdateCharacterHandler handles character updates
type UpdateCharacterHandler struct {
	repo        repository.CharacterRepository
	detailsRepo repository.CharacterDetailsRepository
}

func NewUpdateCharacterHandler(
	repo repository.CharacterRepository,
	detailsRepo repository.CharacterDetailsRepository,
) *UpdateCharacterHandler {
	return &UpdateCharacterHandler{
		repo:        repo,
		detailsRepo: detailsRepo,
	}
}

func (h *UpdateCharacterHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateCharacterCommand)

	character, err := h.repo.FindByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Update character details with guild_id
	if updateCmd.GuildID != nil {
		details, err := h.detailsRepo.FindByCharacterID(ctx, updateCmd.AggregateID)
		if err != nil {
			return fmt.Errorf("failed to find character details: %w", err)
		}

		details.GuildID = updateCmd.GuildID
		details.UpdatedAt = time.Now()

		_, err = h.detailsRepo.Update(ctx, details)
		if err != nil {
			return fmt.Errorf("failed to update character details: %w", err)
		}
	}

	character.UpdatedAt = time.Now()
	_, err = h.repo.Update(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	return nil
}

// DeleteCharacterHandler handles character deletion
type DeleteCharacterHandler struct {
	repo          repository.CharacterRepository
	detailsRepo   repository.CharacterDetailsRepository
	equipmentRepo repository.CharacterEquipmentRepository
	statsRepo     repository.CharacterStatsRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteCharacterHandler(
	repo repository.CharacterRepository,
	detailsRepo repository.CharacterDetailsRepository,
	equipmentRepo repository.CharacterEquipmentRepository,
	statsRepo repository.CharacterStatsRepository,
	kafkaProducer *kafka.Producer,
) *DeleteCharacterHandler {
	return &DeleteCharacterHandler{
		repo:          repo,
		detailsRepo:   detailsRepo,
		equipmentRepo: equipmentRepo,
		statsRepo:     statsRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteCharacterHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteCharacterCommand)

	// Get character for event
	character, err := h.repo.FindByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Delete related data first
	_ = h.detailsRepo.Delete(ctx, deleteCmd.AggregateID)
	_ = h.equipmentRepo.Delete(ctx, deleteCmd.AggregateID)
	_ = h.statsRepo.Delete(ctx, deleteCmd.AggregateID)

	// Delete character (soft delete)
	if err := h.repo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	// Publish character deleted event
	if h.kafkaProducer != nil {
		event := kafka.WarcraftCharacterDeletedEvent{
			CharacterID: character.ID,
			Name:        character.Name,
			Realm:       character.Realm,
			Region:      character.Region,
			DeletedAt:   time.Now(),
		}
		if err := h.kafkaProducer.PublishWarcraftCharacterDeleted("warcraft.character.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish character deleted event: %v\n", err)
		}
	}

	return nil
}

// RefreshCharacterHandler handles character refresh from Blizzard API
type RefreshCharacterHandler struct {
	repo          repository.CharacterRepository
	detailsRepo   repository.CharacterDetailsRepository
	raceRepo      repository.RaceRepository
	classRepo     repository.ClassRepository
	factionRepo   repository.FactionRepository
	guildRepo     repository.GuildRepository
	blizzardClient *blizzard.Client
	kafkaProducer *kafka.Producer
}

func NewRefreshCharacterHandler(
	repo repository.CharacterRepository,
	detailsRepo repository.CharacterDetailsRepository,
	raceRepo repository.RaceRepository,
	classRepo repository.ClassRepository,
	factionRepo repository.FactionRepository,
	guildRepo repository.GuildRepository,
	blizzardClient *blizzard.Client,
	kafkaProducer *kafka.Producer,
) *RefreshCharacterHandler {
	return &RefreshCharacterHandler{
		repo:          repo,
		detailsRepo:   detailsRepo,
		raceRepo:      raceRepo,
		classRepo:     classRepo,
		factionRepo:   factionRepo,
		guildRepo:     guildRepo,
		blizzardClient: blizzardClient,
		kafkaProducer: kafkaProducer,
	}
}

func (h *RefreshCharacterHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	refreshCmd := cmd.(*RefreshCharacterCommand)

	character, err := h.repo.FindByID(ctx, refreshCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Fetch fresh data from Blizzard API
	profile, err := h.blizzardClient.GetCharacter(ctx, character.Name, character.Realm, character.Region)
	if err != nil {
		return fmt.Errorf("failed to refresh character from Blizzard API: %w", err)
	}

	// Update character details
	if err := h.createOrUpdateCharacterDetails(ctx, character.ID, profile); err != nil {
		return fmt.Errorf("failed to update character details: %w", err)
	}

	character.UpdatedAt = time.Now()
	_, err = h.repo.Update(ctx, character)
	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	// Publish character synced event
	if h.kafkaProducer != nil {
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
		if err := h.kafkaProducer.PublishWarcraftCharacterSynced("warcraft.character.synced", event); err != nil {
			fmt.Printf("Warning: Failed to publish character synced event: %v\n", err)
		}
	}

	return nil
}

// Helper function to create or update character details with reference data
func (h *RefreshCharacterHandler) createOrUpdateCharacterDetails(ctx context.Context, characterID string, profile *blizzard.CharacterProfile) error {
	// Get or create Faction
	factionKey := strings.ToLower(profile.FactionType)
	faction, err := h.factionRepo.FindByKey(ctx, factionKey)
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
		faction, err = h.factionRepo.Create(ctx, faction)
		if err != nil {
			return fmt.Errorf("failed to create faction: %w", err)
		}
	}

	// Get or create Race
	raceKey := strings.ToLower(strings.ReplaceAll(profile.RaceName, " ", "-"))
	race, err := h.raceRepo.FindByKey(ctx, raceKey)
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
		race, err = h.raceRepo.Create(ctx, race)
		if err != nil {
			return fmt.Errorf("failed to create race: %w", err)
		}
	}

	// Get or create Class
	classKey := strings.ToLower(strings.ReplaceAll(profile.ClassName, " ", "-"))
	class, err := h.classRepo.FindByKey(ctx, classKey)
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
		class, err = h.classRepo.Create(ctx, class)
		if err != nil {
			return fmt.Errorf("failed to create class: %w", err)
		}
	}

	// Find guild if character has one
	var guildID *string
	if profile.GuildName != nil && profile.GuildRealm != nil {
		guild, err := h.guildRepo.FindByNameRealmRegion(ctx, *profile.GuildName, *profile.GuildRealm, profile.Region)
		if err == nil && guild != nil {
			guildID = &guild.ID
		}
	}

	// Check if details already exist
	now := time.Now()
	existing, err := h.detailsRepo.FindByCharacterID(ctx, characterID)
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

		_, err = h.detailsRepo.Update(ctx, existing)
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

	_, err = h.detailsRepo.Create(ctx, details)
	return err
}
