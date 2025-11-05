package domain

import (
	"time"

	"gorm.io/gorm"
)

// Warehouse represents a store where items are purchased (e.g., "Rewe", "Lidl", "Edeka")
type Warehouse struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Slug      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	ItemDetails []ItemDetail `gorm:"foreignKey:WarehouseID" json:"item_details,omitempty"`
	Receipts    []Receipt    `gorm:"foreignKey:WarehouseID" json:"receipts,omitempty"`
}

func (Warehouse) TableName() string {
	return "warehouses"
}
