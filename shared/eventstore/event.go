package eventstore

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventEnvelope wraps domain events with metadata for event sourcing
type EventEnvelope struct {
	// Event metadata
	EventID       string    `json:"event_id" db:"event_id"`
	EventType     string    `json:"event_type" db:"event_type"`
	AggregateID   string    `json:"aggregate_id" db:"aggregate_id"`
	AggregateType string    `json:"aggregate_type" db:"aggregate_type"`
	Version       int64     `json:"version" db:"version"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`

	// Event payload
	Data json.RawMessage `json:"data" db:"data"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// NewEventEnvelope creates a new event envelope
func NewEventEnvelope(aggregateType, aggregateID, eventType string, version int64, data interface{}) (*EventEnvelope, error) {
	eventData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &EventEnvelope{
		EventID:       uuid.New().String(),
		EventType:     eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Version:       version,
		Timestamp:     time.Now().UTC(),
		Data:          eventData,
		Metadata:      make(map[string]interface{}),
	}, nil
}

// UnmarshalData unmarshals the event data into the provided struct
func (e *EventEnvelope) UnmarshalData(v interface{}) error {
	return json.Unmarshal(e.Data, v)
}

// WithMetadata adds metadata to the event
func (e *EventEnvelope) WithMetadata(key string, value interface{}) *EventEnvelope {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// GetMetadata retrieves metadata from the event
func (e *EventEnvelope) GetMetadata(key string) (interface{}, bool) {
	if e.Metadata == nil {
		return nil, false
	}
	val, ok := e.Metadata[key]
	return val, ok
}

// Event types for common domain events
const (
	// User aggregate events
	EventTypeUserCreated  = "user.created"
	EventTypeUserUpdated  = "user.updated"
	EventTypeUserDeleted  = "user.deleted"
	EventTypeUserActivated = "user.activated"
	EventTypeUserDeactivated = "user.deactivated"

	// Auth aggregate events
	EventTypeAuthRegistered    = "auth.registered"
	EventTypeAuthLogin         = "auth.login"
	EventTypeAuthLogout        = "auth.logout"
	EventTypeAuthTokenRefreshed = "auth.token.refreshed"
	EventTypeAuthTokenRevoked  = "auth.token.revoked"

	// Blog aggregate events
	EventTypeCategoryCreated = "blog.category.created"
	EventTypeCategoryUpdated = "blog.category.updated"
	EventTypeCategoryDeleted = "blog.category.deleted"

	EventTypePostCreated   = "blog.post.created"
	EventTypePostUpdated   = "blog.post.updated"
	EventTypePostPublished = "blog.post.published"
	EventTypePostDeleted   = "blog.post.deleted"

	// Link aggregate events
	EventTypeLinkCreated     = "link.created"
	EventTypeLinkUpdated     = "link.updated"
	EventTypeLinkDeleted     = "link.deleted"
	EventTypeLinkActivated   = "link.activated"
	EventTypeLinkDeactivated = "link.deactivated"
	EventTypeLinkClicked     = "link.clicked"

	// Foodfolio aggregate events
	EventTypeFoodfolioItemCreated   = "foodfolio.item.created"
	EventTypeFoodfolioItemUpdated   = "foodfolio.item.updated"
	EventTypeFoodfolioItemDeleted   = "foodfolio.item.deleted"
	EventTypeFoodfolioItemConsumed  = "foodfolio.item.consumed"
	EventTypeFoodfolioItemExpired   = "foodfolio.item.expired"
	EventTypeFoodfolioItemExpiringSoon = "foodfolio.item.expiring_soon"
)

// Aggregate types
const (
	AggregateTypeUser     = "user"
	AggregateTypeAuth     = "auth"
	AggregateTypeCategory = "category"
	AggregateTypePost     = "post"
	AggregateTypeLink     = "link"
	AggregateTypeFoodfolioItem = "foodfolio_item"
)
