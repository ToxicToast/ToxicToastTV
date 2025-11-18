package domain

import (
	"time"
)

// Category represents a blog category
// Pure domain model - NO infrastructure dependencies
type Category struct {
	ID          string
	Name        string
	Slug        string
	Description string
	ParentID    *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	// Relations (for domain logic, not persistence)
	Parent   *Category
	Children []Category
}
