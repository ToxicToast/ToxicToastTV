package domain

import (
	"time"
)

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

// Post represents a blog post
// Pure domain model - NO infrastructure dependencies
type Post struct {
	ID              string
	Title           string
	Slug            string
	Content         string
	Excerpt         string
	Markdown        string
	HTML            string
	Status          PostStatus
	Featured        bool
	AuthorID        string
	FeaturedImageID *string
	ReadingTime     int
	ViewCount       int
	PublishedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time

	// SEO Fields
	MetaTitle       string
	MetaDescription string
	MetaKeywords    string
	OGTitle         string
	OGDescription   string
	OGImage         string
	CanonicalURL    string

	// Relations (for domain logic, not persistence)
	Categories []Category
	Tags       []Tag
}
