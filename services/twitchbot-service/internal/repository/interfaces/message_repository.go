package interfaces

import (
	"context"
	"toxictoast/services/twitchbot-service/internal/domain"
)

// MessageRepository defines the interface for message data access
type MessageRepository interface {
	// Create creates a new message
	Create(ctx context.Context, message *domain.Message) error

	// GetByID retrieves a message by ID
	GetByID(ctx context.Context, id string) (*domain.Message, error)

	// List retrieves messages with pagination and filtering
	List(ctx context.Context, offset, limit int, streamID, userID string, includeDeleted bool) ([]*domain.Message, int64, error)

	// Search searches messages by query
	Search(ctx context.Context, query string, streamID, userID string, offset, limit int) ([]*domain.Message, int64, error)

	// GetStats retrieves message statistics for a stream
	GetStats(ctx context.Context, streamID string) (totalMessages int64, uniqueUsers int64, mostActiveUser string, mostActiveUserCount int64, err error)

	// Delete soft deletes a message
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a message
	HardDelete(ctx context.Context, id string) error
}
