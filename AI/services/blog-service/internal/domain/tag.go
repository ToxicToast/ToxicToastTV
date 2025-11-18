package domain

import (
	"time"
)

// Tag represents a blog tag
// Pure domain model - NO infrastructure dependencies
type Tag struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
