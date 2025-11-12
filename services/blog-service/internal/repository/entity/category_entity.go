package entity

import (
	"time"

	"gorm.io/gorm"
)

// CategoryEntity is the database entity for categories
// Contains GORM tags and infrastructure concerns
type CategoryEntity struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Slug        string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Description string         `gorm:"type:text"`
	ParentID    *string        `gorm:"type:uuid"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Parent   *CategoryEntity  `gorm:"foreignKey:ParentID"`
	Children []CategoryEntity `gorm:"foreignKey:ParentID"`
}

// TableName sets the table name with service prefix
func (CategoryEntity) TableName() string {
	return "blog_categories"
}
