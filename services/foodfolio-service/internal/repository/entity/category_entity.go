package entity

import (
	"time"

	"gorm.io/gorm"
)

type CategoryEntity struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Slug      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	ParentID  *string        `gorm:"type:uuid"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Parent   *CategoryEntity  `gorm:"foreignKey:ParentID"`
	Children []CategoryEntity `gorm:"foreignKey:ParentID"`
	Items    []ItemEntity     `gorm:"foreignKey:CategoryID"`
}

func (CategoryEntity) TableName() string {
	return "foodfolio_categories"
}
