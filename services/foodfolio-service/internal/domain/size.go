package domain

import (
	"time"

	"gorm.io/gorm"
)

// Size represents product size (e.g., "500ml", "1L", "250g", "1kg")
type Size struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Value     float64        `gorm:"not null" json:"value"`                              // Numeric value (e.g., 500, 1, 250)
	Unit      string         `gorm:"type:varchar(50);not null" json:"unit"`              // Unit (e.g., "ml", "L", "g", "kg")
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	ItemVariants []ItemVariant `gorm:"foreignKey:SizeID" json:"item_variants,omitempty"`
}

func (Size) TableName() string {
	return "sizes"
}
