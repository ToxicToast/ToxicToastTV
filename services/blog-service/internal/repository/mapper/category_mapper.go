package mapper

import (
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository/entity"

	"gorm.io/gorm"
)

// CategoryToEntity converts domain model to database entity
func CategoryToEntity(category *domain.Category) *entity.CategoryEntity {
	if category == nil {
		return nil
	}

	e := &entity.CategoryEntity{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		ParentID:    category.ParentID,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	// Convert DeletedAt
	if category.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *category.DeletedAt,
			Valid: true,
		}
	}

	// Convert Parent (avoid infinite recursion)
	if category.Parent != nil {
		e.Parent = CategoryToEntity(category.Parent)
	}

	// Convert Children (avoid infinite recursion)
	if len(category.Children) > 0 {
		e.Children = make([]entity.CategoryEntity, 0, len(category.Children))
		for _, child := range category.Children {
			e.Children = append(e.Children, *CategoryToEntity(&child))
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
		ID:          e.ID,
		Name:        e.Name,
		Slug:        e.Slug,
		Description: e.Description,
		ParentID:    e.ParentID,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		category.DeletedAt = &deletedAt
	}

	// Convert Parent (avoid infinite recursion)
	if e.Parent != nil {
		category.Parent = CategoryToDomain(e.Parent)
	}

	// Convert Children (avoid infinite recursion)
	if len(e.Children) > 0 {
		category.Children = make([]domain.Category, 0, len(e.Children))
		for _, child := range e.Children {
			category.Children = append(category.Children, *CategoryToDomain(&child))
		}
	}

	return category
}

// CategoriesToDomain converts slice of entities to domain models
func CategoriesToDomain(entities []entity.CategoryEntity) []*domain.Category {
	categories := make([]*domain.Category, 0, len(entities))
	for _, e := range entities {
		categories = append(categories, CategoryToDomain(&e))
	}
	return categories
}
