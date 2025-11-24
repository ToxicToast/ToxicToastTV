package entity

import (
	"time"

	"gorm.io/gorm"
)

// PermissionEntity represents a permission in the database
type PermissionEntity struct {
	ID          string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Resource    string `gorm:"type:varchar(100);not null;index:idx_resource_action"`
	Action      string `gorm:"type:varchar(100);not null;index:idx_resource_action"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for PermissionEntity
func (PermissionEntity) TableName() string {
	return "azkaban_permissions"
}
