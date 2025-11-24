package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// ShoppinglistItemToEntity converts domain model to database entity
func ShoppinglistItemToEntity(item *domain.ShoppinglistItem) *entity.ShoppinglistItemEntity {
	if item == nil {
		return nil
	}

	e := &entity.ShoppinglistItemEntity{
		ID:             item.ID,
		ShoppinglistID: item.ShoppinglistID,
		ItemVariantID:  item.ItemVariantID,
		Quantity:       item.Quantity,
		IsPurchased:    item.IsPurchased,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}

	// Convert DeletedAt
	if item.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *item.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// ShoppinglistItemToDomain converts database entity to domain model
func ShoppinglistItemToDomain(e *entity.ShoppinglistItemEntity) *domain.ShoppinglistItem {
	if e == nil {
		return nil
	}

	item := &domain.ShoppinglistItem{
		ID:             e.ID,
		ShoppinglistID: e.ShoppinglistID,
		ItemVariantID:  e.ItemVariantID,
		Quantity:       e.Quantity,
		IsPurchased:    e.IsPurchased,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		item.DeletedAt = &deletedAt
	}

	// Convert ItemVariant
	if e.ItemVariant != nil {
		item.ItemVariant = ItemVariantToDomain(e.ItemVariant)
	}

	return item
}

// ShoppinglistItemsToDomain converts slice of entities to domain models
func ShoppinglistItemsToDomain(entities []*entity.ShoppinglistItemEntity) []*domain.ShoppinglistItem {
	items := make([]*domain.ShoppinglistItem, 0, len(entities))
	for _, e := range entities {
		items = append(items, ShoppinglistItemToDomain(e))
	}
	return items
}
