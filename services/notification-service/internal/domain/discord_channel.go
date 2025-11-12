package domain

import (
	"strings"
	"time"
)

// DiscordChannel represents a Discord channel with webhook configuration
// Pure domain model - NO infrastructure dependencies
type DiscordChannel struct {
	ID          string
	Name        string
	WebhookURL  string
	EventTypes  string // Comma-separated (e.g., "blog.*,twitchbot.streams")
	Color       int    // Discord embed color (default: blue)
	Active      bool
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	// Statistics
	TotalNotifications   int
	SuccessNotifications int
	FailedNotifications  int
	LastNotificationAt   *time.Time
	LastSuccessAt        *time.Time
	LastFailureAt        *time.Time
}

// MatchesEvent checks if the channel should receive this event type
func (c *DiscordChannel) MatchesEvent(eventType string) bool {
	if c.EventTypes == "" || c.EventTypes == "*" {
		return true
	}

	types := strings.Split(c.EventTypes, ",")
	for _, t := range types {
		t = strings.TrimSpace(t)
		if matchesPattern(eventType, t) {
			return true
		}
	}

	return false
}

// matchesPattern checks if event type matches pattern (e.g., "blog.*" matches "blog.post.created")
func matchesPattern(eventType, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Wildcard matching
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(eventType) >= len(prefix) && eventType[:len(prefix)] == prefix
	}

	return eventType == pattern
}
