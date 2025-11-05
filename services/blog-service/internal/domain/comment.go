package domain

import (
	"time"

	"gorm.io/gorm"
)

type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"
	CommentStatusApproved CommentStatus = "approved"
	CommentStatusSpam     CommentStatus = "spam"
	CommentStatusTrash    CommentStatus = "trash"
)

type Comment struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	PostID      string         `gorm:"type:uuid;not null;index" json:"post_id"`
	ParentID    *string        `gorm:"type:uuid" json:"parent_id"`
	AuthorName  string         `gorm:"type:varchar(255);not null" json:"author_name"`
	AuthorEmail string         `gorm:"type:varchar(255);not null" json:"author_email"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Status      CommentStatus  `gorm:"type:varchar(50);default:'pending'" json:"status"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Post    *Post     `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}
