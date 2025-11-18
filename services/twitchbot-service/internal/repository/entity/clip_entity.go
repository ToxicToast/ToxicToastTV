package entity

import (
	"time"

	"gorm.io/gorm"
)

// ClipEntity is the database entity for Twitch clips
// Contains GORM tags and infrastructure concerns
type ClipEntity struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StreamID        string         `gorm:"type:uuid;not null;index"`
	TwitchClipID    string         `gorm:"type:varchar(100);not null;uniqueIndex"`
	Title           string         `gorm:"type:varchar(255);not null"`
	URL             string         `gorm:"type:text;not null"`
	EmbedURL        string         `gorm:"type:text"`
	ThumbnailURL    string         `gorm:"type:text"`
	CreatorName     string         `gorm:"type:varchar(100);not null"`
	CreatorID       string         `gorm:"type:varchar(100);not null;index"`
	ViewCount       int            `gorm:"default:0"`
	DurationSeconds int            `gorm:"not null"`
	CreatedAtTwitch time.Time      `gorm:"not null"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	// Relations
	Stream  *StreamEntity `gorm:"foreignKey:StreamID"`
	Creator *ViewerEntity `gorm:"foreignKey:CreatorID;references:TwitchID"`
}

// TableName sets the table name with service prefix
func (ClipEntity) TableName() string {
	return "twitchbot_clips"
}
