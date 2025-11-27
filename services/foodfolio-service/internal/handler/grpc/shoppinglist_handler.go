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

type ShoppinglistHandler struct {
	pb.UnimplementedShoppinglistServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewShoppinglistHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *ShoppinglistHandler {
	return &ShoppinglistHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *ShoppinglistHandler) CreateShoppinglist(ctx context.Context, req *pb.CreateShoppinglistRequest) (*pb.CreateShoppinglistResponse, error) {
	cmd := &command.CreateShoppinglistCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the created shoppinglist
	qry := &query.GetShoppinglistByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "shoppinglist not found")
	}

	list := result.(*domain.Shoppinglist)

	return &pb.CreateShoppinglistResponse{
		Shoppinglist: mapper.ShoppinglistToProto(list),
	}, nil
}

func (h *ShoppinglistHandler) GetShoppinglist(ctx context.Context, req *pb.IdRequest) (*pb.GetShoppinglistResponse, error) {
	qry := &query.GetShoppinglistByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "shoppinglist not found")
	}

	list := result.(*domain.Shoppinglist)

	return &pb.GetShoppinglistResponse{
		Shoppinglist: mapper.ShoppinglistToProto(list),
	}, nil
}

func (h *ShoppinglistHandler) ListShoppinglists(ctx context.Context, req *pb.ListShoppinglistsRequest) (*pb.ListShoppinglistsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	qry := &query.ListShoppinglistsQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           int(page),
		PageSize:       int(pageSize),
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListShoppinglistsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListShoppinglistsResponse{
		Shoppinglists: mapper.ShoppinglistsToProto(listResult.Shoppinglists),
		Total:         int32(listResult.Total),
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    int32(totalPages),
	}, nil
}

func (h *ShoppinglistHandler) UpdateShoppinglist(ctx context.Context, req *pb.UpdateShoppinglistRequest) (*pb.UpdateShoppinglistResponse, error) {
	cmd := &command.UpdateShoppinglistCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated shoppinglist
	qry := &query.GetShoppinglistByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "shoppinglist not found")
	}

	list := result.(*domain.Shoppinglist)

	return &pb.UpdateShoppinglistResponse{
		Shoppinglist: mapper.ShoppinglistToProto(list),
	}, nil
}

func (h *ShoppinglistHandler) DeleteShoppinglist(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteShoppinglistCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Shoppinglist deleted successfully",
	}, nil
}

func (h *ShoppinglistHandler) AddItemToShoppinglist(ctx context.Context, req *pb.AddItemToShoppinglistRequest) (*pb.AddItemToShoppinglistResponse, error) {
	cmd := &command.AddItemToShoppinglistCommand{
		BaseCommand:    cqrs.BaseCommand{},
		ShoppinglistID: req.ShoppinglistId,
		VariantID:      req.ItemVariantId,
		Quantity:       int(req.Quantity),
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: Query individual item when GetShoppinglistItemByIDQuery is implemented
	// For now, return a basic response
	return &pb.AddItemToShoppinglistResponse{
		Item: &pb.ShoppinglistItem{
			Id:            cmd.AggregateID,
			ShoppinglistId: req.ShoppinglistId,
			ItemVariantId: req.ItemVariantId,
			Quantity:      req.Quantity,
			IsPurchased:   false,
		},
	}, nil
}

func (h *ShoppinglistHandler) RemoveItemFromShoppinglist(ctx context.Context, req *pb.RemoveItemFromShoppinglistRequest) (*pb.DeleteResponse, error) {
	cmd := &command.RemoveItemFromShoppinglistCommand{
		BaseCommand:    cqrs.BaseCommand{AggregateID: req.ItemId},
		ShoppinglistID: req.ShoppinglistId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Item removed from shoppinglist successfully",
	}, nil
}

func (h *ShoppinglistHandler) UpdateShoppinglistItem(ctx context.Context, req *pb.UpdateShoppinglistItemRequest) (*pb.UpdateShoppinglistItemResponse, error) {
	quantity := 1
	if req.Quantity != nil {
		quantity = int(*req.Quantity)
	}

	isPurchased := false
	if req.IsPurchased != nil {
		isPurchased = *req.IsPurchased
	}

	cmd := &command.UpdateShoppinglistItemCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Quantity:    quantity,
		IsPurchased: isPurchased,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: Query individual item when GetShoppinglistItemByIDQuery is implemented
	// For now, return a basic response
	return &pb.UpdateShoppinglistItemResponse{
		Item: &pb.ShoppinglistItem{
			Id:          req.Id,
			Quantity:    int32(quantity),
			IsPurchased: isPurchased,
		},
	}, nil
}

func (h *ShoppinglistHandler) MarkItemPurchased(ctx context.Context, req *pb.MarkItemPurchasedRequest) (*pb.MarkItemPurchasedResponse, error) {
	cmd := &command.MarkItemPurchasedCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: Query individual item when GetShoppinglistItemByIDQuery is implemented
	// For now, return a basic response
	return &pb.MarkItemPurchasedResponse{
		Item: &pb.ShoppinglistItem{
			Id:          req.Id,
			IsPurchased: true,
		},
	}, nil
}

func (h *ShoppinglistHandler) MarkAllItemsPurchased(ctx context.Context, req *pb.MarkAllItemsPurchasedRequest) (*pb.MarkAllItemsPurchasedResponse, error) {
	cmd := &command.MarkAllItemsPurchasedCommand{
		BaseCommand:    cqrs.BaseCommand{},
		ShoppinglistID: req.ShoppinglistId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MarkAllItemsPurchasedResponse{
		UpdatedCount: int32(cmd.ItemsMarked),
	}, nil
}

func (h *ShoppinglistHandler) ClearPurchasedItems(ctx context.Context, req *pb.ClearPurchasedItemsRequest) (*pb.ClearPurchasedItemsResponse, error) {
	cmd := &command.ClearPurchasedItemsCommand{
		BaseCommand:    cqrs.BaseCommand{},
		ShoppinglistID: req.ShoppinglistId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ClearPurchasedItemsResponse{
		RemovedCount: int32(cmd.ItemsCleared),
	}, nil
}

func (h *ShoppinglistHandler) GenerateFromLowStock(ctx context.Context, req *pb.GenerateFromLowStockRequest) (*pb.GenerateFromLowStockResponse, error) {
	cmd := &command.GenerateFromLowStockCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the created shoppinglist
	qry := &query.GetShoppinglistByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "shoppinglist not found")
	}

	list := result.(*domain.Shoppinglist)

	return &pb.GenerateFromLowStockResponse{
		Shoppinglist: mapper.ShoppinglistToProto(list),
		ItemsAdded:   int32(cmd.ItemsAdded),
	}, nil
}
