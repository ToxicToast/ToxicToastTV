package entity

import (
	"time"

	"gorm.io/gorm"
)

// TagEntity is the database entity for tags
// Contains GORM tags and infrastructure concerns
type TagEntity struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Slug      string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName sets the table name with service prefix
func (TagEntity) TableName() string {
	return "blog_tags"
}
