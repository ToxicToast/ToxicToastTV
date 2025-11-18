package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// TypeToEntity converts domain model to database entity
func TypeToEntity(t *domain.Type) *entity.TypeEntity {
	if t == nil {
		return nil
	}

	e := &entity.TypeEntity{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}

	// Convert DeletedAt
	if t.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *t.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// TypeToDomain converts database entity to domain model
func TypeToDomain(e *entity.TypeEntity) *domain.Type {
	if e == nil {
		return nil
	}

	t := &domain.Type{
		ID:        e.ID,
		Name:      e.Name,
		Slug:      e.Slug,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		t.DeletedAt = &deletedAt
	}

	return t
}

// TypesToDomain converts slice of entities to domain models
func TypesToDomain(entities []*entity.TypeEntity) []*domain.Type {
	types := make([]*domain.Type, 0, len(entities))
	for _, e := range entities {
		types = append(types, TypeToDomain(e))
	}
	return types
}
