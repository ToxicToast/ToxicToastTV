package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio"
	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
)

type ItemDetailHandler struct {
	pb.UnimplementedItemDetailServiceServer
	detailUC usecase.ItemDetailUseCase
}

func NewItemDetailHandler(detailUC usecase.ItemDetailUseCase) *ItemDetailHandler {
	return &ItemDetailHandler{
		detailUC: detailUC,
	}
}

func (h *ItemDetailHandler) CreateItemDetail(ctx context.Context, req *pb.CreateItemDetailRequest) (*pb.CreateItemDetailResponse, error) {
	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		expiry := req.ExpiryDate.AsTime()
		expiryDate = &expiry
	}

	var articleNumber *string
	if req.ArticleNumber != nil {
		articleNumber = req.ArticleNumber
	}

	detail, err := h.detailUC.CreateItemDetail(
		ctx,
		req.ItemVariantId,
		req.WarehouseId,
		req.LocationId,
		articleNumber,
		req.PurchasePrice,
		req.PurchaseDate.AsTime(),
		expiryDate,
		req.HasDeposit,
		req.IsFrozen,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateItemDetailResponse{
		ItemDetail: mapper.ItemDetailToProto(detail),
	}, nil
}

func (h *ItemDetailHandler) BatchCreateItemDetails(ctx context.Context, req *pb.BatchCreateItemDetailsRequest) (*pb.BatchCreateItemDetailsResponse, error) {
	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		expiry := req.ExpiryDate.AsTime()
		expiryDate = &expiry
	}

	var articleNumber *string
	if req.ArticleNumber != nil {
		articleNumber = req.ArticleNumber
	}

	createdDetails, err := h.detailUC.BatchCreateItemDetails(
		ctx,
		req.ItemVariantId,
		req.WarehouseId,
		req.LocationId,
		articleNumber,
		req.PurchasePrice,
		req.PurchaseDate.AsTime(),
		expiryDate,
		req.HasDeposit,
		req.IsFrozen,
		int(req.Quantity),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.BatchCreateItemDetailsResponse{
		ItemDetails:  mapper.ItemDetailsToProto(createdDetails),
		CreatedCount: int32(len(createdDetails)),
	}, nil
}

func (h *ItemDetailHandler) GetItemDetail(ctx context.Context, req *pb.IdRequest) (*pb.GetItemDetailResponse, error) {
	detail, err := h.detailUC.GetItemDetailByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemDetailNotFound {
			return nil, status.Error(codes.NotFound, "item detail not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetItemDetailResponse{
		ItemDetail: mapper.ItemDetailToProto(detail),
	}, nil
}

func (h *ItemDetailHandler) ListItemDetails(ctx context.Context, req *pb.ListItemDetailsRequest) (*pb.ListItemDetailsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	var variantID, warehouseID, locationID *string
	var isOpened, hasDeposit, isFrozen *bool

	if req.ItemVariantId != nil {
		variantID = req.ItemVariantId
	}
	if req.WarehouseId != nil {
		warehouseID = req.WarehouseId
	}
	if req.LocationId != nil {
		locationID = req.LocationId
	}
	if req.IsOpened != nil {
		isOpened = req.IsOpened
	}
	if req.HasDeposit != nil {
		hasDeposit = req.HasDeposit
	}
	if req.IsFrozen != nil {
		isFrozen = req.IsFrozen
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	details, total, err := h.detailUC.ListItemDetails(ctx, int(page), int(pageSize), variantID, warehouseID, locationID, isOpened, hasDeposit, isFrozen, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListItemDetailsResponse{
		ItemDetails: mapper.ItemDetailsToProto(details),
		Total:       int32(total),
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  int32(totalPages),
	}, nil
}

func (h *ItemDetailHandler) UpdateItemDetail(ctx context.Context, req *pb.UpdateItemDetailRequest) (*pb.UpdateItemDetailResponse, error) {
	// Get existing to use as defaults
	existing, err := h.detailUC.GetItemDetailByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemDetailNotFound {
			return nil, status.Error(codes.NotFound, "item detail not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	locationID := existing.LocationID
	if req.LocationId != nil {
		locationID = *req.LocationId
	}

	var articleNumber *string
	if req.ArticleNumber != nil {
		articleNumber = req.ArticleNumber
	} else {
		articleNumber = existing.ArticleNumber
	}

	purchasePrice := existing.PurchasePrice
	if req.PurchasePrice != nil {
		purchasePrice = *req.PurchasePrice
	}

	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		expiry := req.ExpiryDate.AsTime()
		expiryDate = &expiry
	} else {
		expiryDate = existing.ExpiryDate
	}

	hasDeposit := existing.HasDeposit
	if req.HasDeposit != nil {
		hasDeposit = *req.HasDeposit
	}

	isFrozen := existing.IsFrozen
	if req.IsFrozen != nil {
		isFrozen = *req.IsFrozen
	}

	detail, err := h.detailUC.UpdateItemDetail(ctx, req.Id, locationID, articleNumber, purchasePrice, expiryDate, hasDeposit, isFrozen)
	if err != nil {
		if err == usecase.ErrItemDetailNotFound {
			return nil, status.Error(codes.NotFound, "item detail not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateItemDetailResponse{
		ItemDetail: mapper.ItemDetailToProto(detail),
	}, nil
}

func (h *ItemDetailHandler) DeleteItemDetail(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	err := h.detailUC.DeleteItemDetail(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemDetailNotFound {
			return nil, status.Error(codes.NotFound, "item detail not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Item detail deleted successfully",
	}, nil
}

func (h *ItemDetailHandler) OpenItem(ctx context.Context, req *pb.OpenItemRequest) (*pb.OpenItemResponse, error) {
	detail, err := h.detailUC.OpenItem(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemDetailNotFound {
			return nil, status.Error(codes.NotFound, "item detail not found")
		}
		if err == usecase.ErrItemAlreadyOpened {
			return nil, status.Error(codes.FailedPrecondition, "item already opened")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.OpenItemResponse{
		ItemDetail: mapper.ItemDetailToProto(detail),
	}, nil
}

func (h *ItemDetailHandler) MoveItems(ctx context.Context, req *pb.MoveItemsRequest) (*pb.MoveItemsResponse, error) {
	movedItems, err := h.detailUC.MoveItems(ctx, req.ItemDetailIds, req.NewLocationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MoveItemsResponse{
		ItemDetails: mapper.ItemDetailsToProto(movedItems),
		MovedCount:  int32(len(movedItems)),
	}, nil
}

func (h *ItemDetailHandler) GetExpiringItems(ctx context.Context, req *pb.GetExpiringItemsRequest) (*pb.GetExpiringItemsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	days := int(req.Days)
	if days < 1 {
		days = 7 // Default to 7 days
	}

	details, total, err := h.detailUC.GetExpiringItems(ctx, days, int(page), int(pageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetExpiringItemsResponse{
		ItemDetails:    mapper.ItemDetailsToProto(details),
		Total:          int32(total),
		Page:           page,
		PageSize:       pageSize,
		TotalPages:     int32(totalPages),
		TotalExpiring:  int32(total),
	}, nil
}

func (h *ItemDetailHandler) GetExpiredItems(ctx context.Context, req *pb.GetExpiredItemsRequest) (*pb.GetExpiredItemsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	details, total, err := h.detailUC.GetExpiredItems(ctx, int(page), int(pageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetExpiredItemsResponse{
		ItemDetails:   mapper.ItemDetailsToProto(details),
		Total:         int32(total),
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    int32(totalPages),
		TotalExpired:  int32(total),
	}, nil
}

func (h *ItemDetailHandler) GetItemsWithDeposit(ctx context.Context, req *pb.GetItemsWithDepositRequest) (*pb.GetItemsWithDepositResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	details, total, err := h.detailUC.GetItemsWithDeposit(ctx, int(page), int(pageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetItemsWithDepositResponse{
		ItemDetails:       mapper.ItemDetailsToProto(details),
		Total:             int32(total),
		Page:              page,
		PageSize:          pageSize,
		TotalPages:        int32(totalPages),
		TotalDepositValue: 0.0, // TODO: Calculate actual deposit value
	}, nil
}

func (h *ItemDetailHandler) GetItemsByLocation(ctx context.Context, req *pb.GetItemsByLocationRequest) (*pb.GetItemsByLocationResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	// Get all items for location
	allDetails, err := h.detailUC.GetByLocation(ctx, req.LocationId, req.IncludeChildren)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Manual pagination
	total := int64(len(allDetails))
	offset := (int(page) - 1) * int(pageSize)
	end := offset + int(pageSize)

	if offset > len(allDetails) {
		offset = len(allDetails)
	}
	if end > len(allDetails) {
		end = len(allDetails)
	}

	paginatedDetails := allDetails[offset:end]
	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetItemsByLocationResponse{
		ItemDetails: mapper.ItemDetailsToProto(paginatedDetails),
		Total:       int32(total),
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  int32(totalPages),
	}, nil
}
