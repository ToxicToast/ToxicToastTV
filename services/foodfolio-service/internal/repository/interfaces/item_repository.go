package interfaces

import (
	"context"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// ItemRepository defines the interface for item data access
type ItemRepository interface {
	// Create creates a new item
	Create(ctx context.Context, item *domain.Item) error

	// GetByID retrieves an item by ID
	GetByID(ctx context.Context, id string) (*domain.Item, error)

	// GetBySlug retrieves an item by slug
	GetBySlug(ctx context.Context, slug string) (*domain.Item, error)

	// GetWithVariants retrieves an item with all its variants
	GetWithVariants(ctx context.Context, id string, includeDetails bool) (*domain.Item, error)

	// List retrieves items with pagination and filtering
	List(ctx context.Context, offset, limit int, categoryID, companyID, typeID, search *string, includeDeleted bool) ([]*domain.Item, int64, error)

	// Search searches items by query
	Search(ctx context.Context, query string, offset, limit int, categoryID, companyID *string) ([]*domain.Item, int64, error)

	// Update updates an existing item
	Update(ctx context.Context, item *domain.Item) error

	// Delete soft deletes an item
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes an item
	HardDelete(ctx context.Context, id string) error
}
