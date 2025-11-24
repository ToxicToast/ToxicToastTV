package domain

import "time"

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// User represents a user in the system
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	FirstName    string
	LastName     string
	AvatarURL    string
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLogin    *time.Time
	DeletedAt    *time.Time
}

// FullName returns the user's full name
func (u *User) FullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	if u.FirstName != "" {
		return u.FirstName
	}
	if u.LastName != "" {
		return u.LastName
	}
	return u.Username
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}
