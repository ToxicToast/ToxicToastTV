package entity

import (
	"time"

	"gorm.io/gorm"
)

// MessageEntity is the database entity for chat messages
// Contains GORM tags and infrastructure concerns
type MessageEntity struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StreamID      string         `gorm:"type:uuid;not null;index"`
	UserID        string         `gorm:"type:varchar(100);not null;index"`
	Username      string         `gorm:"type:varchar(100);not null"`
	DisplayName   string         `gorm:"type:varchar(100);not null"`
	Message       string         `gorm:"type:text;not null"`
	IsModerator   bool           `gorm:"default:false"`
	IsSubscriber  bool           `gorm:"default:false"`
	IsVIP         bool           `gorm:"default:false"`
	IsBroadcaster bool           `gorm:"default:false"`
	SentAt        time.Time      `gorm:"not null;index"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relations
	Stream *StreamEntity `gorm:"foreignKey:StreamID"`
	Viewer *ViewerEntity `gorm:"foreignKey:UserID;references:TwitchID"`
}

// TableName sets the table name with service prefix
func (MessageEntity) TableName() string {
	return "twitchbot_messages"
}
