package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/auth-service/internal/domain"
	"toxictoast/services/auth-service/internal/repository/interfaces"
)

// CreatePermissionCommand creates a new permission
type CreatePermissionCommand struct {
	cqrs.BaseCommand
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

func (c *CreatePermissionCommand) CommandName() string {
	return "create_permission"
}

func (c *CreatePermissionCommand) Validate() error {
	if c.Resource == "" {
		return errors.New("resource is required")
	}
	if c.Action == "" {
		return errors.New("action is required")
	}
	return nil
}

// UpdatePermissionCommand updates an existing permission
type UpdatePermissionCommand struct {
	cqrs.BaseCommand
	Resource    *string `json:"resource,omitempty"`
	Action      *string `json:"action,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (c *UpdatePermissionCommand) CommandName() string {
	return "update_permission"
}

func (c *UpdatePermissionCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("permission_id is required")
	}
	return nil
}

// DeletePermissionCommand deletes a permission
type DeletePermissionCommand struct {
	cqrs.BaseCommand
}

func (c *DeletePermissionCommand) CommandName() string {
	return "delete_permission"
}

func (c *DeletePermissionCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("permission_id is required")
	}
	return nil
}

// Command Handlers

// CreatePermissionHandler handles permission creation
type CreatePermissionHandler struct {
	permissionRepo interfaces.PermissionRepository
}

func NewCreatePermissionHandler(permissionRepo interfaces.PermissionRepository) *CreatePermissionHandler {
	return &CreatePermissionHandler{permissionRepo: permissionRepo}
}

func (h *CreatePermissionHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreatePermissionCommand)

	// Check if permission with same resource:action exists
	existing, err := h.permissionRepo.GetByResourceAction(ctx, createCmd.Resource, createCmd.Action)
	if err != nil {
		return fmt.Errorf("failed to check existing permission: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("permission '%s:%s' already exists", createCmd.Resource, createCmd.Action)
	}

	permission := &domain.Permission{
		ID:          uuid.New().String(),
		Resource:    createCmd.Resource,
		Action:      createCmd.Action,
		Description: createCmd.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.permissionRepo.Create(ctx, permission); err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// UpdatePermissionHandler handles permission updates
type UpdatePermissionHandler struct {
	permissionRepo interfaces.PermissionRepository
}

func NewUpdatePermissionHandler(permissionRepo interfaces.PermissionRepository) *UpdatePermissionHandler {
	return &UpdatePermissionHandler{permissionRepo: permissionRepo}
}

func (h *UpdatePermissionHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdatePermissionCommand)

	// Get existing permission
	permission, err := h.permissionRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}
	if permission == nil {
		return fmt.Errorf("permission not found")
	}

	// Update fields
	if updateCmd.Resource != nil && *updateCmd.Resource != "" {
		permission.Resource = *updateCmd.Resource
	}

	if updateCmd.Action != nil && *updateCmd.Action != "" {
		permission.Action = *updateCmd.Action
	}

	// Check if updated resource:action conflicts with another permission
	if updateCmd.Resource != nil || updateCmd.Action != nil {
		existing, err := h.permissionRepo.GetByResourceAction(ctx, permission.Resource, permission.Action)
		if err != nil {
			return fmt.Errorf("failed to check existing permission: %w", err)
		}
		if existing != nil && existing.ID != updateCmd.AggregateID {
			return fmt.Errorf("permission '%s:%s' already exists", permission.Resource, permission.Action)
		}
	}

	if updateCmd.Description != nil {
		permission.Description = *updateCmd.Description
	}

	permission.UpdatedAt = time.Now()

	if err := h.permissionRepo.Update(ctx, permission); err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

// DeletePermissionHandler handles permission deletion
type DeletePermissionHandler struct {
	permissionRepo interfaces.PermissionRepository
}

func NewDeletePermissionHandler(permissionRepo interfaces.PermissionRepository) *DeletePermissionHandler {
	return &DeletePermissionHandler{permissionRepo: permissionRepo}
}

func (h *DeletePermissionHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeletePermissionCommand)

	// Check if permission exists
	permission, err := h.permissionRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}
	if permission == nil {
		return fmt.Errorf("permission not found")
	}

	return h.permissionRepo.Delete(ctx, deleteCmd.AggregateID)
}
