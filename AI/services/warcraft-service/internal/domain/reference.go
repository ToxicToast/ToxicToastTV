package domain

import "time"

// Race represents a playable race in WoW
type Race struct {
	ID        string
	Key       string // e.g. "human", "orc"
	Name      string // e.g. "Human", "Orc"
	FactionID string // Foreign key to Faction
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Class represents a playable class in WoW
type Class struct {
	ID        string
	Key       string // e.g. "warrior", "mage"
	Name      string // e.g. "Warrior", "Mage"
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Faction represents Alliance or Horde
type Faction struct {
	ID        string
	Key       string // "alliance" or "horde"
	Name      string // "Alliance" or "Horde"
	CreatedAt time.Time
	UpdatedAt time.Time
}
