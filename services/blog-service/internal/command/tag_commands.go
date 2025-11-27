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

// CreateTagCommand creates a new tag
type CreateTagCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *CreateTagCommand) CommandName() string {
	return "create_tag"
}

func (c *CreateTagCommand) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// UpdateTagCommand updates an existing tag
type UpdateTagCommand struct {
	cqrs.BaseCommand
	Name string `json:"name"`
}

func (c *UpdateTagCommand) CommandName() string {
	return "update_tag"
}

func (c *UpdateTagCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("tag_id is required")
	}
	if c.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// DeleteTagCommand deletes a tag
type DeleteTagCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteTagCommand) CommandName() string {
	return "delete_tag"
}

func (c *DeleteTagCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("tag_id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

// CreateTagHandler handles tag creation
type CreateTagHandler struct {
	tagRepo       repository.TagRepository
	kafkaProducer *kafka.Producer
}

func NewCreateTagHandler(
	tagRepo repository.TagRepository,
	kafkaProducer *kafka.Producer,
) *CreateTagHandler {
	return &CreateTagHandler{
		tagRepo:       tagRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *CreateTagHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateTagCommand)

	// Generate unique slug from name
	slug := h.generateUniqueSlug(ctx, createCmd.Name)

	// Generate UUID for tag
	tagID := uuid.New().String()

	// Create tag entity
	tag := &domain.Tag{
		ID:   tagID,
		Name: createCmd.Name,
		Slug: slug,
	}

	// Save to database
	if err := h.tagRepo.Create(ctx, tag); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	// Store tag ID in command result
	createCmd.AggregateID = tagID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.TagCreatedEvent{
			TagID:     tag.ID,
			Name:      tag.Name,
			Slug:      tag.Slug,
			CreatedAt: tag.CreatedAt,
		}
		if err := h.kafkaProducer.PublishTagCreated("blog.tag.created", event); err != nil {
			fmt.Printf("Warning: Failed to publish tag created event: %v\n", err)
		}
	}

	return nil
}

func (h *CreateTagHandler) generateUniqueSlug(ctx context.Context, name string) string {
	exists := func(slug string) bool {
		exists, err := h.tagRepo.SlugExists(ctx, slug)
		if err != nil {
			return false
		}
		return exists
	}

	return utils.GenerateUniqueSlug(name, exists)
}

// UpdateTagHandler handles tag updates
type UpdateTagHandler struct {
	tagRepo       repository.TagRepository
	kafkaProducer *kafka.Producer
}

func NewUpdateTagHandler(
	tagRepo repository.TagRepository,
	kafkaProducer *kafka.Producer,
) *UpdateTagHandler {
	return &UpdateTagHandler{
		tagRepo:       tagRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *UpdateTagHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateTagCommand)

	// Get existing tag
	tag, err := h.tagRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get tag: %w", err)
	}

	// Update name and regenerate slug
	tag.Name = updateCmd.Name
	tag.Slug = h.generateUniqueSlug(ctx, tag.Name)

	// Save to database
	if err := h.tagRepo.Update(ctx, tag); err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.TagUpdatedEvent{
			TagID:     tag.ID,
			Name:      tag.Name,
			Slug:      tag.Slug,
			UpdatedAt: tag.UpdatedAt,
		}
		if err := h.kafkaProducer.PublishTagUpdated("blog.tag.updated", event); err != nil {
			fmt.Printf("Warning: Failed to publish tag updated event: %v\n", err)
		}
	}

	return nil
}

func (h *UpdateTagHandler) generateUniqueSlug(ctx context.Context, name string) string {
	exists := func(slug string) bool {
		exists, err := h.tagRepo.SlugExists(ctx, slug)
		if err != nil {
			return false
		}
		return exists
	}

	return utils.GenerateUniqueSlug(name, exists)
}

// DeleteTagHandler handles tag deletion
type DeleteTagHandler struct {
	tagRepo       repository.TagRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteTagHandler(
	tagRepo repository.TagRepository,
	kafkaProducer *kafka.Producer,
) *DeleteTagHandler {
	return &DeleteTagHandler{
		tagRepo:       tagRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteTagHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteTagCommand)

	// Check if tag exists
	_, err := h.tagRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	// Delete from database (soft delete)
	// Note: Many-to-many relationship with posts will be handled by GORM cascading
	if err := h.tagRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.TagDeletedEvent{
			TagID:     deleteCmd.AggregateID,
			DeletedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishTagDeleted("blog.tag.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish tag deleted event: %v\n", err)
		}
	}

	return nil
}
