package domain

import (
	"time"
)

// Media represents an uploaded media file
// Pure domain model - NO infrastructure dependencies
type Media struct {
	ID               string
	Filename         string
	OriginalFilename string
	MimeType         string
	Size             int64
	Path             string
	URL              string
	ThumbnailURL     *string
	Width            int
	Height           int
	UploadedBy       string
	CreatedAt        time.Time
	DeletedAt        *time.Time
}
