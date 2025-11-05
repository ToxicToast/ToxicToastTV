package domain

import (
	"time"

	"gorm.io/gorm"
)

// Item represents a base product (e.g., "Monster Energy", "Coca Cola")
type Item struct {
	ID         string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name       string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug       string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	CategoryID string         `gorm:"type:uuid;not null;index" json:"category_id"`
	CompanyID  string         `gorm:"type:uuid;not null;index" json:"company_id"`
	TypeID     string         `gorm:"type:uuid;not null;index" json:"type_id"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Category     *Category     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Company      *Company      `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Type         *Type         `gorm:"foreignKey:TypeID" json:"type,omitempty"`
	ItemVariants []ItemVariant `gorm:"foreignKey:ItemID" json:"item_variants,omitempty"`
}

func (Item) TableName() string {
	return "items"
}
