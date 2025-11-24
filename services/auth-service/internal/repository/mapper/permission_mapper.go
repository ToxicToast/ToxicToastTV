package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/entity"
)

// PermissionToEntity converts domain model to database entity
func PermissionToEntity(permission *domain.Permission) *entity.PermissionEntity {
	if permission == nil {
		return nil
	}

	e := &entity.PermissionEntity{
		ID:          permission.ID,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}

	// Convert DeletedAt
	if permission.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *permission.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// PermissionToDomain converts database entity to domain model
func PermissionToDomain(e *entity.PermissionEntity) *domain.Permission {
	if e == nil {
		return nil
	}

	permission := &domain.Permission{
		ID:          e.ID,
		Resource:    e.Resource,
		Action:      e.Action,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		permission.DeletedAt = &deletedAt
	}

	return permission
}

// PermissionsToDomain converts slice of entities to domain models
func PermissionsToDomain(entities []*entity.PermissionEntity) []*domain.Permission {
	permissions := make([]*domain.Permission, 0, len(entities))
	for _, e := range entities {
		permissions = append(permissions, PermissionToDomain(e))
	}
	return permissions
}
