package interfaces

import (
	"context"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// ShoppinglistRepository defines the interface for shoppinglist data access
type ShoppinglistRepository interface {
	// Create creates a new shoppinglist
	Create(ctx context.Context, shoppinglist *domain.Shoppinglist) error

	// GetByID retrieves a shoppinglist by ID with items
	GetByID(ctx context.Context, id string) (*domain.Shoppinglist, error)

	// List retrieves shoppinglists with pagination
	List(ctx context.Context, offset, limit int, includeDeleted bool) ([]*domain.Shoppinglist, int64, error)

	// Update updates an existing shoppinglist
	Update(ctx context.Context, shoppinglist *domain.Shoppinglist) error

	// Delete soft deletes a shoppinglist
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a shoppinglist
	HardDelete(ctx context.Context, id string) error

	// AddItem adds an item to a shoppinglist
	AddItem(ctx context.Context, item *domain.ShoppinglistItem) error

	// RemoveItem removes an item from a shoppinglist
	RemoveItem(ctx context.Context, shoppinglistID, itemID string) error

	// UpdateItem updates a shoppinglist item
	UpdateItem(ctx context.Context, item *domain.ShoppinglistItem) error

	// MarkItemPurchased marks an item as purchased
	MarkItemPurchased(ctx context.Context, itemID string) error

	// MarkAllItemsPurchased marks all items in a list as purchased
	MarkAllItemsPurchased(ctx context.Context, shoppinglistID string) (int, error)

	// ClearPurchasedItems removes all purchased items from a list
	ClearPurchasedItems(ctx context.Context, shoppinglistID string) (int, error)

	// GetItem retrieves a specific shoppinglist item
	GetItem(ctx context.Context, itemID string) (*domain.ShoppinglistItem, error)

	// GetItems retrieves all items for a shoppinglist
	GetItems(ctx context.Context, shoppinglistID string) ([]*domain.ShoppinglistItem, error)
}
