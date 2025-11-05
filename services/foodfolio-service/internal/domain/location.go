package domain

import (
	"time"

	"gorm.io/gorm"
)

// Location represents a storage location at home (e.g., "Fridge", "Pantry", "Freezer")
type Location struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Slug      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	ParentID  *string        `gorm:"type:uuid" json:"parent_id"`                         // e.g., "Fridge" -> "Top Shelf"
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Parent      *Location    `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children    []Location   `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	ItemDetails []ItemDetail `gorm:"foreignKey:LocationID" json:"item_details,omitempty"`
}

func (Location) TableName() string {
	return "locations"
}
