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

type CreateSizeCommand struct {
	cqrs.BaseCommand
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func (c *CreateSizeCommand) CommandName() string {
	return "create_size"
}

func (c *CreateSizeCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Unit == "" {
		return errors.New("unit is required")
	}
	return nil
}

type UpdateSizeCommand struct {
	cqrs.BaseCommand
	Name  *string  `json:"name"`
	Value *float64 `json:"value"`
	Unit  *string  `json:"unit"`
}

func (c *UpdateSizeCommand) CommandName() string {
	return "update_size"
}

func (c *UpdateSizeCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

type DeleteSizeCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteSizeCommand) CommandName() string {
	return "delete_size"
}

func (c *DeleteSizeCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type CreateSizeHandler struct {
	sizeRepo interfaces.SizeRepository
}

func NewCreateSizeHandler(sizeRepo interfaces.SizeRepository) *CreateSizeHandler {
	return &CreateSizeHandler{
		sizeRepo: sizeRepo,
	}
}

func (h *CreateSizeHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateSizeCommand)

	sizeEntity := &domain.Size{
		Name:  createCmd.Name,
		Value: createCmd.Value,
		Unit:  createCmd.Unit,
	}

	if err := h.sizeRepo.Create(ctx, sizeEntity); err != nil {
		return err
	}

	createCmd.AggregateID = sizeEntity.ID
	return nil
}

type UpdateSizeHandler struct {
	sizeRepo interfaces.SizeRepository
}

func NewUpdateSizeHandler(sizeRepo interfaces.SizeRepository) *UpdateSizeHandler {
	return &UpdateSizeHandler{
		sizeRepo: sizeRepo,
	}
}

func (h *UpdateSizeHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateSizeCommand)

	sizeEntity, err := h.sizeRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}

	if sizeEntity == nil {
		return errors.New("size not found")
	}

	if updateCmd.Name != nil {
		sizeEntity.Name = *updateCmd.Name
	}

	if updateCmd.Value != nil {
		sizeEntity.Value = *updateCmd.Value
	}

	if updateCmd.Unit != nil {
		sizeEntity.Unit = *updateCmd.Unit
	}

	if err := h.sizeRepo.Update(ctx, sizeEntity); err != nil {
		return err
	}

	return nil
}

type DeleteSizeHandler struct {
	sizeRepo interfaces.SizeRepository
}

func NewDeleteSizeHandler(sizeRepo interfaces.SizeRepository) *DeleteSizeHandler {
	return &DeleteSizeHandler{
		sizeRepo: sizeRepo,
	}
}

func (h *DeleteSizeHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteSizeCommand)

	_, err := h.sizeRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	return h.sizeRepo.Delete(ctx, deleteCmd.AggregateID)
}
