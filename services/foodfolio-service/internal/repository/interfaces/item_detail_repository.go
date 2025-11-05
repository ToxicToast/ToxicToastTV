package interfaces

import (
	"context"
	"time"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// ItemDetailRepository defines the interface for item detail data access
type ItemDetailRepository interface {
	// Create creates a new item detail
	Create(ctx context.Context, detail *domain.ItemDetail) error

	// BatchCreate creates multiple item details (for bulk purchases)
	BatchCreate(ctx context.Context, details []*domain.ItemDetail) error

	// GetByID retrieves an item detail by ID
	GetByID(ctx context.Context, id string) (*domain.ItemDetail, error)

	// List retrieves item details with pagination and filtering
	List(ctx context.Context, offset, limit int, variantID, warehouseID, locationID *string, isOpened, hasDeposit, isFrozen *bool, includeDeleted bool) ([]*domain.ItemDetail, int64, error)

	// GetByVariant retrieves all details for an item variant
	GetByVariant(ctx context.Context, variantID string) ([]*domain.ItemDetail, error)

	// GetByLocation retrieves all details at a location
	GetByLocation(ctx context.Context, locationID string, includeChildren bool) ([]*domain.ItemDetail, error)

	// GetExpiringItems retrieves items expiring within N days
	GetExpiringItems(ctx context.Context, days int, offset, limit int) ([]*domain.ItemDetail, int64, error)

	// GetExpiredItems retrieves all expired items
	GetExpiredItems(ctx context.Context, offset, limit int) ([]*domain.ItemDetail, int64, error)

	// GetItemsWithDeposit retrieves all items with deposit
	GetItemsWithDeposit(ctx context.Context, offset, limit int) ([]*domain.ItemDetail, int64, error)

	// OpenItem marks an item as opened
	OpenItem(ctx context.Context, id string, openedDate time.Time) error

	// MoveItems moves multiple items to a new location
	MoveItems(ctx context.Context, itemIDs []string, newLocationID string) error

	// Update updates an existing item detail
	Update(ctx context.Context, detail *domain.ItemDetail) error

	// Delete soft deletes an item detail
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes an item detail
	HardDelete(ctx context.Context, id string) error

	// CountByVariant counts active details for a variant
	CountByVariant(ctx context.Context, variantID string) (int, error)
}
