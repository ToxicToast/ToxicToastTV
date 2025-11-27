package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	pb "toxictoast/services/foodfolio-service/api/proto"
	"toxictoast/services/foodfolio-service/internal/command"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/query"
)

type ItemDetailHandler struct {
	pb.UnimplementedItemDetailServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewItemDetailHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *ItemDetailHandler {
	return &ItemDetailHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
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

	cmd := &command.CreateItemDetailCommand{
		BaseCommand:   cqrs.BaseCommand{},
		ItemVariantID: req.ItemVariantId,
		WarehouseID:   req.WarehouseId,
		LocationID:    req.LocationId,
		ArticleNumber: articleNumber,
		PurchasePrice: req.PurchasePrice,
		PurchaseDate:  req.PurchaseDate.AsTime(),
		ExpiryDate:    expiryDate,
		HasDeposit:    req.HasDeposit,
		IsFrozen:      req.IsFrozen,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the created detail
	qry := &query.GetItemDetailByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item detail not found")
	}

	detail := result.(*domain.ItemDetail)

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

	cmd := &command.BatchCreateItemDetailsCommand{
		BaseCommand:   cqrs.BaseCommand{},
		ItemVariantID: req.ItemVariantId,
		WarehouseID:   req.WarehouseId,
		LocationID:    req.LocationId,
		ArticleNumber: articleNumber,
		PurchasePrice: req.PurchasePrice,
		PurchaseDate:  req.PurchaseDate.AsTime(),
		ExpiryDate:    expiryDate,
		HasDeposit:    req.HasDeposit,
		IsFrozen:      req.IsFrozen,
		Quantity:      int(req.Quantity),
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query all created details
	createdDetails := make([]*domain.ItemDetail, 0, len(cmd.CreatedIDs))
	for _, id := range cmd.CreatedIDs {
		qry := &query.GetItemDetailByIDQuery{
			BaseQuery: cqrs.BaseQuery{},
			ID:        id,
		}

		result, err := h.queryBus.Dispatch(ctx, qry)
		if err == nil {
			if detail, ok := result.(*domain.ItemDetail); ok {
				createdDetails = append(createdDetails, detail)
			}
		}
	}

	return &pb.BatchCreateItemDetailsResponse{
		ItemDetails:  mapper.ItemDetailsToProto(createdDetails),
		CreatedCount: int32(len(createdDetails)),
	}, nil
}

func (h *ItemDetailHandler) GetItemDetail(ctx context.Context, req *pb.IdRequest) (*pb.GetItemDetailResponse, error) {
	qry := &query.GetItemDetailByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item detail not found")
	}

	detail := result.(*domain.ItemDetail)

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

	qry := &query.ListItemDetailsQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           int(page),
		PageSize:       int(pageSize),
		VariantID:      variantID,
		WarehouseID:    warehouseID,
		LocationID:     locationID,
		IsOpened:       isOpened,
		HasDeposit:     hasDeposit,
		IsFrozen:       isFrozen,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListItemDetailsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListItemDetailsResponse{
		ItemDetails: mapper.ItemDetailsToProto(listResult.ItemDetails),
		Total:       int32(listResult.Total),
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  int32(totalPages),
	}, nil
}

func (h *ItemDetailHandler) UpdateItemDetail(ctx context.Context, req *pb.UpdateItemDetailRequest) (*pb.UpdateItemDetailResponse, error) {
	// Get existing detail to use as defaults
	qry := &query.GetItemDetailByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item detail not found")
	}

	existing := result.(*domain.ItemDetail)

	// Use existing values as defaults
	locationID := existing.LocationID
	if req.LocationId != nil {
		locationID = *req.LocationId
	}

	purchasePrice := existing.PurchasePrice
	if req.PurchasePrice != nil {
		purchasePrice = *req.PurchasePrice
	}

	hasDeposit := existing.HasDeposit
	if req.HasDeposit != nil {
		hasDeposit = *req.HasDeposit
	}

	isFrozen := existing.IsFrozen
	if req.IsFrozen != nil {
		isFrozen = *req.IsFrozen
	}

	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		expiry := req.ExpiryDate.AsTime()
		expiryDate = &expiry
	} else {
		expiryDate = existing.ExpiryDate
	}

	articleNumber := existing.ArticleNumber
	if req.ArticleNumber != nil {
		articleNumber = req.ArticleNumber
	}

	cmd := &command.UpdateItemDetailCommand{
		BaseCommand:   cqrs.BaseCommand{AggregateID: req.Id},
		LocationID:    locationID,
		ArticleNumber: articleNumber,
		PurchasePrice: purchasePrice,
		ExpiryDate:    expiryDate,
		HasDeposit:    hasDeposit,
		IsFrozen:      isFrozen,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated detail
	result, err = h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item detail not found")
	}

	detail := result.(*domain.ItemDetail)

	return &pb.UpdateItemDetailResponse{
		ItemDetail: mapper.ItemDetailToProto(detail),
	}, nil
}

func (h *ItemDetailHandler) DeleteItemDetail(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteItemDetailCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Item detail deleted successfully",
	}, nil
}

func (h *ItemDetailHandler) OpenItem(ctx context.Context, req *pb.OpenItemRequest) (*pb.OpenItemResponse, error) {
	cmd := &command.OpenItemCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the opened item
	qry := &query.GetItemDetailByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item detail not found")
	}

	detail := result.(*domain.ItemDetail)

	return &pb.OpenItemResponse{
		ItemDetail: mapper.ItemDetailToProto(detail),
	}, nil
}

func (h *ItemDetailHandler) MoveItems(ctx context.Context, req *pb.MoveItemsRequest) (*pb.MoveItemsResponse, error) {
	cmd := &command.MoveItemsCommand{
		BaseCommand:   cqrs.BaseCommand{},
		ItemIDs:       req.ItemDetailIds,
		NewLocationID: req.NewLocationId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query all moved items
	movedItems := make([]*domain.ItemDetail, 0, len(req.ItemDetailIds))
	for _, id := range req.ItemDetailIds {
		qry := &query.GetItemDetailByIDQuery{
			BaseQuery: cqrs.BaseQuery{},
			ID:        id,
		}

		result, err := h.queryBus.Dispatch(ctx, qry)
		if err == nil {
			if detail, ok := result.(*domain.ItemDetail); ok {
				movedItems = append(movedItems, detail)
			}
		}
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

	qry := &query.GetExpiringItemsQuery{
		BaseQuery: cqrs.BaseQuery{},
		Days:      days,
		Page:      int(page),
		PageSize:  int(pageSize),
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListItemDetailsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetExpiringItemsResponse{
		ItemDetails:   mapper.ItemDetailsToProto(listResult.ItemDetails),
		Total:         int32(listResult.Total),
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    int32(totalPages),
		TotalExpiring: int32(listResult.Total),
	}, nil
}

func (h *ItemDetailHandler) GetExpiredItems(ctx context.Context, req *pb.GetExpiredItemsRequest) (*pb.GetExpiredItemsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	qry := &query.GetExpiredItemsQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      int(page),
		PageSize:  int(pageSize),
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListItemDetailsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetExpiredItemsResponse{
		ItemDetails:  mapper.ItemDetailsToProto(listResult.ItemDetails),
		Total:        int32(listResult.Total),
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   int32(totalPages),
		TotalExpired: int32(listResult.Total),
	}, nil
}

func (h *ItemDetailHandler) GetItemsWithDeposit(ctx context.Context, req *pb.GetItemsWithDepositRequest) (*pb.GetItemsWithDepositResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	qry := &query.GetItemsWithDepositQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      int(page),
		PageSize:  int(pageSize),
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListItemDetailsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetItemsWithDepositResponse{
		ItemDetails:       mapper.ItemDetailsToProto(listResult.ItemDetails),
		Total:             int32(listResult.Total),
		Page:              page,
		PageSize:          pageSize,
		TotalPages:        int32(totalPages),
		TotalDepositValue: 0.0, // TODO: Calculate actual deposit value
	}, nil
}

func (h *ItemDetailHandler) GetItemsByLocation(ctx context.Context, req *pb.GetItemsByLocationRequest) (*pb.GetItemsByLocationResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	qry := &query.GetItemDetailsByLocationQuery{
		BaseQuery:       cqrs.BaseQuery{},
		LocationID:      req.LocationId,
		IncludeChildren: req.IncludeChildren,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	allDetails := result.([]*domain.ItemDetail)

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
