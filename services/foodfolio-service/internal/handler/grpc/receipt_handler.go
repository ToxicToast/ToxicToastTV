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
	// Create receipt with image path (simplified - actual file upload would be handled separately)
	imagePath := "uploads/receipts/" + time.Now().Format("20060102150405") + ".jpg"

	receipt, err := h.receiptUC.CreateReceipt(
		ctx,
		req.WarehouseId,
		time.Now(),
		0.0, // Will be extracted from OCR
		&imagePath,
		nil, // OCR text will be added later
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UploadReceiptResponse{
		Receipt:  mapper.ReceiptToProto(receipt),
		UploadId: receipt.ID,
	}, nil
}

func (h *ReceiptHandler) ProcessReceipt(ctx context.Context, req *pb.ProcessReceiptRequest) (*pb.ProcessReceiptResponse, error) {
	receipt, err := h.receiptUC.GetReceiptByID(ctx, req.ReceiptId)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// In a real implementation, this would trigger OCR processing
	// For now, just return the receipt
	return &pb.ProcessReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
		OcrText: "",
	}, nil
}

func (h *ReceiptHandler) CreateReceipt(ctx context.Context, req *pb.CreateReceiptRequest) (*pb.CreateReceiptResponse, error) {
	receipt, err := h.receiptUC.CreateReceipt(
		ctx,
		req.WarehouseId,
		req.ScanDate.AsTime(),
		req.TotalPrice,
		nil, // imagePath
		nil, // ocrText
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

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	receipts, total, err := h.receiptUC.ListReceipts(ctx, page, pageSize, warehouseID, startDate, endDate, includeDeleted)
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

	warehouseID := existing.WarehouseID
	if req.WarehouseId != nil {
		warehouseID = *req.WarehouseId
	}

	scanDate := existing.ScanDate
	if req.ScanDate != nil {
		scanDate = req.ScanDate.AsTime()
	}

	totalPrice := existing.TotalPrice
	if req.TotalPrice != nil {
		totalPrice = *req.TotalPrice
	}

	receipt, err := h.receiptUC.UpdateReceipt(ctx, req.Id, warehouseID, scanDate, totalPrice)
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
	var articleNumber *string
	if req.ArticleNumber != nil {
		articleNumber = req.ArticleNumber
	}

	var itemVariantID *string
	if req.ItemVariantId != nil {
		itemVariantID = req.ItemVariantId
	}

	totalPrice := req.UnitPrice * float64(req.Quantity)

	item, err := h.receiptUC.AddItemToReceipt(
		ctx,
		req.ReceiptId,
		req.ItemName,
		int(req.Quantity),
		req.UnitPrice,
		totalPrice,
		articleNumber,
		itemVariantID,
	)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddItemToReceiptResponse{
		Item: mapper.ReceiptItemToProto(item),
	}, nil
}

func (h *ReceiptHandler) UpdateReceiptItem(ctx context.Context, req *pb.UpdateReceiptItemRequest) (*pb.UpdateReceiptItemResponse, error) {
	// Set defaults for optional fields
	itemName := ""
	if req.ItemName != nil {
		itemName = *req.ItemName
	}

	quantity := 1
	if req.Quantity != nil {
		quantity = int(*req.Quantity)
	}

	unitPrice := 0.0
	if req.UnitPrice != nil {
		unitPrice = *req.UnitPrice
	}

	// Calculate total price
	totalPrice := unitPrice * float64(quantity)

	var articleNumber *string
	if req.ArticleNumber != nil {
		articleNumber = req.ArticleNumber
	}

	item, err := h.receiptUC.UpdateReceiptItem(ctx, req.Id, itemName, quantity, unitPrice, totalPrice, articleNumber)
	if err != nil {
		if err == usecase.ErrReceiptItemNotFound {
			return nil, status.Error(codes.NotFound, "receipt item not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateReceiptItemResponse{
		Item: mapper.ReceiptItemToProto(item),
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
		Item: mapper.ReceiptItemToProto(item),
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
	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		expiry := req.ExpiryDate.AsTime()
		expiryDate = &expiry
	}

	createdCount, err := h.receiptUC.CreateInventoryFromReceipt(ctx, req.ReceiptId, req.LocationId, expiryDate, req.OnlyMatched)
	if err != nil {
		if err == usecase.ErrReceiptNotFound {
			return nil, status.Error(codes.NotFound, "receipt not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateInventoryFromReceiptResponse{
		CreatedCount: int32(createdCount),
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
