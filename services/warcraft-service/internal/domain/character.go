package domain

import (
	"time"
)

type Character struct {
	ID        string
	Name      string  // Lowercase name for Blizzard API
	Realm     string  // Slug realm for Blizzard API
	Region    string  // Region (us, eu, kr, tw, cn)
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type CharacterDetails struct {
	ID                string
	CharacterID       string
	DisplayName       string  // Formatted character name from Blizzard
	DisplayRealm      string  // Formatted realm name from Blizzard
	Level             int
	ItemLevel         int
	ClassID           string  // Foreign key to Class
	RaceID            string  // Foreign key to Race
	FactionID         string  // Foreign key to Faction
	GuildID           *string // Foreign key to Guild (optional)
	ThumbnailURL      *string
	AchievementPoints int
	LastSyncedAt      *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CharacterEquipment struct {
	ID            string
	CharacterID   string
	EquipmentJSON []byte // Flexible JSON storage for equipment data
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CharacterStats struct {
	ID          string
	CharacterID string
	StatsJSON   []byte // Flexible JSON storage for stats data
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
