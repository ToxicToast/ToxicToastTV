package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository/entity"
	"toxictoast/services/blog-service/internal/repository/mapper"
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
	e := mapper.PostToEntity(post)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *postRepository) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	var e entity.PostEntity
	err := r.db.WithContext(ctx).
		Preload("Categories").
		Preload("Tags").
		First(&e, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}

	return mapper.PostToDomain(&e), nil
}

func (r *postRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	var e entity.PostEntity
	err := r.db.WithContext(ctx).
		Preload("Categories").
		Preload("Tags").
		First(&e, "slug = ?", slug).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}

	return mapper.PostToDomain(&e), nil
}

func (r *postRepository) Update(ctx context.Context, post *domain.Post) error {
	e := mapper.PostToEntity(post)
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *postRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.PostEntity{}, "id = ?", id).Error
}

func (r *postRepository) List(ctx context.Context, filters PostFilters) ([]domain.Post, int64, error) {
	var entities []entity.PostEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.PostEntity{})

	// Apply filters
	if filters.CategoryID != nil {
		query = query.Joins("JOIN blog_post_categories ON blog_post_categories.post_entity_id = blog_posts.id").
			Where("blog_post_categories.category_entity_id = ?", *filters.CategoryID)
	}

	if filters.TagID != nil {
		query = query.Joins("JOIN blog_post_tags ON blog_post_tags.post_entity_id = blog_posts.id").
			Where("blog_post_tags.tag_entity_id = ?", *filters.TagID)
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
	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	posts := make([]domain.Post, 0, len(entities))
	for _, e := range entities {
		posts = append(posts, *mapper.PostToDomain(&e))
	}

	return posts, total, nil
}

func (r *postRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.PostEntity{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *postRepository) IncrementViewCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&entity.PostEntity{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
