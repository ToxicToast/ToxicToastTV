package entity

import (
	"time"

	"gorm.io/gorm"
)

type WarehouseEntity struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Slug      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	ItemDetails []ItemDetailEntity `gorm:"foreignKey:WarehouseID"`
	Receipts    []ReceiptEntity    `gorm:"foreignKey:WarehouseID"`
}

func (WarehouseEntity) TableName() string {
	return "foodfolio_warehouses"
}
