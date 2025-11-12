package entity

import (
	"time"

	"gorm.io/gorm"
)

// NotificationAttemptEntity is the database entity for notification attempts
// Contains GORM tags and infrastructure concerns
type NotificationAttemptEntity struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	NotificationID   string         `gorm:"type:uuid;not null;index"`
	AttemptNumber    int            `gorm:"not null"`
	ResponseStatus   int
	ResponseBody     string `gorm:"type:text"`
	DiscordMessageID string `gorm:"type:varchar(255)"`
	Success          bool
	Error            string    `gorm:"type:text"`
	DurationMs       int
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	// Relations
	Notification *NotificationEntity `gorm:"foreignKey:NotificationID"`
}

// TableName sets the table name with service prefix
func (NotificationAttemptEntity) TableName() string {
	return "notification_attempts"
}
