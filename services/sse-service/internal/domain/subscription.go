package domain

import (
	"time"
)

// Client represents a connected SSE client
type Client struct {
	ID             string
	Channel        chan *Event
	Filter         SubscriptionFilter
	ConnectedAt    time.Time
	LastEventAt    time.Time
	EventsReceived int64
	UserAgent      string
	RemoteAddr     string
}

// SubscriptionFilter defines what events a client wants to receive
type SubscriptionFilter struct {
	EventTypes []EventType `json:"event_types"` // e.g., ["blog.post.created", "twitchbot.*"]
	Sources    []string    `json:"sources"`     // e.g., ["blog-service", "twitchbot-service"]
}

// NewClient creates a new SSE client
func NewClient(id string, filter SubscriptionFilter, userAgent, remoteAddr string, bufferSize int) *Client {
	return &Client{
		ID:             id,
		Channel:        make(chan *Event, bufferSize),
		Filter:         filter,
		ConnectedAt:    time.Now(),
		LastEventAt:    time.Now(),
		EventsReceived: 0,
		UserAgent:      userAgent,
		RemoteAddr:     remoteAddr,
	}
}

// SendEvent sends an event to the client (non-blocking)
func (c *Client) SendEvent(event *Event) bool {
	// Check if event matches filter
	if !event.MatchesFilter(c.Filter) {
		return false
	}

	// Try to send, but don't block
	select {
	case c.Channel <- event:
		c.LastEventAt = time.Now()
		c.EventsReceived++
		return true
	default:
		// Channel full, skip this event
		return false
	}
}

// Close closes the client's channel
func (c *Client) Close() {
	close(c.Channel)
}

// Stats returns client statistics
type ClientStats struct {
	ID             string    `json:"id"`
	ConnectedAt    time.Time `json:"connected_at"`
	LastEventAt    time.Time `json:"last_event_at"`
	EventsReceived int64     `json:"events_received"`
	Filter         SubscriptionFilter `json:"filter"`
	UserAgent      string    `json:"user_agent"`
	RemoteAddr     string    `json:"remote_addr"`
}

// GetStats returns statistics about this client
func (c *Client) GetStats() ClientStats {
	return ClientStats{
		ID:             c.ID,
		ConnectedAt:    c.ConnectedAt,
		LastEventAt:    c.LastEventAt,
		EventsReceived: c.EventsReceived,
		Filter:         c.Filter,
		UserAgent:      c.UserAgent,
		RemoteAddr:     c.RemoteAddr,
	}
}
