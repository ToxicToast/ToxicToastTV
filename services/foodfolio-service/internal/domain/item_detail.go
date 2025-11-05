package domain

import (
	"time"

	"gorm.io/gorm"
)

// ItemDetail represents one physical item/unit
// Each can/bottle/package is tracked individually
type ItemDetail struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ItemVariantID string         `gorm:"type:uuid;not null;index" json:"item_variant_id"`
	WarehouseID   string         `gorm:"type:uuid;not null;index" json:"warehouse_id"`             // Where purchased
	LocationID    string         `gorm:"type:uuid;not null;index" json:"location_id"`              // Where stored
	ArticleNumber *string        `gorm:"type:varchar(255)" json:"article_number"`                  // Shop's article number
	PurchasePrice float64        `gorm:"type:decimal(10,2);not null" json:"purchase_price"`        // Actual paid price
	PurchaseDate  time.Time      `gorm:"not null" json:"purchase_date"`                            // When bought
	ExpiryDate    *time.Time     `json:"expiry_date"`                                              // MHD (best before date)
	OpenedDate    *time.Time     `json:"opened_date"`                                              // When opened
	IsOpened      bool           `gorm:"not null;default:false" json:"is_opened"`                  // Currently opened
	HasDeposit    bool           `gorm:"not null;default:false" json:"has_deposit"`                // Pfand
	IsFrozen      bool           `gorm:"not null;default:false" json:"is_frozen"`                  // Currently frozen
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	ItemVariant *ItemVariant `gorm:"foreignKey:ItemVariantID" json:"item_variant,omitempty"`
	Warehouse   *Warehouse   `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Location    *Location    `gorm:"foreignKey:LocationID" json:"location,omitempty"`
}

func (ItemDetail) TableName() string {
	return "item_details"
}

// IsExpired checks if the item has expired
func (id *ItemDetail) IsExpired() bool {
	if id.ExpiryDate == nil {
		return false
	}
	return time.Now().After(*id.ExpiryDate)
}

// IsExpiringSoon checks if item expires within given days
func (id *ItemDetail) IsExpiringSoon(days int) bool {
	if id.ExpiryDate == nil {
		return false
	}
	threshold := time.Now().AddDate(0, 0, days)
	return id.ExpiryDate.Before(threshold)
}

// IsConsumed checks if item is likely consumed (soft-deleted or opened long ago)
func (id *ItemDetail) IsConsumed() bool {
	// If soft-deleted, it's consumed
	if id.DeletedAt.Valid {
		return true
	}
	// If opened more than 30 days ago, consider consumed
	if id.OpenedDate != nil && time.Since(*id.OpenedDate) > 30*24*time.Hour {
		return true
	}
	return false
}
