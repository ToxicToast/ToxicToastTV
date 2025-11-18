package entity

import (
	"time"

	"gorm.io/gorm"
)

type ReceiptEntity struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WarehouseID string         `gorm:"type:uuid;not null;index"`
	ScanDate    time.Time      `gorm:"not null"`
	TotalPrice  float64        `gorm:"type:decimal(10,2);not null"`
	ImagePath   *string        `gorm:"type:varchar(500)"`
	OCRText     *string        `gorm:"type:text"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Warehouse *WarehouseEntity     `gorm:"foreignKey:WarehouseID"`
	Items     []*ReceiptItemEntity `gorm:"foreignKey:ReceiptID"`
}

func (ReceiptEntity) TableName() string {
	return "foodfolio_receipts"
}
