package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

type ReceiptHandler struct {
	pb.UnimplementedReceiptServiceServer
	receiptUC usecase.ReceiptUseCase
}

func NewReceiptHandler(receiptUC usecase.ReceiptUseCase) *ReceiptHandler {
	return &ReceiptHandler{
		receiptUC: receiptUC,
	}
}

func (h *ReceiptHandler) UploadReceipt(ctx context.Context, req *pb.UploadReceiptRequest) (*pb.UploadReceiptResponse, error) {
	receipt, err := h.receiptUC.UploadReceipt(ctx, req.WarehouseId, req.ImageData, req.Filename)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UploadReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
	}, nil
}

func (h *ReceiptHandler) ProcessReceipt(ctx context.Context, req *pb.IdRequest) (*pb.ProcessReceiptResponse, error) {
	receipt, err := h.receiptUC.ProcessReceipt(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ProcessReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
	}, nil
}

func (h *ReceiptHandler) CreateReceipt(ctx context.Context, req *pb.CreateReceiptRequest) (*pb.CreateReceiptResponse, error) {
	var imagePath *string
	if req.ImagePath != nil {
		imagePath = req.ImagePath
	}

	receipt, err := h.receiptUC.CreateReceipt(
		ctx,
		req.WarehouseId,
		req.PurchaseDate.AsTime(),
		req.TotalAmount,
		imagePath,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
	}, nil
}

func (h *ReceiptHandler) GetReceipt(ctx context.Context, req *pb.IdRequest) (*pb.GetReceiptResponse, error) {
	receipt, err := h.receiptUC.GetReceiptByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
	}, nil
}

func (h *ReceiptHandler) ListReceipts(ctx context.Context, req *pb.ListReceiptsRequest) (*pb.ListReceiptsResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	var warehouseID *string
	if req.WarehouseId != nil {
		warehouseID = req.WarehouseId
	}

	var startDate, endDate *time.Time
	if req.StartDate != nil {
		start := req.StartDate.AsTime()
		startDate = &start
	}
	if req.EndDate != nil {
		end := req.EndDate.AsTime()
		endDate = &end
	}

	var isProcessed, isInventoryCreated *bool
	if req.IsProcessed != nil {
		isProcessed = req.IsProcessed
	}
	if req.IsInventoryCreated != nil {
		isInventoryCreated = req.IsInventoryCreated
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	receipts, total, err := h.receiptUC.ListReceipts(ctx, page, pageSize, warehouseID, startDate, endDate, isProcessed, isInventoryCreated, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListReceiptsResponse{
		Receipts:   mapper.ReceiptsToProto(receipts),
		Pagination: mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}

func (h *ReceiptHandler) UpdateReceipt(ctx context.Context, req *pb.UpdateReceiptRequest) (*pb.UpdateReceiptResponse, error) {
	// Get existing to use as defaults
	existing, err := h.receiptUC.GetReceiptByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	purchaseDate := existing.PurchaseDate
	if req.PurchaseDate != nil {
		purchaseDate = req.PurchaseDate.AsTime()
	}

	totalAmount := existing.TotalAmount
	if req.TotalAmount != nil {
		totalAmount = *req.TotalAmount
	}

	receipt, err := h.receiptUC.UpdateReceipt(ctx, req.Id, purchaseDate, totalAmount)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
	}, nil
}

func (h *ReceiptHandler) DeleteReceipt(ctx context.Context, req *pb.IdRequest) (*pb.SuccessResponse, error) {
	err := h.receiptUC.DeleteReceipt(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SuccessResponse{
		Success: true,
		Message: "Receipt deleted successfully",
	}, nil
}

func (h *ReceiptHandler) AddItemToReceipt(ctx context.Context, req *pb.AddItemToReceiptRequest) (*pb.AddItemToReceiptResponse, error) {
	item, err := h.receiptUC.AddItemToReceipt(
		ctx,
		req.ReceiptId,
		req.ItemName,
		int(req.Quantity),
		req.UnitPrice,
		req.TotalPrice,
	)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddItemToReceiptResponse{
		ReceiptItem: mapper.ReceiptItemToProto(item),
	}, nil
}

func (h *ReceiptHandler) UpdateReceiptItem(ctx context.Context, req *pb.UpdateReceiptItemRequest) (*pb.UpdateReceiptItemResponse, error) {
	// Get existing to use as defaults
	existing, err := h.receiptUC.GetReceiptItemByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrReceiptItemNotFound {
			return nil, status.Error(codes.NotFound, "receipt item not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	itemName := existing.ItemName
	if req.ItemName != nil {
		itemName = *req.ItemName
	}

	quantity := existing.Quantity
	if req.Quantity != nil {
		quantity = int(*req.Quantity)
	}

	unitPrice := existing.UnitPrice
	if req.UnitPrice != nil {
		unitPrice = *req.UnitPrice
	}

	totalPrice := existing.TotalPrice
	if req.TotalPrice != nil {
		totalPrice = *req.TotalPrice
	}

	item, err := h.receiptUC.UpdateReceiptItem(ctx, req.Id, itemName, quantity, unitPrice, totalPrice)
	if err != nil {
		if err == usecase.ErrReceiptItemNotFound {
			return nil, status.Error(codes.NotFound, "receipt item not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateReceiptItemResponse{
		ReceiptItem: mapper.ReceiptItemToProto(item),
	}, nil
}

func (h *ReceiptHandler) MatchReceiptItem(ctx context.Context, req *pb.MatchReceiptItemRequest) (*pb.MatchReceiptItemResponse, error) {
	item, err := h.receiptUC.MatchReceiptItem(ctx, req.ReceiptItemId, req.ItemVariantId)
	if err != nil {
		if err == usecase.ErrReceiptItemNotFound {
			return nil, status.Error(codes.NotFound, "receipt item not found")
		}
		if err == usecase.ErrItemVariantNotFound {
			return nil, status.Error(codes.NotFound, "item variant not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MatchReceiptItemResponse{
		ReceiptItem: mapper.ReceiptItemToProto(item),
	}, nil
}

func (h *ReceiptHandler) AutoMatchReceiptItems(ctx context.Context, req *pb.AutoMatchReceiptItemsRequest) (*pb.AutoMatchReceiptItemsResponse, error) {
	similarityThreshold := float64(req.SimilarityThreshold)
	if similarityThreshold <= 0 {
		similarityThreshold = 0.7 // Default threshold
	}

	matched, unmatched, err := h.receiptUC.AutoMatchReceiptItems(ctx, req.ReceiptId, similarityThreshold)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AutoMatchReceiptItemsResponse{
		MatchedCount:   int32(matched),
		UnmatchedCount: int32(unmatched),
	}, nil
}

func (h *ReceiptHandler) CreateInventoryFromReceipt(ctx context.Context, req *pb.CreateInventoryFromReceiptRequest) (*pb.CreateInventoryFromReceiptResponse, error) {
	createdCount, skippedCount, err := h.receiptUC.CreateInventoryFromReceipt(ctx, req.ReceiptId, req.LocationId)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateInventoryFromReceiptResponse{
		Success:      true,
		Message:      "Inventory created from receipt",
		CreatedCount: int32(createdCount),
		SkippedCount: int32(skippedCount),
	}, nil
}

func (h *ReceiptHandler) GetReceiptStatistics(ctx context.Context, req *pb.GetReceiptStatisticsRequest) (*pb.GetReceiptStatisticsResponse, error) {
	var warehouseID *string
	if req.WarehouseId != nil {
		warehouseID = req.WarehouseId
	}

	var startDate, endDate *time.Time
	if req.StartDate != nil {
		start := req.StartDate.AsTime()
		startDate = &start
	}
	if req.EndDate != nil {
		end := req.EndDate.AsTime()
		endDate = &end
	}

	stats, err := h.receiptUC.GetStatistics(ctx, warehouseID, startDate, endDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &pb.GetReceiptStatisticsResponse{
		TotalReceipts:      int32(stats["total_receipts"].(int64)),
		TotalAmount:        stats["total_amount"].(float64),
		AverageAmount:      stats["average_amount"].(float64),
		TotalItems:         int32(stats["total_items"].(int64)),
		AverageItemsPerReceipt: stats["average_items"].(float64),
	}

	return response, nil
}
