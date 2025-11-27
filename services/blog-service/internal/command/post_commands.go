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
	"toxictoast/services/blog-service/pkg/utils"
)

// ============================================================================
// Commands
// ============================================================================

// CreatePostCommand creates a new blog post
type CreatePostCommand struct {
	cqrs.BaseCommand
	Title           string   `json:"title"`
	Content         string   `json:"content"`
	Excerpt         string   `json:"excerpt"`
	CategoryIDs     []string `json:"category_ids"`
	TagIDs          []string `json:"tag_ids"`
	FeaturedImageID *string  `json:"featured_image_id"`
	Featured        bool     `json:"featured"`
	AuthorID        string   `json:"author_id"`
	SEO             SEOData  `json:"seo"`
}

func (c *CreatePostCommand) CommandName() string {
	return "create_post"
}

func (c *CreatePostCommand) Validate() error {
	if c.Title == "" {
		return errors.New("title is required")
	}
	if c.Content == "" {
		return errors.New("content is required")
	}
	if c.AuthorID == "" {
		return errors.New("author_id is required")
	}
	return nil
}

// UpdatePostCommand updates an existing post
type UpdatePostCommand struct {
	cqrs.BaseCommand
	Title           *string  `json:"title,omitempty"`
	Content         *string  `json:"content,omitempty"`
	Excerpt         *string  `json:"excerpt,omitempty"`
	CategoryIDs     []string `json:"category_ids,omitempty"`
	TagIDs          []string `json:"tag_ids,omitempty"`
	FeaturedImageID *string  `json:"featured_image_id,omitempty"`
	Featured        *bool    `json:"featured,omitempty"`
	SEO             *SEOData `json:"seo,omitempty"`
}

func (c *UpdatePostCommand) CommandName() string {
	return "update_post"
}

func (c *UpdatePostCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("post_id is required")
	}
	return nil
}

// DeletePostCommand deletes a post
type DeletePostCommand struct {
	cqrs.BaseCommand
}

func (c *DeletePostCommand) CommandName() string {
	return "delete_post"
}

func (c *DeletePostCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("post_id is required")
	}
	return nil
}

// PublishPostCommand publishes a draft post
type PublishPostCommand struct {
	cqrs.BaseCommand
}

func (c *PublishPostCommand) CommandName() string {
	return "publish_post"
}

func (c *PublishPostCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("post_id is required")
	}
	return nil
}

// IncrementPostViewCountCommand increments the view count of a post
type IncrementPostViewCountCommand struct {
	cqrs.BaseCommand
}

func (c *IncrementPostViewCountCommand) CommandName() string {
	return "increment_post_view_count"
}

func (c *IncrementPostViewCountCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("post_id is required")
	}
	return nil
}

// PublishScheduledPostCommand publishes a scheduled post
type PublishScheduledPostCommand struct {
	cqrs.BaseCommand
	Post *domain.Post `json:"-"` // Populated by scheduler
}

func (c *PublishScheduledPostCommand) CommandName() string {
	return "publish_scheduled_post"
}

func (c *PublishScheduledPostCommand) Validate() error {
	if c.Post == nil {
		return errors.New("post is required")
	}
	return nil
}

// SEOData contains SEO metadata
type SEOData struct {
	MetaTitle       string   `json:"meta_title"`
	MetaDescription string   `json:"meta_description"`
	MetaKeywords    []string `json:"meta_keywords"`
	OGTitle         string   `json:"og_title"`
	OGDescription   string   `json:"og_description"`
	OGImage         string   `json:"og_image"`
	CanonicalURL    string   `json:"canonical_url"`
}

// ============================================================================
// Command Handlers
// ============================================================================

// CreatePostHandler handles post creation
type CreatePostHandler struct {
	postRepo      repository.PostRepository
	categoryRepo  repository.CategoryRepository
	tagRepo       repository.TagRepository
	kafkaProducer *kafka.Producer
}

func NewCreatePostHandler(
	postRepo repository.PostRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
	kafkaProducer *kafka.Producer,
) *CreatePostHandler {
	return &CreatePostHandler{
		postRepo:      postRepo,
		categoryRepo:  categoryRepo,
		tagRepo:       tagRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *CreatePostHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreatePostCommand)

	// Generate unique slug from title
	slug := h.generateUniqueSlug(ctx, createCmd.Title)

	// Process markdown to HTML
	html := utils.MarkdownToHTML(createCmd.Content)

	// Generate excerpt if not provided
	excerpt := createCmd.Excerpt
	if excerpt == "" {
		excerpt = utils.GenerateExcerpt(html, 200)
	}

	// Calculate reading time
	readingTime := utils.CalculateReadingTime(createCmd.Content)

	// Generate UUID for post
	postID := uuid.New().String()

	// Create post entity
	post := &domain.Post{
		ID:              postID,
		Title:           createCmd.Title,
		Slug:            slug,
		Content:         createCmd.Content,
		Excerpt:         excerpt,
		Markdown:        createCmd.Content,
		HTML:            html,
		Status:          domain.PostStatusDraft,
		Featured:        createCmd.Featured,
		AuthorID:        createCmd.AuthorID,
		FeaturedImageID: createCmd.FeaturedImageID,
		ReadingTime:     readingTime,
		ViewCount:       0,
		MetaTitle:       createCmd.SEO.MetaTitle,
		MetaDescription: createCmd.SEO.MetaDescription,
		MetaKeywords:    joinStrings(createCmd.SEO.MetaKeywords, ","),
		OGTitle:         createCmd.SEO.OGTitle,
		OGDescription:   createCmd.SEO.OGDescription,
		OGImage:         createCmd.SEO.OGImage,
		CanonicalURL:    createCmd.SEO.CanonicalURL,
	}

	// Load categories if provided
	if len(createCmd.CategoryIDs) > 0 {
		categories, err := h.getCategoriesByIDs(ctx, createCmd.CategoryIDs)
		if err != nil {
			return fmt.Errorf("failed to load categories: %w", err)
		}
		post.Categories = categories
	}

	// Load tags if provided
	if len(createCmd.TagIDs) > 0 {
		tags, err := h.tagRepo.GetByIDs(ctx, createCmd.TagIDs)
		if err != nil {
			return fmt.Errorf("failed to load tags: %w", err)
		}
		post.Tags = tags
	}

	// Save to database
	if err := h.postRepo.Create(ctx, post); err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	// Store post ID in command result
	createCmd.AggregateID = postID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.PostCreatedEvent{
			PostID:    post.ID,
			Title:     post.Title,
			Slug:      post.Slug,
			AuthorID:  post.AuthorID,
			Status:    string(post.Status),
			CreatedAt: post.CreatedAt,
		}
		if err := h.kafkaProducer.PublishPostCreated("blog.post.created", event); err != nil {
			fmt.Printf("Warning: Failed to publish post created event: %v\n", err)
		}
	}

	return nil
}

func (h *CreatePostHandler) generateUniqueSlug(ctx context.Context, title string) string {
	exists := func(slug string) bool {
		exists, err := h.postRepo.SlugExists(ctx, slug)
		if err != nil {
			return false
		}
		return exists
	}

	return utils.GenerateUniqueSlug(title, exists)
}

func (h *CreatePostHandler) getCategoriesByIDs(ctx context.Context, ids []string) ([]domain.Category, error) {
	categories := []domain.Category{}
	for _, id := range ids {
		category, err := h.categoryRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *category)
	}
	return categories, nil
}

// UpdatePostHandler handles post updates
type UpdatePostHandler struct {
	postRepo      repository.PostRepository
	categoryRepo  repository.CategoryRepository
	tagRepo       repository.TagRepository
	kafkaProducer *kafka.Producer
}

func NewUpdatePostHandler(
	postRepo repository.PostRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
	kafkaProducer *kafka.Producer,
) *UpdatePostHandler {
	return &UpdatePostHandler{
		postRepo:      postRepo,
		categoryRepo:  categoryRepo,
		tagRepo:       tagRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *UpdatePostHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdatePostCommand)

	// Get existing post
	post, err := h.postRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	// Update fields if provided
	if updateCmd.Title != nil {
		post.Title = *updateCmd.Title
		// Regenerate slug if title changed
		post.Slug = h.generateUniqueSlug(ctx, post.Title)
	}

	if updateCmd.Content != nil {
		post.Content = *updateCmd.Content
		post.Markdown = *updateCmd.Content
		post.HTML = utils.MarkdownToHTML(*updateCmd.Content)
		post.ReadingTime = utils.CalculateReadingTime(*updateCmd.Content)
	}

	if updateCmd.Excerpt != nil {
		post.Excerpt = *updateCmd.Excerpt
	}

	if updateCmd.Featured != nil {
		post.Featured = *updateCmd.Featured
	}

	if updateCmd.FeaturedImageID != nil {
		post.FeaturedImageID = updateCmd.FeaturedImageID
	}

	// Update SEO if provided
	if updateCmd.SEO != nil {
		post.MetaTitle = updateCmd.SEO.MetaTitle
		post.MetaDescription = updateCmd.SEO.MetaDescription
		post.MetaKeywords = joinStrings(updateCmd.SEO.MetaKeywords, ",")
		post.OGTitle = updateCmd.SEO.OGTitle
		post.OGDescription = updateCmd.SEO.OGDescription
		post.OGImage = updateCmd.SEO.OGImage
		post.CanonicalURL = updateCmd.SEO.CanonicalURL
	}

	// Update categories if provided
	if len(updateCmd.CategoryIDs) > 0 {
		categories, err := h.getCategoriesByIDs(ctx, updateCmd.CategoryIDs)
		if err != nil {
			return fmt.Errorf("failed to load categories: %w", err)
		}
		post.Categories = categories
	}

	// Update tags if provided
	if len(updateCmd.TagIDs) > 0 {
		tags, err := h.tagRepo.GetByIDs(ctx, updateCmd.TagIDs)
		if err != nil {
			return fmt.Errorf("failed to load tags: %w", err)
		}
		post.Tags = tags
	}

	// Save to database
	if err := h.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.PostUpdatedEvent{
			PostID:    post.ID,
			Title:     post.Title,
			Slug:      post.Slug,
			UpdatedAt: post.UpdatedAt,
		}
		if err := h.kafkaProducer.PublishPostUpdated("blog.post.updated", event); err != nil {
			fmt.Printf("Warning: Failed to publish post updated event: %v\n", err)
		}
	}

	return nil
}

func (h *UpdatePostHandler) generateUniqueSlug(ctx context.Context, title string) string {
	exists := func(slug string) bool {
		exists, err := h.postRepo.SlugExists(ctx, slug)
		if err != nil {
			return false
		}
		return exists
	}

	return utils.GenerateUniqueSlug(title, exists)
}

func (h *UpdatePostHandler) getCategoriesByIDs(ctx context.Context, ids []string) ([]domain.Category, error) {
	categories := []domain.Category{}
	for _, id := range ids {
		category, err := h.categoryRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *category)
	}
	return categories, nil
}

// DeletePostHandler handles post deletion
type DeletePostHandler struct {
	postRepo      repository.PostRepository
	kafkaProducer *kafka.Producer
}

func NewDeletePostHandler(
	postRepo repository.PostRepository,
	kafkaProducer *kafka.Producer,
) *DeletePostHandler {
	return &DeletePostHandler{
		postRepo:      postRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeletePostHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeletePostCommand)

	// Check if post exists
	_, err := h.postRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	// Delete from database (soft delete)
	if err := h.postRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.PostDeletedEvent{
			PostID:    deleteCmd.AggregateID,
			DeletedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishPostDeleted("blog.post.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish post deleted event: %v\n", err)
		}
	}

	return nil
}

// PublishPostHandler handles post publishing
type PublishPostHandler struct {
	postRepo      repository.PostRepository
	kafkaProducer *kafka.Producer
}

func NewPublishPostHandler(
	postRepo repository.PostRepository,
	kafkaProducer *kafka.Producer,
) *PublishPostHandler {
	return &PublishPostHandler{
		postRepo:      postRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *PublishPostHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	publishCmd := cmd.(*PublishPostCommand)

	// Get post
	post, err := h.postRepo.GetByID(ctx, publishCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	// Check if already published
	if post.Status == domain.PostStatusPublished {
		return nil
	}

	// Update status and published date
	post.Status = domain.PostStatusPublished
	now := time.Now()
	post.PublishedAt = &now

	// Save to database
	if err := h.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to publish post: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.PostPublishedEvent{
			PostID:      post.ID,
			Title:       post.Title,
			Slug:        post.Slug,
			AuthorID:    post.AuthorID,
			PublishedAt: *post.PublishedAt,
		}
		if err := h.kafkaProducer.PublishPostPublished("blog.post.published", event); err != nil {
			fmt.Printf("Warning: Failed to publish post published event: %v\n", err)
		}
	}

	return nil
}

// IncrementPostViewCountHandler handles view count increment
type IncrementPostViewCountHandler struct {
	postRepo repository.PostRepository
}

func NewIncrementPostViewCountHandler(
	postRepo repository.PostRepository,
) *IncrementPostViewCountHandler {
	return &IncrementPostViewCountHandler{
		postRepo: postRepo,
	}
}

func (h *IncrementPostViewCountHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	incrementCmd := cmd.(*IncrementPostViewCountCommand)

	// Atomic increment in database
	if err := h.postRepo.IncrementViewCount(ctx, incrementCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

// PublishScheduledPostHandler handles scheduled post publishing
type PublishScheduledPostHandler struct {
	postRepo      repository.PostRepository
	kafkaProducer *kafka.Producer
}

func NewPublishScheduledPostHandler(
	postRepo repository.PostRepository,
	kafkaProducer *kafka.Producer,
) *PublishScheduledPostHandler {
	return &PublishScheduledPostHandler{
		postRepo:      postRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *PublishScheduledPostHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	scheduledCmd := cmd.(*PublishScheduledPostCommand)

	// Change status to published
	scheduledCmd.Post.Status = domain.PostStatusPublished
	scheduledCmd.Post.UpdatedAt = time.Now()

	// Update in database
	if err := h.postRepo.Update(ctx, scheduledCmd.Post); err != nil {
		return fmt.Errorf("failed to publish scheduled post: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.BlogPostScheduledPublishedEvent{
			PostID:      scheduledCmd.Post.ID,
			Title:       scheduledCmd.Post.Title,
			Slug:        scheduledCmd.Post.Slug,
			ScheduledAt: scheduledCmd.Post.PublishedAt,
			PublishedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishBlogPostScheduledPublished("blog.post.scheduled.published", event); err != nil {
			fmt.Printf("Warning: Failed to publish scheduled post event: %v\n", err)
		}
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

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
