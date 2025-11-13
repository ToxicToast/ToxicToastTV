package domain

import (
	"time"
)

type Guild struct {
	ID                string
	Name              string
	Realm             string
	Region            string
	FactionID         string // Foreign key to Faction
	MemberCount       int
	AchievementPoints int
	LastSyncedAt      *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type GuildMember struct {
	ID             string
	GuildID        string
	CharacterName  string
	CharacterRealm string
	Rank           int
	Level          int
	CharacterClass string
	Race           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
