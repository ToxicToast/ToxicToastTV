package usecase

import (
	"context"
	"fmt"
	"time"

	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// RBACUseCase handles RBAC business logic
type RBACUseCase struct {
	userRoleRepo       interfaces.UserRoleRepository
	rolePermissionRepo interfaces.RolePermissionRepository
	roleRepo           interfaces.RoleRepository
	permissionRepo     interfaces.PermissionRepository
}

// NewRBACUseCase creates a new RBAC use case
func NewRBACUseCase(
	userRoleRepo interfaces.UserRoleRepository,
	rolePermissionRepo interfaces.RolePermissionRepository,
	roleRepo interfaces.RoleRepository,
	permissionRepo interfaces.PermissionRepository,
) *RBACUseCase {
	return &RBACUseCase{
		userRoleRepo:       userRoleRepo,
		rolePermissionRepo: rolePermissionRepo,
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
	}
}

// AssignRole assigns a role to a user
func (uc *RBACUseCase) AssignRole(ctx context.Context, userID, roleID string) error {
	// Check if role exists
	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Check if already assigned
	hasRole, err := uc.userRoleRepo.HasRole(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to check role assignment: %w", err)
	}
	if hasRole {
		return fmt.Errorf("user already has this role")
	}

	// Assign role
	userRole := &domain.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
	}

	return uc.userRoleRepo.AssignRole(ctx, userRole)
}

// RevokeRole removes a role from a user
func (uc *RBACUseCase) RevokeRole(ctx context.Context, userID, roleID string) error {
	// Check if user has the role
	hasRole, err := uc.userRoleRepo.HasRole(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to check role assignment: %w", err)
	}
	if !hasRole {
		return fmt.Errorf("user does not have this role")
	}

	return uc.userRoleRepo.RevokeRole(ctx, userID, roleID)
}

// AssignPermission assigns a permission to a role
func (uc *RBACUseCase) AssignPermission(ctx context.Context, roleID, permissionID string) error {
	// Check if role exists
	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Check if permission exists
	permission, err := uc.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}
	if permission == nil {
		return fmt.Errorf("permission not found")
	}

	// Check if already assigned
	hasPermission, err := uc.rolePermissionRepo.HasPermission(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check permission assignment: %w", err)
	}
	if hasPermission {
		return fmt.Errorf("role already has this permission")
	}

	// Assign permission
	rolePermission := &domain.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedAt:    time.Now(),
	}

	return uc.rolePermissionRepo.AssignPermission(ctx, rolePermission)
}

// RevokePermission removes a permission from a role
func (uc *RBACUseCase) RevokePermission(ctx context.Context, roleID, permissionID string) error {
	// Check if role has the permission
	hasPermission, err := uc.rolePermissionRepo.HasPermission(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check permission assignment: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("role does not have this permission")
	}

	return uc.rolePermissionRepo.RevokePermission(ctx, roleID, permissionID)
}

// CheckPermission checks if a user has a specific permission
func (uc *RBACUseCase) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// Get user permissions
	permissions, err := uc.rolePermissionRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// Check if user has the permission
	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			return true, nil
		}
	}

	return false, nil
}

// GetUserRoles returns all roles assigned to a user
func (uc *RBACUseCase) GetUserRoles(ctx context.Context, userID string) ([]*domain.Role, error) {
	return uc.userRoleRepo.GetUserRoles(ctx, userID)
}

// GetUserPermissions returns all permissions for a user (via their roles)
func (uc *RBACUseCase) GetUserPermissions(ctx context.Context, userID string) ([]*domain.Permission, error) {
	return uc.rolePermissionRepo.GetUserPermissions(ctx, userID)
}

// GetRolePermissions returns all permissions assigned to a role
func (uc *RBACUseCase) GetRolePermissions(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	return uc.rolePermissionRepo.GetRolePermissions(ctx, roleID)
}
