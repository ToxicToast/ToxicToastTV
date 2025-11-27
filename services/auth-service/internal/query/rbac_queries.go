package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// GetUserRolesQuery retrieves all roles assigned to a user
type GetUserRolesQuery struct {
	cqrs.BaseQuery
	UserID string `json:"user_id"`
}

func (q *GetUserRolesQuery) QueryName() string {
	return "get_user_roles"
}

func (q *GetUserRolesQuery) Validate() error {
	if q.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// GetUserPermissionsQuery retrieves all permissions for a user (via their roles)
type GetUserPermissionsQuery struct {
	cqrs.BaseQuery
	UserID string `json:"user_id"`
}

func (q *GetUserPermissionsQuery) QueryName() string {
	return "get_user_permissions"
}

func (q *GetUserPermissionsQuery) Validate() error {
	if q.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// GetRolePermissionsQuery retrieves all permissions assigned to a role
type GetRolePermissionsQuery struct {
	cqrs.BaseQuery
	RoleID string `json:"role_id"`
}

func (q *GetRolePermissionsQuery) QueryName() string {
	return "get_role_permissions"
}

func (q *GetRolePermissionsQuery) Validate() error {
	if q.RoleID == "" {
		return errors.New("role_id is required")
	}
	return nil
}

// CheckPermissionQuery checks if a user has a specific permission
type CheckPermissionQuery struct {
	cqrs.BaseQuery
	UserID   string `json:"user_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

func (q *CheckPermissionQuery) QueryName() string {
	return "check_permission"
}

func (q *CheckPermissionQuery) Validate() error {
	if q.UserID == "" {
		return errors.New("user_id is required")
	}
	if q.Resource == "" {
		return errors.New("resource is required")
	}
	if q.Action == "" {
		return errors.New("action is required")
	}
	return nil
}

// CheckPermissionResult contains the result of a permission check
type CheckPermissionResult struct {
	HasPermission bool
}

// Query Handlers

// GetUserRolesHandler handles user roles retrieval
type GetUserRolesHandler struct {
	userRoleRepo interfaces.UserRoleRepository
}

func NewGetUserRolesHandler(userRoleRepo interfaces.UserRoleRepository) *GetUserRolesHandler {
	return &GetUserRolesHandler{userRoleRepo: userRoleRepo}
}

func (h *GetUserRolesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetUserRolesQuery)
	return h.userRoleRepo.GetUserRoles(ctx, q.UserID)
}

// GetUserPermissionsHandler handles user permissions retrieval
type GetUserPermissionsHandler struct {
	rolePermissionRepo interfaces.RolePermissionRepository
}

func NewGetUserPermissionsHandler(rolePermissionRepo interfaces.RolePermissionRepository) *GetUserPermissionsHandler {
	return &GetUserPermissionsHandler{rolePermissionRepo: rolePermissionRepo}
}

func (h *GetUserPermissionsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetUserPermissionsQuery)
	return h.rolePermissionRepo.GetUserPermissions(ctx, q.UserID)
}

// GetRolePermissionsHandler handles role permissions retrieval
type GetRolePermissionsHandler struct {
	rolePermissionRepo interfaces.RolePermissionRepository
}

func NewGetRolePermissionsHandler(rolePermissionRepo interfaces.RolePermissionRepository) *GetRolePermissionsHandler {
	return &GetRolePermissionsHandler{rolePermissionRepo: rolePermissionRepo}
}

func (h *GetRolePermissionsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetRolePermissionsQuery)
	return h.rolePermissionRepo.GetRolePermissions(ctx, q.RoleID)
}

// CheckPermissionHandler handles permission checking
type CheckPermissionHandler struct {
	rolePermissionRepo interfaces.RolePermissionRepository
}

func NewCheckPermissionHandler(rolePermissionRepo interfaces.RolePermissionRepository) *CheckPermissionHandler {
	return &CheckPermissionHandler{rolePermissionRepo: rolePermissionRepo}
}

func (h *CheckPermissionHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*CheckPermissionQuery)

	// Get user permissions
	permissions, err := h.rolePermissionRepo.GetUserPermissions(ctx, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// Check if user has the permission
	for _, perm := range permissions {
		if perm.Resource == q.Resource && perm.Action == q.Action {
			return &CheckPermissionResult{HasPermission: true}, nil
		}
	}

	return &CheckPermissionResult{HasPermission: false}, nil
}
