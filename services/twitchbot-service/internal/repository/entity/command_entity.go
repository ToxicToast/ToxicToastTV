package entity

import (
	"time"

	"gorm.io/gorm"
)

// CommandEntity is the database entity for chat commands
// Contains GORM tags and infrastructure concerns
type CommandEntity struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name            string         `gorm:"type:varchar(100);not null;uniqueIndex"`
	Description     string         `gorm:"type:text"`
	Response        string         `gorm:"type:text;not null"`
	IsActive        bool           `gorm:"default:true"`
	ModeratorOnly   bool           `gorm:"default:false"`
	SubscriberOnly  bool           `gorm:"default:false"`
	CooldownSeconds int            `gorm:"default:0"`
	UsageCount      int            `gorm:"default:0"`
	LastUsed        *time.Time
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// TableName sets the table name with service prefix
func (CommandEntity) TableName() string {
	return "twitchbot_commands"
}
