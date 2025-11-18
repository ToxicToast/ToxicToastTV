package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// SizeToEntity converts domain model to database entity
func SizeToEntity(size *domain.Size) *entity.SizeEntity {
	if size == nil {
		return nil
	}

	e := &entity.SizeEntity{
		ID:        size.ID,
		Name:      size.Name,
		Value:     size.Value,
		Unit:      size.Unit,
		CreatedAt: size.CreatedAt,
		UpdatedAt: size.UpdatedAt,
	}

	// Convert DeletedAt
	if size.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *size.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// SizeToDomain converts database entity to domain model
func SizeToDomain(e *entity.SizeEntity) *domain.Size {
	if e == nil {
		return nil
	}

	size := &domain.Size{
		ID:        e.ID,
		Name:      e.Name,
		Value:     e.Value,
		Unit:      e.Unit,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		size.DeletedAt = &deletedAt
	}

	return size
}

// SizesToDomain converts slice of entities to domain models
func SizesToDomain(entities []*entity.SizeEntity) []*domain.Size {
	sizes := make([]*domain.Size, 0, len(entities))
	for _, e := range entities {
		sizes = append(sizes, SizeToDomain(e))
	}
	return sizes
}
