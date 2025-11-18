package entity

import (
	"time"

	"gorm.io/gorm"
)

type ShoppinglistItemEntity struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ShoppinglistID string         `gorm:"type:uuid;not null;index"`
	ItemVariantID  string         `gorm:"type:uuid;not null;index"`
	Quantity       int            `gorm:"not null;default:1"`
	IsPurchased    bool           `gorm:"not null;default:false"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// Relations
	Shoppinglist *ShoppinglistEntity `gorm:"foreignKey:ShoppinglistID"`
	ItemVariant  *ItemVariantEntity  `gorm:"foreignKey:ItemVariantID"`
}

func (ShoppinglistItemEntity) TableName() string {
	return "foodfolio_shoppinglist_items"
}
