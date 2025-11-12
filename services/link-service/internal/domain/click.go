package domain

import (
	"time"
)

type Click struct {
	ID         string     `json:"id"`
	LinkID     string     `json:"link_id"`
	IPAddress  string     `json:"ip_address"`
	UserAgent  string     `json:"user_agent"`
	Referer    *string    `json:"referer,omitempty"`
	Country    *string    `json:"country,omitempty"`
	City       *string    `json:"city,omitempty"`
	DeviceType *string    `json:"device_type,omitempty"`
	ClickedAt  time.Time  `json:"clicked_at"`
	CreatedAt  time.Time  `json:"created_at"`

	// Domain relationship (no GORM constraint)
	Link *Link `json:"link,omitempty"`
}
