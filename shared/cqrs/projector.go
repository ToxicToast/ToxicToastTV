package cqrs

import (
	"context"
	"fmt"
	"log"

	"github.com/toxictoast/toxictoastgo/shared/eventstore"
)

// Projector projects events into read models
type Projector interface {
	// ProjectEvent projects a single event
	ProjectEvent(ctx context.Context, event *eventstore.EventEnvelope) error

	// GetEventTypes returns the event types this projector handles
	GetEventTypes() []string

	// GetProjectorName returns the name of the projector
	GetProjectorName() string
}

// ProjectorManager manages multiple projectors
type ProjectorManager struct {
	projectors map[string][]Projector // event type -> projectors
	eventStore eventstore.EventStore
}

// NewProjectorManager creates a new projector manager
func NewProjectorManager(eventStore eventstore.EventStore) *ProjectorManager {
	return &ProjectorManager{
		projectors: make(map[string][]Projector),
		eventStore: eventStore,
	}
}

// RegisterProjector registers a projector for specific event types
func (m *ProjectorManager) RegisterProjector(projector Projector) {
	for _, eventType := range projector.GetEventTypes() {
		m.projectors[eventType] = append(m.projectors[eventType], projector)
		log.Printf("Registered projector '%s' for event type '%s'", projector.GetProjectorName(), eventType)
	}
}

// ProjectEvent projects a single event to all registered projectors
func (m *ProjectorManager) ProjectEvent(ctx context.Context, event *eventstore.EventEnvelope) error {
	projectors, ok := m.projectors[event.EventType]
	if !ok {
		// No projectors for this event type
		return nil
	}

	for _, projector := range projectors {
		if err := projector.ProjectEvent(ctx, event); err != nil {
			log.Printf("Error projecting event %s with projector %s: %v",
				event.EventType, projector.GetProjectorName(), err)
			return fmt.Errorf("failed to project event with %s: %w", projector.GetProjectorName(), err)
		}
	}

	return nil
}

// RebuildProjections rebuilds all projections from event store
// Useful for recovering read models or adding new projectors
func (m *ProjectorManager) RebuildProjections(ctx context.Context, aggregateType string) error {
	log.Printf("Rebuilding projections for aggregate type: %s", aggregateType)

	offset := 0
	limit := 100

	for {
		events, err := m.eventStore.GetAllEvents(ctx, aggregateType, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to get events: %w", err)
		}

		if len(events) == 0 {
			break
		}

		for _, event := range events {
			if err := m.ProjectEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to project event %s: %w", event.EventID, err)
			}
		}

		offset += len(events)

		if len(events) < limit {
			break
		}
	}

	log.Printf("Finished rebuilding projections for %s", aggregateType)
	return nil
}

// StartEventStreamProjection continuously projects events from the event stream
// This is useful for real-time read model updates
func (m *ProjectorManager) StartEventStreamProjection(ctx context.Context, pollInterval int64) {
	lastTimestamp := int64(0)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping event stream projection")
				return
			default:
				events, err := m.eventStore.GetEventStream(ctx, lastTimestamp, 100)
				if err != nil {
					log.Printf("Error fetching event stream: %v", err)
					continue
				}

				for _, event := range events {
					if err := m.ProjectEvent(ctx, event); err != nil {
						log.Printf("Error projecting event from stream: %v", err)
					}
					lastTimestamp = event.Timestamp.Unix()
				}

				// Sleep before next poll
				// time.Sleep(time.Duration(pollInterval) * time.Second)
			}
		}
	}()
}
