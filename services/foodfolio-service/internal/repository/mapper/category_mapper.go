package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// CategoryToEntity converts domain model to database entity
func CategoryToEntity(category *domain.Category) *entity.CategoryEntity {
	if category == nil {
		return nil
	}

	e := &entity.CategoryEntity{
		ID:       category.ID,
		Name:     category.Name,
		Slug:     category.Slug,
		ParentID: category.ParentID,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}

	// Convert DeletedAt
	if category.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *category.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// CategoryToDomain converts database entity to domain model
func CategoryToDomain(e *entity.CategoryEntity) *domain.Category {
	if e == nil {
		return nil
	}

	category := &domain.Category{
		ID:       e.ID,
		Name:     e.Name,
		Slug:     e.Slug,
		ParentID: e.ParentID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		category.DeletedAt = &deletedAt
	}

	return category
}

// CategoriesToDomain converts slice of entities to domain models
func CategoriesToDomain(entities []*entity.CategoryEntity) []*domain.Category {
	categories := make([]*domain.Category, 0, len(entities))
	for _, e := range entities {
		categories = append(categories, CategoryToDomain(e))
	}
	return categories
}
