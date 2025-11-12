package entity

import (
	"time"

	"gorm.io/gorm"
)

type DeliveryEntity struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WebhookID     string         `gorm:"type:uuid;not null;index"`
	EventID       string         `gorm:"type:varchar(255);not null"`
	EventType     string         `gorm:"type:varchar(255);not null;index"`
	EventPayload  string         `gorm:"type:text;not null"`
	Status        string         `gorm:"type:varchar(50);not null;default:'pending';index"`
	AttemptCount  int            `gorm:"default:0"`
	NextRetryAt   *time.Time
	LastAttemptAt *time.Time
	CompletedAt   *time.Time
	LastError     string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relations
	Webhook  *WebhookEntity         `gorm:"foreignKey:WebhookID"`
	Attempts []DeliveryAttemptEntity `gorm:"foreignKey:DeliveryID"`
}

func (DeliveryEntity) TableName() string {
	return "webhook_deliveries"
}
