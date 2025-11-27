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

type CreateLocationCommand struct {
	cqrs.BaseCommand
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

func (c *CreateLocationCommand) CommandName() string {
	return "create_location"
}

func (c *CreateLocationCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type UpdateLocationCommand struct {
	cqrs.BaseCommand
	Name     *string `json:"name"`
	ParentID *string `json:"parent_id"`
}

func (c *UpdateLocationCommand) CommandName() string {
	return "update_location"
}

func (c *UpdateLocationCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

type DeleteLocationCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteLocationCommand) CommandName() string {
	return "delete_location"
}

func (c *DeleteLocationCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type CreateLocationHandler struct {
	locationRepo interfaces.LocationRepository
}

func NewCreateLocationHandler(locationRepo interfaces.LocationRepository) *CreateLocationHandler {
	return &CreateLocationHandler{
		locationRepo: locationRepo,
	}
}

func (h *CreateLocationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateLocationCommand)

	locationEntity := &domain.Location{
		Name:     createCmd.Name,
		ParentID: createCmd.ParentID,
	}

	if err := h.locationRepo.Create(ctx, locationEntity); err != nil {
		return err
	}

	createCmd.AggregateID = locationEntity.ID
	return nil
}

type UpdateLocationHandler struct {
	locationRepo interfaces.LocationRepository
}

func NewUpdateLocationHandler(locationRepo interfaces.LocationRepository) *UpdateLocationHandler {
	return &UpdateLocationHandler{
		locationRepo: locationRepo,
	}
}

func (h *UpdateLocationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateLocationCommand)

	locationEntity, err := h.locationRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}

	if locationEntity == nil {
		return errors.New("location not found")
	}

	if updateCmd.Name != nil {
		locationEntity.Name = *updateCmd.Name
	}

	if updateCmd.ParentID != nil {
		locationEntity.ParentID = updateCmd.ParentID
	}

	if err := h.locationRepo.Update(ctx, locationEntity); err != nil {
		return err
	}

	return nil
}

type DeleteLocationHandler struct {
	locationRepo interfaces.LocationRepository
}

func NewDeleteLocationHandler(locationRepo interfaces.LocationRepository) *DeleteLocationHandler {
	return &DeleteLocationHandler{
		locationRepo: locationRepo,
	}
}

func (h *DeleteLocationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteLocationCommand)

	_, err := h.locationRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	return h.locationRepo.Delete(ctx, deleteCmd.AggregateID)
}
