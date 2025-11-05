package interfaces

import (
	"context"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// SizeRepository defines the interface for size data access
type SizeRepository interface {
	// Create creates a new size
	Create(ctx context.Context, size *domain.Size) error

	// GetByID retrieves a size by ID
	GetByID(ctx context.Context, id string) (*domain.Size, error)

	// GetByName retrieves a size by name
	GetByName(ctx context.Context, name string) (*domain.Size, error)

	// List retrieves sizes with pagination and filtering
	List(ctx context.Context, offset, limit int, unit string, minValue, maxValue *float64, includeDeleted bool) ([]*domain.Size, int64, error)

	// Update updates an existing size
	Update(ctx context.Context, size *domain.Size) error

	// Delete soft deletes a size
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a size
	HardDelete(ctx context.Context, id string) error
}
