package entity

import (
	"time"

	"gorm.io/gorm"
)

type LinkEntity struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OriginalURL string         `gorm:"type:text;not null"`
	ShortCode   string         `gorm:"type:varchar(10);uniqueIndex;not null"`
	CustomAlias *string        `gorm:"type:varchar(50);uniqueIndex"`
	Title       *string        `gorm:"type:varchar(255)"`
	Description *string        `gorm:"type:text"`
	ExpiresAt   *time.Time
	IsActive    bool      `gorm:"default:true"`
	ClickCount  int       `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (LinkEntity) TableName() string {
	return "link_links"
}
