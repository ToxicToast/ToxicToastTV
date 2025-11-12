package mapper

import (
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository/entity"

	"gorm.io/gorm"
)

// CommentToEntity converts domain model to database entity
func CommentToEntity(comment *domain.Comment) *entity.CommentEntity {
	if comment == nil {
		return nil
	}

	e := &entity.CommentEntity{
		ID:          comment.ID,
		PostID:      comment.PostID,
		ParentID:    comment.ParentID,
		AuthorName:  comment.AuthorName,
		AuthorEmail: comment.AuthorEmail,
		Content:     comment.Content,
		Status:      string(comment.Status),
		CreatedAt:   comment.CreatedAt,
		UpdatedAt:   comment.UpdatedAt,
	}

	// Convert DeletedAt
	if comment.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *comment.DeletedAt,
			Valid: true,
		}
	}

	// Convert Parent (avoid infinite recursion)
	if comment.Parent != nil {
		e.Parent = CommentToEntity(comment.Parent)
	}

	// Convert Replies (avoid infinite recursion)
	if len(comment.Replies) > 0 {
		e.Replies = make([]entity.CommentEntity, 0, len(comment.Replies))
		for _, reply := range comment.Replies {
			e.Replies = append(e.Replies, *CommentToEntity(&reply))
		}
	}

	return e
}

// CommentToDomain converts database entity to domain model
func CommentToDomain(e *entity.CommentEntity) *domain.Comment {
	if e == nil {
		return nil
	}

	comment := &domain.Comment{
		ID:          e.ID,
		PostID:      e.PostID,
		ParentID:    e.ParentID,
		AuthorName:  e.AuthorName,
		AuthorEmail: e.AuthorEmail,
		Content:     e.Content,
		Status:      domain.CommentStatus(e.Status),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		comment.DeletedAt = &deletedAt
	}

	// Convert Parent (avoid infinite recursion)
	if e.Parent != nil {
		comment.Parent = CommentToDomain(e.Parent)
	}

	// Convert Replies (avoid infinite recursion)
	if len(e.Replies) > 0 {
		comment.Replies = make([]domain.Comment, 0, len(e.Replies))
		for _, reply := range e.Replies {
			comment.Replies = append(comment.Replies, *CommentToDomain(&reply))
		}
	}

	// Convert Post if present
	if e.Post != nil {
		comment.Post = PostToDomain(e.Post)
	}

	return comment
}

// CommentsToDomain converts slice of entities to domain models
func CommentsToDomain(entities []entity.CommentEntity) []*domain.Comment {
	comments := make([]*domain.Comment, 0, len(entities))
	for _, e := range entities {
		comments = append(comments, CommentToDomain(&e))
	}
	return comments
}
