package domain

import (
	"time"

	"gorm.io/gorm"
)

type Link struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OriginalURL string         `gorm:"type:text;not null" json:"original_url"`
	ShortCode   string         `gorm:"type:varchar(10);uniqueIndex;not null" json:"short_code"`
	CustomAlias *string        `gorm:"type:varchar(50);uniqueIndex" json:"custom_alias,omitempty"`
	Title       *string        `gorm:"type:varchar(255)" json:"title,omitempty"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	ClickCount  int            `gorm:"default:0" json:"click_count"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Link) TableName() string {
	return "links"
}

// IsExpired checks if the link has expired
func (l *Link) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*l.ExpiresAt)
}

// IsAvailable checks if the link is available for use
func (l *Link) IsAvailable() bool {
	return l.IsActive && !l.IsExpired()
}
