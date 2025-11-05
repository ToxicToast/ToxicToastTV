package domain

import (
	"time"

	"gorm.io/gorm"
)

type Media struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Filename         string         `gorm:"type:varchar(255);not null" json:"filename"`
	OriginalFilename string         `gorm:"type:varchar(255);not null" json:"original_filename"`
	MimeType         string         `gorm:"type:varchar(100);not null" json:"mime_type"`
	Size             int64          `gorm:"not null" json:"size"`
	Path             string         `gorm:"type:varchar(500);not null" json:"path"`
	URL              string         `gorm:"type:varchar(500);not null" json:"url"`
	ThumbnailURL     *string        `gorm:"type:varchar(500)" json:"thumbnail_url"`
	Width            int            `gorm:"default:0" json:"width"`
	Height           int            `gorm:"default:0" json:"height"`
	UploadedBy       string         `gorm:"type:varchar(255);not null" json:"uploaded_by"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Media) TableName() string {
	return "media"
}
