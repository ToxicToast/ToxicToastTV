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

type shoppinglistRepository struct {
	db *gorm.DB
}

// NewShoppinglistRepository creates a new shoppinglist repository instance
func NewShoppinglistRepository(db *gorm.DB) interfaces.ShoppinglistRepository {
	return &shoppinglistRepository{db: db}
}

func (r *shoppinglistRepository) Create(ctx context.Context, shoppinglist *domain.Shoppinglist) error {
	e := mapper.ShoppinglistToEntity(shoppinglist)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *shoppinglistRepository) GetByID(ctx context.Context, id string) (*domain.Shoppinglist, error) {
	var e entity.ShoppinglistEntity
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.ItemVariant").
		Preload("Items.ItemVariant.Item").
		Preload("Items.ItemVariant.Size").
		First(&e, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ShoppinglistToDomain(&e), nil
}

func (r *shoppinglistRepository) List(ctx context.Context, offset, limit int, includeDeleted bool) ([]*domain.Shoppinglist, int64, error) {
	var entities []*entity.ShoppinglistEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ShoppinglistEntity{})

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
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ShoppinglistsToDomain(entities), total, nil
}

func (r *shoppinglistRepository) Update(ctx context.Context, shoppinglist *domain.Shoppinglist) error {
	e := mapper.ShoppinglistToEntity(shoppinglist)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *shoppinglistRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ShoppinglistEntity{}, "id = ?", id).Error
}

func (r *shoppinglistRepository) HardDelete(ctx context.Context, id string) error {
	// Also delete associated items
	r.db.WithContext(ctx).Unscoped().Where("shoppinglist_id = ?", id).Delete(&entity.ShoppinglistItemEntity{})
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.ShoppinglistEntity{}, "id = ?", id).Error
}

func (r *shoppinglistRepository) AddItem(ctx context.Context, item *domain.ShoppinglistItem) error {
	e := mapper.ShoppinglistItemToEntity(item)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *shoppinglistRepository) RemoveItem(ctx context.Context, shoppinglistID, itemID string) error {
	return r.db.WithContext(ctx).
		Where("shoppinglist_id = ? AND id = ?", shoppinglistID, itemID).
		Delete(&entity.ShoppinglistItemEntity{}).Error
}

func (r *shoppinglistRepository) UpdateItem(ctx context.Context, item *domain.ShoppinglistItem) error {
	e := mapper.ShoppinglistItemToEntity(item)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *shoppinglistRepository) MarkItemPurchased(ctx context.Context, itemID string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ShoppinglistItemEntity{}).
		Where("id = ?", itemID).
		Update("is_purchased", true).Error
}

func (r *shoppinglistRepository) MarkAllItemsPurchased(ctx context.Context, shoppinglistID string) (int, error) {
	result := r.db.WithContext(ctx).
		Model(&entity.ShoppinglistItemEntity{}).
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
		Delete(&entity.ShoppinglistItemEntity{})

	if result.Error != nil {
		return 0, result.Error
	}

	return int(result.RowsAffected), nil
}

func (r *shoppinglistRepository) GetItem(ctx context.Context, itemID string) (*domain.ShoppinglistItem, error) {
	var e entity.ShoppinglistItemEntity
	err := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		First(&e, "id = ?", itemID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ShoppinglistItemToDomain(&e), nil
}

func (r *shoppinglistRepository) GetItems(ctx context.Context, shoppinglistID string) ([]*domain.ShoppinglistItem, error) {
	var entities []*entity.ShoppinglistItemEntity

	err := r.db.WithContext(ctx).
		Preload("ItemVariant").
		Preload("ItemVariant.Item").
		Preload("ItemVariant.Size").
		Where("shoppinglist_id = ?", shoppinglistID).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return mapper.ShoppinglistItemsToDomain(entities), nil
}
