package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"toxictoast/services/blog-service/internal/domain"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	GetByID(ctx context.Context, id string) (*domain.Comment, error)
	Update(ctx context.Context, comment *domain.Comment) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters CommentFilters) ([]domain.Comment, int64, error)
	GetReplies(ctx context.Context, parentID string) ([]domain.Comment, error)
}

type CommentFilters struct {
	Page     int
	PageSize int
	PostID   *string
	Status   *domain.CommentStatus
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) GetByID(ctx context.Context, id string) (*domain.Comment, error) {
	var comment domain.Comment
	err := r.db.WithContext(ctx).
		Preload("Post").
		Preload("Parent").
		Preload("Replies").
		First(&comment, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, err
	}

	return &comment, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

func (r *commentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Comment{}, "id = ?", id).Error
}

func (r *commentRepository) List(ctx context.Context, filters CommentFilters) ([]domain.Comment, int64, error) {
	var comments []domain.Comment
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Comment{})

	// Filter by post
	if filters.PostID != nil {
		query = query.Where("post_id = ?", *filters.PostID)
	}

	// Filter by status
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	// Only show top-level comments (not replies)
	query = query.Where("parent_id IS NULL")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	if filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query = query.Offset(offset).Limit(filters.PageSize)
	}

	// Preload replies
	query = query.Preload("Replies")

	// Order by creation date (newest first)
	query = query.Order("created_at DESC")

	// Execute query
	if err := query.Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *commentRepository) GetReplies(ctx context.Context, parentID string) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("created_at ASC").
		Find(&comments).Error

	return comments, err
}
