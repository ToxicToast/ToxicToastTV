package eventstore

import (
	"context"
	"fmt"
)

// Aggregate is the base interface for all event-sourced aggregates
type Aggregate interface {
	// GetID returns the aggregate ID
	GetID() string

	// GetType returns the aggregate type
	GetType() string

	// GetVersion returns the current version of the aggregate
	GetVersion() int64

	// GetUncommittedEvents returns events that haven't been persisted yet
	GetUncommittedEvents() []*EventEnvelope

	// MarkEventsAsCommitted clears the uncommitted events
	MarkEventsAsCommitted()

	// LoadFromHistory reconstructs the aggregate from events
	LoadFromHistory(events []*EventEnvelope) error
}

// BaseAggregate provides common functionality for event-sourced aggregates
type BaseAggregate struct {
	ID                string
	AggregateType     string
	Version           int64
	uncommittedEvents []*EventEnvelope
}

// NewBaseAggregate creates a new base aggregate
func NewBaseAggregate(id, aggregateType string) *BaseAggregate {
	return &BaseAggregate{
		ID:                id,
		AggregateType:     aggregateType,
		Version:           -1,
		uncommittedEvents: []*EventEnvelope{},
	}
}

// GetID returns the aggregate ID
func (a *BaseAggregate) GetID() string {
	return a.ID
}

// GetType returns the aggregate type
func (a *BaseAggregate) GetType() string {
	return a.AggregateType
}

// GetVersion returns the current version
func (a *BaseAggregate) GetVersion() int64 {
	return a.Version
}

// GetUncommittedEvents returns uncommitted events
func (a *BaseAggregate) GetUncommittedEvents() []*EventEnvelope {
	return a.uncommittedEvents
}

// MarkEventsAsCommitted clears uncommitted events
func (a *BaseAggregate) MarkEventsAsCommitted() {
	a.uncommittedEvents = []*EventEnvelope{}
}

// RaiseEvent adds an event to the uncommitted events and applies it
func (a *BaseAggregate) RaiseEvent(eventType string, data interface{}, applyFunc func(*EventEnvelope) error) error {
	a.Version++

	envelope, err := NewEventEnvelope(a.AggregateType, a.ID, eventType, a.Version, data)
	if err != nil {
		a.Version--
		return fmt.Errorf("failed to create event envelope: %w", err)
	}

	// Apply the event to update aggregate state
	if err := applyFunc(envelope); err != nil {
		a.Version--
		return fmt.Errorf("failed to apply event: %w", err)
	}

	a.uncommittedEvents = append(a.uncommittedEvents, envelope)
	return nil
}

// LoadFromHistory reconstructs aggregate state from events
func (a *BaseAggregate) LoadFromHistory(events []*EventEnvelope, applyFunc func(*EventEnvelope) error) error {
	for _, event := range events {
		if err := applyFunc(event); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", event.EventType, err)
		}
		a.Version = event.Version
	}
	return nil
}

// AggregateRepository provides methods to load and save aggregates
type AggregateRepository struct {
	eventStore EventStore
}

// NewAggregateRepository creates a new aggregate repository
func NewAggregateRepository(eventStore EventStore) *AggregateRepository {
	return &AggregateRepository{
		eventStore: eventStore,
	}
}

// Load loads an aggregate from the event store
func (r *AggregateRepository) Load(ctx context.Context, aggregate Aggregate) error {
	events, err := r.eventStore.GetEvents(ctx, aggregate.GetType(), aggregate.GetID())
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return ErrAggregateNotFound
	}

	if err := aggregate.LoadFromHistory(events); err != nil {
		return fmt.Errorf("failed to load from history: %w", err)
	}

	return nil
}

// Save saves an aggregate to the event store
func (r *AggregateRepository) Save(ctx context.Context, aggregate Aggregate) error {
	uncommittedEvents := aggregate.GetUncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil // No events to save
	}

	expectedVersion := aggregate.GetVersion() - int64(len(uncommittedEvents))

	if err := r.eventStore.SaveEvents(ctx, aggregate.GetID(), expectedVersion, uncommittedEvents); err != nil {
		return fmt.Errorf("failed to save events: %w", err)
	}

	aggregate.MarkEventsAsCommitted()
	return nil
}

// Exists checks if an aggregate exists in the event store
func (r *AggregateRepository) Exists(ctx context.Context, aggregateType, aggregateID string) (bool, error) {
	version, err := r.eventStore.GetAggregateVersion(ctx, aggregateType, aggregateID)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return version >= 0, nil
}
