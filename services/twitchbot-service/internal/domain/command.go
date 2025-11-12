package domain

import (
	"time"
)

// Command represents a chat command that can be executed by viewers
// Pure domain model - NO infrastructure dependencies
type Command struct {
	ID              string
	Name            string
	Description     string
	Response        string
	IsActive        bool
	ModeratorOnly   bool
	SubscriberOnly  bool
	CooldownSeconds int
	UsageCount      int
	LastUsed        *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
