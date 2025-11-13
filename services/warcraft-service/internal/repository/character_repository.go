package repository

import (
	"context"
	"toxictoast/services/warcraft-service/internal/domain"
)

type CharacterRepository interface {
	Create(ctx context.Context, character *domain.Character) (*domain.Character, error)
	FindByID(ctx context.Context, id string) (*domain.Character, error)
	FindByNameRealmRegion(ctx context.Context, name, realm, region string) (*domain.Character, error)
	List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*domain.Character, int, error)
	Update(ctx context.Context, character *domain.Character) (*domain.Character, error)
	Delete(ctx context.Context, id string) error
}

type CharacterDetailsRepository interface {
	Create(ctx context.Context, details *domain.CharacterDetails) (*domain.CharacterDetails, error)
	FindByCharacterID(ctx context.Context, characterID string) (*domain.CharacterDetails, error)
	Update(ctx context.Context, details *domain.CharacterDetails) (*domain.CharacterDetails, error)
	Delete(ctx context.Context, characterID string) error
}

type GuildRepository interface {
	Create(ctx context.Context, guild *domain.Guild) (*domain.Guild, error)
	FindByID(ctx context.Context, id string) (*domain.Guild, error)
	FindByNameRealmRegion(ctx context.Context, name, realm, region string) (*domain.Guild, error)
	List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*domain.Guild, int, error)
	Update(ctx context.Context, guild *domain.Guild) (*domain.Guild, error)
	Delete(ctx context.Context, id string) error
}
