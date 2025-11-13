package repository

import (
	"context"
	"toxictoast/services/warcraft-service/internal/domain"
)

type RaceRepository interface {
	Create(ctx context.Context, race *domain.Race) (*domain.Race, error)
	FindByID(ctx context.Context, id string) (*domain.Race, error)
	FindByKey(ctx context.Context, key string) (*domain.Race, error)
	List(ctx context.Context) ([]*domain.Race, error)
	Update(ctx context.Context, race *domain.Race) (*domain.Race, error)
	Delete(ctx context.Context, id string) error
}

type ClassRepository interface {
	Create(ctx context.Context, class *domain.Class) (*domain.Class, error)
	FindByID(ctx context.Context, id string) (*domain.Class, error)
	FindByKey(ctx context.Context, key string) (*domain.Class, error)
	List(ctx context.Context) ([]*domain.Class, error)
	Update(ctx context.Context, class *domain.Class) (*domain.Class, error)
	Delete(ctx context.Context, id string) error
}

type FactionRepository interface {
	Create(ctx context.Context, faction *domain.Faction) (*domain.Faction, error)
	FindByID(ctx context.Context, id string) (*domain.Faction, error)
	FindByKey(ctx context.Context, key string) (*domain.Faction, error)
	List(ctx context.Context) ([]*domain.Faction, error)
	Update(ctx context.Context, faction *domain.Faction) (*domain.Faction, error)
	Delete(ctx context.Context, id string) error
}
