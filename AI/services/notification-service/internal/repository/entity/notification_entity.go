package entity

import (
	"time"

	"gorm.io/gorm"
)

// NotificationEntity is the database entity for notifications
// Contains GORM tags and infrastructure concerns
type NotificationEntity struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ChannelID        string         `gorm:"type:uuid;not null;index"`
	EventID          string         `gorm:"type:varchar(255);not null"`
	EventType        string         `gorm:"type:varchar(255);not null;index"`
	EventPayload     string         `gorm:"type:text;not null"` // JSON string
	DiscordMessageID string         `gorm:"type:varchar(255)"`
	Status           string         `gorm:"type:varchar(50);not null;default:'pending';index"`
	AttemptCount     int            `gorm:"default:0"`
	LastError        string         `gorm:"type:text"`
	SentAt           *time.Time
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	// Relations
	Channel  *DiscordChannelEntity      `gorm:"foreignKey:ChannelID"`
	Attempts []NotificationAttemptEntity `gorm:"foreignKey:NotificationID"`
}

// TableName sets the table name with service prefix
func (NotificationEntity) TableName() string {
	return "notification_notifications"
}
