package entity

import (
	"time"

	"gorm.io/gorm"
)

// UserEntity represents a user in the database
type UserEntity struct {
	ID           string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string `gorm:"type:varchar(255);unique;not null;index"`
	Username     string `gorm:"type:varchar(100);unique;not null;index"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	FirstName    string `gorm:"type:varchar(100)"`
	LastName     string `gorm:"type:varchar(100)"`
	AvatarURL    string `gorm:"type:varchar(500)"`
	Status       string `gorm:"type:varchar(50);not null;default:'active';index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLogin    *time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for UserEntity
func (UserEntity) TableName() string {
	return "azkaban_users"
}
