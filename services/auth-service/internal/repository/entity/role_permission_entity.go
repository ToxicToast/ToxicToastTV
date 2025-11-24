package entity

import "time"

// RolePermissionEntity represents the role-permission relationship in the database
type RolePermissionEntity struct {
	RoleID       string    `gorm:"type:uuid;primaryKey;index:idx_role_permission"`
	PermissionID string    `gorm:"type:uuid;primaryKey;index:idx_role_permission"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

// TableName specifies the table name for RolePermissionEntity
func (RolePermissionEntity) TableName() string {
	return "azkaban_role_permissions"
}
