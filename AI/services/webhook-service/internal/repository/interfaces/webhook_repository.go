package interfaces

import (
	"context"

	"toxictoast/services/webhook-service/internal/domain"
)

type WebhookRepository interface {
	// Create creates a new webhook
	Create(ctx context.Context, webhook *domain.Webhook) error

	// GetByID gets a webhook by ID
	GetByID(ctx context.Context, id string) (*domain.Webhook, error)

	// GetByURL gets a webhook by URL
	GetByURL(ctx context.Context, url string) (*domain.Webhook, error)

	// List lists all webhooks
	List(ctx context.Context, limit, offset int, activeOnly bool) ([]*domain.Webhook, int64, error)

	// Update updates a webhook
	Update(ctx context.Context, webhook *domain.Webhook) error

	// Delete soft deletes a webhook
	Delete(ctx context.Context, id string) error

	// GetActiveWebhooksForEvent gets all active webhooks that match an event type
	GetActiveWebhooksForEvent(ctx context.Context, eventType string) ([]*domain.Webhook, error)

	// UpdateStatistics updates webhook statistics
	UpdateStatistics(ctx context.Context, id string, success bool) error
}
