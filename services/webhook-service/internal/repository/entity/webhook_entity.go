package entity

import (
	"time"

	"gorm.io/gorm"
)

type WebhookEntity struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	URL         string         `gorm:"type:varchar(500);not null;uniqueIndex"`
	Secret      string         `gorm:"type:varchar(255);not null"`
	EventTypes  string         `gorm:"type:text"`
	Description string         `gorm:"type:text"`
	Active      bool           `gorm:"default:true"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Statistics
	TotalDeliveries   int       `gorm:"default:0"`
	SuccessDeliveries int       `gorm:"default:0"`
	FailedDeliveries  int       `gorm:"default:0"`
	LastDeliveryAt    time.Time
	LastSuccessAt     time.Time
	LastFailureAt     time.Time
}

func (WebhookEntity) TableName() string {
	return "webhook_webhooks"
}
