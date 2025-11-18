package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/pkg/config"
	"toxictoast/services/blog-service/pkg/utils"
)

type TagUseCase interface {
	CreateTag(ctx context.Context, input CreateTagInput) (*domain.Tag, error)
	GetTag(ctx context.Context, id string) (*domain.Tag, error)
	GetTagBySlug(ctx context.Context, slug string) (*domain.Tag, error)
	UpdateTag(ctx context.Context, id string, input UpdateTagInput) (*domain.Tag, error)
	DeleteTag(ctx context.Context, id string) error
	ListTags(ctx context.Context, filters repository.TagFilters) ([]domain.Tag, int64, error)
}

type CreateTagInput struct {
	Name string
}

type UpdateTagInput struct {
	Name string
}

type tagUseCase struct {
	repo          repository.TagRepository
	kafkaProducer *kafka.Producer
	config        *config.Config
}

func NewTagUseCase(
	repo repository.TagRepository,
	kafkaProducer *kafka.Producer,
	cfg *config.Config,
) TagUseCase {
	return &tagUseCase{
		repo:          repo,
		kafkaProducer: kafkaProducer,
		config:        cfg,
	}
}

func (uc *tagUseCase) CreateTag(ctx context.Context, input CreateTagInput) (*domain.Tag, error) {
	// Generate slug from name
	slug := uc.generateUniqueSlug(ctx, input.Name)

	// Generate UUID for tag
	tagID := uuid.New().String()

	// Create tag entity
	tag := &domain.Tag{
		ID:   tagID,
		Name: input.Name,
		Slug: slug,
	}

	// Save to database
	if err := uc.repo.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.TagCreatedEvent{
			TagID:     tag.ID,
			Name:      tag.Name,
			Slug:      tag.Slug,
			CreatedAt: tag.CreatedAt,
		}
		if err := uc.kafkaProducer.PublishTagCreated("blog.tag.created", event); err != nil {
			fmt.Printf("Warning: Failed to publish tag created event: %v\n", err)
		}
	}

	return tag, nil
}

func (uc *tagUseCase) GetTag(ctx context.Context, id string) (*domain.Tag, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *tagUseCase) GetTagBySlug(ctx context.Context, slug string) (*domain.Tag, error) {
	return uc.repo.GetBySlug(ctx, slug)
}

func (uc *tagUseCase) UpdateTag(ctx context.Context, id string, input UpdateTagInput) (*domain.Tag, error) {
	// Get existing tag
	tag, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update name and regenerate slug
	tag.Name = input.Name
	tag.Slug = uc.generateUniqueSlug(ctx, tag.Name)

	// Save to database
	if err := uc.repo.Update(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.TagUpdatedEvent{
			TagID:     tag.ID,
			Name:      tag.Name,
			Slug:      tag.Slug,
			UpdatedAt: tag.UpdatedAt,
		}
		if err := uc.kafkaProducer.PublishTagUpdated("blog.tag.updated", event); err != nil {
			fmt.Printf("Warning: Failed to publish tag updated event: %v\n", err)
		}
	}

	return tag, nil
}

func (uc *tagUseCase) DeleteTag(ctx context.Context, id string) error {
	// Check if tag exists
	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	// Note: Many-to-many relationship with posts will be handled by GORM cascading
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.TagDeletedEvent{
			TagID:     id,
			DeletedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishTagDeleted("blog.tag.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish tag deleted event: %v\n", err)
		}
	}

	return nil
}

func (uc *tagUseCase) ListTags(ctx context.Context, filters repository.TagFilters) ([]domain.Tag, int64, error) {
	return uc.repo.List(ctx, filters)
}

// Helper methods

func (uc *tagUseCase) generateUniqueSlug(ctx context.Context, name string) string {
	exists := func(slug string) bool {
		exists, err := uc.repo.SlugExists(ctx, slug)
		if err != nil {
			return false
		}
		return exists
	}

	return utils.GenerateUniqueSlug(name, exists)
}
