package entity

import (
	"time"
)

type ClickEntity struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	LinkID     string    `gorm:"type:uuid;not null;index"`
	IPAddress  string    `gorm:"type:varchar(45);not null"`
	UserAgent  string    `gorm:"type:text"`
	Referer    *string   `gorm:"type:text"`
	Country    *string   `gorm:"type:varchar(100)"`
	City       *string   `gorm:"type:varchar(100)"`
	DeviceType *string   `gorm:"type:varchar(50)"`
	ClickedAt  time.Time `gorm:"not null;index"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`

	// Foreign key relationship
	Link LinkEntity `gorm:"foreignKey:LinkID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (ClickEntity) TableName() string {
	return "link_clicks"
}
