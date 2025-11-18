package entity

import (
	"time"

	"gorm.io/gorm"
)

type Guild struct {
	ID                string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name              string         `gorm:"not null;index:idx_guild_name_realm_region,unique"`
	Realm             string         `gorm:"not null;index:idx_guild_name_realm_region,unique"`
	Region            string         `gorm:"not null;index:idx_guild_name_realm_region,unique"`
	FactionID         string         `gorm:"type:uuid;not null;index"`
	MemberCount       int            `gorm:"default:0"`
	AchievementPoints int            `gorm:"default:0"`
	LastSyncedAt      *time.Time
	CreatedAt         time.Time      `gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

func (Guild) TableName() string {
	return "guilds"
}
