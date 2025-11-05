package domain

import (
	"time"

	"gorm.io/gorm"
)

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

type Post struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Title           string         `gorm:"type:varchar(255);not null" json:"title"`
	Slug            string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Content         string         `gorm:"type:text" json:"content"`
	Excerpt         string         `gorm:"type:text" json:"excerpt"`
	Markdown        string         `gorm:"type:text" json:"markdown"`
	HTML            string         `gorm:"type:text" json:"html"`
	Status          PostStatus     `gorm:"type:varchar(50);default:'draft'" json:"status"`
	Featured        bool           `gorm:"default:false" json:"featured"`
	AuthorID        string         `gorm:"type:varchar(255);not null" json:"author_id"`
	FeaturedImageID *string        `gorm:"type:uuid" json:"featured_image_id"`
	ReadingTime     int            `gorm:"default:0" json:"reading_time"`
	ViewCount       int            `gorm:"default:0" json:"view_count"`
	PublishedAt     *time.Time     `json:"published_at"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// SEO Fields
	MetaTitle       string `gorm:"type:varchar(255)" json:"meta_title"`
	MetaDescription string `gorm:"type:text" json:"meta_description"`
	MetaKeywords    string `gorm:"type:text" json:"meta_keywords"`
	OGTitle         string `gorm:"type:varchar(255)" json:"og_title"`
	OGDescription   string `gorm:"type:text" json:"og_description"`
	OGImage         string `gorm:"type:varchar(500)" json:"og_image"`
	CanonicalURL    string `gorm:"type:varchar(500)" json:"canonical_url"`

	// Relations (using many-to-many through junction tables)
	Categories []Category `gorm:"many2many:post_categories;" json:"categories,omitempty"`
	Tags       []Tag      `gorm:"many2many:post_tags;" json:"tags,omitempty"`
}

func (Post) TableName() string {
	return "posts"
}
