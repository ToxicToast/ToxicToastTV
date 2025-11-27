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

type ItemVariantHandler struct {
	pb.UnimplementedItemVariantServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewItemVariantHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *ItemVariantHandler {
	return &ItemVariantHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *ItemVariantHandler) CreateItemVariant(ctx context.Context, req *pb.CreateItemVariantRequest) (*pb.CreateItemVariantResponse, error) {
	var barcode *string
	if req.Barcode != nil {
		barcode = req.Barcode
	}

	cmd := &command.CreateItemVariantCommand{
		BaseCommand:      cqrs.BaseCommand{},
		ItemID:           req.ItemId,
		SizeID:           req.SizeId,
		VariantName:      req.VariantName,
		Barcode:          barcode,
		MinSKU:           int(req.MinSku),
		MaxSKU:           int(req.MaxSku),
		IsNormallyFrozen: req.IsNormallyFrozen,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateItemVariantResponse{
		ItemVariant: &pb.ItemVariant{
			ItemId:           req.ItemId,
			SizeId:           req.SizeId,
			VariantName:      req.VariantName,
			Barcode:          barcode,
			MinSku:           req.MinSku,
			MaxSku:           req.MaxSku,
			IsNormallyFrozen: req.IsNormallyFrozen,
		},
	}, nil
}

func (h *ItemVariantHandler) GetItemVariant(ctx context.Context, req *pb.IdRequest) (*pb.GetItemVariantResponse, error) {
	qry := &query.GetItemVariantByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item variant not found")
	}

	variant := result.(*domain.ItemVariant)

	return &pb.GetItemVariantResponse{
		ItemVariant: mapper.ItemVariantToProto(variant),
	}, nil
}

func (h *ItemVariantHandler) ListItemVariants(ctx context.Context, req *pb.ListItemVariantsRequest) (*pb.ListItemVariantsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	var itemID, sizeID *string
	var isNormallyFrozen *bool

	if req.ItemId != nil {
		itemID = req.ItemId
	}
	if req.SizeId != nil {
		sizeID = req.SizeId
	}
	if req.IsNormallyFrozen != nil {
		isNormallyFrozen = req.IsNormallyFrozen
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	qry := &query.ListItemVariantsQuery{
		BaseQuery:        cqrs.BaseQuery{},
		Page:             int(page),
		PageSize:         int(pageSize),
		ItemID:           itemID,
		SizeID:           sizeID,
		IsNormallyFrozen: isNormallyFrozen,
		IncludeDeleted:   includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListItemVariantsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListItemVariantsResponse{
		ItemVariants: mapper.ItemVariantsToProto(listResult.ItemVariants),
		Total:        int32(listResult.Total),
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   int32(totalPages),
	}, nil
}

func (h *ItemVariantHandler) UpdateItemVariant(ctx context.Context, req *pb.UpdateItemVariantRequest) (*pb.UpdateItemVariantResponse, error) {
	var minSKU, maxSKU *int
	if req.MinSku != nil {
		val := int(*req.MinSku)
		minSKU = &val
	}
	if req.MaxSku != nil {
		val := int(*req.MaxSku)
		maxSKU = &val
	}

	cmd := &command.UpdateItemVariantCommand{
		BaseCommand:      cqrs.BaseCommand{AggregateID: req.Id},
		VariantName:      req.VariantName,
		Barcode:          req.Barcode,
		MinSKU:           minSKU,
		MaxSKU:           maxSKU,
		IsNormallyFrozen: req.IsNormallyFrozen,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated variant
	qry := &query.GetItemVariantByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item variant not found")
	}

	variant := result.(*domain.ItemVariant)

	return &pb.UpdateItemVariantResponse{
		ItemVariant: mapper.ItemVariantToProto(variant),
	}, nil
}

func (h *ItemVariantHandler) DeleteItemVariant(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteItemVariantCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Item variant deleted successfully",
	}, nil
}

func (h *ItemVariantHandler) GetCurrentStock(ctx context.Context, req *pb.GetCurrentStockRequest) (*pb.GetCurrentStockResponse, error) {
	qry := &query.GetCurrentStockQuery{
		BaseQuery: cqrs.BaseQuery{},
		VariantID: req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item variant not found")
	}

	stockResult := result.(*query.CurrentStockResult)

	return &pb.GetCurrentStockResponse{
		CurrentStock:  int32(stockResult.Stock),
		NeedsRestock:  stockResult.NeedsRestock,
		IsOverstocked: stockResult.IsOverstocked,
	}, nil
}

func (h *ItemVariantHandler) GetLowStockVariants(ctx context.Context, req *pb.GetLowStockVariantsRequest) (*pb.GetLowStockVariantsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	qry := &query.GetLowStockVariantsQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      int(page),
		PageSize:  int(pageSize),
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListItemVariantsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetLowStockVariantsResponse{
		ItemVariants: mapper.ItemVariantsToProto(listResult.ItemVariants),
		Total:        int32(listResult.Total),
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   int32(totalPages),
	}, nil
}

func (h *ItemVariantHandler) GetOverstockedVariants(ctx context.Context, req *pb.GetOverstockedVariantsRequest) (*pb.GetOverstockedVariantsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	qry := &query.GetOverstockedVariantsQuery{
		BaseQuery: cqrs.BaseQuery{},
		Page:      int(page),
		PageSize:  int(pageSize),
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListItemVariantsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetOverstockedVariantsResponse{
		ItemVariants: mapper.ItemVariantsToProto(listResult.ItemVariants),
		Total:        int32(listResult.Total),
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   int32(totalPages),
	}, nil
}

func (h *ItemVariantHandler) ScanBarcode(ctx context.Context, req *pb.ScanBarcodeRequest) (*pb.ScanBarcodeResponse, error) {
	qry := &query.GetItemVariantByBarcodeQuery{
		BaseQuery: cqrs.BaseQuery{},
		Barcode:   req.Barcode,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "item variant not found for barcode")
	}

	variant := result.(*domain.ItemVariant)

	return &pb.ScanBarcodeResponse{
		ItemVariant: mapper.ItemVariantToProto(variant),
	}, nil
}
