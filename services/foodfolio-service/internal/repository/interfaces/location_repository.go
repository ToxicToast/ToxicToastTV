package interfaces

import (
	"context"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// LocationRepository defines the interface for location data access
type LocationRepository interface {
	// Create creates a new location
	Create(ctx context.Context, location *domain.Location) error

	// GetByID retrieves a location by ID
	GetByID(ctx context.Context, id string) (*domain.Location, error)

	// GetBySlug retrieves a location by slug
	GetBySlug(ctx context.Context, slug string) (*domain.Location, error)

	// List retrieves locations with pagination and filtering
	List(ctx context.Context, offset, limit int, parentID *string, includeChildren bool, includeDeleted bool) ([]*domain.Location, int64, error)

	// GetTree retrieves the full location tree or subtree
	GetTree(ctx context.Context, rootID *string, maxDepth int) ([]*domain.Location, error)

	// GetChildren retrieves direct children of a location
	GetChildren(ctx context.Context, parentID string) ([]*domain.Location, error)

	// GetRootLocations retrieves all root locations (no parent)
	GetRootLocations(ctx context.Context) ([]*domain.Location, error)

	// Update updates an existing location
	Update(ctx context.Context, location *domain.Location) error

	// Delete soft deletes a location
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a location
	HardDelete(ctx context.Context, id string) error
}
