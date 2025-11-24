package cqrs

import (
	"context"
	"time"
)

// ReadModel represents a read-optimized projection of aggregate state
// Read models are eventually consistent and optimized for queries
type ReadModel interface {
	// GetID returns the read model ID
	GetID() string

	// GetLastUpdated returns when the read model was last updated
	GetLastUpdated() time.Time
}

// ReadModelRepository provides methods to query read models
type ReadModelRepository interface {
	// FindByID finds a read model by ID
	FindByID(ctx context.Context, id string) (ReadModel, error)

	// FindAll returns all read models with pagination
	FindAll(ctx context.Context, limit, offset int) ([]ReadModel, error)

	// Save saves a read model
	Save(ctx context.Context, model ReadModel) error

	// Delete deletes a read model
	Delete(ctx context.Context, id string) error
}

// BaseReadModel provides common functionality for read models
type BaseReadModel struct {
	ID          string    `json:"id" db:"id"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// NewBaseReadModel creates a new base read model
func NewBaseReadModel(id string) *BaseReadModel {
	return &BaseReadModel{
		ID:          id,
		LastUpdated: time.Now().UTC(),
	}
}

// GetID returns the ID
func (m *BaseReadModel) GetID() string {
	return m.ID
}

// GetLastUpdated returns the last updated timestamp
func (m *BaseReadModel) GetLastUpdated() time.Time {
	return m.LastUpdated
}

// UpdateTimestamp updates the last updated timestamp
func (m *BaseReadModel) UpdateTimestamp() {
	m.LastUpdated = time.Now().UTC()
}
