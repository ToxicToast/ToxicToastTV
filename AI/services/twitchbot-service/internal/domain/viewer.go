package domain

import (
	"time"
)

// Viewer represents a viewer/user who has interacted with the stream
// Pure domain model - NO infrastructure dependencies
type Viewer struct {
	ID                  string
	TwitchID            string
	Username            string
	DisplayName         string
	TotalMessages       int
	TotalStreamsWatched int
	FirstSeen           time.Time
	LastSeen            time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time

	// Relations (for domain logic, not persistence)
	Messages []Message
	Clips    []Clip
}
