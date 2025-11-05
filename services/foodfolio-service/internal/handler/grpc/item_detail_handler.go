package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
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
		req.PurchasePrice,
		req.PurchaseDate.AsTime(),
		expiryDate,
		articleNumber,
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
	details := make([]*domain.ItemDetail, len(req.Details))
	for i, d := range req.Details {
		var expiryDate *time.Time
		if d.ExpiryDate != nil {
			expiry := d.ExpiryDate.AsTime()
			expiryDate = &expiry
		}

		var articleNumber *string
		if d.ArticleNumber != nil {
			articleNumber = d.ArticleNumber
		}

		details[i] = &domain.ItemDetail{
			ItemVariantID: d.ItemVariantId,
			WarehouseID:   d.WarehouseId,
			LocationID:    d.LocationId,
			PurchasePrice: d.PurchasePrice,
			PurchaseDate:  d.PurchaseDate.AsTime(),
			ExpiryDate:    expiryDate,
			ArticleNumber: articleNumber,
			HasDeposit:    d.HasDeposit,
			IsFrozen:      d.IsFrozen,
		}
	}

	createdDetails, err := h.detailUC.BatchCreateItemDetails(ctx, details)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.BatchCreateItemDetailsResponse{
		ItemDetails: mapper.ItemDetailsToProto(createdDetails),
		Count:       int32(len(createdDetails)),
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
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

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

	details, total, err := h.detailUC.ListItemDetails(ctx, page, pageSize, variantID, warehouseID, locationID, isOpened, hasDeposit, isFrozen, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListItemDetailsResponse{
		ItemDetails: mapper.ItemDetailsToProto(details),
		Pagination:  mapper.ToPaginationResponse(page, pageSize, total),
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

	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		expiry := req.ExpiryDate.AsTime()
		expiryDate = &expiry
	} else {
		expiryDate = existing.ExpiryDate
	}

	var articleNumber *string
	if req.ArticleNumber != nil {
		articleNumber = req.ArticleNumber
	} else {
		articleNumber = existing.ArticleNumber
	}

	hasDeposit := existing.HasDeposit
	if req.HasDeposit != nil {
		hasDeposit = *req.HasDeposit
	}

	isFrozen := existing.IsFrozen
	if req.IsFrozen != nil {
		isFrozen = *req.IsFrozen
	}

	detail, err := h.detailUC.UpdateItemDetail(ctx, req.Id, locationID, expiryDate, articleNumber, hasDeposit, isFrozen)
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

func (h *ItemDetailHandler) DeleteItemDetail(ctx context.Context, req *pb.IdRequest) (*pb.SuccessResponse, error) {
	err := h.detailUC.DeleteItemDetail(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemDetailNotFound {
			return nil, status.Error(codes.NotFound, "item detail not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SuccessResponse{
		Success: true,
		Message: "Item detail deleted successfully",
	}, nil
}

func (h *ItemDetailHandler) OpenItem(ctx context.Context, req *pb.IdRequest) (*pb.OpenItemResponse, error) {
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
	count, err := h.detailUC.MoveItems(ctx, req.ItemIds, req.NewLocationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MoveItemsResponse{
		Success:    true,
		Message:    "Items moved successfully",
		ItemsCount: int32(count),
	}, nil
}

func (h *ItemDetailHandler) GetExpiringItems(ctx context.Context, req *pb.GetExpiringItemsRequest) (*pb.GetExpiringItemsResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	days := int(req.Days)
	if days < 1 {
		days = 7 // Default to 7 days
	}

	details, total, err := h.detailUC.GetExpiringItems(ctx, days, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetExpiringItemsResponse{
		ItemDetails: mapper.ItemDetailsToProto(details),
		Pagination:  mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}

func (h *ItemDetailHandler) GetExpiredItems(ctx context.Context, req *pb.GetExpiredItemsRequest) (*pb.GetExpiredItemsResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	details, total, err := h.detailUC.GetExpiredItems(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetExpiredItemsResponse{
		ItemDetails: mapper.ItemDetailsToProto(details),
		Pagination:  mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}

func (h *ItemDetailHandler) GetItemsWithDeposit(ctx context.Context, req *pb.GetItemsWithDepositRequest) (*pb.GetItemsWithDepositResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	details, total, err := h.detailUC.GetItemsWithDeposit(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetItemsWithDepositResponse{
		ItemDetails: mapper.ItemDetailsToProto(details),
		Pagination:  mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}

func (h *ItemDetailHandler) GetItemsByLocation(ctx context.Context, req *pb.GetItemsByLocationRequest) (*pb.GetItemsByLocationResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	details, total, err := h.detailUC.GetItemsByLocation(ctx, req.LocationId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetItemsByLocationResponse{
		ItemDetails: mapper.ItemDetailsToProto(details),
		Pagination:  mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}
