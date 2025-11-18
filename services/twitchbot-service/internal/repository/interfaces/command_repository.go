package interfaces

import (
	"context"
	"toxictoast/services/twitchbot-service/internal/domain"
)

// CommandRepository defines the interface for command data access
type CommandRepository interface {
	// Create creates a new command
	Create(ctx context.Context, command *domain.Command) error

	// GetByID retrieves a command by ID
	GetByID(ctx context.Context, id string) (*domain.Command, error)

	// GetByName retrieves a command by name
	GetByName(ctx context.Context, name string) (*domain.Command, error)

	// List retrieves commands with pagination and filtering
	List(ctx context.Context, offset, limit int, onlyActive bool, includeDeleted bool) ([]*domain.Command, int64, error)

	// Update updates an existing command
	Update(ctx context.Context, command *domain.Command) error

	// IncrementUsage increments the usage count and updates last used timestamp
	IncrementUsage(ctx context.Context, id string) error

	// Delete soft deletes a command
	Delete(ctx context.Context, id string) error

	// HardDelete permanently deletes a command
	HardDelete(ctx context.Context, id string) error
}
