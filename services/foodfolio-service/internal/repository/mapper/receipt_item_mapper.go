package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// ReceiptItemToEntity converts domain model to database entity
func ReceiptItemToEntity(item *domain.ReceiptItem) *entity.ReceiptItemEntity {
	if item == nil {
		return nil
	}

	e := &entity.ReceiptItemEntity{
		ID:            item.ID,
		ReceiptID:     item.ReceiptID,
		ItemVariantID: item.ItemVariantID,
		ItemName:      item.ItemName,
		Quantity:      item.Quantity,
		UnitPrice:     item.UnitPrice,
		TotalPrice:    item.TotalPrice,
		ArticleNumber: item.ArticleNumber,
		IsMatched:     item.IsMatched,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
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

// ReceiptItemToDomain converts database entity to domain model
func ReceiptItemToDomain(e *entity.ReceiptItemEntity) *domain.ReceiptItem {
	if e == nil {
		return nil
	}

	item := &domain.ReceiptItem{
		ID:            e.ID,
		ReceiptID:     e.ReceiptID,
		ItemVariantID: e.ItemVariantID,
		ItemName:      e.ItemName,
		Quantity:      e.Quantity,
		UnitPrice:     e.UnitPrice,
		TotalPrice:    e.TotalPrice,
		ArticleNumber: e.ArticleNumber,
		IsMatched:     e.IsMatched,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		item.DeletedAt = &deletedAt
	}

	return item
}

// ReceiptItemsToDomain converts slice of entities to domain models
func ReceiptItemsToDomain(entities []*entity.ReceiptItemEntity) []*domain.ReceiptItem {
	items := make([]*domain.ReceiptItem, 0, len(entities))
	for _, e := range entities {
		items = append(items, ReceiptItemToDomain(e))
	}
	return items
}
