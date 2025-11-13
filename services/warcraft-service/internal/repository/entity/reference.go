package entity

import (
	"time"
)

type Race struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Key       string    `gorm:"not null;unique;index"`
	Name      string    `gorm:"not null"`
	FactionID string    `gorm:"type:uuid;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Race) TableName() string {
	return "races"
}

type Class struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Key       string    `gorm:"not null;unique;index"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Class) TableName() string {
	return "classes"
}

type Faction struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Key       string    `gorm:"not null;unique;index"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Faction) TableName() string {
	return "factions"
}
