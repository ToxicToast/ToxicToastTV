package interfaces

import (
	"context"
	"toxictoast/services/auth-service/internal/domain"
)

// RolePermissionRepository defines the interface for role-permission relationship data access
type RolePermissionRepository interface {
	AssignPermission(ctx context.Context, rolePermission *domain.RolePermission) error
	RevokePermission(ctx context.Context, roleID, permissionID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*domain.Permission, error)
	GetPermissionRoles(ctx context.Context, permissionID string) ([]*domain.Role, error)
	HasPermission(ctx context.Context, roleID, permissionID string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]*domain.Permission, error)
}
