package interfaces

import (
	"context"

	"toxictoast/services/notification-service/internal/domain"
)

type DiscordChannelRepository interface {
	// Create creates a new Discord channel
	Create(ctx context.Context, channel *domain.DiscordChannel) error

	// GetByID gets a Discord channel by ID
	GetByID(ctx context.Context, id string) (*domain.DiscordChannel, error)

	// GetByWebhookURL gets a Discord channel by webhook URL
	GetByWebhookURL(ctx context.Context, webhookURL string) (*domain.DiscordChannel, error)

	// List lists all Discord channels
	List(ctx context.Context, limit, offset int, activeOnly bool) ([]*domain.DiscordChannel, int64, error)

	// Update updates a Discord channel
	Update(ctx context.Context, channel *domain.DiscordChannel) error

	// Delete soft deletes a Discord channel
	Delete(ctx context.Context, id string) error

	// GetActiveChannelsForEvent gets all active channels that match an event type
	GetActiveChannelsForEvent(ctx context.Context, eventType string) ([]*domain.DiscordChannel, error)

	// UpdateStatistics updates channel statistics
	UpdateStatistics(ctx context.Context, id string, success bool) error
}
