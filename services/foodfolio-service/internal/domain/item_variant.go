package domain

import (
	"time"

	"gorm.io/gorm"
)

// ItemVariant represents a specific variant (flavor + size combination)
// Example: "Monster Energy Original 500ml" vs "Monster Energy Ultra White 500ml"
type ItemVariant struct {
	ID                 string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ItemID             string         `gorm:"type:uuid;not null;index" json:"item_id"`
	SizeID             string         `gorm:"type:uuid;not null;index" json:"size_id"`
	VariantName        string         `gorm:"type:varchar(255);not null" json:"variant_name"`        // e.g., "Original", "Ultra White", "Zero"
	Slug               string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	Barcode            *string        `gorm:"type:varchar(255);uniqueIndex" json:"barcode"`          // EAN/UPC
	MinSKU             int            `gorm:"not null;default:0" json:"min_sku"`                     // Alert when below
	MaxSKU             int            `gorm:"not null;default:0" json:"max_sku"`                     // Alert when above
	IsNormallyFrozen   bool           `gorm:"not null;default:false" json:"is_normally_frozen"`      // Should this be frozen
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Item        *Item        `gorm:"foreignKey:ItemID" json:"item,omitempty"`
	Size        *Size        `gorm:"foreignKey:SizeID" json:"size,omitempty"`
	ItemDetails []ItemDetail `gorm:"foreignKey:ItemVariantID" json:"item_details,omitempty"`
}

func (ItemVariant) TableName() string {
	return "item_variants"
}

// CurrentStock returns the count of active (not opened/consumed) ItemDetails
func (iv *ItemVariant) CurrentStock() int {
	count := 0
	for _, detail := range iv.ItemDetails {
		if !detail.IsOpened && !detail.IsConsumed() {
			count++
		}
	}
	return count
}
