package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"toxictoast/services/blog-service/internal/domain"
)

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, id string) (*domain.Post, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Post, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters PostFilters) ([]domain.Post, int64, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	IncrementViewCount(ctx context.Context, id string) error
}

type PostFilters struct {
	Page       int
	PageSize   int
	CategoryID *string
	TagID      *string
	AuthorID   *string
	Status     *domain.PostStatus
	Featured   *bool
	Search     *string
	SortBy     string
	SortOrder  string
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepository) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	var post domain.Post
	err := r.db.WithContext(ctx).
		Preload("Categories").
		Preload("Tags").
		First(&post, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}

	return &post, nil
}

func (r *postRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	var post domain.Post
	err := r.db.WithContext(ctx).
		Preload("Categories").
		Preload("Tags").
		First(&post, "slug = ?", slug).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}

	return &post, nil
}

func (r *postRepository) Update(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *postRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Post{}, "id = ?", id).Error
}

func (r *postRepository) List(ctx context.Context, filters PostFilters) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Post{})

	// Apply filters
	if filters.CategoryID != nil {
		query = query.Joins("JOIN post_categories ON post_categories.post_id = posts.id").
			Where("post_categories.category_id = ?", *filters.CategoryID)
	}

	if filters.TagID != nil {
		query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Where("post_tags.tag_id = ?", *filters.TagID)
	}

	if filters.AuthorID != nil {
		query = query.Where("author_id = ?", *filters.AuthorID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.Featured != nil {
		query = query.Where("featured = ?", *filters.Featured)
	}

	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + *filters.Search + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR excerpt ILIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}

	sortOrder := "DESC"
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Pagination
	if filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query = query.Offset(offset).Limit(filters.PageSize)
	}

	// Preload associations
	query = query.Preload("Categories").Preload("Tags")

	// Execute query
	if err := query.Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *postRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Post{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *postRepository) IncrementViewCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Post{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
