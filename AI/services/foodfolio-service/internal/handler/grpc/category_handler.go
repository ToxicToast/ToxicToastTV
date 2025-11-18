package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto"
)

type CategoryHandler struct {
	pb.UnimplementedCategoryServiceServer
	categoryUC usecase.CategoryUseCase
}

func NewCategoryHandler(categoryUC usecase.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{
		categoryUC: categoryUC,
	}
}

func (h *CategoryHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	category, err := h.categoryUC.CreateCategory(ctx, req.Name, parentID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCategoryResponse{
		Category: mapper.CategoryToProto(category),
	}, nil
}

func (h *CategoryHandler) GetCategory(ctx context.Context, req *pb.IdRequest) (*pb.GetCategoryResponse, error) {
	category, err := h.categoryUC.GetCategoryByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrCategoryNotFound {
			return nil, status.Error(codes.NotFound, "category not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

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

	categories, total, err := h.categoryUC.ListCategories(ctx, int(page), int(pageSize), parentID, includeChildren, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListCategoriesResponse{
		Categories: mapper.CategoriesToProto(categories),
		Total:      int32(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	var name string
	if req.Name != nil {
		name = *req.Name
	} else {
		// Get existing to keep name
		cat, err := h.categoryUC.GetCategoryByID(ctx, req.Id)
		if err != nil {
			if err == usecase.ErrCategoryNotFound {
				return nil, status.Error(codes.NotFound, "category not found")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		name = cat.Name
	}

	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	category, err := h.categoryUC.UpdateCategory(ctx, req.Id, name, parentID)
	if err != nil {
		if err == usecase.ErrCategoryNotFound {
			return nil, status.Error(codes.NotFound, "category not found")
		}
		if err == usecase.ErrCircularReference {
			return nil, status.Error(codes.InvalidArgument, "circular reference detected")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateCategoryResponse{
		Category: mapper.CategoryToProto(category),
	}, nil
}

func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	err := h.categoryUC.DeleteCategory(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrCategoryNotFound {
			return nil, status.Error(codes.NotFound, "category not found")
		}
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

	categories, err := h.categoryUC.GetCategoryTree(ctx, rootID, maxDepth)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetCategoryTreeResponse{
		Categories: mapper.CategoriesToProto(categories),
	}, nil
}
