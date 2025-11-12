package entity

import (
	"time"

	"gorm.io/gorm"
)

// DiscordChannelEntity is the database entity for Discord channels
// Contains GORM tags and infrastructure concerns
type DiscordChannelEntity struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `gorm:"type:varchar(255);not null"`
	WebhookURL  string         `gorm:"type:varchar(500);not null;uniqueIndex"`
	EventTypes  string         `gorm:"type:text"` // Comma-separated event types
	Color       int            `gorm:"default:3447003"`
	Active      bool           `gorm:"default:true;index"`
	Description string         `gorm:"type:text"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Statistics
	TotalNotifications   int        `gorm:"default:0"`
	SuccessNotifications int        `gorm:"default:0"`
	FailedNotifications  int        `gorm:"default:0"`
	LastNotificationAt   *time.Time
	LastSuccessAt        *time.Time
	LastFailureAt        *time.Time
}

// TableName sets the table name with service prefix
func (DiscordChannelEntity) TableName() string {
	return "notification_channels"
}
