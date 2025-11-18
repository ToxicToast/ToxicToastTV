package entity

import (
	"time"

	"gorm.io/gorm"
)

// CommentEntity is the database entity for comments
// Contains GORM tags and infrastructure concerns
type CommentEntity struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PostID      string         `gorm:"type:uuid;not null;index"`
	ParentID    *string        `gorm:"type:uuid"`
	AuthorName  string         `gorm:"type:varchar(255);not null"`
	AuthorEmail string         `gorm:"type:varchar(255);not null"`
	Content     string         `gorm:"type:text;not null"`
	Status      string         `gorm:"type:varchar(50);default:'pending'"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Post    *PostEntity     `gorm:"foreignKey:PostID"`
	Parent  *CommentEntity  `gorm:"foreignKey:ParentID"`
	Replies []CommentEntity `gorm:"foreignKey:ParentID"`
}

// TableName sets the table name with service prefix
func (CommentEntity) TableName() string {
	return "blog_comments"
}
