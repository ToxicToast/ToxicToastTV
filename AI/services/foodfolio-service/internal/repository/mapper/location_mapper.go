package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/entity"
)

// LocationToEntity converts domain model to database entity
func LocationToEntity(location *domain.Location) *entity.LocationEntity {
	if location == nil {
		return nil
	}

	e := &entity.LocationEntity{
		ID:        location.ID,
		Name:      location.Name,
		Slug:      location.Slug,
		ParentID:  location.ParentID,
		CreatedAt: location.CreatedAt,
		UpdatedAt: location.UpdatedAt,
	}

	// Convert DeletedAt
	if location.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *location.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// LocationToDomain converts database entity to domain model
func LocationToDomain(e *entity.LocationEntity) *domain.Location {
	if e == nil {
		return nil
	}

	location := &domain.Location{
		ID:        e.ID,
		Name:      e.Name,
		Slug:      e.Slug,
		ParentID:  e.ParentID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		location.DeletedAt = &deletedAt
	}

	return location
}

// LocationsToDomain converts slice of entities to domain models
func LocationsToDomain(entities []*entity.LocationEntity) []*domain.Location {
	locations := make([]*domain.Location, 0, len(entities))
	for _, e := range entities {
		locations = append(locations, LocationToDomain(e))
	}
	return locations
}
