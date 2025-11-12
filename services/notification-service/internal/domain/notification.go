package domain

import (
	"time"
)

// NotificationStatus represents the status of a notification delivery
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSuccess NotificationStatus = "success"
	NotificationStatusFailed  NotificationStatus = "failed"
)

// Notification represents a Discord notification delivery
// Pure domain model - NO infrastructure dependencies
type Notification struct {
	ID               string
	ChannelID        string
	EventID          string
	EventType        string
	EventPayload     string // JSON string
	DiscordMessageID string
	Status           NotificationStatus
	AttemptCount     int
	LastError        string
	SentAt           *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time

	// Relations (for domain logic, not persistence)
	Channel  *DiscordChannel
	Attempts []NotificationAttempt
}

// NotificationAttempt represents a single delivery attempt
// Pure domain model - NO infrastructure dependencies
type NotificationAttempt struct {
	ID               string
	NotificationID   string
	AttemptNumber    int
	ResponseStatus   int
	ResponseBody     string
	DiscordMessageID string
	Success          bool
	Error            string
	DurationMs       int
	CreatedAt        time.Time
	DeletedAt        *time.Time

	// Relations (for domain logic, not persistence)
	Notification *Notification
}
