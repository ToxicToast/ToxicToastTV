package domain

import (
	"time"
)

type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"
	CommentStatusApproved CommentStatus = "approved"
	CommentStatusSpam     CommentStatus = "spam"
	CommentStatusTrash    CommentStatus = "trash"
)

// Comment represents a blog post comment
// Pure domain model - NO infrastructure dependencies
type Comment struct {
	ID          string
	PostID      string
	ParentID    *string
	AuthorName  string
	AuthorEmail string
	Content     string
	Status      CommentStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time

	// Relations (for domain logic, not persistence)
	Post    *Post
	Parent  *Comment
	Replies []Comment
}
