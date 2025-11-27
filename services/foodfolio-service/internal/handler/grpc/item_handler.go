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

type ItemHandler struct {
	pb.UnimplementedItemServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewItemHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *ItemHandler {
	return &ItemHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *ItemHandler) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	cmd := &command.CreateItemCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
		CategoryID:  req.CategoryId,
		CompanyID:   req.CompanyId,
		TypeID:      req.TypeId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateItemResponse{
		Item: &pb.Item{
			Name:       req.Name,
			CategoryId: req.CategoryId,
			CompanyId:  req.CompanyId,
			TypeId:     req.TypeId,
		},
	}, nil
}

func (h *ItemHandler) GetItem(ctx context.Context, req *pb.IdRequest) (*pb.GetItemResponse, error) {
	qry := &query.GetItemByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item not found")
	}

	item := result.(*domain.Item)

	return &pb.GetItemResponse{
		Item: mapper.ItemToProto(item),
	}, nil
}

func (h *ItemHandler) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

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

	qry := &query.ListItemsQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           int(page),
		PageSize:       int(pageSize),
		CategoryID:     categoryID,
		CompanyID:      companyID,
		TypeID:         typeID,
		Search:         search,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListItemsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListItemsResponse{
		Items:      mapper.ItemsToProto(listResult.Items),
		Total:      int32(listResult.Total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (h *ItemHandler) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	cmd := &command.UpdateItemCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
		CategoryID:  req.CategoryId,
		CompanyID:   req.CompanyId,
		TypeID:      req.TypeId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated item
	qry := &query.GetItemByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item not found")
	}

	item := result.(*domain.Item)

	return &pb.UpdateItemResponse{
		Item: mapper.ItemToProto(item),
	}, nil
}

func (h *ItemHandler) DeleteItem(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteItemCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Item deleted successfully",
	}, nil
}
