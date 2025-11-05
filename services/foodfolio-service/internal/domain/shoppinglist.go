package domain

import (
	"time"

	"gorm.io/gorm"
)

// Shoppinglist represents a shopping list
type Shoppinglist struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`                   // e.g., "Weekly Shopping", "Party Supplies"
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Items []*ShoppinglistItem `gorm:"foreignKey:ShoppinglistID" json:"items,omitempty"`
}

func (Shoppinglist) TableName() string {
	return "shoppinglists"
}

// ShoppinglistItem represents an item on a shopping list
type ShoppinglistItem struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ShoppinglistID   string         `gorm:"type:uuid;not null;index" json:"shoppinglist_id"`
	ItemVariantID    string         `gorm:"type:uuid;not null;index" json:"item_variant_id"`
	Quantity         int            `gorm:"not null;default:1" json:"quantity"`                      // How many to buy
	IsPurchased      bool           `gorm:"not null;default:false" json:"is_purchased"`              // Already bought
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Shoppinglist *Shoppinglist `gorm:"foreignKey:ShoppinglistID" json:"shoppinglist,omitempty"`
	ItemVariant  *ItemVariant  `gorm:"foreignKey:ItemVariantID" json:"item_variant,omitempty"`
}

func (ShoppinglistItem) TableName() string {
	return "shoppinglist_items"
}
