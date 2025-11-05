package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio"
)

type ItemVariantHandler struct {
	pb.UnimplementedItemVariantServiceServer
	variantUC usecase.ItemVariantUseCase
}

func NewItemVariantHandler(variantUC usecase.ItemVariantUseCase) *ItemVariantHandler {
	return &ItemVariantHandler{
		variantUC: variantUC,
	}
}

func (h *ItemVariantHandler) CreateItemVariant(ctx context.Context, req *pb.CreateItemVariantRequest) (*pb.CreateItemVariantResponse, error) {
	var barcode *string
	if req.Barcode != nil {
		barcode = req.Barcode
	}

	variant, err := h.variantUC.CreateItemVariant(
		ctx,
		req.ItemId,
		req.SizeId,
		req.VariantName,
		barcode,
		int(req.MinSku),
		int(req.MaxSku),
		req.IsNormallyFrozen,
	)
	if err != nil {
		if err == usecase.ErrBarcodeAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "barcode already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateItemVariantResponse{
		ItemVariant: mapper.ItemVariantToProto(variant),
	}, nil
}

func (h *ItemVariantHandler) GetItemVariant(ctx context.Context, req *pb.IdRequest) (*pb.GetItemVariantResponse, error) {
	variant, err := h.variantUC.GetItemVariantByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemVariantNotFound {
			return nil, status.Error(codes.NotFound, "item variant not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

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

	variants, total, err := h.variantUC.ListItemVariants(ctx, int(page), int(pageSize), itemID, sizeID, isNormallyFrozen, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListItemVariantsResponse{
		ItemVariants: mapper.ItemVariantsToProto(variants),
		Total:        int32(total),
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   int32(totalPages),
	}, nil
}

func (h *ItemVariantHandler) UpdateItemVariant(ctx context.Context, req *pb.UpdateItemVariantRequest) (*pb.UpdateItemVariantResponse, error) {
	// Get existing to use as defaults
	existing, err := h.variantUC.GetItemVariantByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemVariantNotFound {
			return nil, status.Error(codes.NotFound, "item variant not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	var variantName string
	var barcode *string
	var minSku, maxSku int
	var isNormallyFrozen bool

	if req.VariantName != nil {
		variantName = *req.VariantName
	} else {
		variantName = existing.VariantName
	}

	if req.Barcode != nil {
		barcode = req.Barcode
	} else {
		barcode = existing.Barcode
	}

	if req.MinSku != nil {
		minSku = int(*req.MinSku)
	} else {
		minSku = existing.MinSKU
	}

	if req.MaxSku != nil {
		maxSku = int(*req.MaxSku)
	} else {
		maxSku = existing.MaxSKU
	}

	if req.IsNormallyFrozen != nil {
		isNormallyFrozen = *req.IsNormallyFrozen
	} else {
		isNormallyFrozen = existing.IsNormallyFrozen
	}

	variant, err := h.variantUC.UpdateItemVariant(ctx, req.Id, variantName, barcode, minSku, maxSku, isNormallyFrozen)
	if err != nil {
		if err == usecase.ErrItemVariantNotFound {
			return nil, status.Error(codes.NotFound, "item variant not found")
		}
		if err == usecase.ErrBarcodeAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "barcode already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateItemVariantResponse{
		ItemVariant: mapper.ItemVariantToProto(variant),
	}, nil
}

func (h *ItemVariantHandler) DeleteItemVariant(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	err := h.variantUC.DeleteItemVariant(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemVariantNotFound {
			return nil, status.Error(codes.NotFound, "item variant not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Item variant deleted successfully",
	}, nil
}

func (h *ItemVariantHandler) GetCurrentStock(ctx context.Context, req *pb.GetCurrentStockRequest) (*pb.GetCurrentStockResponse, error) {
	stock, needsRestock, isOverstocked, err := h.variantUC.GetCurrentStock(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrItemVariantNotFound {
			return nil, status.Error(codes.NotFound, "item variant not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetCurrentStockResponse{
		CurrentStock:   int32(stock),
		NeedsRestock:   needsRestock,
		IsOverstocked:  isOverstocked,
	}, nil
}

func (h *ItemVariantHandler) GetLowStockVariants(ctx context.Context, req *pb.GetLowStockVariantsRequest) (*pb.GetLowStockVariantsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	variants, total, err := h.variantUC.GetLowStockVariants(ctx, int(page), int(pageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetLowStockVariantsResponse{
		ItemVariants: mapper.ItemVariantsToProto(variants),
		Total:        int32(total),
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   int32(totalPages),
	}, nil
}

func (h *ItemVariantHandler) GetOverstockedVariants(ctx context.Context, req *pb.GetOverstockedVariantsRequest) (*pb.GetOverstockedVariantsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	variants, total, err := h.variantUC.GetOverstockedVariants(ctx, int(page), int(pageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.GetOverstockedVariantsResponse{
		ItemVariants: mapper.ItemVariantsToProto(variants),
		Total:        int32(total),
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   int32(totalPages),
	}, nil
}

func (h *ItemVariantHandler) ScanBarcode(ctx context.Context, req *pb.ScanBarcodeRequest) (*pb.ScanBarcodeResponse, error) {
	variant, err := h.variantUC.GetItemVariantByBarcode(ctx, req.Barcode)
	if err != nil {
		if err == usecase.ErrItemVariantNotFound {
			return nil, status.Error(codes.NotFound, "item variant not found for barcode")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ScanBarcodeResponse{
		ItemVariant: mapper.ItemVariantToProto(variant),
	}, nil
}
