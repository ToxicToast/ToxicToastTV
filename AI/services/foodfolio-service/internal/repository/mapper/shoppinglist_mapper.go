package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// ShoppinglistToEntity converts domain model to database entity
func ShoppinglistToEntity(list *domain.Shoppinglist) *entity.ShoppinglistEntity {
	if list == nil {
		return nil
	}

	e := &entity.ShoppinglistEntity{
		ID:        list.ID,
		Name:      list.Name,
		CreatedAt: list.CreatedAt,
		UpdatedAt: list.UpdatedAt,
	}

	// Convert DeletedAt
	if list.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *list.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// ShoppinglistToDomain converts database entity to domain model
func ShoppinglistToDomain(e *entity.ShoppinglistEntity) *domain.Shoppinglist {
	if e == nil {
		return nil
	}

	list := &domain.Shoppinglist{
		ID:        e.ID,
		Name:      e.Name,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		list.DeletedAt = &deletedAt
	}

	return list
}

// ShoppinglistsToDomain converts slice of entities to domain models
func ShoppinglistsToDomain(entities []*entity.ShoppinglistEntity) []*domain.Shoppinglist {
	lists := make([]*domain.Shoppinglist, 0, len(entities))
	for _, e := range entities {
		lists = append(lists, ShoppinglistToDomain(e))
	}
	return lists
}
