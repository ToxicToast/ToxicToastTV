package domain

import "time"

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID    string
	RoleID    string
	CreatedAt time.Time
}
