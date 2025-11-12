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

type sizeRepository struct {
	db *gorm.DB
}

// NewSizeRepository creates a new size repository instance
func NewSizeRepository(db *gorm.DB) interfaces.SizeRepository {
	return &sizeRepository{db: db}
}

func (r *sizeRepository) Create(ctx context.Context, size *domain.Size) error {
	e := mapper.SizeToEntity(size)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *sizeRepository) GetByID(ctx context.Context, id string) (*domain.Size, error) {
	var e entity.SizeEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.SizeToDomain(&e), nil
}

func (r *sizeRepository) GetByName(ctx context.Context, name string) (*domain.Size, error) {
	var e entity.SizeEntity
	err := r.db.WithContext(ctx).First(&e, "name = ?", name).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.SizeToDomain(&e), nil
}

func (r *sizeRepository) List(ctx context.Context, offset, limit int, unit string, minValue, maxValue *float64, includeDeleted bool) ([]*domain.Size, int64, error) {
	var entities []*entity.SizeEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.SizeEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if unit != "" {
		query = query.Where("unit = ?", unit)
	}

	if minValue != nil {
		query = query.Where("value >= ?", *minValue)
	}

	if maxValue != nil {
		query = query.Where("value <= ?", *maxValue)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("value ASC").Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.SizesToDomain(entities), total, nil
}

func (r *sizeRepository) Update(ctx context.Context, size *domain.Size) error {
	e := mapper.SizeToEntity(size)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *sizeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.SizeEntity{}, "id = ?", id).Error
}

func (r *sizeRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.SizeEntity{}, "id = ?", id).Error
}
