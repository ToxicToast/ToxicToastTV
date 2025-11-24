package eventstore

import (
	"context"
	"errors"
)

var (
	// ErrConcurrencyConflict is returned when an optimistic locking conflict occurs
	ErrConcurrencyConflict = errors.New("concurrency conflict: aggregate version mismatch")

	// ErrAggregateNotFound is returned when an aggregate is not found
	ErrAggregateNotFound = errors.New("aggregate not found")

	// ErrInvalidVersion is returned when an invalid version is provided
	ErrInvalidVersion = errors.New("invalid version number")
)

// EventStore defines the interface for event store implementations
type EventStore interface {
	// SaveEvents saves one or more events for an aggregate with optimistic locking
	// expectedVersion is the expected current version of the aggregate
	// Returns ErrConcurrencyConflict if the version doesn't match
	SaveEvents(ctx context.Context, aggregateID string, expectedVersion int64, events []*EventEnvelope) error

	// GetEvents retrieves all events for an aggregate in order
	GetEvents(ctx context.Context, aggregateType, aggregateID string) ([]*EventEnvelope, error)

	// GetEventsSince retrieves events for an aggregate since a specific version
	GetEventsSince(ctx context.Context, aggregateType, aggregateID string, sinceVersion int64) ([]*EventEnvelope, error)

	// GetEventsByType retrieves events of a specific type for an aggregate
	GetEventsByType(ctx context.Context, aggregateType, aggregateID, eventType string) ([]*EventEnvelope, error)

	// GetAllEvents retrieves all events of a specific aggregate type
	// Useful for rebuilding read models
	GetAllEvents(ctx context.Context, aggregateType string, limit, offset int) ([]*EventEnvelope, error)

	// GetEventStream retrieves events across all aggregates ordered by timestamp
	// Useful for projections and event replay
	GetEventStream(ctx context.Context, sinceTimestamp int64, limit int) ([]*EventEnvelope, error)

	// GetAggregateVersion gets the current version of an aggregate
	GetAggregateVersion(ctx context.Context, aggregateType, aggregateID string) (int64, error)

	// Close closes the event store connection
	Close() error
}

// Snapshot represents a snapshot of an aggregate state at a specific version
type Snapshot struct {
	AggregateID   string `json:"aggregate_id" db:"aggregate_id"`
	AggregateType string `json:"aggregate_type" db:"aggregate_type"`
	Version       int64  `json:"version" db:"version"`
	State         []byte `json:"state" db:"state"`
	CreatedAt     int64  `json:"created_at" db:"created_at"`
}

// SnapshotStore defines the interface for snapshot storage
// Snapshots are optional optimizations to avoid replaying all events
type SnapshotStore interface {
	// SaveSnapshot saves a snapshot of an aggregate
	SaveSnapshot(ctx context.Context, snapshot *Snapshot) error

	// GetSnapshot retrieves the latest snapshot for an aggregate
	GetSnapshot(ctx context.Context, aggregateType, aggregateID string) (*Snapshot, error)

	// DeleteSnapshot deletes snapshots for an aggregate
	DeleteSnapshot(ctx context.Context, aggregateType, aggregateID string) error
}
