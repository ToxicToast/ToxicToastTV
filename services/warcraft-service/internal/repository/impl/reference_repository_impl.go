package impl

import (
	"context"

	"gorm.io/gorm"
	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/repository"
	"toxictoast/services/warcraft-service/internal/repository/entity"
	"toxictoast/services/warcraft-service/internal/repository/mapper"
)

// Race Repository

type raceRepositoryImpl struct {
	db *gorm.DB
}

func NewRaceRepository(db *gorm.DB) repository.RaceRepository {
	return &raceRepositoryImpl{db: db}
}

func (r *raceRepositoryImpl) Create(ctx context.Context, race *domain.Race) (*domain.Race, error) {
	e := mapper.RaceToEntity(race)
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return nil, err
	}
	return mapper.RaceToDomain(e), nil
}

func (r *raceRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Race, error) {
	var e entity.Race
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.RaceToDomain(&e), nil
}

func (r *raceRepositoryImpl) FindByKey(ctx context.Context, key string) (*domain.Race, error) {
	var e entity.Race
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.RaceToDomain(&e), nil
}

func (r *raceRepositoryImpl) List(ctx context.Context) ([]*domain.Race, error) {
	var entities []entity.Race
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}

	races := make([]*domain.Race, len(entities))
	for i, e := range entities {
		races[i] = mapper.RaceToDomain(&e)
	}
	return races, nil
}

func (r *raceRepositoryImpl) Update(ctx context.Context, race *domain.Race) (*domain.Race, error) {
	e := mapper.RaceToEntity(race)
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return nil, err
	}
	return mapper.RaceToDomain(e), nil
}

func (r *raceRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Race{}, "id = ?", id).Error
}

// Class Repository

type classRepositoryImpl struct {
	db *gorm.DB
}

func NewClassRepository(db *gorm.DB) repository.ClassRepository {
	return &classRepositoryImpl{db: db}
}

func (r *classRepositoryImpl) Create(ctx context.Context, class *domain.Class) (*domain.Class, error) {
	e := mapper.ClassToEntity(class)
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return nil, err
	}
	return mapper.ClassToDomain(e), nil
}

func (r *classRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Class, error) {
	var e entity.Class
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.ClassToDomain(&e), nil
}

func (r *classRepositoryImpl) FindByKey(ctx context.Context, key string) (*domain.Class, error) {
	var e entity.Class
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.ClassToDomain(&e), nil
}

func (r *classRepositoryImpl) List(ctx context.Context) ([]*domain.Class, error) {
	var entities []entity.Class
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}

	classes := make([]*domain.Class, len(entities))
	for i, e := range entities {
		classes[i] = mapper.ClassToDomain(&e)
	}
	return classes, nil
}

func (r *classRepositoryImpl) Update(ctx context.Context, class *domain.Class) (*domain.Class, error) {
	e := mapper.ClassToEntity(class)
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return nil, err
	}
	return mapper.ClassToDomain(e), nil
}

func (r *classRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Class{}, "id = ?", id).Error
}

// Faction Repository

type factionRepositoryImpl struct {
	db *gorm.DB
}

func NewFactionRepository(db *gorm.DB) repository.FactionRepository {
	return &factionRepositoryImpl{db: db}
}

func (r *factionRepositoryImpl) Create(ctx context.Context, faction *domain.Faction) (*domain.Faction, error) {
	e := mapper.FactionToEntity(faction)
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return nil, err
	}
	return mapper.FactionToDomain(e), nil
}

func (r *factionRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Faction, error) {
	var e entity.Faction
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.FactionToDomain(&e), nil
}

func (r *factionRepositoryImpl) FindByKey(ctx context.Context, key string) (*domain.Faction, error) {
	var e entity.Faction
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&e).Error; err != nil {
		return nil, err
	}
	return mapper.FactionToDomain(&e), nil
}

func (r *factionRepositoryImpl) List(ctx context.Context) ([]*domain.Faction, error) {
	var entities []entity.Faction
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}

	factions := make([]*domain.Faction, len(entities))
	for i, e := range entities {
		factions[i] = mapper.FactionToDomain(&e)
	}
	return factions, nil
}

func (r *factionRepositoryImpl) Update(ctx context.Context, faction *domain.Faction) (*domain.Faction, error) {
	e := mapper.FactionToEntity(faction)
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return nil, err
	}
	return mapper.FactionToDomain(e), nil
}

func (r *factionRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Faction{}, "id = ?", id).Error
}
