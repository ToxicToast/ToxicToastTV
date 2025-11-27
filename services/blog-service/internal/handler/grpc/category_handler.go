package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/toxictoast/toxictoastgo/shared/auth"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	pb "toxictoast/services/blog-service/api/proto"
	"toxictoast/services/blog-service/internal/command"
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/query"
	"toxictoast/services/blog-service/internal/repository"
)

type CategoryHandler struct {
	pb.UnimplementedBlogServiceServer
	commandBus  *cqrs.CommandBus
	queryBus    *cqrs.QueryBus
	authEnabled bool
}

func NewCategoryHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus, authEnabled bool) *CategoryHandler {
	return &CategoryHandler{
		commandBus:  commandBus,
		queryBus:    queryBus,
		authEnabled: authEnabled,
	}
}

func (h *CategoryHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CategoryResponse, error) {
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	cmd := &command.CreateCategoryCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
		Description: req.Description,
		ParentID:    stringPtrToOptional(req.ParentId),
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}

	getQuery := &query.GetCategoryByIDQuery{
		BaseQuery:  cqrs.BaseQuery{},
		CategoryID: cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created category: %v", err)
	}

	category := result.(*domain.Category)

	return &pb.CategoryResponse{
		Category: domainCategoryToProto(category),
	}, nil
}

func (h *CategoryHandler) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.CategoryResponse, error) {
	var category *domain.Category
	var err error
	var result interface{}

	switch identifier := req.Identifier.(type) {
	case *pb.GetCategoryRequest_Id:
		getQuery := &query.GetCategoryByIDQuery{
			BaseQuery:  cqrs.BaseQuery{},
			CategoryID: identifier.Id,
		}
		result, err = h.queryBus.Dispatch(ctx, getQuery)
		if err == nil {
			category = result.(*domain.Category)
		}

	case *pb.GetCategoryRequest_Slug:
		getQuery := &query.GetCategoryBySlugQuery{
			BaseQuery: cqrs.BaseQuery{},
			Slug:      identifier.Slug,
		}
		result, err = h.queryBus.Dispatch(ctx, getQuery)
		if err == nil {
			category = result.(*domain.Category)
		}

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
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	cmd := &command.UpdateCategoryCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        stringPtrFromOptional(req.Name),
		Description: stringPtrFromOptional(req.Description),
		ParentID:    stringPtrToOptional(req.ParentId),
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update category: %v", err)
	}

	getQuery := &query.GetCategoryByIDQuery{
		BaseQuery:  cqrs.BaseQuery{},
		CategoryID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, getQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve updated category: %v", err)
	}

	category := result.(*domain.Category)

	return &pb.CategoryResponse{
		Category: domainCategoryToProto(category),
	}, nil
}

func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteResponse, error) {
	_, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	cmd := &command.DeleteCategoryCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete category: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Category deleted successfully",
	}, nil
}

func (h *CategoryHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	filters := repository.CategoryFilters{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		ParentID: stringPtrFromOptional(req.ParentId),
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 100
	}

	listQuery := &query.ListCategoriesQuery{
		BaseQuery: cqrs.BaseQuery{},
		Filters:   filters,
	}

	result, err := h.queryBus.Dispatch(ctx, listQuery)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list categories: %v", err)
	}

	listResult := result.(*query.ListCategoriesResult)

	protoCategories := make([]*pb.Category, len(listResult.Categories))
	for i, cat := range listResult.Categories {
		protoCategories[i] = domainCategoryToProto(&cat)
	}

	return &pb.ListCategoriesResponse{
		Categories: protoCategories,
		Total:      int32(listResult.Total),
	}, nil
}

func domainCategoryToProto(category *domain.Category) *pb.Category {
	if category == nil {
		return nil
	}

	protoCategory := &pb.Category{
		Id:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		ParentId:    category.ParentID,
		CreatedAt:   timestamppb.New(category.CreatedAt),
		UpdatedAt:   timestamppb.New(category.UpdatedAt),
	}

	return protoCategory
}

func stringPtrToOptional(s *string) *string {
	if s == nil {
		return nil
	}
	return s
}

func (h *CategoryHandler) requireAuth(ctx context.Context) (*auth.UserContext, error) {
	if !h.authEnabled {
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
