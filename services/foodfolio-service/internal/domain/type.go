package domain

import (
	"time"

	"gorm.io/gorm"
)

// Type represents packaging type (e.g., "Can", "PET Bottle", "Box", "Bag", "Glass")
type Type struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Slug      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Items []Item `gorm:"foreignKey:TypeID" json:"items,omitempty"`
}

func (Type) TableName() string {
	return "types"
}
