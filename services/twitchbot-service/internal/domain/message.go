package domain

import (
	"time"
)

// Message represents a chat message in a stream
// Pure domain model - NO infrastructure dependencies
type Message struct {
	ID            string
	StreamID      string
	UserID        string
	Username      string
	DisplayName   string
	Message       string
	IsModerator   bool
	IsSubscriber  bool
	IsVIP         bool
	IsBroadcaster bool
	SentAt        time.Time
	CreatedAt     time.Time
	DeletedAt     *time.Time

	// Relations (for domain logic, not persistence)
	Stream *Stream
	Viewer *Viewer
}
