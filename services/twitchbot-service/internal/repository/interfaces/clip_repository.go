package interfaces

import (
	"context"
	"toxictoast/services/twitchbot-service/internal/domain"
)

// ClipRepository defines the interface for clip data access
type ClipRepository interface {
	// Create creates a new clip
	Create(ctx context.Context, clip *domain.Clip) error

	// GetByID retrieves a clip by ID
	GetByID(ctx context.Context, id string) (*domain.Clip, error)

	// GetByTwitchClipID retrieves a clip by Twitch clip ID
	GetByTwitchClipID(ctx context.Context, twitchClipID string) (*domain.Clip, error)

	// List retrieves clips with pagination and filtering
	List(ctx context.Context, offset, limit int, streamID, orderBy string, includeDeleted bool) ([]*domain.Clip, int64, error)

	// Update updates an existing clip
	Update(ctx context.Context, clip *domain.Clip) error

	// Delete soft deletes a clip
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a clip
	HardDelete(ctx context.Context, id string) error
}
