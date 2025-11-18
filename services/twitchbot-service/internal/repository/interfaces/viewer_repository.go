package interfaces

import (
	"context"
	"toxictoast/services/twitchbot-service/internal/domain"
)

// ViewerRepository defines the interface for viewer data access
type ViewerRepository interface {
	// Create creates a new viewer
	Create(ctx context.Context, viewer *domain.Viewer) error

	// GetByID retrieves a viewer by ID
	GetByID(ctx context.Context, id string) (*domain.Viewer, error)

	// GetByTwitchID retrieves a viewer by Twitch ID
	GetByTwitchID(ctx context.Context, twitchID string) (*domain.Viewer, error)

	// List retrieves viewers with pagination and filtering
	List(ctx context.Context, offset, limit int, orderBy string, includeDeleted bool) ([]*domain.Viewer, int64, error)

	// Update updates an existing viewer
	Update(ctx context.Context, viewer *domain.Viewer) error

	// Delete soft deletes a viewer
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a viewer
	HardDelete(ctx context.Context, id string) error
}
