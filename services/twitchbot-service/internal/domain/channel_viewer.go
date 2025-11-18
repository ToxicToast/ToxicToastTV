package domain

import (
	"time"
)

// ChannelViewer represents a viewer's presence in a specific channel
// This allows tracking which viewers are in which channels independently
// Pure domain model - NO infrastructure dependencies
type ChannelViewer struct {
	ID          string
	Channel     string
	TwitchID    string
	Username    string
	DisplayName string
	FirstSeen   time.Time
	LastSeen    time.Time
	IsModerator bool
	IsVIP       bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	// Relation to global Viewer (optional, for domain logic)
	Viewer *Viewer
}
