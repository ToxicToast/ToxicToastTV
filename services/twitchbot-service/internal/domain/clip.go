package domain

import (
	"time"
)

// Clip represents a Twitch clip created during a stream
// Pure domain model - NO infrastructure dependencies
type Clip struct {
	ID              string
	StreamID        string
	TwitchClipID    string
	Title           string
	URL             string
	EmbedURL        string
	ThumbnailURL    string
	CreatorName     string
	CreatorID       string
	ViewCount       int
	DurationSeconds int
	CreatedAtTwitch time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time

	// Relations (for domain logic, not persistence)
	Stream  *Stream
	Creator *Viewer
}
