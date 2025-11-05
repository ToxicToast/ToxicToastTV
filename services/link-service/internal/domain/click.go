package domain

import (
	"time"
)

type Click struct {
	ID         string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	LinkID     string     `gorm:"type:uuid;not null;index" json:"link_id"`
	IPAddress  string     `gorm:"type:varchar(45);not null" json:"ip_address"`
	UserAgent  string     `gorm:"type:text" json:"user_agent"`
	Referer    *string    `gorm:"type:text" json:"referer,omitempty"`
	Country    *string    `gorm:"type:varchar(100)" json:"country,omitempty"`
	City       *string    `gorm:"type:varchar(100)" json:"city,omitempty"`
	DeviceType *string    `gorm:"type:varchar(50)" json:"device_type,omitempty"`
	ClickedAt  time.Time  `gorm:"not null;index" json:"clicked_at"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Foreign key relationship
	Link Link `gorm:"foreignKey:LinkID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"link,omitempty"`
}

func (Click) TableName() string {
	return "clicks"
}
