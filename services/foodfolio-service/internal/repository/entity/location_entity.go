package entity

import (
	"time"

	"gorm.io/gorm"
)

type LocationEntity struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Slug      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	ParentID  *string        `gorm:"type:uuid"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Parent      *LocationEntity    `gorm:"foreignKey:ParentID"`
	Children    []LocationEntity   `gorm:"foreignKey:ParentID"`
	ItemDetails []ItemDetailEntity `gorm:"foreignKey:LocationID"`
}

func (LocationEntity) TableName() string {
	return "foodfolio_locations"
}
