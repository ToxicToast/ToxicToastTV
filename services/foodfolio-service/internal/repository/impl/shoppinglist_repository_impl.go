package impl

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

type shoppinglistRepository struct {
	db *gorm.DB
}

// NewShoppinglistRepository creates a new shoppinglist repository instance
func NewShoppinglistRepository(db *gorm.DB) interfaces.ShoppinglistRepository {
	return &shoppinglistRepository{db: db}
}

func (r *shoppinglistRepository) Create(ctx context.Context, shoppinglist *domain.Shoppinglist) error {
	return r.db.WithContext(ctx).Create(shoppinglist).Error
}

func (r *shoppinglistRepository) GetByID(ctx context.Context, id string) (*domain.Shoppinglist, error) {
	var shoppinglist domain.Shoppinglist
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.ItemVariant").
		Preload("Items.ItemVariant.Item").
		Preload("Items.ItemVariant.Size").
		First(&shoppinglist, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &shoppinglist, nil
}

func (r *shoppinglistRepository) List(ctx context.Context, offset, limit int, includeDeleted bool) ([]*domain.Shoppinglist, int64, error) {
	var shoppinglists []*domain.Shoppinglist
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Shoppinglist{})

	if includeDeleted {
		query = query.Unscoped()
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Items").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&shoppinglists).Error; err != nil {
		return nil, 0, err
	}

	return shoppinglists, total, nil
}

func (r *shoppinglistRepository) Update(ctx context.Context, shoppinglist *domain.Shoppinglist) error {
	return r.db.WithContext(ctx).Save(shoppinglist).Error
}

func (r *shoppinglistRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Shoppinglist{}, "id = ?", id).Error
}

func (r *shoppinglistRepository) HardDelete(ctx context.Context, id string) error {
	// Also delete associated items
	r.db.WithContext(ctx).Unscoped().Where("shoppinglist_id = ?", id).Delete(&domain.ShoppinglistItem{})
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.Shoppinglist{}, "id = ?", id).Error
}

func (r *shoppinglistRepository) AddItem(ctx context.Context, item *domain.ShoppinglistItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *shoppinglistRepository) RemoveItem(ctx context.Context, shoppinglistID, itemID string) error {
	return r.db.WithContext(ctx).
		Where("shoppinglist_id = ? AND id = ?", shoppinglistID, itemID).
		Delete(&domain.ShoppinglistItem{}).Error
}

func (r *shoppinglistRepository) UpdateItem(ctx context.Context, item *domain.ShoppinglistItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *shoppinglistRepository) MarkItemPurchased(ctx context.Context, itemID string) error {
	return r.db.WithContext(ctx).
		Model(&domain.ShoppinglistItem{}).
		Where("id = ?", itemID).
		Update("is_purchased", true).Error
}

func (r *shoppinglistRepository) MarkAllItemsPurchased(ctx context.Context, shoppinglistID string) (int, error) {
	result := r.db.WithContext(ctx).
		Model(&domain.ShoppinglistItem{}).
		Where("shoppinglist_id = ? AND is_purchased = ?", shoppinglistID, false).
		Update("is_purchased", true)

	if result.Error != nil {
		return 0, result.Error
	}

	return int(result.RowsAffected), nil
}

func (r *shoppinglistRepository) ClearPurchasedItems(ctx context.Context, shoppinglistID string) (int, error) {
	result := r.db.WithContext(ctx).
		Where("shoppinglist_id = ? AND is_purchased = ?", shoppinglistID, true).
		Delete(&domain.ShoppinglistItem{})

	if result.Error != nil {
		return 0, result.Error
	}

	return int(result.RowsAffected), nil
}

func (r *shoppinglistRepository) GetItem(ctx context.Context, itemID string) (*domain.ShoppinglistItem, error) {
	var item domain.ShoppinglistItem
	err := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		First(&item, "id = ?", itemID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *shoppinglistRepository) GetItems(ctx context.Context, shoppinglistID string) ([]*domain.ShoppinglistItem, error) {
	var items []*domain.ShoppinglistItem

	err := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Where("shoppinglist_id = ?", shoppinglistID).
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}
