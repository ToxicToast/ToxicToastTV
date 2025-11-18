package domain

import (
	"time"
)

// Stream represents a Twitch stream session
// Pure domain model - NO infrastructure dependencies
type Stream struct {
	ID             string
	Title          string
	GameName       string
	GameID         string
	StartedAt      time.Time
	EndedAt        *time.Time
	PeakViewers    int
	AverageViewers int
	TotalMessages  int
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time

	// Relations (for domain logic, not persistence)
	Messages []Message
	Clips    []Clip
}
