package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
)

// ============================================================================
// Mock Repositories
// ============================================================================

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, post *domain.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) Update(ctx context.Context, post *domain.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}

func (m *MockPostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}

func (m *MockPostRepository) List(ctx context.Context, filters repository.PostFilters) ([]domain.Post, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]domain.Post), args.Get(1).(int64), args.Error(2)
}

func (m *MockPostRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostRepository) IncrementViewCount(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepository) List(ctx context.Context, filters repository.CategoryFilters) ([]domain.Category, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]domain.Category), args.Get(1).(int64), args.Error(2)
}

func (m *MockCategoryRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

func (m *MockCategoryRepository) GetChildren(ctx context.Context, parentID string) ([]domain.Category, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Category), args.Error(1)
}

type MockTagRepository struct {
	mock.Mock
}

func (m *MockTagRepository) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tag), args.Error(1)
}

func (m *MockTagRepository) GetByIDs(ctx context.Context, ids []string) ([]domain.Tag, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Tag), args.Error(1)
}

func (m *MockTagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *MockTagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *MockTagRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTagRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tag, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tag), args.Error(1)
}

func (m *MockTagRepository) List(ctx context.Context, filters repository.TagFilters) ([]domain.Tag, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]domain.Tag), args.Get(1).(int64), args.Error(2)
}

func (m *MockTagRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

// ============================================================================
// Command Validation Tests
// ============================================================================

func TestCreatePostCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *CreatePostCommand
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: &CreatePostCommand{
				Title:    "Test Post",
				Content:  "Test Content",
				AuthorID: "author-123",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			cmd: &CreatePostCommand{
				Title:    "",
				Content:  "Test Content",
				AuthorID: "author-123",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			cmd: &CreatePostCommand{
				Title:    "Test Post",
				Content:  "",
				AuthorID: "author-123",
			},
			wantErr: true,
		},
		{
			name: "missing author_id",
			cmd: &CreatePostCommand{
				Title:    "Test Post",
				Content:  "Test Content",
				AuthorID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdatePostCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *UpdatePostCommand
		wantErr bool
	}{
		{
			name: "valid command with aggregate id",
			cmd: &UpdatePostCommand{
				BaseCommand: cqrs.BaseCommand{AggregateID: "post-123"},
			},
			wantErr: false,
		},
		{
			name: "missing aggregate id",
			cmd: &UpdatePostCommand{
				BaseCommand: cqrs.BaseCommand{AggregateID: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeletePostCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *DeletePostCommand
		wantErr bool
	}{
		{
			name: "valid command with aggregate id",
			cmd: &DeletePostCommand{
				BaseCommand: cqrs.BaseCommand{AggregateID: "post-123"},
			},
			wantErr: false,
		},
		{
			name: "missing aggregate id",
			cmd: &DeletePostCommand{
				BaseCommand: cqrs.BaseCommand{AggregateID: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPublishPostCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *PublishPostCommand
		wantErr bool
	}{
		{
			name: "valid command with aggregate id",
			cmd: &PublishPostCommand{
				BaseCommand: cqrs.BaseCommand{AggregateID: "post-123"},
			},
			wantErr: false,
		},
		{
			name: "missing aggregate id",
			cmd: &PublishPostCommand{
				BaseCommand: cqrs.BaseCommand{AggregateID: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Command Handler Tests
// ============================================================================

func TestCreatePostHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful post creation", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockTagRepo := new(MockTagRepository)

		mockPostRepo.On("SlugExists", ctx, mock.AnythingOfType("string")).Return(false, nil)
		mockPostRepo.On("Create", ctx, mock.AnythingOfType("*domain.Post")).Return(nil)

		handler := NewCreatePostHandler(mockPostRepo, mockCategoryRepo, mockTagRepo, nil)

		cmd := &CreatePostCommand{
			Title:    "Test Post",
			Content:  "Test Content",
			AuthorID: "author-123",
			SEO:      SEOData{},
		}

		err := handler.Handle(ctx, cmd)

		assert.NoError(t, err)
		assert.NotEmpty(t, cmd.AggregateID)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("repository error during creation", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockTagRepo := new(MockTagRepository)

		mockPostRepo.On("SlugExists", ctx, mock.AnythingOfType("string")).Return(false, nil)
		mockPostRepo.On("Create", ctx, mock.AnythingOfType("*domain.Post")).Return(errors.New("database error"))

		handler := NewCreatePostHandler(mockPostRepo, mockCategoryRepo, mockTagRepo, nil)

		cmd := &CreatePostCommand{
			Title:    "Test Post",
			Content:  "Test Content",
			AuthorID: "author-123",
			SEO:      SEOData{},
		}

		err := handler.Handle(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create post")
	})

	t.Run("create post with categories and tags", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockTagRepo := new(MockTagRepository)

		mockPostRepo.On("SlugExists", ctx, mock.AnythingOfType("string")).Return(false, nil)
		mockPostRepo.On("Create", ctx, mock.AnythingOfType("*domain.Post")).Return(nil)

		mockCategoryRepo.On("GetByID", ctx, "cat-1").Return(&domain.Category{ID: "cat-1", Name: "Category 1"}, nil)
		mockTagRepo.On("GetByIDs", ctx, []string{"tag-1", "tag-2"}).Return([]domain.Tag{
			{ID: "tag-1", Name: "Tag 1"},
			{ID: "tag-2", Name: "Tag 2"},
		}, nil)

		handler := NewCreatePostHandler(mockPostRepo, mockCategoryRepo, mockTagRepo, nil)

		cmd := &CreatePostCommand{
			Title:       "Test Post",
			Content:     "Test Content",
			AuthorID:    "author-123",
			CategoryIDs: []string{"cat-1"},
			TagIDs:      []string{"tag-1", "tag-2"},
			SEO:         SEOData{},
		}

		err := handler.Handle(ctx, cmd)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestDeletePostHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful post deletion", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)

		existingPost := &domain.Post{
			ID:    "post-123",
			Title: "Test Post",
		}

		mockPostRepo.On("GetByID", ctx, "post-123").Return(existingPost, nil)
		mockPostRepo.On("Delete", ctx, "post-123").Return(nil)

		handler := NewDeletePostHandler(mockPostRepo, nil)

		cmd := &DeletePostCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: "post-123"},
		}

		err := handler.Handle(ctx, cmd)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("post not found", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)

		mockPostRepo.On("GetByID", ctx, "non-existent").Return(nil, errors.New("not found"))

		handler := NewDeletePostHandler(mockPostRepo, nil)

		cmd := &DeletePostCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: "non-existent"},
		}

		err := handler.Handle(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")
	})
}

func TestPublishPostHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful post publishing", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)

		existingPost := &domain.Post{
			ID:     "post-123",
			Title:  "Test Post",
			Status: domain.PostStatusDraft,
		}

		mockPostRepo.On("GetByID", ctx, "post-123").Return(existingPost, nil)
		mockPostRepo.On("Update", ctx, mock.MatchedBy(func(post *domain.Post) bool {
			return post.ID == "post-123" && post.Status == domain.PostStatusPublished && post.PublishedAt != nil
		})).Return(nil)

		handler := NewPublishPostHandler(mockPostRepo, nil)

		cmd := &PublishPostCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: "post-123"},
		}

		err := handler.Handle(ctx, cmd)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("already published post", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)

		now := time.Now()
		existingPost := &domain.Post{
			ID:          "post-123",
			Title:       "Test Post",
			Status:      domain.PostStatusPublished,
			PublishedAt: &now,
		}

		mockPostRepo.On("GetByID", ctx, "post-123").Return(existingPost, nil)

		handler := NewPublishPostHandler(mockPostRepo, nil)

		cmd := &PublishPostCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: "post-123"},
		}

		err := handler.Handle(ctx, cmd)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
		mockPostRepo.AssertNotCalled(t, "Update")
	})
}

func TestIncrementPostViewCountHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful view count increment", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)

		mockPostRepo.On("IncrementViewCount", ctx, "post-123").Return(nil)

		handler := NewIncrementPostViewCountHandler(mockPostRepo)

		cmd := &IncrementPostViewCountCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: "post-123"},
		}

		err := handler.Handle(ctx, cmd)

		assert.NoError(t, err)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("repository error during increment", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)

		mockPostRepo.On("IncrementViewCount", ctx, "post-123").Return(errors.New("database error"))

		handler := NewIncrementPostViewCountHandler(mockPostRepo)

		cmd := &IncrementPostViewCountCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: "post-123"},
		}

		err := handler.Handle(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to increment view count")
	})
}

func TestPublishScheduledPostHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful scheduled post publishing", func(t *testing.T) {
		mockPostRepo := new(MockPostRepository)

		now := time.Now()
		scheduledPost := &domain.Post{
			ID:          "post-123",
			Title:       "Scheduled Post",
			Status:      domain.PostStatusDraft,
			PublishedAt: &now,
		}

		mockPostRepo.On("Update", ctx, mock.MatchedBy(func(post *domain.Post) bool {
			return post.ID == "post-123" && post.Status == domain.PostStatusPublished
		})).Return(nil)

		handler := NewPublishScheduledPostHandler(mockPostRepo, nil)

		cmd := &PublishScheduledPostCommand{
			Post: scheduledPost,
		}

		err := handler.Handle(ctx, cmd)

		assert.NoError(t, err)
		assert.Equal(t, domain.PostStatusPublished, scheduledPost.Status)
		mockPostRepo.AssertExpectations(t)
	})
}
