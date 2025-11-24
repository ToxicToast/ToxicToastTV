package interfaces

import (
	"context"
	"toxictoast/services/auth-service/internal/domain"
)

// UserRoleRepository defines the interface for user-role relationship data access
type UserRoleRepository interface {
	AssignRole(ctx context.Context, userRole *domain.UserRole) error
	RevokeRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*domain.Role, error)
	GetRoleUsers(ctx context.Context, roleID string) ([]string, error)
	HasRole(ctx context.Context, userID, roleID string) (bool, error)
}
