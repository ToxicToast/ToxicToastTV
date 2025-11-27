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

type CreateCategoryCommand struct {
	cqrs.BaseCommand
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

func (c *CreateCategoryCommand) CommandName() string {
	return "create_category"
}

func (c *CreateCategoryCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type UpdateCategoryCommand struct {
	cqrs.BaseCommand
	Name     *string `json:"name"`
	ParentID *string `json:"parent_id"`
}

func (c *UpdateCategoryCommand) CommandName() string {
	return "update_category"
}

func (c *UpdateCategoryCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

type DeleteCategoryCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteCategoryCommand) CommandName() string {
	return "delete_category"
}

func (c *DeleteCategoryCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

type CreateCategoryHandler struct {
	categoryRepo interfaces.CategoryRepository
}

func NewCreateCategoryHandler(categoryRepo interfaces.CategoryRepository) *CreateCategoryHandler {
	return &CreateCategoryHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *CreateCategoryHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateCategoryCommand)

	categoryEntity := &domain.Category{
		Name:     createCmd.Name,
		ParentID: createCmd.ParentID,
	}

	if err := h.categoryRepo.Create(ctx, categoryEntity); err != nil {
		return err
	}

	createCmd.AggregateID = categoryEntity.ID
	return nil
}

type UpdateCategoryHandler struct {
	categoryRepo interfaces.CategoryRepository
}

func NewUpdateCategoryHandler(categoryRepo interfaces.CategoryRepository) *UpdateCategoryHandler {
	return &UpdateCategoryHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *UpdateCategoryHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateCategoryCommand)

	categoryEntity, err := h.categoryRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}

	if categoryEntity == nil {
		return errors.New("category not found")
	}

	if updateCmd.Name != nil {
		categoryEntity.Name = *updateCmd.Name
	}

	if updateCmd.ParentID != nil {
		categoryEntity.ParentID = updateCmd.ParentID
	}

	if err := h.categoryRepo.Update(ctx, categoryEntity); err != nil {
		return err
	}

	return nil
}

type DeleteCategoryHandler struct {
	categoryRepo interfaces.CategoryRepository
}

func NewDeleteCategoryHandler(categoryRepo interfaces.CategoryRepository) *DeleteCategoryHandler {
	return &DeleteCategoryHandler{
		categoryRepo: categoryRepo,
	}
}

func (h *DeleteCategoryHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteCategoryCommand)

	_, err := h.categoryRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	return h.categoryRepo.Delete(ctx, deleteCmd.AggregateID)
}
