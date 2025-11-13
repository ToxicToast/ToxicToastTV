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
	ID          string
	CharacterID string
	Head        *EquipmentSlot
	Neck        *EquipmentSlot
	Shoulder    *EquipmentSlot
	Back        *EquipmentSlot
	Chest       *EquipmentSlot
	Wrist       *EquipmentSlot
	Hands       *EquipmentSlot
	Waist       *EquipmentSlot
	Legs        *EquipmentSlot
	Feet        *EquipmentSlot
	Finger1     *EquipmentSlot
	Finger2     *EquipmentSlot
	Trinket1    *EquipmentSlot
	Trinket2    *EquipmentSlot
	MainHand    *EquipmentSlot
	OffHand     *EquipmentSlot
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EquipmentSlot struct {
	ItemID    int
	Name      string
	ItemLevel int
	Quality   string
	Icon      string
}

type CharacterStats struct {
	ID             string
	CharacterID    string
	Health         int
	Strength       int
	Agility        int
	Intellect      int
	Stamina        int
	CriticalStrike int
	Haste          int
	Mastery        int
	Versatility    int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
