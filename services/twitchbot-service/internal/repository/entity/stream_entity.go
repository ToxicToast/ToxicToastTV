package entity

import (
	"time"

	"gorm.io/gorm"
)

// StreamEntity is the database entity for Twitch stream sessions
// Contains GORM tags and infrastructure concerns
type StreamEntity struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title          string         `gorm:"type:varchar(255);not null"`
	GameName       string         `gorm:"type:varchar(255)"`
	GameID         string         `gorm:"type:varchar(100)"`
	StartedAt      time.Time      `gorm:"not null"`
	EndedAt        *time.Time
	PeakViewers    int            `gorm:"default:0"`
	AverageViewers int            `gorm:"default:0"`
	TotalMessages  int            `gorm:"default:0"`
	IsActive       bool           `gorm:"default:true"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// Relations
	Messages []MessageEntity `gorm:"foreignKey:StreamID"`
	Clips    []ClipEntity    `gorm:"foreignKey:StreamID"`
}

// TableName sets the table name with service prefix
func (StreamEntity) TableName() string {
	return "twitchbot_streams"
}
