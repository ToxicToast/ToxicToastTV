package entity

import (
	"time"

	"gorm.io/gorm"
)

type SizeEntity struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Value     float64        `gorm:"not null"`
	Unit      string         `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	ItemVariants []ItemVariantEntity `gorm:"foreignKey:SizeID"`
}

func (SizeEntity) TableName() string {
	return "foodfolio_sizes"
}
