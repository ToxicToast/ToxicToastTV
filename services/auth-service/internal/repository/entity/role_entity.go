package entity

import (
	"time"

	"gorm.io/gorm"
)

// RoleEntity represents a role in the database
type RoleEntity struct {
	ID          string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string `gorm:"type:varchar(100);unique;not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for RoleEntity
func (RoleEntity) TableName() string {
	return "azkaban_roles"
}
