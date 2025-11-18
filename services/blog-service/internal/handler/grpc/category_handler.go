package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/toxictoast/toxictoastgo/shared/auth"

	pb "toxictoast/services/blog-service/api/proto"
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/internal/usecase"
)

type CategoryHandler struct {
	pb.UnimplementedBlogServiceServer
	categoryUseCase usecase.CategoryUseCase
	authEnabled     bool
}

func NewCategoryHandler(categoryUseCase usecase.CategoryUseCase, authEnabled bool) *CategoryHandler {
	return &CategoryHandler{
		categoryUseCase: categoryUseCase,
		authEnabled:     authEnabled,
	}
}

func (h *CategoryHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CategoryResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Convert request to use case input
	input := usecase.CreateCategoryInput{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    stringPtrFromOptional(req.ParentId),
	}

	// Create category
	category, err := h.categoryUseCase.CreateCategory(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}

	return &pb.CategoryResponse{
		Category: domainCategoryToProto(category),
	}, nil
}

func (h *CategoryHandler) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.CategoryResponse, error) {
	var category *domain.Category
	var err error

	switch identifier := req.Identifier.(type) {
	case *pb.GetCategoryRequest_Id:
		category, err = h.categoryUseCase.GetCategory(ctx, identifier.Id)
	case *pb.GetCategoryRequest_Slug:
		category, err = h.categoryUseCase.GetCategoryBySlug(ctx, identifier.Slug)
	default:
		return nil, status.Error(codes.InvalidArgument, "must provide either id or slug")
	}

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "category not found: %v", err)
	}

	return &pb.CategoryResponse{
		Category: domainCategoryToProto(category),
	}, nil
}

func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.CategoryResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Convert request to use case input
	input := usecase.UpdateCategoryInput{
		Name:        stringPtrFromOptional(req.Name),
		Description: stringPtrFromOptional(req.Description),
		ParentID:    stringPtrFromOptional(req.ParentId),
	}

	// Update category
	category, err := h.categoryUseCase.UpdateCategory(ctx, req.Id, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update category: %v", err)
	}

	return &pb.CategoryResponse{
		Category: domainCategoryToProto(category),
	}, nil
}

func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteResponse, error) {
	// Get user from context (authenticated by middleware)
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Delete category
	if err := h.categoryUseCase.DeleteCategory(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete category: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Category deleted successfully",
	}, nil
}

func (h *CategoryHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	// Convert request to filters
	filters := repository.CategoryFilters{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		ParentID: stringPtrFromOptional(req.ParentId),
	}

	// Default pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 50 // Higher default for categories
	}

	// List categories
	categories, total, err := h.categoryUseCase.ListCategories(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list categories: %v", err)
	}

	// Convert to proto
	protoCategories := make([]*pb.Category, len(categories))
	for i, category := range categories {
		protoCategories[i] = domainCategoryToProto(&category)
	}

	return &pb.ListCategoriesResponse{
		Categories: protoCategories,
		Total:      int32(total),
	}, nil
}

// Helper functions for conversion

func domainCategoryToProto(category *domain.Category) *pb.Category {
	if category == nil {
		return nil
	}

	protoCategory := &pb.Category{
		Id:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		CreatedAt:   timestamppb.New(category.CreatedAt),
		UpdatedAt:   timestamppb.New(category.UpdatedAt),
	}

	// Add parent ID if exists
	if category.ParentID != nil {
		protoCategory.ParentId = stringPtrToOptional(category.ParentID)
	}

	return protoCategory
}

func stringPtrToOptional(s *string) *string {
	if s == nil {
		return nil
	}
	return s
}

// requireAuth checks authentication if enabled, returns user context or error
func (h *CategoryHandler) requireAuth(ctx context.Context) (*auth.UserContext, error) {
	if !h.authEnabled {
		// Return a dummy user context when auth is disabled
		return &auth.UserContext{
			UserID:   "test-user",
			Username: "test",
			Email:    "test@example.com",
			Roles:    []string{"admin"},
		}, nil
	}

	user, err := auth.GetUserContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	return user, nil
}
