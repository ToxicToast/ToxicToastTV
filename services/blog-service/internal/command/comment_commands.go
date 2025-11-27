package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
)

// ============================================================================
// Commands
// ============================================================================

// CreateCommentCommand creates a new comment
type CreateCommentCommand struct {
	cqrs.BaseCommand
	PostID      string  `json:"post_id"`
	ParentID    *string `json:"parent_id"`
	AuthorName  string  `json:"author_name"`
	AuthorEmail string  `json:"author_email"`
	Content     string  `json:"content"`
}

func (c *CreateCommentCommand) CommandName() string {
	return "create_comment"
}

func (c *CreateCommentCommand) Validate() error {
	if c.PostID == "" {
		return errors.New("post_id is required")
	}
	if c.AuthorName == "" {
		return errors.New("author_name is required")
	}
	if c.AuthorEmail == "" {
		return errors.New("author_email is required")
	}
	if c.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

// UpdateCommentCommand updates an existing comment
type UpdateCommentCommand struct {
	cqrs.BaseCommand
	Content string `json:"content"`
}

func (c *UpdateCommentCommand) CommandName() string {
	return "update_comment"
}

func (c *UpdateCommentCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("comment_id is required")
	}
	if c.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

// DeleteCommentCommand deletes a comment
type DeleteCommentCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteCommentCommand) CommandName() string {
	return "delete_comment"
}

func (c *DeleteCommentCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("comment_id is required")
	}
	return nil
}

// ModerateCommentCommand moderates a comment (change status)
type ModerateCommentCommand struct {
	cqrs.BaseCommand
	Status domain.CommentStatus `json:"status"`
}

func (c *ModerateCommentCommand) CommandName() string {
	return "moderate_comment"
}

func (c *ModerateCommentCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("comment_id is required")
	}

	// Validate status
	validStatuses := map[domain.CommentStatus]bool{
		domain.CommentStatusPending:  true,
		domain.CommentStatusApproved: true,
		domain.CommentStatusSpam:     true,
		domain.CommentStatusTrash:    true,
	}

	if !validStatuses[c.Status] {
		return fmt.Errorf("invalid comment status: %s", c.Status)
	}

	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

// CreateCommentHandler handles comment creation
type CreateCommentHandler struct {
	commentRepo   repository.CommentRepository
	postRepo      repository.PostRepository
	kafkaProducer *kafka.Producer
}

func NewCreateCommentHandler(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	kafkaProducer *kafka.Producer,
) *CreateCommentHandler {
	return &CreateCommentHandler{
		commentRepo:   commentRepo,
		postRepo:      postRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *CreateCommentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateCommentCommand)

	// Validate that post exists
	_, err := h.postRepo.GetByID(ctx, createCmd.PostID)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	// Validate parent comment exists if provided
	if createCmd.ParentID != nil {
		parent, err := h.commentRepo.GetByID(ctx, *createCmd.ParentID)
		if err != nil {
			return fmt.Errorf("parent comment not found: %w", err)
		}

		// Ensure parent belongs to the same post
		if parent.PostID != createCmd.PostID {
			return fmt.Errorf("parent comment does not belong to the same post")
		}
	}

	// Generate UUID for comment
	commentID := uuid.New().String()

	// Create comment entity (default status: pending for moderation)
	comment := &domain.Comment{
		ID:          commentID,
		PostID:      createCmd.PostID,
		ParentID:    createCmd.ParentID,
		AuthorName:  createCmd.AuthorName,
		AuthorEmail: createCmd.AuthorEmail,
		Content:     createCmd.Content,
		Status:      domain.CommentStatusPending,
	}

	// Save to database
	if err := h.commentRepo.Create(ctx, comment); err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	// Store comment ID in command result
	createCmd.AggregateID = commentID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.CommentCreatedEvent{
			CommentID:   comment.ID,
			PostID:      comment.PostID,
			AuthorName:  comment.AuthorName,
			AuthorEmail: comment.AuthorEmail,
			Content:     comment.Content,
			Status:      string(comment.Status),
			CreatedAt:   comment.CreatedAt,
		}
		if err := h.kafkaProducer.PublishCommentCreated("blog.comment.created", event); err != nil {
			fmt.Printf("Warning: Failed to publish comment created event: %v\n", err)
		}
	}

	return nil
}

// UpdateCommentHandler handles comment updates
type UpdateCommentHandler struct {
	commentRepo repository.CommentRepository
}

func NewUpdateCommentHandler(
	commentRepo repository.CommentRepository,
) *UpdateCommentHandler {
	return &UpdateCommentHandler{
		commentRepo: commentRepo,
	}
}

func (h *UpdateCommentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateCommentCommand)

	// Get existing comment
	comment, err := h.commentRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// Update content
	comment.Content = updateCmd.Content

	// Save to database
	if err := h.commentRepo.Update(ctx, comment); err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	return nil
}

// DeleteCommentHandler handles comment deletion
type DeleteCommentHandler struct {
	commentRepo   repository.CommentRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteCommentHandler(
	commentRepo repository.CommentRepository,
	kafkaProducer *kafka.Producer,
) *DeleteCommentHandler {
	return &DeleteCommentHandler{
		commentRepo:   commentRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteCommentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteCommentCommand)

	// Check if comment exists
	comment, err := h.commentRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("comment not found: %w", err)
	}

	// Check if comment has replies
	replies, err := h.commentRepo.GetReplies(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to check replies: %w", err)
	}

	if len(replies) > 0 {
		return fmt.Errorf("cannot delete comment with replies")
	}

	// Delete from database (soft delete)
	if err := h.commentRepo.Delete(ctx, comment.ID); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.CommentDeletedEvent{
			CommentID: comment.ID,
			PostID:    comment.PostID,
			DeletedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishCommentDeleted("blog.comment.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish comment deleted event: %v\n", err)
		}
	}

	return nil
}

// ModerateCommentHandler handles comment moderation (status changes)
type ModerateCommentHandler struct {
	commentRepo   repository.CommentRepository
	kafkaProducer *kafka.Producer
}

func NewModerateCommentHandler(
	commentRepo repository.CommentRepository,
	kafkaProducer *kafka.Producer,
) *ModerateCommentHandler {
	return &ModerateCommentHandler{
		commentRepo:   commentRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *ModerateCommentHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	moderateCmd := cmd.(*ModerateCommentCommand)

	// Get existing comment
	comment, err := h.commentRepo.GetByID(ctx, moderateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// Update status
	comment.Status = moderateCmd.Status

	// Save to database
	if err := h.commentRepo.Update(ctx, comment); err != nil {
		return fmt.Errorf("failed to moderate comment: %w", err)
	}

	// Publish Kafka event based on status
	if h.kafkaProducer != nil {
		switch moderateCmd.Status {
		case domain.CommentStatusApproved:
			event := kafka.CommentApprovedEvent{
				CommentID:  comment.ID,
				PostID:     comment.PostID,
				ApprovedAt: time.Now(),
			}
			if err := h.kafkaProducer.PublishCommentApproved("blog.comment.approved", event); err != nil {
				fmt.Printf("Warning: Failed to publish comment approved event: %v\n", err)
			}
		case domain.CommentStatusSpam, domain.CommentStatusTrash:
			event := kafka.CommentRejectedEvent{
				CommentID:  comment.ID,
				PostID:     comment.PostID,
				Reason:     string(moderateCmd.Status),
				RejectedAt: time.Now(),
			}
			if err := h.kafkaProducer.PublishCommentRejected("blog.comment.rejected", event); err != nil {
				fmt.Printf("Warning: Failed to publish comment rejected event: %v\n", err)
			}
		}
	}

	return nil
}
