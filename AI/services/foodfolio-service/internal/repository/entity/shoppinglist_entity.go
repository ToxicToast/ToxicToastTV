package entity

import (
	"time"

	"gorm.io/gorm"
)

type ShoppinglistEntity struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Items []*ShoppinglistItemEntity `gorm:"foreignKey:ShoppinglistID"`
}

func (ShoppinglistEntity) TableName() string {
	return "foodfolio_shoppinglists"
}
