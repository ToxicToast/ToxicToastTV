package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/pkg/config"
)

type CommentUseCase interface {
	CreateComment(ctx context.Context, input CreateCommentInput) (*domain.Comment, error)
	GetComment(ctx context.Context, id string) (*domain.Comment, error)
	UpdateComment(ctx context.Context, id string, input UpdateCommentInput) (*domain.Comment, error)
	DeleteComment(ctx context.Context, id string) error
	ListComments(ctx context.Context, filters repository.CommentFilters) ([]domain.Comment, int64, error)
	ModerateComment(ctx context.Context, id string, status domain.CommentStatus) (*domain.Comment, error)
}

type CreateCommentInput struct {
	PostID      string
	ParentID    *string
	AuthorName  string
	AuthorEmail string
	Content     string
}

type UpdateCommentInput struct {
	Content string
}

type commentUseCase struct {
	repo          repository.CommentRepository
	postRepo      repository.PostRepository
	kafkaProducer *kafka.Producer
	config        *config.Config
}

func NewCommentUseCase(
	repo repository.CommentRepository,
	postRepo repository.PostRepository,
	kafkaProducer *kafka.Producer,
	cfg *config.Config,
) CommentUseCase {
	return &commentUseCase{
		repo:          repo,
		postRepo:      postRepo,
		kafkaProducer: kafkaProducer,
		config:        cfg,
	}
}

func (uc *commentUseCase) CreateComment(ctx context.Context, input CreateCommentInput) (*domain.Comment, error) {
	// Validate that post exists
	_, err := uc.postRepo.GetByID(ctx, input.PostID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// Validate parent comment exists if provided
	if input.ParentID != nil {
		parent, err := uc.repo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent comment not found: %w", err)
		}

		// Ensure parent belongs to the same post
		if parent.PostID != input.PostID {
			return nil, fmt.Errorf("parent comment does not belong to the same post")
		}
	}

	// Generate UUID for comment
	commentID := uuid.New().String()

	// Create comment entity (default status: pending for moderation)
	comment := &domain.Comment{
		ID:          commentID,
		PostID:      input.PostID,
		ParentID:    input.ParentID,
		AuthorName:  input.AuthorName,
		AuthorEmail: input.AuthorEmail,
		Content:     input.Content,
		Status:      domain.CommentStatusPending,
	}

	// Save to database
	if err := uc.repo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.CommentCreatedEvent{
			CommentID:   comment.ID,
			PostID:      comment.PostID,
			AuthorName:  comment.AuthorName,
			AuthorEmail: comment.AuthorEmail,
			Content:     comment.Content,
			Status:      string(comment.Status),
			CreatedAt:   comment.CreatedAt,
		}
		if err := uc.kafkaProducer.PublishCommentCreated("blog.comment.created", event); err != nil {
			fmt.Printf("Warning: Failed to publish comment created event: %v\n", err)
		}
	}

	return comment, nil
}

func (uc *commentUseCase) GetComment(ctx context.Context, id string) (*domain.Comment, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *commentUseCase) UpdateComment(ctx context.Context, id string, input UpdateCommentInput) (*domain.Comment, error) {
	// Get existing comment
	comment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update content
	comment.Content = input.Content

	// Save to database
	if err := uc.repo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return comment, nil
}

func (uc *commentUseCase) DeleteComment(ctx context.Context, id string) error {
	// Check if comment exists
	comment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if comment has replies
	replies, err := uc.repo.GetReplies(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check replies: %w", err)
	}

	if len(replies) > 0 {
		return fmt.Errorf("cannot delete comment with replies")
	}

	// Delete from database
	if err := uc.repo.Delete(ctx, comment.ID); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.CommentDeletedEvent{
			CommentID: comment.ID,
			PostID:    comment.PostID,
			DeletedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishCommentDeleted("blog.comment.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish comment deleted event: %v\n", err)
		}
	}

	return nil
}

func (uc *commentUseCase) ListComments(ctx context.Context, filters repository.CommentFilters) ([]domain.Comment, int64, error) {
	return uc.repo.List(ctx, filters)
}

func (uc *commentUseCase) ModerateComment(ctx context.Context, id string, status domain.CommentStatus) (*domain.Comment, error) {
	// Get existing comment
	comment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate status
	validStatuses := map[domain.CommentStatus]bool{
		domain.CommentStatusPending:  true,
		domain.CommentStatusApproved: true,
		domain.CommentStatusSpam:     true,
		domain.CommentStatusTrash:    true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid comment status: %s", status)
	}

	// Update status
	comment.Status = status

	// Save to database
	if err := uc.repo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to moderate comment: %w", err)
	}

	// Publish Kafka event based on status
	if uc.kafkaProducer != nil {
		switch status {
		case domain.CommentStatusApproved:
			event := kafka.CommentApprovedEvent{
				CommentID:  comment.ID,
				PostID:     comment.PostID,
				ApprovedAt: time.Now(),
			}
			if err := uc.kafkaProducer.PublishCommentApproved("blog.comment.approved", event); err != nil {
				fmt.Printf("Warning: Failed to publish comment approved event: %v\n", err)
			}
		case domain.CommentStatusSpam, domain.CommentStatusTrash:
			event := kafka.CommentRejectedEvent{
				CommentID:  comment.ID,
				PostID:     comment.PostID,
				Reason:     string(status),
				RejectedAt: time.Now(),
			}
			if err := uc.kafkaProducer.PublishCommentRejected("blog.comment.rejected", event); err != nil {
				fmt.Printf("Warning: Failed to publish comment rejected event: %v\n", err)
			}
		}
	}

	return comment, nil
}
