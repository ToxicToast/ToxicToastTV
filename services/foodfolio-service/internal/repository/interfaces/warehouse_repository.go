package interfaces

import (
	"context"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// WarehouseRepository defines the interface for warehouse data access
type WarehouseRepository interface {
	// Create creates a new warehouse
	Create(ctx context.Context, warehouse *domain.Warehouse) error

	// GetByID retrieves a warehouse by ID
	GetByID(ctx context.Context, id string) (*domain.Warehouse, error)

	// GetBySlug retrieves a warehouse by slug
	GetBySlug(ctx context.Context, slug string) (*domain.Warehouse, error)

	// List retrieves warehouses with pagination and filtering
	List(ctx context.Context, offset, limit int, search string, includeDeleted bool) ([]*domain.Warehouse, int64, error)

	// Update updates an existing warehouse
	Update(ctx context.Context, warehouse *domain.Warehouse) error

	// Delete soft deletes a warehouse
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a warehouse
	HardDelete(ctx context.Context, id string) error
}
