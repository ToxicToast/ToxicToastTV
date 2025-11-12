package entity

import (
	"time"

	"gorm.io/gorm"
)

// ViewerEntity is the database entity for viewers/users
// Contains GORM tags and infrastructure concerns
type ViewerEntity struct {
	ID                  string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TwitchID            string         `gorm:"type:varchar(100);not null;uniqueIndex"`
	Username            string         `gorm:"type:varchar(100);not null"`
	DisplayName         string         `gorm:"type:varchar(100);not null"`
	TotalMessages       int            `gorm:"default:0"`
	TotalStreamsWatched int            `gorm:"default:0"`
	FirstSeen           time.Time      `gorm:"not null"`
	LastSeen            time.Time      `gorm:"not null"`
	CreatedAt           time.Time      `gorm:"autoCreateTime"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime"`
	DeletedAt           gorm.DeletedAt `gorm:"index"`

	// Relations
	Messages []MessageEntity `gorm:"foreignKey:UserID;references:TwitchID"`
	Clips    []ClipEntity    `gorm:"foreignKey:CreatorID;references:TwitchID"`
}

// TableName sets the table name with service prefix
func (ViewerEntity) TableName() string {
	return "twitchbot_viewers"
}
