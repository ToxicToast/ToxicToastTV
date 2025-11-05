package interfaces

import (
	"context"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	// Create creates a new category
	Create(ctx context.Context, category *domain.Category) error

	// GetByID retrieves a category by ID
	GetByID(ctx context.Context, id string) (*domain.Category, error)

	// GetBySlug retrieves a category by slug
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)

	// List retrieves categories with pagination and filtering
	List(ctx context.Context, offset, limit int, parentID *string, includeChildren bool, includeDeleted bool) ([]*domain.Category, int64, error)

	// GetTree retrieves the full category tree or subtree
	GetTree(ctx context.Context, rootID *string, maxDepth int) ([]*domain.Category, error)

	// GetChildren retrieves direct children of a category
	GetChildren(ctx context.Context, parentID string) ([]*domain.Category, error)

	// GetRootCategories retrieves all root categories (no parent)
	GetRootCategories(ctx context.Context) ([]*domain.Category, error)

	// Update updates an existing category
	Update(ctx context.Context, category *domain.Category) error

	// Delete soft deletes a category
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a category
	HardDelete(ctx context.Context, id string) error
}
