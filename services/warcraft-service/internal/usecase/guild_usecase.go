package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/repository"
	"toxictoast/services/warcraft-service/pkg/blizzard"
)

type GuildUseCase struct {
	repo          repository.GuildRepository
	blizzardClient *blizzard.Client
}

func NewGuildUseCase(repo repository.GuildRepository, blizzardClient *blizzard.Client) *GuildUseCase {
	return &GuildUseCase{
		repo:          repo,
		blizzardClient: blizzardClient,
	}
}

func (uc *GuildUseCase) CreateGuild(ctx context.Context, name, realm, region string) (*domain.Guild, error) {
	// Check if guild already exists
	existing, _ := uc.repo.FindByNameRealmRegion(ctx, name, realm, region)
	if existing != nil {
		return nil, errors.New("guild already exists")
	}

	// TODO: Fetch from Blizzard API when implemented
	// FactionID should be determined from API response
	guild := &domain.Guild{
		ID:                uuid.New().String(),
		Name:              name,
		Realm:             realm,
		Region:            region,
		FactionID:         "", // TODO: Determine from Blizzard API
		MemberCount:       0,
		AchievementPoints: 0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return uc.repo.Create(ctx, guild)
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
	return uc.repo.Delete(ctx, id)
}

func (uc *GuildUseCase) RefreshGuild(ctx context.Context, id string) (*domain.Guild, error) {
	guild, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: Fetch fresh data from Blizzard API when implemented
	now := time.Now()
	guild.LastSyncedAt = &now
	guild.UpdatedAt = now

	return uc.repo.Update(ctx, guild)
}

func (uc *GuildUseCase) GetGuildRoster(ctx context.Context, guildID string, page, pageSize int) ([]domain.GuildMember, int, error) {
	// TODO: Implement guild roster fetching
	return nil, 0, errors.New("guild roster not yet implemented")
}
