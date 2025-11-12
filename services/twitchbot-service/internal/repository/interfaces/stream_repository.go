package interfaces

import (
	"context"
	"toxictoast/services/twitchbot-service/internal/domain"
)

// StreamRepository defines the interface for stream data access
type StreamRepository interface {
	// Create creates a new stream
	Create(ctx context.Context, stream *domain.Stream) error

	// GetByID retrieves a stream by ID
	GetByID(ctx context.Context, id string) (*domain.Stream, error)

	// List retrieves streams with pagination and filtering
	List(ctx context.Context, offset, limit int, onlyActive bool, gameName string, includeDeleted bool) ([]*domain.Stream, int64, error)

	// GetActive retrieves the currently active stream
	GetActive(ctx context.Context) (*domain.Stream, error)

	// Update updates an existing stream
	Update(ctx context.Context, stream *domain.Stream) error

	// EndStream marks a stream as ended
	EndStream(ctx context.Context, id string) error

	// Delete soft deletes a stream
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a stream
	HardDelete(ctx context.Context, id string) error
}
