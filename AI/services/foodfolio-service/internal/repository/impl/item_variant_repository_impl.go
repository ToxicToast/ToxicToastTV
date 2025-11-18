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

type itemVariantRepository struct {
	db *gorm.DB
}

// NewItemVariantRepository creates a new item variant repository instance
func NewItemVariantRepository(db *gorm.DB) interfaces.ItemVariantRepository {
	return &itemVariantRepository{db: db}
}

func (r *itemVariantRepository) Create(ctx context.Context, variant *domain.ItemVariant) error {
	e := mapper.ItemVariantToEntity(variant)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *itemVariantRepository) GetByID(ctx context.Context, id string) (*domain.ItemVariant, error) {
	var e entity.ItemVariantEntity
	err := r.db.WithContext(ctx).
		Preload("Item").
		Preload("Item.Category").
		Preload("Item.Company").
		Preload("Item.Type").
		Preload("Size").
		First(&e, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ItemVariantToDomain(&e), nil
}

func (r *itemVariantRepository) GetByBarcode(ctx context.Context, barcode string) (*domain.ItemVariant, error) {
	var e entity.ItemVariantEntity
	err := r.db.WithContext(ctx).
		Preload("Item").
		Preload("Item.Category").
		Preload("Item.Company").
		Preload("Item.Type").
		Preload("Size").
		First(&e, "barcode = ?", barcode).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ItemVariantToDomain(&e), nil
}

func (r *itemVariantRepository) List(ctx context.Context, offset, limit int, itemID, sizeID *string, isNormallyFrozen *bool, includeDeleted bool) ([]*domain.ItemVariant, int64, error) {
	var entities []*entity.ItemVariantEntity
	var total int64

	// Build base query with filters
	applyFilters := func(query *gorm.DB) *gorm.DB {
		if includeDeleted {
			query = query.Unscoped()
		}

		if itemID != nil && *itemID != "" {
			query = query.Where("item_id = ?", *itemID)
		}

		if sizeID != nil && *sizeID != "" {
			query = query.Where("size_id = ?", *sizeID)
		}

		if isNormallyFrozen != nil {
			query = query.Where("is_normally_frozen = ?", *isNormallyFrozen)
		}

		return query
	}

	// Count total
	countQuery := applyFilters(r.db.WithContext(ctx).Model(&entity.ItemVariantEntity{}))
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with preloads
	fetchQuery := applyFilters(r.db.WithContext(ctx).Model(&entity.ItemVariantEntity{}))
	if err := fetchQuery.
		Preload("Item").
		Preload("Item.Category").
		Preload("Item.Company").
		Preload("Item.Type").
		Preload("Size").
		Offset(offset).
		Limit(limit).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ItemVariantsToDomain(entities), total, nil
}

func (r *itemVariantRepository) GetByItem(ctx context.Context, itemID string) ([]*domain.ItemVariant, error) {
	var entities []*entity.ItemVariantEntity

	err := r.db.WithContext(ctx).
		Preload("Size").
		Where("item_id = ?", itemID).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return mapper.ItemVariantsToDomain(entities), nil
}

func (r *itemVariantRepository) GetLowStockVariants(ctx context.Context, offset, limit int) ([]*domain.ItemVariant, int64, error) {
	var entities []*entity.ItemVariantEntity
	var total int64

	// Subquery to count active item details per variant
	subQuery := r.db.Model(&entity.ItemDetailEntity{}).
		Select("item_variant_id, COUNT(*) as stock_count").
		Where("is_opened = ? AND (expiry_date IS NULL OR expiry_date > NOW())", false).
		Group("item_variant_id")

	query := r.db.WithContext(ctx).
		Table("item_variants").
		Joins("LEFT JOIN (?) as stock ON stock.item_variant_id = item_variants.id", subQuery).
		Where("COALESCE(stock.stock_count, 0) < item_variants.min_sku")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with preloads
	var variantIDs []string
	if err := query.Offset(offset).Limit(limit).Pluck("item_variants.id", &variantIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(variantIDs) > 0 {
		if err := r.db.WithContext(ctx).
			Preload("Item").
			Preload("Item.Category").
			Preload("Item.Company").
			Preload("Size").
			Where("id IN ?", variantIDs).
			Find(&entities).Error; err != nil {
			return nil, 0, err
		}
	}

	return mapper.ItemVariantsToDomain(entities), total, nil
}

func (r *itemVariantRepository) GetOverstockedVariants(ctx context.Context, offset, limit int) ([]*domain.ItemVariant, int64, error) {
	var entities []*entity.ItemVariantEntity
	var total int64

	// Subquery to count active item details per variant
	subQuery := r.db.Model(&entity.ItemDetailEntity{}).
		Select("item_variant_id, COUNT(*) as stock_count").
		Where("is_opened = ? AND (expiry_date IS NULL OR expiry_date > NOW())", false).
		Group("item_variant_id")

	query := r.db.WithContext(ctx).
		Table("item_variants").
		Joins("LEFT JOIN (?) as stock ON stock.item_variant_id = item_variants.id", subQuery).
		Where("COALESCE(stock.stock_count, 0) > item_variants.max_sku AND item_variants.max_sku > 0")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with preloads
	var variantIDs []string
	if err := query.Offset(offset).Limit(limit).Pluck("item_variants.id", &variantIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(variantIDs) > 0 {
		if err := r.db.WithContext(ctx).
			Preload("Item").
			Preload("Item.Category").
			Preload("Item.Company").
			Preload("Size").
			Where("id IN ?", variantIDs).
			Find(&entities).Error; err != nil {
			return nil, 0, err
		}
	}

	return mapper.ItemVariantsToDomain(entities), total, nil
}

func (r *itemVariantRepository) GetCurrentStock(ctx context.Context, variantID string) (int, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&entity.ItemDetailEntity{}).
		Where("item_variant_id = ? AND is_opened = ? AND (expiry_date IS NULL OR expiry_date > NOW())", variantID, false).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *itemVariantRepository) Update(ctx context.Context, variant *domain.ItemVariant) error {
	e := mapper.ItemVariantToEntity(variant)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *itemVariantRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ItemVariantEntity{}, "id = ?", id).Error
}

func (r *itemVariantRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.ItemVariantEntity{}, "id = ?", id).Error
}
