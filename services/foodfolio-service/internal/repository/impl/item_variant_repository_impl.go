package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

type itemVariantRepository struct {
	db *gorm.DB
}

// NewItemVariantRepository creates a new item variant repository instance
func NewItemVariantRepository(db *gorm.DB) interfaces.ItemVariantRepository {
	return &itemVariantRepository{db: db}
}

func (r *itemVariantRepository) Create(ctx context.Context, variant *domain.ItemVariant) error {
	return r.db.WithContext(ctx).Create(variant).Error
}

func (r *itemVariantRepository) GetByID(ctx context.Context, id string) (*domain.ItemVariant, error) {
	var variant domain.ItemVariant
	err := r.db.WithContext(ctx).
		Preload("Item").
		Preload("Item.Category").
		Preload("Item.Company").
		Preload("Item.Type").
		Preload("Size").
		First(&variant, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &variant, nil
}

func (r *itemVariantRepository) GetByBarcode(ctx context.Context, barcode string) (*domain.ItemVariant, error) {
	var variant domain.ItemVariant
	err := r.db.WithContext(ctx).
		Preload("Item").
		Preload("Item.Category").
		Preload("Item.Company").
		Preload("Item.Type").
		Preload("Size").
		First(&variant, "barcode = ?", barcode).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &variant, nil
}

func (r *itemVariantRepository) List(ctx context.Context, offset, limit int, itemID, sizeID *string, isNormallyFrozen *bool, includeDeleted bool) ([]*domain.ItemVariant, int64, error) {
	var variants []*domain.ItemVariant
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.ItemVariant{})

	if includeDeleted {
		query = query.Unscoped()
	}

	// Filters
	if itemID != nil && *itemID != "" {
		query = query.Where("item_id = ?", *itemID)
	}

	if sizeID != nil && *sizeID != "" {
		query = query.Where("size_id = ?", *sizeID)
	}

	if isNormallyFrozen != nil {
		query = query.Where("is_normally_frozen = ?", *isNormallyFrozen)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Item").
		Preload("Size").
		Offset(offset).
		Limit(limit).
		Find(&variants).Error; err != nil {
		return nil, 0, err
	}

	return variants, total, nil
}

func (r *itemVariantRepository) GetByItem(ctx context.Context, itemID string) ([]*domain.ItemVariant, error) {
	var variants []*domain.ItemVariant

	err := r.db.WithContext(ctx).
		Preload("Size").
		Where("item_id = ?", itemID).
		Find(&variants).Error

	if err != nil {
		return nil, err
	}

	return variants, nil
}

func (r *itemVariantRepository) GetLowStockVariants(ctx context.Context, offset, limit int) ([]*domain.ItemVariant, int64, error) {
	var variants []*domain.ItemVariant
	var total int64

	// Subquery to count active item details per variant
	subQuery := r.db.Model(&domain.ItemDetail{}).
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
			Find(&variants).Error; err != nil {
			return nil, 0, err
		}
	}

	return variants, total, nil
}

func (r *itemVariantRepository) GetOverstockedVariants(ctx context.Context, offset, limit int) ([]*domain.ItemVariant, int64, error) {
	var variants []*domain.ItemVariant
	var total int64

	// Subquery to count active item details per variant
	subQuery := r.db.Model(&domain.ItemDetail{}).
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
			Find(&variants).Error; err != nil {
			return nil, 0, err
		}
	}

	return variants, total, nil
}

func (r *itemVariantRepository) GetCurrentStock(ctx context.Context, variantID string) (int, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&domain.ItemDetail{}).
		Where("item_variant_id = ? AND is_opened = ? AND (expiry_date IS NULL OR expiry_date > NOW())", variantID, false).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *itemVariantRepository) Update(ctx context.Context, variant *domain.ItemVariant) error {
	return r.db.WithContext(ctx).Save(variant).Error
}

func (r *itemVariantRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.ItemVariant{}, "id = ?", id).Error
}

func (r *itemVariantRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.ItemVariant{}, "id = ?", id).Error
}
