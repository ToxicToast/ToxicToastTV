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

// CreateGuildCommand creates a new guild by fetching data from Blizzard API
type CreateGuildCommand struct {
	cqrs.BaseCommand
	Name   string `json:"name"`
	Realm  string `json:"realm"`
	Region string `json:"region"`
}

func (c *CreateGuildCommand) CommandName() string {
	return "create_guild"
}

func (c *CreateGuildCommand) Validate() error {
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

// UpdateGuildCommand updates a guild
type UpdateGuildCommand struct {
	cqrs.BaseCommand
}

func (c *UpdateGuildCommand) CommandName() string {
	return "update_guild"
}

func (c *UpdateGuildCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("guild_id is required")
	}
	return nil
}

// DeleteGuildCommand soft-deletes a guild
type DeleteGuildCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteGuildCommand) CommandName() string {
	return "delete_guild"
}

func (c *DeleteGuildCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("guild_id is required")
	}
	return nil
}

// RefreshGuildCommand refreshes guild data from Blizzard API
type RefreshGuildCommand struct {
	cqrs.BaseCommand
}

func (c *RefreshGuildCommand) CommandName() string {
	return "refresh_guild"
}

func (c *RefreshGuildCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("guild_id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

// CreateGuildHandler handles guild creation
type CreateGuildHandler struct {
	repo           repository.GuildRepository
	factionRepo    repository.FactionRepository
	blizzardClient *blizzard.Client
	kafkaProducer  *kafka.Producer
}

func NewCreateGuildHandler(
	repo repository.GuildRepository,
	factionRepo repository.FactionRepository,
	blizzardClient *blizzard.Client,
	kafkaProducer *kafka.Producer,
) *CreateGuildHandler {
	return &CreateGuildHandler{
		repo:           repo,
		factionRepo:    factionRepo,
		blizzardClient: blizzardClient,
		kafkaProducer:  kafkaProducer,
	}
}

func (h *CreateGuildHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateGuildCommand)

	// Check if guild already exists
	existing, _ := h.repo.FindByNameRealmRegion(ctx, createCmd.Name, createCmd.Realm, createCmd.Region)
	if existing != nil {
		return errors.New("guild already exists")
	}

	// Fetch from Blizzard API
	profile, err := h.blizzardClient.GetGuild(ctx, createCmd.Name, createCmd.Realm, createCmd.Region)
	if err != nil {
		return fmt.Errorf("failed to fetch guild from Blizzard API: %w", err)
	}

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

	guild, err = h.repo.Create(ctx, guild)
	if err != nil {
		return fmt.Errorf("failed to create guild: %w", err)
	}

	// Set AggregateID for the command
	createCmd.AggregateID = guild.ID

	// Publish guild created event
	if h.kafkaProducer != nil {
		event := kafka.WarcraftGuildCreatedEvent{
			GuildID:   guild.ID,
			Name:      guild.Name,
			Realm:     guild.Realm,
			Region:    guild.Region,
			Faction:   faction.Name,
			CreatedAt: guild.CreatedAt,
		}
		if err := h.kafkaProducer.PublishWarcraftGuildCreated("warcraft.guild.created", event); err != nil {
			fmt.Printf("Warning: Failed to publish guild created event: %v\n", err)
		}
	}

	return nil
}

// UpdateGuildHandler handles guild updates
type UpdateGuildHandler struct {
	repo repository.GuildRepository
}

func NewUpdateGuildHandler(repo repository.GuildRepository) *UpdateGuildHandler {
	return &UpdateGuildHandler{
		repo: repo,
	}
}

func (h *UpdateGuildHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateGuildCommand)

	guild, err := h.repo.FindByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("guild not found: %w", err)
	}

	guild.UpdatedAt = time.Now()
	_, err = h.repo.Update(ctx, guild)
	if err != nil {
		return fmt.Errorf("failed to update guild: %w", err)
	}

	return nil
}

// DeleteGuildHandler handles guild deletion
type DeleteGuildHandler struct {
	repo          repository.GuildRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteGuildHandler(
	repo repository.GuildRepository,
	kafkaProducer *kafka.Producer,
) *DeleteGuildHandler {
	return &DeleteGuildHandler{
		repo:          repo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteGuildHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteGuildCommand)

	// Get guild for event
	guild, err := h.repo.FindByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("guild not found: %w", err)
	}

	// Delete guild
	if err := h.repo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete guild: %w", err)
	}

	// Publish guild deleted event
	if h.kafkaProducer != nil {
		event := kafka.WarcraftGuildDeletedEvent{
			GuildID:   guild.ID,
			Name:      guild.Name,
			Realm:     guild.Realm,
			Region:    guild.Region,
			DeletedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishWarcraftGuildDeleted("warcraft.guild.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish guild deleted event: %v\n", err)
		}
	}

	return nil
}

// RefreshGuildHandler handles guild refresh from Blizzard API
type RefreshGuildHandler struct {
	repo           repository.GuildRepository
	factionRepo    repository.FactionRepository
	blizzardClient *blizzard.Client
	kafkaProducer  *kafka.Producer
}

func NewRefreshGuildHandler(
	repo repository.GuildRepository,
	factionRepo repository.FactionRepository,
	blizzardClient *blizzard.Client,
	kafkaProducer *kafka.Producer,
) *RefreshGuildHandler {
	return &RefreshGuildHandler{
		repo:           repo,
		factionRepo:    factionRepo,
		blizzardClient: blizzardClient,
		kafkaProducer:  kafkaProducer,
	}
}

func (h *RefreshGuildHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	refreshCmd := cmd.(*RefreshGuildCommand)

	guild, err := h.repo.FindByID(ctx, refreshCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("guild not found: %w", err)
	}

	// Fetch fresh data from Blizzard API
	profile, err := h.blizzardClient.GetGuild(ctx, guild.Name, guild.Realm, guild.Region)
	if err != nil {
		return fmt.Errorf("failed to refresh guild from Blizzard API: %w", err)
	}

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

	// Update guild
	now := time.Now()
	guild.FactionID = faction.ID
	guild.MemberCount = profile.MemberCount
	guild.AchievementPoints = profile.AchievementPoints
	guild.LastSyncedAt = &now
	guild.UpdatedAt = now

	guild, err = h.repo.Update(ctx, guild)
	if err != nil {
		return fmt.Errorf("failed to update guild: %w", err)
	}

	// Publish guild synced event
	if h.kafkaProducer != nil {
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
		if err := h.kafkaProducer.PublishWarcraftGuildSynced("warcraft.guild.synced", event); err != nil {
			fmt.Printf("Warning: Failed to publish guild synced event: %v\n", err)
		}
	}

	return nil
}
