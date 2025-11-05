package usecase

import (
	"context"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/pkg/config"
	"toxictoast/services/blog-service/pkg/utils"
)

type CategoryUseCase interface {
	CreateCategory(ctx context.Context, input CreateCategoryInput) (*domain.Category, error)
	GetCategory(ctx context.Context, id string) (*domain.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error)
	UpdateCategory(ctx context.Context, id string, input UpdateCategoryInput) (*domain.Category, error)
	DeleteCategory(ctx context.Context, id string) error
	ListCategories(ctx context.Context, filters repository.CategoryFilters) ([]domain.Category, int64, error)
	GetChildren(ctx context.Context, parentID string) ([]domain.Category, error)
}

type CreateCategoryInput struct {
	Name        string
	Description string
	ParentID    *string
}

type UpdateCategoryInput struct {
	Name        *string
	Description *string
	ParentID    *string
}

type categoryUseCase struct {
	repo          repository.CategoryRepository
	kafkaProducer *kafka.Producer
	config        *config.Config
}

func NewCategoryUseCase(
	repo repository.CategoryRepository,
	kafkaProducer *kafka.Producer,
	cfg *config.Config,
) CategoryUseCase {
	return &categoryUseCase{
		repo:          repo,
		kafkaProducer: kafkaProducer,
		config:        cfg,
	}
}

func (uc *categoryUseCase) CreateCategory(ctx context.Context, input CreateCategoryInput) (*domain.Category, error) {
	// Generate slug from name
	slug := uc.generateUniqueSlug(ctx, input.Name)

	// Validate parent exists if provided
	if input.ParentID != nil {
		_, err := uc.repo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
	}

	// Create category entity
	category := &domain.Category{
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		ParentID:    input.ParentID,
	}

	// Save to database
	if err := uc.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Publish Kafka event (optional - no event defined yet, skip for now)
	// if uc.kafkaProducer != nil { ... }

	return category, nil
}

func (uc *categoryUseCase) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *categoryUseCase) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	return uc.repo.GetBySlug(ctx, slug)
}

func (uc *categoryUseCase) UpdateCategory(ctx context.Context, id string, input UpdateCategoryInput) (*domain.Category, error) {
	// Get existing category
	category, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		category.Name = *input.Name
		// Regenerate slug if name changed
		category.Slug = uc.generateUniqueSlug(ctx, category.Name)
	}

	if input.Description != nil {
		category.Description = *input.Description
	}

	if input.ParentID != nil {
		// Validate parent exists
		if *input.ParentID != "" {
			_, err := uc.repo.GetByID(ctx, *input.ParentID)
			if err != nil {
				return nil, fmt.Errorf("parent category not found: %w", err)
			}

			// Prevent circular reference (category can't be its own parent)
			if *input.ParentID == id {
				return nil, fmt.Errorf("category cannot be its own parent")
			}
		}
		category.ParentID = input.ParentID
	}

	// Save to database
	if err := uc.repo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

func (uc *categoryUseCase) DeleteCategory(ctx context.Context, id string) error {
	// Check if category exists
	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if category has children
	children, err := uc.repo.GetChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}

	if len(children) > 0 {
		return fmt.Errorf("cannot delete category with children")
	}

	// Delete from database
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

func (uc *categoryUseCase) ListCategories(ctx context.Context, filters repository.CategoryFilters) ([]domain.Category, int64, error) {
	return uc.repo.List(ctx, filters)
}

func (uc *categoryUseCase) GetChildren(ctx context.Context, parentID string) ([]domain.Category, error) {
	return uc.repo.GetChildren(ctx, parentID)
}

// Helper methods

func (uc *categoryUseCase) generateUniqueSlug(ctx context.Context, name string) string {
	exists := func(slug string) bool {
		exists, err := uc.repo.SlugExists(ctx, slug)
		if err != nil {
			return false
		}
		return exists
	}

	return utils.GenerateUniqueSlug(name, exists)
}
