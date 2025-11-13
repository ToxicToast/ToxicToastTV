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

type CharacterUseCase struct {
	repo          repository.CharacterRepository
	blizzardClient *blizzard.Client
}

func NewCharacterUseCase(repo repository.CharacterRepository, blizzardClient *blizzard.Client) *CharacterUseCase {
	return &CharacterUseCase{
		repo:          repo,
		blizzardClient: blizzardClient,
	}
}

func (uc *CharacterUseCase) CreateCharacter(ctx context.Context, name, realm, region string) (*domain.Character, error) {
	// Check if character already exists
	existing, _ := uc.repo.FindByNameRealmRegion(ctx, name, realm, region)
	if existing != nil {
		return nil, errors.New("character already exists")
	}

	// TODO: Fetch from Blizzard API when implemented
	// For now, create with placeholder data
	character := &domain.Character{
		ID:        uuid.New().String(),
		Name:      name,
		Realm:     realm,
		Region:    region,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return uc.repo.Create(ctx, character)
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

	// TODO: Update character details (guild_id belongs to CharacterDetails now)
	// Need to implement CharacterDetailsRepository to update the guild_id
	// For now, just update the character timestamp
	character.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, character)
}

func (uc *CharacterUseCase) DeleteCharacter(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *CharacterUseCase) RefreshCharacter(ctx context.Context, id string) (*domain.Character, error) {
	character, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: Fetch fresh data from Blizzard API when implemented
	// blizzardData, err := uc.blizzardClient.GetCharacter(ctx, character.Name, character.Realm, character.Region)
	// if err != nil {
	//     return nil, err
	// }
	// Update character details with fresh data (LastSyncedAt belongs to CharacterDetails now)

	character.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, character)
}

func (uc *CharacterUseCase) GetCharacterEquipment(ctx context.Context, characterID string) (*domain.CharacterEquipment, error) {
	// TODO: Implement equipment fetching from Blizzard API
	return nil, errors.New("character equipment not yet implemented")
}

func (uc *CharacterUseCase) GetCharacterStats(ctx context.Context, characterID string) (*domain.CharacterStats, error) {
	// TODO: Implement stats fetching from Blizzard API
	return nil, errors.New("character stats not yet implemented")
}
