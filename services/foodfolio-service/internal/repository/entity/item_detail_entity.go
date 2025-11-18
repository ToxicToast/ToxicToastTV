package entity

import (
	"time"

	"gorm.io/gorm"
)

type ItemDetailEntity struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ItemVariantID string         `gorm:"type:uuid;not null;index"`
	WarehouseID   string         `gorm:"type:uuid;not null;index"`
	LocationID    string         `gorm:"type:uuid;not null;index"`
	ArticleNumber *string        `gorm:"type:varchar(255)"`
	PurchasePrice float64        `gorm:"type:decimal(10,2);not null"`
	PurchaseDate  time.Time      `gorm:"not null"`
	ExpiryDate    *time.Time
	OpenedDate    *time.Time
	IsOpened      bool           `gorm:"not null;default:false"`
	HasDeposit    bool           `gorm:"not null;default:false"`
	IsFrozen      bool           `gorm:"not null;default:false"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relations
	ItemVariant *ItemVariantEntity `gorm:"foreignKey:ItemVariantID"`
	Warehouse   *WarehouseEntity   `gorm:"foreignKey:WarehouseID"`
	Location    *LocationEntity    `gorm:"foreignKey:LocationID"`
}

func (ItemDetailEntity) TableName() string {
	return "foodfolio_item_details"
}
