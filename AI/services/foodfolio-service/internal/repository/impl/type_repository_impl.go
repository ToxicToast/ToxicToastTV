package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
	"toxictoast/services/foodfolio-service/internal/repository/mapper"
)

type typeRepository struct {
	db *gorm.DB
}

// NewTypeRepository creates a new type repository instance
func NewTypeRepository(db *gorm.DB) interfaces.TypeRepository {
	return &typeRepository{db: db}
}

func (r *typeRepository) Create(ctx context.Context, typeEntity *domain.Type) error {
	typeEntity.Slug = generateSlug(typeEntity.Name)
	e := mapper.TypeToEntity(typeEntity)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *typeRepository) GetByID(ctx context.Context, id string) (*domain.Type, error) {
	var e entity.TypeEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.TypeToDomain(&e), nil
}

func (r *typeRepository) GetBySlug(ctx context.Context, slug string) (*domain.Type, error) {
	var e entity.TypeEntity
	err := r.db.WithContext(ctx).First(&e, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.TypeToDomain(&e), nil
}

func (r *typeRepository) List(ctx context.Context, offset, limit int, search string, includeDeleted bool) ([]*domain.Type, int64, error) {
	var entities []*entity.TypeEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.TypeEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.TypesToDomain(entities), total, nil
}

func (r *typeRepository) Update(ctx context.Context, typeEntity *domain.Type) error {
	typeEntity.Slug = generateSlug(typeEntity.Name)
	e := mapper.TypeToEntity(typeEntity)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *typeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.TypeEntity{}, "id = ?", id).Error
}

func (r *typeRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.TypeEntity{}, "id = ?", id).Error
}
