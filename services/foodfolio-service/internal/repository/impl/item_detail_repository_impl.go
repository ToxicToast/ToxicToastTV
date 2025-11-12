package impl

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
	"toxictoast/services/foodfolio-service/internal/repository/mapper"
)

type itemDetailRepository struct {
	db *gorm.DB
}

// NewItemDetailRepository creates a new item detail repository instance
func NewItemDetailRepository(db *gorm.DB) interfaces.ItemDetailRepository {
	return &itemDetailRepository{db: db}
}

func (r *itemDetailRepository) Create(ctx context.Context, detail *domain.ItemDetail) error {
	e := mapper.ItemDetailToEntity(detail)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *itemDetailRepository) BatchCreate(ctx context.Context, details []*domain.ItemDetail) error {
	entities := make([]*entity.ItemDetailEntity, len(details))
	for i, d := range details {
		entities[i] = mapper.ItemDetailToEntity(d)
	}
	return r.db.WithContext(ctx).CreateInBatches(entities, 100).Error
}

func (r *itemDetailRepository) GetByID(ctx context.Context, id string) (*domain.ItemDetail, error) {
	var e entity.ItemDetailEntity
	err := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Preload("Warehouse").
		Preload("Location").
		First(&e, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ItemDetailToDomain(&e), nil
}

func (r *itemDetailRepository) List(ctx context.Context, offset, limit int, variantID, warehouseID, locationID *string, isOpened, hasDeposit, isFrozen *bool, includeDeleted bool) ([]*domain.ItemDetail, int64, error) {
	var entities []*entity.ItemDetailEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ItemDetailEntity{})

	if includeDeleted {
		query = query.Unscoped()
	}

	// Filters
	if variantID != nil && *variantID != "" {
		query = query.Where("item_variant_id = ?", *variantID)
	}

	if warehouseID != nil && *warehouseID != "" {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}

	if locationID != nil && *locationID != "" {
		query = query.Where("location_id = ?", *locationID)
	}

	if isOpened != nil {
		query = query.Where("is_opened = ?", *isOpened)
	}

	if hasDeposit != nil {
		query = query.Where("has_deposit = ?", *hasDeposit)
	}

	if isFrozen != nil {
		query = query.Where("is_frozen = ?", *isFrozen)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Preload("Warehouse").
		Preload("Location").
		Offset(offset).
		Limit(limit).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ItemDetailsToDomain(entities), total, nil
}

func (r *itemDetailRepository) GetByVariant(ctx context.Context, variantID string) ([]*domain.ItemDetail, error) {
	var entities []*entity.ItemDetailEntity

	err := r.db.WithContext(ctx).
		Preload("Warehouse").
		Preload("Location").
		Where("item_variant_id = ?", variantID).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return mapper.ItemDetailsToDomain(entities), nil
}

func (r *itemDetailRepository) GetByLocation(ctx context.Context, locationID string, includeChildren bool) ([]*domain.ItemDetail, error) {
	var entities []*entity.ItemDetailEntity

	query := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Preload("Warehouse").
		Preload("Location")

	if includeChildren {
		// Get all location IDs including children (simplified - in production use recursive CTE)
		var locationIDs []string
		locationIDs = append(locationIDs, locationID)

		// Get direct children
		var children []*entity.LocationEntity
		r.db.Where("parent_id = ?", locationID).Find(&children)
		for _, child := range children {
			locationIDs = append(locationIDs, child.ID)
		}

		query = query.Where("location_id IN ?", locationIDs)
	} else {
		query = query.Where("location_id = ?", locationID)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}

	return mapper.ItemDetailsToDomain(entities), nil
}

func (r *itemDetailRepository) GetExpiringItems(ctx context.Context, days int, offset, limit int) ([]*domain.ItemDetail, int64, error) {
	var entities []*entity.ItemDetailEntity
	var total int64

	threshold := time.Now().AddDate(0, 0, days)

	query := r.db.WithContext(ctx).Model(&entity.ItemDetailEntity{}).
		Where("expiry_date IS NOT NULL").
		Where("expiry_date <= ?", threshold).
		Where("expiry_date > ?", time.Now()).
		Where("is_opened = ?", false)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Preload("Warehouse").
		Preload("Location").
		Order("expiry_date ASC").
		Offset(offset).
		Limit(limit).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ItemDetailsToDomain(entities), total, nil
}

func (r *itemDetailRepository) GetExpiredItems(ctx context.Context, offset, limit int) ([]*domain.ItemDetail, int64, error) {
	var entities []*entity.ItemDetailEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ItemDetailEntity{}).
		Where("expiry_date IS NOT NULL").
		Where("expiry_date < ?", time.Now()).
		Where("is_opened = ?", false)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Preload("Warehouse").
		Preload("Location").
		Order("expiry_date ASC").
		Offset(offset).
		Limit(limit).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ItemDetailsToDomain(entities), total, nil
}

func (r *itemDetailRepository) GetItemsWithDeposit(ctx context.Context, offset, limit int) ([]*domain.ItemDetail, int64, error) {
	var entities []*entity.ItemDetailEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ItemDetailEntity{}).
		Where("has_deposit = ?", true).
		Where("is_opened = ?", false)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Preload("Warehouse").
		Preload("Location").
		Offset(offset).
		Limit(limit).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ItemDetailsToDomain(entities), total, nil
}

func (r *itemDetailRepository) OpenItem(ctx context.Context, id string, openedDate time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entity.ItemDetailEntity{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_opened":   true,
			"opened_date": openedDate,
		}).Error
}

func (r *itemDetailRepository) MoveItems(ctx context.Context, itemIDs []string, newLocationID string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ItemDetailEntity{}).
		Where("id IN ?", itemIDs).
		Update("location_id", newLocationID).Error
}

func (r *itemDetailRepository) Update(ctx context.Context, detail *domain.ItemDetail) error {
	e := mapper.ItemDetailToEntity(detail)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *itemDetailRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ItemDetailEntity{}, "id = ?", id).Error
}

func (r *itemDetailRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.ItemDetailEntity{}, "id = ?", id).Error
}

func (r *itemDetailRepository) CountByVariant(ctx context.Context, variantID string) (int, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&entity.ItemDetailEntity{}).
		Where("item_variant_id = ? AND is_opened = ? AND (expiry_date IS NULL OR expiry_date > ?)", variantID, false, time.Now()).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}
