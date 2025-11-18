package entity

import (
	"time"

	"gorm.io/gorm"
)

type ItemVariantEntity struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ItemID           string         `gorm:"type:uuid;not null;index"`
	SizeID           string         `gorm:"type:uuid;not null;index"`
	VariantName      string         `gorm:"type:varchar(255);not null"`
	Slug             string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Barcode          *string        `gorm:"type:varchar(255);uniqueIndex"`
	MinSKU           int            `gorm:"not null;default:0"`
	MaxSKU           int            `gorm:"not null;default:0"`
	IsNormallyFrozen bool           `gorm:"not null;default:false"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	// Relations
	Item        *ItemEntity        `gorm:"foreignKey:ItemID"`
	Size        *SizeEntity        `gorm:"foreignKey:SizeID"`
	ItemDetails []ItemDetailEntity `gorm:"foreignKey:ItemVariantID"`
}

func (ItemVariantEntity) TableName() string {
	return "foodfolio_item_variants"
}
