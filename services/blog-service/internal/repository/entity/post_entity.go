package entity

import (
	"time"

	"gorm.io/gorm"
)

// PostEntity is the database entity for blog posts
// Contains GORM tags and infrastructure concerns
type PostEntity struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title           string         `gorm:"type:varchar(255);not null"`
	Slug            string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Content         string         `gorm:"type:text"`
	Excerpt         string         `gorm:"type:text"`
	Markdown        string         `gorm:"type:text"`
	HTML            string         `gorm:"type:text"`
	Status          string         `gorm:"type:varchar(50);default:'draft'"`
	Featured        bool           `gorm:"default:false"`
	AuthorID        string         `gorm:"type:varchar(255);not null"`
	FeaturedImageID *string        `gorm:"type:uuid"`
	ReadingTime     int            `gorm:"default:0"`
	ViewCount       int            `gorm:"default:0"`
	PublishedAt     *time.Time
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	// SEO Fields
	MetaTitle       string `gorm:"type:varchar(255)"`
	MetaDescription string `gorm:"type:text"`
	MetaKeywords    string `gorm:"type:text"`
	OGTitle         string `gorm:"type:varchar(255)"`
	OGDescription   string `gorm:"type:text"`
	OGImage         string `gorm:"type:varchar(500)"`
	CanonicalURL    string `gorm:"type:varchar(500)"`

	// Relations
	Categories []CategoryEntity `gorm:"many2many:blog_post_categories;"`
	Tags       []TagEntity      `gorm:"many2many:blog_post_tags;"`
}

// TableName sets the table name with service prefix
func (PostEntity) TableName() string {
	return "blog_posts"
}
