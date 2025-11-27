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

// CreateRoleCommand creates a new role
type CreateRoleCommand struct {
	cqrs.BaseCommand
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *CreateRoleCommand) CommandName() string {
	return "create_role"
}

func (c *CreateRoleCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// UpdateRoleCommand updates an existing role
type UpdateRoleCommand struct {
	cqrs.BaseCommand
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (c *UpdateRoleCommand) CommandName() string {
	return "update_role"
}

func (c *UpdateRoleCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("role_id is required")
	}
	return nil
}

// DeleteRoleCommand deletes a role
type DeleteRoleCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteRoleCommand) CommandName() string {
	return "delete_role"
}

func (c *DeleteRoleCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("role_id is required")
	}
	return nil
}

// Command Handlers

// CreateRoleHandler handles role creation
type CreateRoleHandler struct {
	roleRepo interfaces.RoleRepository
}

func NewCreateRoleHandler(roleRepo interfaces.RoleRepository) *CreateRoleHandler {
	return &CreateRoleHandler{roleRepo: roleRepo}
}

func (h *CreateRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateRoleCommand)

	// Check if role with same name exists
	existing, err := h.roleRepo.GetByName(ctx, createCmd.Name)
	if err != nil {
		return fmt.Errorf("failed to check existing role: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("role with name '%s' already exists", createCmd.Name)
	}

	role := &domain.Role{
		ID:          uuid.New().String(),
		Name:        createCmd.Name,
		Description: createCmd.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.roleRepo.Create(ctx, role); err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// UpdateRoleHandler handles role updates
type UpdateRoleHandler struct {
	roleRepo interfaces.RoleRepository
}

func NewUpdateRoleHandler(roleRepo interfaces.RoleRepository) *UpdateRoleHandler {
	return &UpdateRoleHandler{roleRepo: roleRepo}
}

func (h *UpdateRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateRoleCommand)

	// Get existing role
	role, err := h.roleRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Update fields
	if updateCmd.Name != nil && *updateCmd.Name != "" {
		// Check if new name conflicts with another role
		existing, err := h.roleRepo.GetByName(ctx, *updateCmd.Name)
		if err != nil {
			return fmt.Errorf("failed to check existing role: %w", err)
		}
		if existing != nil && existing.ID != updateCmd.AggregateID {
			return fmt.Errorf("role with name '%s' already exists", *updateCmd.Name)
		}
		role.Name = *updateCmd.Name
	}

	if updateCmd.Description != nil {
		role.Description = *updateCmd.Description
	}

	role.UpdatedAt = time.Now()

	if err := h.roleRepo.Update(ctx, role); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// DeleteRoleHandler handles role deletion
type DeleteRoleHandler struct {
	roleRepo interfaces.RoleRepository
}

func NewDeleteRoleHandler(roleRepo interfaces.RoleRepository) *DeleteRoleHandler {
	return &DeleteRoleHandler{roleRepo: roleRepo}
}

func (h *DeleteRoleHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteRoleCommand)

	// Check if role exists
	role, err := h.roleRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	return h.roleRepo.Delete(ctx, deleteCmd.AggregateID)
}
