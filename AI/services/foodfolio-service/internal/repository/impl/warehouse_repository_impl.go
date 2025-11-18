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

type warehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository creates a new warehouse repository instance
func NewWarehouseRepository(db *gorm.DB) interfaces.WarehouseRepository {
	return &warehouseRepository{db: db}
}

func (r *warehouseRepository) Create(ctx context.Context, warehouse *domain.Warehouse) error {
	warehouse.Slug = generateSlug(warehouse.Name)
	e := mapper.WarehouseToEntity(warehouse)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *warehouseRepository) GetByID(ctx context.Context, id string) (*domain.Warehouse, error) {
	var e entity.WarehouseEntity
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.WarehouseToDomain(&e), nil
}

func (r *warehouseRepository) GetBySlug(ctx context.Context, slug string) (*domain.Warehouse, error) {
	var e entity.WarehouseEntity
	err := r.db.WithContext(ctx).First(&e, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.WarehouseToDomain(&e), nil
}

func (r *warehouseRepository) List(ctx context.Context, offset, limit int, search string, includeDeleted bool) ([]*domain.Warehouse, int64, error) {
	var entities []*entity.WarehouseEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.WarehouseEntity{})

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

	return mapper.WarehousesToDomain(entities), total, nil
}

func (r *warehouseRepository) Update(ctx context.Context, warehouse *domain.Warehouse) error {
	warehouse.Slug = generateSlug(warehouse.Name)
	e := mapper.WarehouseToEntity(warehouse)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *warehouseRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.WarehouseEntity{}, "id = ?", id).Error
}

func (r *warehouseRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.WarehouseEntity{}, "id = ?", id).Error
}
