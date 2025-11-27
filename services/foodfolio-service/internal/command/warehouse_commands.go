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

type CreateWarehouseCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *CreateWarehouseCommand) CommandName() string {
	return "create_warehouse"
}

func (c *CreateWarehouseCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type UpdateWarehouseCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *UpdateWarehouseCommand) CommandName() string {
	return "update_warehouse"
}

func (c *UpdateWarehouseCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type DeleteWarehouseCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteWarehouseCommand) CommandName() string {
	return "delete_warehouse"
}

func (c *DeleteWarehouseCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type CreateWarehouseHandler struct {
	warehouseRepo interfaces.WarehouseRepository
}

func NewCreateWarehouseHandler(warehouseRepo interfaces.WarehouseRepository) *CreateWarehouseHandler {
	return &CreateWarehouseHandler{
		warehouseRepo: warehouseRepo,
	}
}

func (h *CreateWarehouseHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateWarehouseCommand)

	warehouseEntity := &domain.Warehouse{
		Name: createCmd.Name,
	}

	if err := h.warehouseRepo.Create(ctx, warehouseEntity); err != nil {
		return err
	}

	createCmd.AggregateID = warehouseEntity.ID
	return nil
}

type UpdateWarehouseHandler struct {
	warehouseRepo interfaces.WarehouseRepository
}

func NewUpdateWarehouseHandler(warehouseRepo interfaces.WarehouseRepository) *UpdateWarehouseHandler {
	return &UpdateWarehouseHandler{
		warehouseRepo: warehouseRepo,
	}
}

func (h *UpdateWarehouseHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateWarehouseCommand)

	warehouseEntity, err := h.warehouseRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}

	if warehouseEntity == nil {
		return errors.New("warehouse not found")
	}

	warehouseEntity.Name = updateCmd.Name

	if err := h.warehouseRepo.Update(ctx, warehouseEntity); err != nil {
		return err
	}

	return nil
}

type DeleteWarehouseHandler struct {
	warehouseRepo interfaces.WarehouseRepository
}

func NewDeleteWarehouseHandler(warehouseRepo interfaces.WarehouseRepository) *DeleteWarehouseHandler {
	return &DeleteWarehouseHandler{
		warehouseRepo: warehouseRepo,
	}
}

func (h *DeleteWarehouseHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteWarehouseCommand)

	_, err := h.warehouseRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	return h.warehouseRepo.Delete(ctx, deleteCmd.AggregateID)
}
