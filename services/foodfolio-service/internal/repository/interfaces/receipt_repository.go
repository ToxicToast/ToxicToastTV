package interfaces

import (
	"context"
	"time"
	"toxictoast/services/foodfolio-service/internal/domain"
)

// ReceiptRepository defines the interface for receipt data access
type ReceiptRepository interface {
	// Create creates a new receipt
	Create(ctx context.Context, receipt *domain.Receipt) error

	// GetByID retrieves a receipt by ID with items
	GetByID(ctx context.Context, id string) (*domain.Receipt, error)

	// List retrieves receipts with pagination and filtering
	List(ctx context.Context, offset, limit int, warehouseID *string, startDate, endDate *time.Time, includeDeleted bool) ([]*domain.Receipt, int64, error)

	// Update updates an existing receipt
	Update(ctx context.Context, receipt *domain.Receipt) error

	// Delete soft deletes a receipt
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a receipt
	HardDelete(ctx context.Context, id string) error

	// AddItem adds an item to a receipt
	AddItem(ctx context.Context, item *domain.ReceiptItem) error

	// UpdateItem updates a receipt item
	UpdateItem(ctx context.Context, item *domain.ReceiptItem) error

	// MatchItem matches a receipt item to an item variant
	MatchItem(ctx context.Context, receiptItemID, itemVariantID string) error

	// GetItem retrieves a specific receipt item
	GetItem(ctx context.Context, itemID string) (*domain.ReceiptItem, error)

	// GetItems retrieves all items for a receipt
	GetItems(ctx context.Context, receiptID string) ([]*domain.ReceiptItem, error)

	// GetUnmatchedItems retrieves all unmatched items for a receipt
	GetUnmatchedItems(ctx context.Context, receiptID string) ([]*domain.ReceiptItem, error)

	// GetTotalAmount calculates total amount for filtered receipts
	GetTotalAmount(ctx context.Context, warehouseID *string, startDate, endDate *time.Time) (float64, error)

	// GetStatistics retrieves receipt statistics
	GetStatistics(ctx context.Context, warehouseID *string, startDate, endDate *time.Time) (map[string]interface{}, error)
}
