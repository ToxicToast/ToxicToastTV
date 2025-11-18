package broker

import (
	"sync"

	"toxictoast/services/sse-service/internal/domain"
)

// EventHistory maintains a circular buffer of recent events
type EventHistory struct {
	events   []*domain.Event
	size     int
	position int
	mu       sync.RWMutex
}

// NewEventHistory creates a new event history buffer
func NewEventHistory(size int) *EventHistory {
	return &EventHistory{
		events:   make([]*domain.Event, 0, size),
		size:     size,
		position: 0,
	}
}

// Add adds an event to the history
func (h *EventHistory) Add(event *domain.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.events) < h.size {
		// Buffer not full yet, just append
		h.events = append(h.events, event)
	} else {
		// Buffer full, replace oldest
		h.events[h.position] = event
		h.position = (h.position + 1) % h.size
	}
}

// GetAll returns all events in chronological order
func (h *EventHistory) GetAll() []*domain.Event {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.events) == 0 {
		return []*domain.Event{}
	}

	// If buffer is not full yet, return in order
	if len(h.events) < h.size {
		result := make([]*domain.Event, len(h.events))
		copy(result, h.events)
		return result
	}

	// Buffer is full, reconstruct chronological order
	result := make([]*domain.Event, h.size)
	for i := 0; i < h.size; i++ {
		idx := (h.position + i) % h.size
		result[i] = h.events[idx]
	}
	return result
}

// GetSince returns events since a specific event ID
func (h *EventHistory) GetSince(lastEventID string) []*domain.Event {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if lastEventID == "" {
		return h.getAllUnlocked()
	}

	allEvents := h.getAllUnlocked()

	// Find the position of lastEventID
	startIdx := -1
	for i, event := range allEvents {
		if event.ID == lastEventID {
			startIdx = i + 1 // Start from next event
			break
		}
	}

	// If not found or no events after it, return empty
	if startIdx < 0 || startIdx >= len(allEvents) {
		return []*domain.Event{}
	}

	// Return events after lastEventID
	result := make([]*domain.Event, len(allEvents)-startIdx)
	copy(result, allEvents[startIdx:])
	return result
}

// getAllUnlocked returns all events in chronological order (caller must hold lock)
func (h *EventHistory) getAllUnlocked() []*domain.Event {
	if len(h.events) == 0 {
		return []*domain.Event{}
	}

	if len(h.events) < h.size {
		result := make([]*domain.Event, len(h.events))
		copy(result, h.events)
		return result
	}

	result := make([]*domain.Event, h.size)
	for i := 0; i < h.size; i++ {
		idx := (h.position + i) % h.size
		result[i] = h.events[idx]
	}
	return result
}

// Size returns the number of events in history
func (h *EventHistory) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.events)
}

// Clear clears all events from history
func (h *EventHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = make([]*domain.Event, 0, h.size)
	h.position = 0
}
