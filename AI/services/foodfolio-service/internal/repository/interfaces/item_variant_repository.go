package interfaces

import (
	"context"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// ItemVariantRepository defines the interface for item variant data access
type ItemVariantRepository interface {
	// Create creates a new item variant
	Create(ctx context.Context, variant *domain.ItemVariant) error

	// GetByID retrieves an item variant by ID
	GetByID(ctx context.Context, id string) (*domain.ItemVariant, error)

	// GetByBarcode retrieves an item variant by barcode
	GetByBarcode(ctx context.Context, barcode string) (*domain.ItemVariant, error)

	// List retrieves item variants with pagination and filtering
	List(ctx context.Context, offset, limit int, itemID, sizeID *string, isNormallyFrozen *bool, includeDeleted bool) ([]*domain.ItemVariant, int64, error)

	// GetByItem retrieves all variants for an item
	GetByItem(ctx context.Context, itemID string) ([]*domain.ItemVariant, error)

	// GetLowStockVariants retrieves variants with stock below MinSKU
	GetLowStockVariants(ctx context.Context, offset, limit int) ([]*domain.ItemVariant, int64, error)

	// GetOverstockedVariants retrieves variants with stock above MaxSKU
	GetOverstockedVariants(ctx context.Context, offset, limit int) ([]*domain.ItemVariant, int64, error)

	// GetCurrentStock calculates current stock for a variant
	GetCurrentStock(ctx context.Context, variantID string) (int, error)

	// Update updates an existing item variant
	Update(ctx context.Context, variant *domain.ItemVariant) error

	// Delete soft deletes an item variant
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes an item variant
	HardDelete(ctx context.Context, id string) error
}
