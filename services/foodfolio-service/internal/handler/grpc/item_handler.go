package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

type ItemHandler struct {
	pb.UnimplementedItemServiceServer
	itemUC usecase.ItemUseCase
}

func NewItemHandler(itemUC usecase.ItemUseCase) *ItemHandler {
	return &ItemHandler{
		itemUC: itemUC,
	}
}

func (h *ItemHandler) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	item, err := h.itemUC.CreateItem(ctx, req.Name, req.CategoryId, req.CompanyId, req.TypeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateItemResponse{
		Item: mapper.ItemToProto(item),
	}, nil
}

func (h *ItemHandler) GetItem(ctx context.Context, req *pb.IdRequest) (*pb.GetItemResponse, error) {
	item, err := h.itemUC.GetItemByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemNotFound {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetItemResponse{
		Item: mapper.ItemToProto(item),
	}, nil
}

func (h *ItemHandler) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	var categoryID, companyID, typeID, search *string
	if req.CategoryId != nil {
		categoryID = req.CategoryId
	}
	if req.CompanyId != nil {
		companyID = req.CompanyId
	}
	if req.TypeId != nil {
		typeID = req.TypeId
	}
	if req.Search != nil {
		search = req.Search
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	items, total, err := h.itemUC.ListItems(ctx, page, pageSize, categoryID, companyID, typeID, search, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListItemsResponse{
		Items:      mapper.ItemsToProto(items),
		Pagination: mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}

func (h *ItemHandler) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	var name, categoryID, companyID, typeID string

	// Get existing to use as defaults
	existing, err := h.itemUC.GetItemByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemNotFound {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if req.Name != nil {
		name = *req.Name
	} else {
		name = existing.Name
	}

	if req.CategoryId != nil {
		categoryID = *req.CategoryId
	} else {
		categoryID = existing.CategoryID
	}

	if req.CompanyId != nil {
		companyID = *req.CompanyId
	} else {
		companyID = existing.CompanyID
	}

	if req.TypeId != nil {
		typeID = *req.TypeId
	} else {
		typeID = existing.TypeID
	}

	item, err := h.itemUC.UpdateItem(ctx, req.Id, name, categoryID, companyID, typeID)
	if err != nil {
		if err == usecase.ErrItemNotFound {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateItemResponse{
		Item: mapper.ItemToProto(item),
	}, nil
}

func (h *ItemHandler) DeleteItem(ctx context.Context, req *pb.IdRequest) (*pb.SuccessResponse, error) {
	err := h.itemUC.DeleteItem(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemNotFound {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SuccessResponse{
		Success: true,
		Message: "Item deleted successfully",
	}, nil
}

func (h *ItemHandler) SearchItems(ctx context.Context, req *pb.SearchItemsRequest) (*pb.SearchItemsResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	var categoryID, companyID *string
	if req.CategoryId != nil {
		categoryID = req.CategoryId
	}
	if req.CompanyId != nil {
		companyID = req.CompanyId
	}

	items, total, err := h.itemUC.SearchItems(ctx, req.Query, page, pageSize, categoryID, companyID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SearchItemsResponse{
		Items:      mapper.ItemsToProto(items),
		Pagination: mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}
