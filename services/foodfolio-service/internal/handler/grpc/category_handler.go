package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	pb "toxictoast/services/foodfolio-service/api/proto"
	"toxictoast/services/foodfolio-service/internal/command"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/query"
)

type CategoryHandler struct {
	pb.UnimplementedCategoryServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewCategoryHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *CategoryHandler {
	return &CategoryHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *CategoryHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	cmd := &command.CreateCategoryCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
		ParentID:    parentID,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCategoryResponse{
		Category: &pb.Category{
			Name:     req.Name,
			ParentId: parentID,
		},
	}, nil
}

func (h *CategoryHandler) GetCategory(ctx context.Context, req *pb.IdRequest) (*pb.GetCategoryResponse, error) {
	qry := &query.GetCategoryByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "category not found")
	}

	category := result.(*domain.Category)

	return &pb.GetCategoryResponse{
		Category: mapper.CategoryToProto(category),
	}, nil
}

func (h *CategoryHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	includeChildren := req.IncludeChildren

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	qry := &query.ListCategoriesQuery{
		BaseQuery:       cqrs.BaseQuery{},
		Page:            int(page),
		PageSize:        int(pageSize),
		ParentID:        parentID,
		IncludeChildren: includeChildren,
		IncludeDeleted:  includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListCategoriesResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListCategoriesResponse{
		Categories: mapper.CategoriesToProto(listResult.Categories),
		Total:      int32(listResult.Total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	cmd := &command.UpdateCategoryCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
		ParentID:    req.ParentId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated category
	qry := &query.GetCategoryByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "category not found")
	}

	category := result.(*domain.Category)

	return &pb.UpdateCategoryResponse{
		Category: mapper.CategoryToProto(category),
	}, nil
}

func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteCategoryCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Category deleted successfully",
	}, nil
}

func (h *CategoryHandler) GetCategoryTree(ctx context.Context, req *pb.GetCategoryTreeRequest) (*pb.GetCategoryTreeResponse, error) {
	var rootID *string
	if req.RootId != nil {
		rootID = req.RootId
	}

	maxDepth := int(req.MaxDepth)

	qry := &query.GetCategoryTreeQuery{
		BaseQuery: cqrs.BaseQuery{},
		RootID:    rootID,
		MaxDepth:  maxDepth,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	categories := result.([]*domain.Category)

	return &pb.GetCategoryTreeResponse{
		Categories: mapper.CategoriesToProto(categories),
	}, nil
}
