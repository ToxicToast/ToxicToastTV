package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
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
	return r.db.WithContext(ctx).Create(warehouse).Error
}

func (r *warehouseRepository) GetByID(ctx context.Context, id string) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	err := r.db.WithContext(ctx).First(&warehouse, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *warehouseRepository) GetBySlug(ctx context.Context, slug string) (*domain.Warehouse, error) {
	var warehouse domain.Warehouse
	err := r.db.WithContext(ctx).First(&warehouse, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &warehouse, nil
}

func (r *warehouseRepository) List(ctx context.Context, offset, limit int, search string, includeDeleted bool) ([]*domain.Warehouse, int64, error) {
	var warehouses []*domain.Warehouse
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Warehouse{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&warehouses).Error; err != nil {
		return nil, 0, err
	}

	return warehouses, total, nil
}

func (r *warehouseRepository) Update(ctx context.Context, warehouse *domain.Warehouse) error {
	warehouse.Slug = generateSlug(warehouse.Name)
	return r.db.WithContext(ctx).Save(warehouse).Error
}

func (r *warehouseRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Warehouse{}, "id = ?", id).Error
}

func (r *warehouseRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.Warehouse{}, "id = ?", id).Error
}
