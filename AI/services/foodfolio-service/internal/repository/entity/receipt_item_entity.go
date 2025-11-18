package entity

import (
	"time"

	"gorm.io/gorm"
)

type ReceiptItemEntity struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ReceiptID     string         `gorm:"type:uuid;not null;index"`
	ItemVariantID *string        `gorm:"type:uuid;index"`
	ItemName      string         `gorm:"type:varchar(255);not null"`
	Quantity      int            `gorm:"not null;default:1"`
	UnitPrice     float64        `gorm:"type:decimal(10,2);not null"`
	TotalPrice    float64        `gorm:"type:decimal(10,2);not null"`
	ArticleNumber *string        `gorm:"type:varchar(255)"`
	IsMatched     bool           `gorm:"not null;default:false"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relations
	Receipt     *ReceiptEntity     `gorm:"foreignKey:ReceiptID"`
	ItemVariant *ItemVariantEntity `gorm:"foreignKey:ItemVariantID"`
}

func (ReceiptItemEntity) TableName() string {
	return "foodfolio_receipt_items"
}
