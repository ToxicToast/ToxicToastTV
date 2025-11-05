package interfaces

import (
	"context"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// TypeRepository defines the interface for type data access
type TypeRepository interface {
	// Create creates a new type
	Create(ctx context.Context, typeEntity *domain.Type) error

	// GetByID retrieves a type by ID
	GetByID(ctx context.Context, id string) (*domain.Type, error)

	// GetBySlug retrieves a type by slug
	GetBySlug(ctx context.Context, slug string) (*domain.Type, error)

	// List retrieves types with pagination and filtering
	List(ctx context.Context, offset, limit int, search string, includeDeleted bool) ([]*domain.Type, int64, error)

	// Update updates an existing type
	Update(ctx context.Context, typeEntity *domain.Type) error

	// Delete soft deletes a type
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a type
	HardDelete(ctx context.Context, id string) error
}
