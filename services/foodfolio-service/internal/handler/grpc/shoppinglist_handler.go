package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto"
)

type ShoppinglistHandler struct {
	pb.UnimplementedShoppinglistServiceServer
	shoppinglistUC usecase.ShoppinglistUseCase
}

func NewShoppinglistHandler(shoppinglistUC usecase.ShoppinglistUseCase) *ShoppinglistHandler {
	return &ShoppinglistHandler{
		shoppinglistUC: shoppinglistUC,
	}
}

func (h *ShoppinglistHandler) CreateShoppinglist(ctx context.Context, req *pb.CreateShoppinglistRequest) (*pb.CreateShoppinglistResponse, error) {
	list, err := h.shoppinglistUC.CreateShoppinglist(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateShoppinglistResponse{
		Shoppinglist: mapper.ShoppinglistToProto(list),
	}, nil
}

func (h *ShoppinglistHandler) GetShoppinglist(ctx context.Context, req *pb.IdRequest) (*pb.GetShoppinglistResponse, error) {
	list, err := h.shoppinglistUC.GetShoppinglistByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrShoppinglistNotFound {
			return nil, status.Error(codes.NotFound, "shoppinglist not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

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

	lists, total, err := h.shoppinglistUC.ListShoppinglists(ctx, int(page), int(pageSize), includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListShoppinglistsResponse{
		Shoppinglists: mapper.ShoppinglistsToProto(lists),
		Total:         int32(total),
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    int32(totalPages),
	}, nil
}

func (h *ShoppinglistHandler) UpdateShoppinglist(ctx context.Context, req *pb.UpdateShoppinglistRequest) (*pb.UpdateShoppinglistResponse, error) {
	list, err := h.shoppinglistUC.UpdateShoppinglist(ctx, req.Id, req.Name)
	if err != nil {
		if err == usecase.ErrShoppinglistNotFound {
			return nil, status.Error(codes.NotFound, "shoppinglist not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateShoppinglistResponse{
		Shoppinglist: mapper.ShoppinglistToProto(list),
	}, nil
}

func (h *ShoppinglistHandler) DeleteShoppinglist(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	err := h.shoppinglistUC.DeleteShoppinglist(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrShoppinglistNotFound {
			return nil, status.Error(codes.NotFound, "shoppinglist not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Shoppinglist deleted successfully",
	}, nil
}

func (h *ShoppinglistHandler) AddItemToShoppinglist(ctx context.Context, req *pb.AddItemToShoppinglistRequest) (*pb.AddItemToShoppinglistResponse, error) {
	item, err := h.shoppinglistUC.AddItemToShoppinglist(ctx, req.ShoppinglistId, req.ItemVariantId, int(req.Quantity))
	if err != nil {
		if err == usecase.ErrShoppinglistNotFound {
			return nil, status.Error(codes.NotFound, "shoppinglist not found")
		}
		if err == usecase.ErrItemVariantNotFound {
			return nil, status.Error(codes.NotFound, "item variant not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddItemToShoppinglistResponse{
		Item: mapper.ShoppinglistItemToProto(item),
	}, nil
}

func (h *ShoppinglistHandler) RemoveItemFromShoppinglist(ctx context.Context, req *pb.RemoveItemFromShoppinglistRequest) (*pb.DeleteResponse, error) {
	err := h.shoppinglistUC.RemoveItemFromShoppinglist(ctx, req.ShoppinglistId, req.ItemId)
	if err != nil {
		if err == usecase.ErrShoppinglistItemNotFound {
			return nil, status.Error(codes.NotFound, "shoppinglist item not found")
		}
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

	item, err := h.shoppinglistUC.UpdateShoppinglistItem(ctx, req.Id, quantity, isPurchased)
	if err != nil {
		if err == usecase.ErrShoppinglistItemNotFound {
			return nil, status.Error(codes.NotFound, "shoppinglist item not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateShoppinglistItemResponse{
		Item: mapper.ShoppinglistItemToProto(item),
	}, nil
}

func (h *ShoppinglistHandler) MarkItemPurchased(ctx context.Context, req *pb.MarkItemPurchasedRequest) (*pb.MarkItemPurchasedResponse, error) {
	item, err := h.shoppinglistUC.MarkItemPurchased(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrShoppinglistItemNotFound {
			return nil, status.Error(codes.NotFound, "shoppinglist item not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MarkItemPurchasedResponse{
		Item: mapper.ShoppinglistItemToProto(item),
	}, nil
}

func (h *ShoppinglistHandler) MarkAllItemsPurchased(ctx context.Context, req *pb.MarkAllItemsPurchasedRequest) (*pb.MarkAllItemsPurchasedResponse, error) {
	count, err := h.shoppinglistUC.MarkAllItemsPurchased(ctx, req.ShoppinglistId)
	if err != nil {
		if err == usecase.ErrShoppinglistNotFound {
			return nil, status.Error(codes.NotFound, "shoppinglist not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MarkAllItemsPurchasedResponse{
		UpdatedCount: int32(count),
	}, nil
}

func (h *ShoppinglistHandler) ClearPurchasedItems(ctx context.Context, req *pb.ClearPurchasedItemsRequest) (*pb.ClearPurchasedItemsResponse, error) {
	count, err := h.shoppinglistUC.ClearPurchasedItems(ctx, req.ShoppinglistId)
	if err != nil {
		if err == usecase.ErrShoppinglistNotFound {
			return nil, status.Error(codes.NotFound, "shoppinglist not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ClearPurchasedItemsResponse{
		RemovedCount: int32(count),
	}, nil
}

func (h *ShoppinglistHandler) GenerateFromLowStock(ctx context.Context, req *pb.GenerateFromLowStockRequest) (*pb.GenerateFromLowStockResponse, error) {
	list, itemsAdded, err := h.shoppinglistUC.GenerateFromLowStock(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GenerateFromLowStockResponse{
		Shoppinglist: mapper.ShoppinglistToProto(list),
		ItemsAdded:   int32(itemsAdded),
	}, nil
}
