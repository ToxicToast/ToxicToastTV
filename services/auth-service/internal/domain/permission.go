package domain

import "time"

// Permission represents a permission in the system
// Format: resource:action (e.g., "blog:create", "user:delete")
type Permission struct {
	ID          string
	Resource    string // e.g., "blog", "user", "comment"
	Action      string // e.g., "create", "read", "update", "delete"
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// String returns the permission as a string (resource:action)
func (p *Permission) String() string {
	return p.Resource + ":" + p.Action
}
