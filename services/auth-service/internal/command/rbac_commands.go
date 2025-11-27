package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// AssignRoleCommand assigns a role to a user
type AssignRoleCommand struct {
	cqrs.BaseCommand
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

func (c *AssignRoleCommand) CommandName() string {
	return "assign_role"
}

func (c *AssignRoleCommand) Validate() error {
	if c.UserID == "" {
		return errors.New("user_id is required")
	}
	if c.RoleID == "" {
		return errors.New("role_id is required")
	}
	return nil
}

// RevokeRoleCommand removes a role from a user
type RevokeRoleCommand struct {
	cqrs.BaseCommand
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

func (c *RevokeRoleCommand) CommandName() string {
	return "revoke_role"
}

func (c *RevokeRoleCommand) Validate() error {
	if c.UserID == "" {
		return errors.New("user_id is required")
	}
	if c.RoleID == "" {
		return errors.New("role_id is required")
	}
	return nil
}

// AssignPermissionCommand assigns a permission to a role
type AssignPermissionCommand struct {
	cqrs.BaseCommand
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
}

func (c *AssignPermissionCommand) CommandName() string {
	return "assign_permission"
}

func (c *AssignPermissionCommand) Validate() error {
	if c.RoleID == "" {
		return errors.New("role_id is required")
	}
	if c.PermissionID == "" {
		return errors.New("permission_id is required")
	}
	return nil
}

// RevokePermissionCommand removes a permission from a role
type RevokePermissionCommand struct {
	cqrs.BaseCommand
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
}

func (c *RevokePermissionCommand) CommandName() string {
	return "revoke_permission"
}

func (c *RevokePermissionCommand) Validate() error {
	if c.RoleID == "" {
		return errors.New("role_id is required")
	}
	if c.PermissionID == "" {
		return errors.New("permission_id is required")
	}
	return nil
}

// Command Handlers

// AssignRoleHandler handles role assignment to users
type AssignRoleHandler struct {
	userRoleRepo interfaces.UserRoleRepository
	roleRepo     interfaces.RoleRepository
}

func NewAssignRoleHandler(userRoleRepo interfaces.UserRoleRepository, roleRepo interfaces.RoleRepository) *AssignRoleHandler {
	return &AssignRoleHandler{
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
	}
}

func (h *AssignRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	assignCmd := cmd.(*AssignRoleCommand)

	// Check if role exists
	role, err := h.roleRepo.GetByID(ctx, assignCmd.RoleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Check if already assigned
	hasRole, err := h.userRoleRepo.HasRole(ctx, assignCmd.UserID, assignCmd.RoleID)
	if err != nil {
		return fmt.Errorf("failed to check role assignment: %w", err)
	}
	if hasRole {
		return fmt.Errorf("user already has this role")
	}

	// Assign role
	userRole := &domain.UserRole{
		UserID:    assignCmd.UserID,
		RoleID:    assignCmd.RoleID,
		CreatedAt: time.Now(),
	}

	return h.userRoleRepo.AssignRole(ctx, userRole)
}

// RevokeRoleHandler handles role revocation from users
type RevokeRoleHandler struct {
	userRoleRepo interfaces.UserRoleRepository
}

func NewRevokeRoleHandler(userRoleRepo interfaces.UserRoleRepository) *RevokeRoleHandler {
	return &RevokeRoleHandler{userRoleRepo: userRoleRepo}
}

func (h *RevokeRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	revokeCmd := cmd.(*RevokeRoleCommand)

	// Check if user has the role
	hasRole, err := h.userRoleRepo.HasRole(ctx, revokeCmd.UserID, revokeCmd.RoleID)
	if err != nil {
		return fmt.Errorf("failed to check role assignment: %w", err)
	}
	if !hasRole {
		return fmt.Errorf("user does not have this role")
	}

	return h.userRoleRepo.RevokeRole(ctx, revokeCmd.UserID, revokeCmd.RoleID)
}

// AssignPermissionHandler handles permission assignment to roles
type AssignPermissionHandler struct {
	rolePermissionRepo interfaces.RolePermissionRepository
	roleRepo           interfaces.RoleRepository
	permissionRepo     interfaces.PermissionRepository
}

func NewAssignPermissionHandler(
	rolePermissionRepo interfaces.RolePermissionRepository,
	roleRepo interfaces.RoleRepository,
	permissionRepo interfaces.PermissionRepository,
) *AssignPermissionHandler {
	return &AssignPermissionHandler{
		rolePermissionRepo: rolePermissionRepo,
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
	}
}

func (h *AssignPermissionHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	assignCmd := cmd.(*AssignPermissionCommand)

	// Check if role exists
	role, err := h.roleRepo.GetByID(ctx, assignCmd.RoleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Check if permission exists
	permission, err := h.permissionRepo.GetByID(ctx, assignCmd.PermissionID)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}
	if permission == nil {
		return fmt.Errorf("permission not found")
	}

	// Check if already assigned
	hasPermission, err := h.rolePermissionRepo.HasPermission(ctx, assignCmd.RoleID, assignCmd.PermissionID)
	if err != nil {
		return fmt.Errorf("failed to check permission assignment: %w", err)
	}
	if hasPermission {
		return fmt.Errorf("role already has this permission")
	}

	// Assign permission
	rolePermission := &domain.RolePermission{
		RoleID:       assignCmd.RoleID,
		PermissionID: assignCmd.PermissionID,
		CreatedAt:    time.Now(),
	}

	return h.rolePermissionRepo.AssignPermission(ctx, rolePermission)
}

// RevokePermissionHandler handles permission revocation from roles
type RevokePermissionHandler struct {
	rolePermissionRepo interfaces.RolePermissionRepository
}

func NewRevokePermissionHandler(rolePermissionRepo interfaces.RolePermissionRepository) *RevokePermissionHandler {
	return &RevokePermissionHandler{rolePermissionRepo: rolePermissionRepo}
}

func (h *RevokePermissionHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	revokeCmd := cmd.(*RevokePermissionCommand)

	// Check if role has the permission
	hasPermission, err := h.rolePermissionRepo.HasPermission(ctx, revokeCmd.RoleID, revokeCmd.PermissionID)
	if err != nil {
		return fmt.Errorf("failed to check permission assignment: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("role does not have this permission")
	}

	return h.rolePermissionRepo.RevokePermission(ctx, revokeCmd.RoleID, revokeCmd.PermissionID)
}
