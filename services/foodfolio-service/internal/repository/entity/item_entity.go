package entity

import (
	"time"

	"gorm.io/gorm"
)

type ItemEntity struct {
	ID         string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name       string         `gorm:"type:varchar(255);not null"`
	Slug       string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	CategoryID string         `gorm:"type:uuid;not null;index"`
	CompanyID  string         `gorm:"type:uuid;not null;index"`
	TypeID     string         `gorm:"type:uuid;not null;index"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	// Relations
	Category     *CategoryEntity     `gorm:"foreignKey:CategoryID"`
	Company      *CompanyEntity      `gorm:"foreignKey:CompanyID"`
	Type         *TypeEntity         `gorm:"foreignKey:TypeID"`
	ItemVariants []ItemVariantEntity `gorm:"foreignKey:ItemID"`
}

func (ItemEntity) TableName() string {
	return "foodfolio_items"
}
