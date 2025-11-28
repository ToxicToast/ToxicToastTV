package query

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
)

// ============================================================================
// Mock Repository
// ============================================================================

type MockPostRepository struct {
	mock.Mock
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

func (m *MockPostRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostRepository) IncrementViewCount(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ============================================================================
// Query Validation Tests
// ============================================================================

func TestGetPostByIDQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   *GetPostByIDQuery
		wantErr bool
	}{
		{
			name: "valid query",
			query: &GetPostByIDQuery{
				PostID: "post-123",
			},
			wantErr: false,
		},
		{
			name: "missing post_id",
			query: &GetPostByIDQuery{
				PostID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPostByIDQuery_QueryName(t *testing.T) {
	query := &GetPostByIDQuery{}
	assert.Equal(t, "get_post_by_id", query.QueryName())
}

func TestGetPostBySlugQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   *GetPostBySlugQuery
		wantErr bool
	}{
		{
			name: "valid query",
			query: &GetPostBySlugQuery{
				Slug: "test-post",
			},
			wantErr: false,
		},
		{
			name: "missing slug",
			query: &GetPostBySlugQuery{
				Slug: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPostBySlugQuery_QueryName(t *testing.T) {
	query := &GetPostBySlugQuery{}
	assert.Equal(t, "get_post_by_slug", query.QueryName())
}

func TestListPostsQuery_Validate(t *testing.T) {
	query := &ListPostsQuery{
		Filters: repository.PostFilters{
			Page:     1,
			PageSize: 10,
		},
	}

	// ListPostsQuery should always validate successfully
	err := query.Validate()
	assert.NoError(t, err)
}

func TestListPostsQuery_QueryName(t *testing.T) {
	query := &ListPostsQuery{}
	assert.Equal(t, "list_posts", query.QueryName())
}

// ============================================================================
// Query Handler Tests
// ============================================================================

func TestGetPostByIDHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful post retrieval", func(t *testing.T) {
		mockRepo := new(MockPostRepository)

		expectedPost := &domain.Post{
			ID:    "post-123",
			Title: "Test Post",
			Slug:  "test-post",
		}

		mockRepo.On("GetByID", ctx, "post-123").Return(expectedPost, nil)

		handler := NewGetPostByIDHandler(mockRepo)

		query := &GetPostByIDQuery{
			PostID: "post-123",
		}

		result, err := handler.Handle(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		post := result.(*domain.Post)
		assert.Equal(t, "post-123", post.ID)
		assert.Equal(t, "Test Post", post.Title)

		mockRepo.AssertExpectations(t)
	})

	t.Run("post not found", func(t *testing.T) {
		mockRepo := new(MockPostRepository)

		mockRepo.On("GetByID", ctx, "non-existent").Return(nil, errors.New("not found"))

		handler := NewGetPostByIDHandler(mockRepo)

		query := &GetPostByIDQuery{
			PostID: "non-existent",
		}

		result, err := handler.Handle(ctx, query)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get post")
	})
}

func TestGetPostBySlugHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful post retrieval by slug", func(t *testing.T) {
		mockRepo := new(MockPostRepository)

		expectedPost := &domain.Post{
			ID:    "post-123",
			Title: "Test Post",
			Slug:  "test-post",
		}

		mockRepo.On("GetBySlug", ctx, "test-post").Return(expectedPost, nil)

		handler := NewGetPostBySlugHandler(mockRepo)

		query := &GetPostBySlugQuery{
			Slug: "test-post",
		}

		result, err := handler.Handle(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		post := result.(*domain.Post)
		assert.Equal(t, "post-123", post.ID)
		assert.Equal(t, "test-post", post.Slug)

		mockRepo.AssertExpectations(t)
	})

	t.Run("post not found by slug", func(t *testing.T) {
		mockRepo := new(MockPostRepository)

		mockRepo.On("GetBySlug", ctx, "non-existent-slug").Return(nil, errors.New("not found"))

		handler := NewGetPostBySlugHandler(mockRepo)

		query := &GetPostBySlugQuery{
			Slug: "non-existent-slug",
		}

		result, err := handler.Handle(ctx, query)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get post by slug")
	})
}

func TestListPostsHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful posts listing", func(t *testing.T) {
		mockRepo := new(MockPostRepository)

		expectedPosts := []domain.Post{
			{ID: "post-1", Title: "Post 1", Slug: "post-1"},
			{ID: "post-2", Title: "Post 2", Slug: "post-2"},
			{ID: "post-3", Title: "Post 3", Slug: "post-3"},
		}

		filters := repository.PostFilters{
			Page:     1,
			PageSize: 10,
		}

		mockRepo.On("List", ctx, filters).Return(expectedPosts, int64(3), nil)

		handler := NewListPostsHandler(mockRepo)

		query := &ListPostsQuery{
			Filters: filters,
		}

		result, err := handler.Handle(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		listResult := result.(*ListPostsResult)
		assert.Len(t, listResult.Posts, 3)
		assert.Equal(t, int64(3), listResult.Total)
		assert.Equal(t, "post-1", listResult.Posts[0].ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty posts list", func(t *testing.T) {
		mockRepo := new(MockPostRepository)

		filters := repository.PostFilters{
			Page:     1,
			PageSize: 10,
		}

		mockRepo.On("List", ctx, filters).Return([]domain.Post{}, int64(0), nil)

		handler := NewListPostsHandler(mockRepo)

		query := &ListPostsQuery{
			Filters: filters,
		}

		result, err := handler.Handle(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		listResult := result.(*ListPostsResult)
		assert.Len(t, listResult.Posts, 0)
		assert.Equal(t, int64(0), listResult.Total)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error during listing", func(t *testing.T) {
		mockRepo := new(MockPostRepository)

		filters := repository.PostFilters{
			Page:     1,
			PageSize: 10,
		}

		mockRepo.On("List", ctx, filters).Return([]domain.Post{}, int64(0), errors.New("database error"))

		handler := NewListPostsHandler(mockRepo)

		query := &ListPostsQuery{
			Filters: filters,
		}

		result, err := handler.Handle(ctx, query)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to list posts")
	})

	t.Run("listing with pagination", func(t *testing.T) {
		mockRepo := new(MockPostRepository)

		expectedPosts := []domain.Post{
			{ID: "post-11", Title: "Post 11", Slug: "post-11"},
			{ID: "post-12", Title: "Post 12", Slug: "post-12"},
		}

		filters := repository.PostFilters{
			Page:     2,
			PageSize: 10,
		}

		mockRepo.On("List", ctx, filters).Return(expectedPosts, int64(25), nil)

		handler := NewListPostsHandler(mockRepo)

		query := &ListPostsQuery{
			Filters: filters,
		}

		result, err := handler.Handle(ctx, query)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		listResult := result.(*ListPostsResult)
		assert.Len(t, listResult.Posts, 2)
		assert.Equal(t, int64(25), listResult.Total)

		mockRepo.AssertExpectations(t)
	})
}
