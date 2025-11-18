package entity

import (
	"time"

	"gorm.io/gorm"
)

// ChannelViewerEntity is the database entity for channel-specific viewer presence
// Contains GORM tags and infrastructure concerns
type ChannelViewerEntity struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Channel     string         `gorm:"type:varchar(100);not null;index:idx_channel_viewer_unique,unique"`
	TwitchID    string         `gorm:"type:varchar(100);not null;index:idx_channel_viewer_unique,unique"`
	Username    string         `gorm:"type:varchar(100);not null"`
	DisplayName string         `gorm:"type:varchar(100);not null"`
	FirstSeen   time.Time      `gorm:"not null"`
	LastSeen    time.Time      `gorm:"not null"`
	IsModerator bool           `gorm:"default:false"`
	IsVIP       bool           `gorm:"default:false"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relation to global Viewer (optional)
	Viewer *ViewerEntity `gorm:"foreignKey:TwitchID;references:TwitchID"`
}

// TableName sets the table name with service prefix
func (ChannelViewerEntity) TableName() string {
	return "twitchbot_channel_viewers"
}
