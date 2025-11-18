package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// ItemToEntity converts domain model to database entity
func ItemToEntity(item *domain.Item) *entity.ItemEntity {
	if item == nil {
		return nil
	}

	e := &entity.ItemEntity{
		ID:         item.ID,
		Name:       item.Name,
		Slug:       item.Slug,
		CategoryID: item.CategoryID,
		CompanyID:  item.CompanyID,
		TypeID:     item.TypeID,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
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

// ItemToDomain converts database entity to domain model
func ItemToDomain(e *entity.ItemEntity) *domain.Item {
	if e == nil {
		return nil
	}

	item := &domain.Item{
		ID:         e.ID,
		Name:       e.Name,
		Slug:       e.Slug,
		CategoryID: e.CategoryID,
		CompanyID:  e.CompanyID,
		TypeID:     e.TypeID,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		item.DeletedAt = &deletedAt
	}

	return item
}

// ItemsToDomain converts slice of entities to domain models
func ItemsToDomain(entities []*entity.ItemEntity) []*domain.Item {
	items := make([]*domain.Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, ItemToDomain(e))
	}
	return items
}
