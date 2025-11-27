package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/pkg/utils"
)

// ============================================================================
// Commands
// ============================================================================

// CreateCategoryCommand creates a new category
type CreateCategoryCommand struct {
	cqrs.BaseCommand
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id"`
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

// UpdateCategoryCommand updates an existing category
type UpdateCategoryCommand struct {
	cqrs.BaseCommand
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
}

func (c *UpdateCategoryCommand) CommandName() string {
	return "update_category"
}

func (c *UpdateCategoryCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("category_id is required")
	}
	return nil
}

// DeleteCategoryCommand deletes a category
type DeleteCategoryCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteCategoryCommand) CommandName() string {
	return "delete_category"
}

func (c *DeleteCategoryCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("category_id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

// CreateCategoryHandler handles category creation
type CreateCategoryHandler struct {
	categoryRepo  repository.CategoryRepository
	kafkaProducer *kafka.Producer
}

func NewCreateCategoryHandler(
	categoryRepo repository.CategoryRepository,
	kafkaProducer *kafka.Producer,
) *CreateCategoryHandler {
	return &CreateCategoryHandler{
		categoryRepo:  categoryRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *CreateCategoryHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateCategoryCommand)

	// Generate unique slug from name
	slug := h.generateUniqueSlug(ctx, createCmd.Name)

	// Validate parent exists if provided
	if createCmd.ParentID != nil {
		_, err := h.categoryRepo.GetByID(ctx, *createCmd.ParentID)
		if err != nil {
			return fmt.Errorf("parent category not found: %w", err)
		}
	}

	// Generate UUID for category
	categoryID := uuid.New().String()

	// Create category entity
	category := &domain.Category{
		ID:          categoryID,
		Name:        createCmd.Name,
		Slug:        slug,
		Description: createCmd.Description,
		ParentID:    createCmd.ParentID,
	}

	// Save to database
	if err := h.categoryRepo.Create(ctx, category); err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	// Store category ID in command result
	createCmd.AggregateID = categoryID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.CategoryCreatedEvent{
			CategoryID:  category.ID,
			Name:        category.Name,
			Slug:        category.Slug,
			Description: category.Description,
			ParentID:    category.ParentID,
			CreatedAt:   category.CreatedAt,
		}
		if err := h.kafkaProducer.PublishCategoryCreated("blog.category.created", event); err != nil {
			fmt.Printf("Warning: Failed to publish category created event: %v\n", err)
		}
	}

	return nil
}

func (h *CreateCategoryHandler) generateUniqueSlug(ctx context.Context, name string) string {
	exists := func(slug string) bool {
		exists, err := h.categoryRepo.SlugExists(ctx, slug)
		if err != nil {
			return false
		}
		return exists
	}

	return utils.GenerateUniqueSlug(name, exists)
}

// UpdateCategoryHandler handles category updates
type UpdateCategoryHandler struct {
	categoryRepo  repository.CategoryRepository
	kafkaProducer *kafka.Producer
}

func NewUpdateCategoryHandler(
	categoryRepo repository.CategoryRepository,
	kafkaProducer *kafka.Producer,
) *UpdateCategoryHandler {
	return &UpdateCategoryHandler{
		categoryRepo:  categoryRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *UpdateCategoryHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateCategoryCommand)

	// Get existing category
	category, err := h.categoryRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}

	// Update fields if provided
	if updateCmd.Name != nil {
		category.Name = *updateCmd.Name
		// Regenerate slug if name changed
		category.Slug = h.generateUniqueSlug(ctx, category.Name)
	}

	if updateCmd.Description != nil {
		category.Description = *updateCmd.Description
	}

	if updateCmd.ParentID != nil {
		// Validate parent exists
		if *updateCmd.ParentID != "" {
			_, err := h.categoryRepo.GetByID(ctx, *updateCmd.ParentID)
			if err != nil {
				return fmt.Errorf("parent category not found: %w", err)
			}

			// Prevent circular reference (category can't be its own parent)
			if *updateCmd.ParentID == updateCmd.AggregateID {
				return fmt.Errorf("category cannot be its own parent")
			}
		}
		category.ParentID = updateCmd.ParentID
	}

	// Save to database
	if err := h.categoryRepo.Update(ctx, category); err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.CategoryUpdatedEvent{
			CategoryID:  category.ID,
			Name:        category.Name,
			Slug:        category.Slug,
			Description: category.Description,
			ParentID:    category.ParentID,
			UpdatedAt:   category.UpdatedAt,
		}
		if err := h.kafkaProducer.PublishCategoryUpdated("blog.category.updated", event); err != nil {
			fmt.Printf("Warning: Failed to publish category updated event: %v\n", err)
		}
	}

	return nil
}

func (h *UpdateCategoryHandler) generateUniqueSlug(ctx context.Context, name string) string {
	exists := func(slug string) bool {
		exists, err := h.categoryRepo.SlugExists(ctx, slug)
		if err != nil {
			return false
		}
		return exists
	}

	return utils.GenerateUniqueSlug(name, exists)
}

// DeleteCategoryHandler handles category deletion
type DeleteCategoryHandler struct {
	categoryRepo  repository.CategoryRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteCategoryHandler(
	categoryRepo repository.CategoryRepository,
	kafkaProducer *kafka.Producer,
) *DeleteCategoryHandler {
	return &DeleteCategoryHandler{
		categoryRepo:  categoryRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteCategoryHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteCategoryCommand)

	// Check if category exists
	_, err := h.categoryRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Check if category has children
	children, err := h.categoryRepo.GetChildren(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}

	if len(children) > 0 {
		return fmt.Errorf("cannot delete category with children")
	}

	// Delete from database (soft delete)
	if err := h.categoryRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.CategoryDeletedEvent{
			CategoryID: deleteCmd.AggregateID,
			DeletedAt:  time.Now(),
		}
		if err := h.kafkaProducer.PublishCategoryDeleted("blog.category.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish category deleted event: %v\n", err)
		}
	}

	return nil
}
