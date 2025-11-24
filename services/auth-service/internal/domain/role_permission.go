package domain

import "time"

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       string
	PermissionID string
	CreatedAt    time.Time
}
