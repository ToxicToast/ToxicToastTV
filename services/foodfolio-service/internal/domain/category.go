package domain

import (
	"time"

	"gorm.io/gorm"
)

// Category represents a product category (e.g., "Beverages", "Dairy", "Snacks")
type Category struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Slug      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	ParentID  *string        `gorm:"type:uuid" json:"parent_id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Items    []Item     `gorm:"foreignKey:CategoryID" json:"items,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}
