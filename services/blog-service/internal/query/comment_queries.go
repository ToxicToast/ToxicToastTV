package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
)

// ============================================================================
// Queries
// ============================================================================

// GetCommentByIDQuery retrieves a comment by ID
type GetCommentByIDQuery struct {
	cqrs.BaseQuery
	CommentID string `json:"comment_id"`
}

func (q *GetCommentByIDQuery) QueryName() string {
	return "get_comment_by_id"
}

func (q *GetCommentByIDQuery) Validate() error {
	if q.CommentID == "" {
		return errors.New("comment_id is required")
	}
	return nil
}

// ListCommentsQuery retrieves a list of comments with filtering
type ListCommentsQuery struct {
	cqrs.BaseQuery
	Filters repository.CommentFilters `json:"filters"`
}

func (q *ListCommentsQuery) QueryName() string {
	return "list_comments"
}

func (q *ListCommentsQuery) Validate() error {
	// No strict validation required for filtering
	return nil
}

// GetCommentRepliesQuery retrieves replies to a comment
type GetCommentRepliesQuery struct {
	cqrs.BaseQuery
	ParentID string `json:"parent_id"`
}

func (q *GetCommentRepliesQuery) QueryName() string {
	return "get_comment_replies"
}

func (q *GetCommentRepliesQuery) Validate() error {
	if q.ParentID == "" {
		return errors.New("parent_id is required")
	}
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// ListCommentsResult contains the result of listing comments
type ListCommentsResult struct {
	Comments []domain.Comment
	Total    int64
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetCommentByIDHandler handles comment retrieval by ID
type GetCommentByIDHandler struct {
	commentRepo repository.CommentRepository
}

func NewGetCommentByIDHandler(commentRepo repository.CommentRepository) *GetCommentByIDHandler {
	return &GetCommentByIDHandler{
		commentRepo: commentRepo,
	}
}

func (h *GetCommentByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCommentByIDQuery)

	comment, err := h.commentRepo.GetByID(ctx, q.CommentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return comment, nil
}

// ListCommentsHandler handles comment listing with filters
type ListCommentsHandler struct {
	commentRepo repository.CommentRepository
}

func NewListCommentsHandler(commentRepo repository.CommentRepository) *ListCommentsHandler {
	return &ListCommentsHandler{
		commentRepo: commentRepo,
	}
}

func (h *ListCommentsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListCommentsQuery)

	comments, total, err := h.commentRepo.List(ctx, q.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}

	return &ListCommentsResult{
		Comments: comments,
		Total:    total,
	}, nil
}

// GetCommentRepliesHandler handles retrieving replies to a comment
type GetCommentRepliesHandler struct {
	commentRepo repository.CommentRepository
}

func NewGetCommentRepliesHandler(commentRepo repository.CommentRepository) *GetCommentRepliesHandler {
	return &GetCommentRepliesHandler{
		commentRepo: commentRepo,
	}
}

func (h *GetCommentRepliesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCommentRepliesQuery)

	replies, err := h.commentRepo.GetReplies(ctx, q.ParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment replies: %w", err)
	}

	return replies, nil
}
