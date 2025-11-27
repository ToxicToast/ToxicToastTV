package command

import (
	"context"
	"errors"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/repository/interfaces"
)

// ============================================================================
// Commands
// ============================================================================

// CreateTypeCommand creates a new type
type CreateTypeCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *CreateTypeCommand) CommandName() string {
	return "create_type"
}

func (c *CreateTypeCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// UpdateTypeCommand updates a type
type UpdateTypeCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *UpdateTypeCommand) CommandName() string {
	return "update_type"
}

func (c *UpdateTypeCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// DeleteTypeCommand soft-deletes a type
type DeleteTypeCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteTypeCommand) CommandName() string {
	return "delete_type"
}

func (c *DeleteTypeCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

// CreateTypeHandler handles type creation
type CreateTypeHandler struct {
	typeRepo interfaces.TypeRepository
}

func NewCreateTypeHandler(typeRepo interfaces.TypeRepository) *CreateTypeHandler {
	return &CreateTypeHandler{
		typeRepo: typeRepo,
	}
}

func (h *CreateTypeHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateTypeCommand)

	typeEntity := &domain.Type{
		Name: createCmd.Name,
	}

	if err := h.typeRepo.Create(ctx, typeEntity); err != nil {
		return err
	}

	// Set AggregateID to the created entity's ID
	createCmd.AggregateID = typeEntity.ID

	return nil
}

// UpdateTypeHandler handles type updates
type UpdateTypeHandler struct {
	typeRepo interfaces.TypeRepository
}

func NewUpdateTypeHandler(typeRepo interfaces.TypeRepository) *UpdateTypeHandler {
	return &UpdateTypeHandler{
		typeRepo: typeRepo,
	}
}

func (h *UpdateTypeHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateTypeCommand)

	typeEntity, err := h.typeRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}

	if typeEntity == nil {
		return errors.New("type not found")
	}

	typeEntity.Name = updateCmd.Name

	if err := h.typeRepo.Update(ctx, typeEntity); err != nil {
		return err
	}

	return nil
}

// DeleteTypeHandler handles type deletion
type DeleteTypeHandler struct {
	typeRepo interfaces.TypeRepository
}

func NewDeleteTypeHandler(typeRepo interfaces.TypeRepository) *DeleteTypeHandler {
	return &DeleteTypeHandler{
		typeRepo: typeRepo,
	}
}

func (h *DeleteTypeHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteTypeCommand)

	_, err := h.typeRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	return h.typeRepo.Delete(ctx, deleteCmd.AggregateID)
}
