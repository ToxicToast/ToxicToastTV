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

type GuildUseCase struct {
	repo           repository.GuildRepository
	factionRepo    repository.FactionRepository
	blizzardClient *blizzard.Client
	kafkaProducer  *kafka.Producer
}

func NewGuildUseCase(
	repo repository.GuildRepository,
	factionRepo repository.FactionRepository,
	blizzardClient *blizzard.Client,
	kafkaProducer *kafka.Producer,
) *GuildUseCase {
	return &GuildUseCase{
		repo:           repo,
		factionRepo:    factionRepo,
		blizzardClient: blizzardClient,
		kafkaProducer:  kafkaProducer,
	}
}

func (uc *GuildUseCase) CreateGuild(ctx context.Context, name, realm, region string) (*domain.Guild, error) {
	// Check if guild already exists
	existing, _ := uc.repo.FindByNameRealmRegion(ctx, name, realm, region)
	if existing != nil {
		return nil, errors.New("guild already exists")
	}

	// Fetch from Blizzard API
	profile, err := uc.blizzardClient.GetGuild(ctx, name, realm, region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guild from Blizzard API: %w", err)
	}

	// Get or create Faction
	factionKey := strings.ToLower(profile.FactionType)
	faction, err := uc.factionRepo.FindByKey(ctx, factionKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to find faction: %w", err)
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
			return nil, fmt.Errorf("failed to create faction: %w", err)
		}
	}

	// Create Guild
	now := time.Now()
	guild := &domain.Guild{
		ID:                uuid.New().String(),
		Name:              profile.Name,
		Realm:             profile.Realm,
		Region:            profile.Region,
		FactionID:         faction.ID,
		MemberCount:       profile.MemberCount,
		AchievementPoints: profile.AchievementPoints,
		LastSyncedAt:      &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	guild, err = uc.repo.Create(ctx, guild)
	if err != nil {
		return nil, err
	}

	// Publish guild created event
	if uc.kafkaProducer != nil {
		event := kafka.WarcraftGuildCreatedEvent{
			GuildID:   guild.ID,
			Name:      guild.Name,
			Realm:     guild.Realm,
			Region:    guild.Region,
			Faction:   faction.Name,
			CreatedAt: guild.CreatedAt,
		}
		if err := uc.kafkaProducer.PublishWarcraftGuildCreated("warcraft.guild.created", event); err != nil {
			fmt.Printf("Warning: Failed to publish guild created event: %v\n", err)
		}
	}

	return guild, nil
}

func (uc *GuildUseCase) GetGuild(ctx context.Context, id string) (*domain.Guild, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *GuildUseCase) ListGuilds(ctx context.Context, page, pageSize int, region, realm, faction *string) ([]*domain.Guild, int, error) {
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

func (uc *GuildUseCase) UpdateGuild(ctx context.Context, id string) (*domain.Guild, error) {
	guild, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	guild.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, guild)
}

func (uc *GuildUseCase) DeleteGuild(ctx context.Context, id string) error {
	// Get guild for event
	guild, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete guild
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish guild deleted event
	if uc.kafkaProducer != nil {
		event := kafka.WarcraftGuildDeletedEvent{
			GuildID:   guild.ID,
			Name:      guild.Name,
			Realm:     guild.Realm,
			Region:    guild.Region,
			DeletedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishWarcraftGuildDeleted("warcraft.guild.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish guild deleted event: %v\n", err)
		}
	}

	return nil
}

func (uc *GuildUseCase) RefreshGuild(ctx context.Context, id string) (*domain.Guild, error) {
	guild, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Fetch fresh data from Blizzard API
	profile, err := uc.blizzardClient.GetGuild(ctx, guild.Name, guild.Realm, guild.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh guild from Blizzard API: %w", err)
	}

	// Get or create Faction
	factionKey := strings.ToLower(profile.FactionType)
	faction, err := uc.factionRepo.FindByKey(ctx, factionKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to find faction: %w", err)
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
			return nil, fmt.Errorf("failed to create faction: %w", err)
		}
	}

	// Update guild
	now := time.Now()
	guild.FactionID = faction.ID
	guild.MemberCount = profile.MemberCount
	guild.AchievementPoints = profile.AchievementPoints
	guild.LastSyncedAt = &now
	guild.UpdatedAt = now

	guild, err = uc.repo.Update(ctx, guild)
	if err != nil {
		return nil, err
	}

	// Publish guild synced event
	if uc.kafkaProducer != nil {
		event := kafka.WarcraftGuildSyncedEvent{
			GuildID:           guild.ID,
			Name:              guild.Name,
			Realm:             guild.Realm,
			Region:            guild.Region,
			Faction:           faction.Name,
			MemberCount:       guild.MemberCount,
			AchievementPoints: guild.AchievementPoints,
			SyncedAt:          now,
		}
		if err := uc.kafkaProducer.PublishWarcraftGuildSynced("warcraft.guild.synced", event); err != nil {
			fmt.Printf("Warning: Failed to publish guild synced event: %v\n", err)
		}
	}

	return guild, nil
}

func (uc *GuildUseCase) GetGuildRoster(ctx context.Context, guildID string, page, pageSize int) ([]domain.GuildMember, int, error) {
	guild, err := uc.repo.FindByID(ctx, guildID)
	if err != nil {
		return nil, 0, err
	}

	// Fetch roster from Blizzard API
	members, err := uc.blizzardClient.GetGuildRoster(ctx, guild.Name, guild.Realm, guild.Region)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch roster from Blizzard API: %w", err)
	}

	// Apply pagination
	total := len(members)
	start := (page - 1) * pageSize
	if start >= total {
		return []domain.GuildMember{}, total, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return members[start:end], total, nil
}
