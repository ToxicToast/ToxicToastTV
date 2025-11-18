package interfaces

import (
	"context"

	"toxictoast/services/twitchbot-service/internal/domain"
)

type ChannelViewerRepository interface {
	// Create or update a channel viewer
	Upsert(ctx context.Context, channelViewer *domain.ChannelViewer) error

	// Get channel viewer by channel and Twitch ID
	GetByChannelAndTwitchID(ctx context.Context, channel, twitchID string) (*domain.ChannelViewer, error)

	// List all viewers in a channel
	ListByChannel(ctx context.Context, channel string, limit, offset int) ([]*domain.ChannelViewer, int64, error)

	// Update last seen time
	UpdateLastSeen(ctx context.Context, channel, twitchID string) error

	// Delete channel viewer (when they leave)
	Delete(ctx context.Context, channel, twitchID string) error

	// Count viewers in a channel
	CountByChannel(ctx context.Context, channel string) (int64, error)
}
