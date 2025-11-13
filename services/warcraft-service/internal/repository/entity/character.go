package entity

import (
	"time"

	"gorm.io/gorm"
)

type Character struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string         `gorm:"not null;index:idx_character_name_realm_region,unique"`
	Realm     string         `gorm:"not null;index:idx_character_name_realm_region,unique"`
	Region    string         `gorm:"not null;index:idx_character_name_realm_region,unique"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Character) TableName() string {
	return "characters"
}

type CharacterDetails struct {
	ID                string     `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CharacterID       string     `gorm:"type:uuid;not null;unique;index"`
	DisplayName       string     `gorm:"not null"`
	DisplayRealm      string     `gorm:"not null"`
	Level             int        `gorm:"not null"`
	ItemLevel         int        `gorm:"not null"`
	ClassID           string     `gorm:"type:uuid;not null;index"`
	RaceID            string     `gorm:"type:uuid;not null;index"`
	FactionID         string     `gorm:"type:uuid;not null;index"`
	GuildID           *string    `gorm:"type:uuid;index"`
	ThumbnailURL      *string    `gorm:"type:text"`
	AchievementPoints int        `gorm:"default:0"`
	LastSyncedAt      *time.Time
	CreatedAt         time.Time  `gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime"`
}

func (CharacterDetails) TableName() string {
	return "character_details"
}
