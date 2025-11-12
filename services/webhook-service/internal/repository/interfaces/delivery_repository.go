package interfaces

import (
	"context"
	"time"

	"toxictoast/services/webhook-service/internal/domain"
)

type DeliveryRepository interface {
	// Create creates a new delivery
	Create(ctx context.Context, delivery *domain.Delivery) error

	// GetByID gets a delivery by ID
	GetByID(ctx context.Context, id string) (*domain.Delivery, error)

	// List lists deliveries with filters
	List(ctx context.Context, webhookID string, status domain.DeliveryStatus, limit, offset int) ([]*domain.Delivery, int64, error)

	// Update updates a delivery
	Update(ctx context.Context, delivery *domain.Delivery) error

	// GetPendingRetries gets deliveries that need retry
	GetPendingRetries(ctx context.Context, limit int) ([]*domain.Delivery, error)

	// CreateAttempt creates a delivery attempt
	CreateAttempt(ctx context.Context, attempt *domain.DeliveryAttempt) error

	// GetAttempts gets all attempts for a delivery
	GetAttempts(ctx context.Context, deliveryID string) ([]*domain.DeliveryAttempt, error)

	// Delete soft deletes a delivery
	Delete(ctx context.Context, id string) error

	// CleanupOldDeliveries removes deliveries older than the specified duration
	CleanupOldDeliveries(ctx context.Context, olderThan time.Duration) error
}
