package entity

import (
	"time"

	"gorm.io/gorm"
)

// MediaEntity is the database entity for media files
// Contains GORM tags and infrastructure concerns
type MediaEntity struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Filename         string         `gorm:"type:varchar(255);not null"`
	OriginalFilename string         `gorm:"type:varchar(255);not null"`
	MimeType         string         `gorm:"type:varchar(100);not null"`
	Size             int64          `gorm:"not null"`
	Path             string         `gorm:"type:varchar(500);not null"`
	URL              string         `gorm:"type:varchar(500);not null"`
	ThumbnailURL     *string        `gorm:"type:varchar(500)"`
	Width            int            `gorm:"default:0"`
	Height           int            `gorm:"default:0"`
	UploadedBy       string         `gorm:"type:varchar(255);not null"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// TableName sets the table name with service prefix
func (MediaEntity) TableName() string {
	return "blog_media"
}
