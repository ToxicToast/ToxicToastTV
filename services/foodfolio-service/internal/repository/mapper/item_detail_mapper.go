package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// ItemDetailToEntity converts domain model to database entity
func ItemDetailToEntity(detail *domain.ItemDetail) *entity.ItemDetailEntity {
	if detail == nil {
		return nil
	}

	e := &entity.ItemDetailEntity{
		ID:            detail.ID,
		ItemVariantID: detail.ItemVariantID,
		WarehouseID:   detail.WarehouseID,
		LocationID:    detail.LocationID,
		ArticleNumber: detail.ArticleNumber,
		PurchasePrice: detail.PurchasePrice,
		PurchaseDate:  detail.PurchaseDate,
		ExpiryDate:    detail.ExpiryDate,
		OpenedDate:    detail.OpenedDate,
		IsOpened:      detail.IsOpened,
		HasDeposit:    detail.HasDeposit,
		IsFrozen:      detail.IsFrozen,
		CreatedAt:     detail.CreatedAt,
		UpdatedAt:     detail.UpdatedAt,
	}

	// Convert DeletedAt
	if detail.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *detail.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// ItemDetailToDomain converts database entity to domain model
func ItemDetailToDomain(e *entity.ItemDetailEntity) *domain.ItemDetail {
	if e == nil {
		return nil
	}

	detail := &domain.ItemDetail{
		ID:            e.ID,
		ItemVariantID: e.ItemVariantID,
		WarehouseID:   e.WarehouseID,
		LocationID:    e.LocationID,
		ArticleNumber: e.ArticleNumber,
		PurchasePrice: e.PurchasePrice,
		PurchaseDate:  e.PurchaseDate,
		ExpiryDate:    e.ExpiryDate,
		OpenedDate:    e.OpenedDate,
		IsOpened:      e.IsOpened,
		HasDeposit:    e.HasDeposit,
		IsFrozen:      e.IsFrozen,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		detail.DeletedAt = &deletedAt
	}

	return detail
}

// ItemDetailsToDomain converts slice of entities to domain models
func ItemDetailsToDomain(entities []*entity.ItemDetailEntity) []*domain.ItemDetail {
	details := make([]*domain.ItemDetail, 0, len(entities))
	for _, e := range entities {
		details = append(details, ItemDetailToDomain(e))
	}
	return details
}
