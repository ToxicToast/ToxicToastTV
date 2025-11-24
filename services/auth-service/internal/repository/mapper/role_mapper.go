package mapper

import (
	"gorm.io/gorm"
	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/entity"
)

// RoleToEntity converts domain model to database entity
func RoleToEntity(role *domain.Role) *entity.RoleEntity {
	if role == nil {
		return nil
	}

	e := &entity.RoleEntity{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}

	// Convert DeletedAt
	if role.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *role.DeletedAt,
			Valid: true,
		}
	}

	return e
}

// RoleToDomain converts database entity to domain model
func RoleToDomain(e *entity.RoleEntity) *domain.Role {
	if e == nil {
		return nil
	}

	role := &domain.Role{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		role.DeletedAt = &deletedAt
	}

	return role
}

// RolesToDomain converts slice of entities to domain models
func RolesToDomain(entities []*entity.RoleEntity) []*domain.Role {
	roles := make([]*domain.Role, 0, len(entities))
	for _, e := range entities {
		roles = append(roles, RoleToDomain(e))
	}
	return roles
}
