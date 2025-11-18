package entity

import (
	"time"

	"gorm.io/gorm"
)

type TypeEntity struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	Slug      string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Items []ItemEntity `gorm:"foreignKey:TypeID"`
}

func (TypeEntity) TableName() string {
	return "foodfolio_types"
}
