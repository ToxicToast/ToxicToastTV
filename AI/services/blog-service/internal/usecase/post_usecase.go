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
	"toxictoast/services/blog-service/pkg/utils"
)

type PostUseCase interface {
	CreatePost(ctx context.Context, input CreatePostInput) (*domain.Post, error)
	GetPost(ctx context.Context, id string) (*domain.Post, error)
	GetPostBySlug(ctx context.Context, slug string) (*domain.Post, error)
	UpdatePost(ctx context.Context, id string, input UpdatePostInput) (*domain.Post, error)
	DeletePost(ctx context.Context, id string) error
	ListPosts(ctx context.Context, filters repository.PostFilters) ([]domain.Post, int64, error)
	PublishPost(ctx context.Context, id string) (*domain.Post, error)
	IncrementViewCount(ctx context.Context, id string) error
	PublishScheduledPost(ctx context.Context, post *domain.Post) error
}

type CreatePostInput struct {
	Title           string
	Content         string
	Excerpt         string
	CategoryIDs     []string
	TagIDs          []string
	FeaturedImageID *string
	Featured        bool
	AuthorID        string
	SEO             SEOInput
}

type UpdatePostInput struct {
	Title           *string
	Content         *string
	Excerpt         *string
	CategoryIDs     []string
	TagIDs          []string
	FeaturedImageID *string
	Featured        *bool
	SEO             *SEOInput
}

type SEOInput struct {
	MetaTitle       string
	MetaDescription string
	MetaKeywords    []string
	OGTitle         string
	OGDescription   string
	OGImage         string
	CanonicalURL    string
}

type postUseCase struct {
	postRepo      repository.PostRepository
	categoryRepo  repository.CategoryRepository
	tagRepo       repository.TagRepository
	kafkaProducer *kafka.Producer
	config        *config.Config
}

func NewPostUseCase(
	postRepo repository.PostRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
	kafkaProducer *kafka.Producer,
	cfg *config.Config,
) PostUseCase {
	return &postUseCase{
		postRepo:      postRepo,
		categoryRepo:  categoryRepo,
		tagRepo:       tagRepo,
		kafkaProducer: kafkaProducer,
		config:        cfg,
	}
}

func (uc *postUseCase) CreatePost(ctx context.Context, input CreatePostInput) (*domain.Post, error) {
	// Generate slug from title
	slug := uc.generateUniqueSlug(ctx, input.Title)

	// Process markdown to HTML
	html := utils.MarkdownToHTML(input.Content)

	// Generate excerpt if not provided
	excerpt := input.Excerpt
	if excerpt == "" {
		excerpt = utils.GenerateExcerpt(html, 200)
	}

	// Calculate reading time
	readingTime := utils.CalculateReadingTime(input.Content)

	// Generate UUID for post
	postID := uuid.New().String()

	// Create post entity
	post := &domain.Post{
		ID:              postID,
		Title:           input.Title,
		Slug:            slug,
		Content:         input.Content,
		Excerpt:         excerpt,
		Markdown:        input.Content,
		HTML:            html,
		Status:          domain.PostStatusDraft,
		Featured:        input.Featured,
		AuthorID:        input.AuthorID,
		FeaturedImageID: input.FeaturedImageID,
		ReadingTime:     readingTime,
		ViewCount:       0,
		MetaTitle:       input.SEO.MetaTitle,
		MetaDescription: input.SEO.MetaDescription,
		MetaKeywords:    joinStrings(input.SEO.MetaKeywords, ","),
		OGTitle:         input.SEO.OGTitle,
		OGDescription:   input.SEO.OGDescription,
		OGImage:         input.SEO.OGImage,
		CanonicalURL:    input.SEO.CanonicalURL,
	}

	// Load categories if provided
	if len(input.CategoryIDs) > 0 {
		categories, err := uc.getCategoriesByIDs(ctx, input.CategoryIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load categories: %w", err)
		}
		post.Categories = categories
	}

	// Load tags if provided
	if len(input.TagIDs) > 0 {
		tags, err := uc.tagRepo.GetByIDs(ctx, input.TagIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load tags: %w", err)
		}
		post.Tags = tags
	}

	// Save to database
	if err := uc.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.PostCreatedEvent{
			PostID:    post.ID,
			Title:     post.Title,
			Slug:      post.Slug,
			AuthorID:  post.AuthorID,
			Status:    string(post.Status),
			CreatedAt: post.CreatedAt,
		}
		if err := uc.kafkaProducer.PublishPostCreated("blog.post.created", event); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: Failed to publish post created event: %v\n", err)
		}
	}

	return post, nil
}

func (uc *postUseCase) GetPost(ctx context.Context, id string) (*domain.Post, error) {
	return uc.postRepo.GetByID(ctx, id)
}

func (uc *postUseCase) GetPostBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	return uc.postRepo.GetBySlug(ctx, slug)
}

func (uc *postUseCase) UpdatePost(ctx context.Context, id string, input UpdatePostInput) (*domain.Post, error) {
	// Get existing post
	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Title != nil {
		post.Title = *input.Title
		// Regenerate slug if title changed
		post.Slug = uc.generateUniqueSlug(ctx, post.Title)
	}

	if input.Content != nil {
		post.Content = *input.Content
		post.Markdown = *input.Content
		post.HTML = utils.MarkdownToHTML(*input.Content)
		post.ReadingTime = utils.CalculateReadingTime(*input.Content)
	}

	if input.Excerpt != nil {
		post.Excerpt = *input.Excerpt
	}

	if input.Featured != nil {
		post.Featured = *input.Featured
	}

	if input.FeaturedImageID != nil {
		post.FeaturedImageID = input.FeaturedImageID
	}

	// Update SEO if provided
	if input.SEO != nil {
		post.MetaTitle = input.SEO.MetaTitle
		post.MetaDescription = input.SEO.MetaDescription
		post.MetaKeywords = joinStrings(input.SEO.MetaKeywords, ",")
		post.OGTitle = input.SEO.OGTitle
		post.OGDescription = input.SEO.OGDescription
		post.OGImage = input.SEO.OGImage
		post.CanonicalURL = input.SEO.CanonicalURL
	}

	// Update categories if provided
	if len(input.CategoryIDs) > 0 {
		categories, err := uc.getCategoriesByIDs(ctx, input.CategoryIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load categories: %w", err)
		}
		post.Categories = categories
	}

	// Update tags if provided
	if len(input.TagIDs) > 0 {
		tags, err := uc.tagRepo.GetByIDs(ctx, input.TagIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load tags: %w", err)
		}
		post.Tags = tags
	}

	// Save to database
	if err := uc.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.PostUpdatedEvent{
			PostID:    post.ID,
			Title:     post.Title,
			Slug:      post.Slug,
			UpdatedAt: post.UpdatedAt,
		}
		if err := uc.kafkaProducer.PublishPostUpdated("blog.post.updated", event); err != nil {
			fmt.Printf("Warning: Failed to publish post updated event: %v\n", err)
		}
	}

	return post, nil
}

func (uc *postUseCase) DeletePost(ctx context.Context, id string) error {
	// Check if post exists
	_, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := uc.postRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.PostDeletedEvent{
			PostID:    id,
			DeletedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishPostDeleted("blog.post.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish post deleted event: %v\n", err)
		}
	}

	return nil
}

func (uc *postUseCase) ListPosts(ctx context.Context, filters repository.PostFilters) ([]domain.Post, int64, error) {
	return uc.postRepo.List(ctx, filters)
}

func (uc *postUseCase) PublishPost(ctx context.Context, id string) (*domain.Post, error) {
	// Get post
	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if already published
	if post.Status == domain.PostStatusPublished {
		return post, nil
	}

	// Update status and published date
	post.Status = domain.PostStatusPublished
	now := time.Now()
	post.PublishedAt = &now

	// Save to database
	if err := uc.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to publish post: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.PostPublishedEvent{
			PostID:      post.ID,
			Title:       post.Title,
			Slug:        post.Slug,
			AuthorID:    post.AuthorID,
			PublishedAt: *post.PublishedAt,
		}
		if err := uc.kafkaProducer.PublishPostPublished("blog.post.published", event); err != nil {
			fmt.Printf("Warning: Failed to publish post published event: %v\n", err)
		}
	}

	return post, nil
}

func (uc *postUseCase) IncrementViewCount(ctx context.Context, id string) error {
	return uc.postRepo.IncrementViewCount(ctx, id)
}

// PublishScheduledPost publishes a post that was scheduled for publishing
func (uc *postUseCase) PublishScheduledPost(ctx context.Context, post *domain.Post) error {
	// Change status to published
	post.Status = domain.PostStatusPublished
	post.UpdatedAt = time.Now()

	// Update in database
	if err := uc.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to publish scheduled post: %w", err)
	}

	// Publish Kafka event
	if uc.kafkaProducer != nil {
		event := kafka.BlogPostScheduledPublishedEvent{
			PostID:      post.ID,
			Title:       post.Title,
			Slug:        post.Slug,
			ScheduledAt: post.PublishedAt,
			PublishedAt: time.Now(),
		}
		if err := uc.kafkaProducer.PublishBlogPostScheduledPublished("blog.post.scheduled.published", event); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to publish scheduled post event: %v\n", err)
		}
	}

	return nil
}

// Helper methods

func (uc *postUseCase) generateUniqueSlug(ctx context.Context, title string) string {
	exists := func(slug string) bool {
		exists, err := uc.postRepo.SlugExists(ctx, slug)
		if err != nil {
			return false
		}
		return exists
	}

	return utils.GenerateUniqueSlug(title, exists)
}

func (uc *postUseCase) getCategoriesByIDs(ctx context.Context, ids []string) ([]domain.Category, error) {
	categories := []domain.Category{}
	for _, id := range ids {
		category, err := uc.categoryRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *category)
	}
	return categories, nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
