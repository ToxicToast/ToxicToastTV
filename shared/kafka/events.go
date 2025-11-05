package kafka

import "time"

// Post Events
type PostCreatedEvent struct {
	PostID       string    `json:"post_id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	AuthorID     string    `json:"author_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type PostUpdatedEvent struct {
	PostID       string    `json:"post_id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PostPublishedEvent struct {
	PostID       string    `json:"post_id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	AuthorID     string    `json:"author_id"`
	PublishedAt  time.Time `json:"published_at"`
}

type PostDeletedEvent struct {
	PostID       string    `json:"post_id"`
	DeletedAt    time.Time `json:"deleted_at"`
}

// Comment Events
type CommentCreatedEvent struct {
	CommentID    string    `json:"comment_id"`
	PostID       string    `json:"post_id"`
	AuthorName   string    `json:"author_name"`
	AuthorEmail  string    `json:"author_email"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type CommentModeratedEvent struct {
	CommentID    string    `json:"comment_id"`
	PostID       string    `json:"post_id"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	ModeratedAt  time.Time `json:"moderated_at"`
}

type CommentDeletedEvent struct {
	CommentID    string    `json:"comment_id"`
	PostID       string    `json:"post_id"`
	DeletedAt    time.Time `json:"deleted_at"`
}

// Media Events
type MediaUploadedEvent struct {
	MediaID      string    `json:"media_id"`
	Filename     string    `json:"filename"`
	MimeType     string    `json:"mime_type"`
	Size         int64     `json:"size"`
	UploadedBy   string    `json:"uploaded_by"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

type MediaDeletedEvent struct {
	MediaID      string    `json:"media_id"`
	DeletedAt    time.Time `json:"deleted_at"`
}
