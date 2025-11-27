package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	pb "toxictoast/services/foodfolio-service/api/proto"
	"toxictoast/services/foodfolio-service/internal/command"
	"toxictoast/services/foodfolio-service/internal/domain"
	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/query"
	"toxictoast/services/foodfolio-service/pkg/ocr"
)

type ReceiptHandler struct {
	pb.UnimplementedReceiptServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewReceiptHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *ReceiptHandler {
	return &ReceiptHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *ReceiptHandler) UploadReceipt(ctx context.Context, req *pb.UploadReceiptRequest) (*pb.UploadReceiptResponse, error) {
	// Validate input
	if req.ImageData == nil || len(req.ImageData) == 0 {
		return nil, status.Error(codes.InvalidArgument, "image_data is required")
	}

	// Create receipt first to get ID
	cmd := &command.CreateReceiptCommand{
		BaseCommand: cqrs.BaseCommand{},
		WarehouseID: req.WarehouseId,
		ScanDate:    time.Now(),
		TotalPrice:  0.0, // Will be updated after OCR
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create receipt: %v", err))
	}

	receiptID := cmd.AggregateID

	// Save image file
	imagePath, err := ocr.SaveImageFile(req.ImageData, receiptID)
	if err != nil {
		log.Printf("Failed to save image file for receipt %s: %v", receiptID, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to save image: %v", err))
	}

	log.Printf("Saved receipt image to: %s", imagePath)

	// Perform OCR
	ocrData, err := ocr.ParseReceipt(imagePath)
	if err != nil {
		log.Printf("OCR failed for receipt %s: %v", receiptID, err)
		// Still update with image path even if OCR fails
		updateCmd := &command.UpdateReceiptOCRDataCommand{
			BaseCommand: cqrs.BaseCommand{AggregateID: receiptID},
			ImagePath:   imagePath,
			OCRText:     "",
			TotalPrice:  0.0,
		}
		if updateErr := h.commandBus.Dispatch(ctx, updateCmd); updateErr != nil {
			log.Printf("Failed to update receipt with image path: %v", updateErr)
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("OCR failed: %v", err))
	}

	log.Printf("OCR completed for receipt %s: found %d items, total: %.2f", receiptID, len(ocrData.Items), ocrData.TotalPrice)

	// Update receipt with OCR data
	updateCmd := &command.UpdateReceiptOCRDataCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: receiptID},
		ImagePath:   imagePath,
		OCRText:     ocrData.RawText,
		TotalPrice:  ocrData.TotalPrice,
	}

	if err := h.commandBus.Dispatch(ctx, updateCmd); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update receipt: %v", err))
	}

	// Create receipt items
	for _, item := range ocrData.Items {
		var articleNumber *string
		if item.ArticleNumber != "" {
			articleNumber = &item.ArticleNumber
		}

		addItemCmd := &command.AddItemToReceiptCommand{
			BaseCommand:   cqrs.BaseCommand{},
			ReceiptID:     receiptID,
			ItemName:      item.Name,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			TotalPrice:    item.TotalPrice,
			ArticleNumber: articleNumber,
			ItemVariantID: nil, // Will be auto-matched next
		}

		if err := h.commandBus.Dispatch(ctx, addItemCmd); err != nil {
			log.Printf("Failed to add item '%s' to receipt %s: %v", item.Name, receiptID, err)
			// Continue with other items
		}
	}

	// Auto-match receipt items with existing variants
	autoMatchCmd := &command.AutoMatchReceiptItemsCommand{
		BaseCommand:         cqrs.BaseCommand{},
		ReceiptID:           receiptID,
		SimilarityThreshold: 0.7,
	}

	if err := h.commandBus.Dispatch(ctx, autoMatchCmd); err != nil {
		log.Printf("Auto-matching failed for receipt %s: %v", receiptID, err)
	} else {
		log.Printf("Auto-matching completed for receipt %s: %d matched, %d unmatched", receiptID, autoMatchCmd.MatchedCount, autoMatchCmd.UnmatchedCount)
	}

	// Reload receipt with items
	qry := &query.GetReceiptByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        receiptID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to reload receipt: %v", err))
	}

	receipt := result.(*domain.Receipt)

	return &pb.UploadReceiptResponse{
		Receipt:  mapper.ReceiptToProto(receipt),
		UploadId: receipt.ID,
	}, nil
}

func (h *ReceiptHandler) ProcessReceipt(ctx context.Context, req *pb.ProcessReceiptRequest) (*pb.ProcessReceiptResponse, error) {
	qry := &query.GetReceiptByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.ReceiptId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "receipt not found")
	}

	receipt := result.(*domain.Receipt)

	// In a real implementation, this would trigger OCR processing
	// For now, just return the receipt
	ocrText := ""
	if receipt.OCRText != nil {
		ocrText = *receipt.OCRText
	}

	return &pb.ProcessReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
		OcrText: ocrText,
	}, nil
}

func (h *ReceiptHandler) CreateReceipt(ctx context.Context, req *pb.CreateReceiptRequest) (*pb.CreateReceiptResponse, error) {
	cmd := &command.CreateReceiptCommand{
		BaseCommand: cqrs.BaseCommand{},
		WarehouseID: req.WarehouseId,
		ScanDate:    req.ScanDate.AsTime(),
		TotalPrice:  req.TotalPrice,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the created receipt
	qry := &query.GetReceiptByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "receipt not found")
	}

	receipt := result.(*domain.Receipt)

	return &pb.CreateReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
	}, nil
}

func (h *ReceiptHandler) GetReceipt(ctx context.Context, req *pb.IdRequest) (*pb.GetReceiptResponse, error) {
	qry := &query.GetReceiptByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "receipt not found")
	}

	receipt := result.(*domain.Receipt)

	return &pb.GetReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
	}, nil
}

func (h *ReceiptHandler) ListReceipts(ctx context.Context, req *pb.ListReceiptsRequest) (*pb.ListReceiptsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

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

	qry := &query.ListReceiptsQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           int(page),
		PageSize:       int(pageSize),
		WarehouseID:    warehouseID,
		StartDate:      startDate,
		EndDate:        endDate,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListReceiptsResult)

	// Calculate total amount
	var totalAmount float64
	for _, receipt := range listResult.Receipts {
		totalAmount += receipt.TotalPrice
	}

	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListReceiptsResponse{
		Receipts:    mapper.ReceiptsToProto(listResult.Receipts),
		Total:       int32(listResult.Total),
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  int32(totalPages),
		TotalAmount: totalAmount,
	}, nil
}

func (h *ReceiptHandler) UpdateReceipt(ctx context.Context, req *pb.UpdateReceiptRequest) (*pb.UpdateReceiptResponse, error) {
	// Get existing receipt to use as defaults
	qry := &query.GetReceiptByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "receipt not found")
	}

	existing := result.(*domain.Receipt)

	// Use existing values as defaults
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

	cmd := &command.UpdateReceiptCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		WarehouseID: warehouseID,
		ScanDate:    scanDate,
		TotalPrice:  totalPrice,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated receipt
	result, err = h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "receipt not found")
	}

	receipt := result.(*domain.Receipt)

	return &pb.UpdateReceiptResponse{
		Receipt: mapper.ReceiptToProto(receipt),
	}, nil
}

func (h *ReceiptHandler) DeleteReceipt(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteReceiptCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
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

	cmd := &command.AddItemToReceiptCommand{
		BaseCommand:   cqrs.BaseCommand{},
		ReceiptID:     req.ReceiptId,
		ItemName:      req.ItemName,
		Quantity:      int(req.Quantity),
		UnitPrice:     req.UnitPrice,
		TotalPrice:    totalPrice,
		ArticleNumber: articleNumber,
		ItemVariantID: itemVariantID,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: Query individual item when GetReceiptItemByIDQuery is implemented
	// For now, return a basic response
	return &pb.AddItemToReceiptResponse{
		Item: &pb.ReceiptItem{
			Id:            cmd.AggregateID,
			ReceiptId:     req.ReceiptId,
			ItemName:      req.ItemName,
			Quantity:      req.Quantity,
			UnitPrice:     req.UnitPrice,
			TotalPrice:    totalPrice,
			ArticleNumber: articleNumber,
			ItemVariantId: itemVariantID,
		},
	}, nil
}

func (h *ReceiptHandler) UpdateReceiptItem(ctx context.Context, req *pb.UpdateReceiptItemRequest) (*pb.UpdateReceiptItemResponse, error) {
	// UpdateReceiptItemCommand expects non-pointer types
	// We need to get existing item first or use defaults
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

	totalPrice := unitPrice * float64(quantity)

	cmd := &command.UpdateReceiptItemCommand{
		BaseCommand:   cqrs.BaseCommand{AggregateID: req.Id},
		ItemName:      itemName,
		Quantity:      quantity,
		UnitPrice:     unitPrice,
		TotalPrice:    totalPrice,
		ArticleNumber: req.ArticleNumber,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: Query individual item when GetReceiptItemByIDQuery is implemented
	// For now, return a basic response
	return &pb.UpdateReceiptItemResponse{
		Item: &pb.ReceiptItem{
			Id:            req.Id,
			ItemName:      itemName,
			Quantity:      int32(quantity),
			UnitPrice:     unitPrice,
			TotalPrice:    totalPrice,
			ArticleNumber: req.ArticleNumber,
		},
	}, nil
}

func (h *ReceiptHandler) MatchReceiptItem(ctx context.Context, req *pb.MatchReceiptItemRequest) (*pb.MatchReceiptItemResponse, error) {
	cmd := &command.MatchReceiptItemCommand{
		BaseCommand:   cqrs.BaseCommand{AggregateID: req.ReceiptItemId},
		ItemVariantID: req.ItemVariantId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: Query individual item when GetReceiptItemByIDQuery is implemented
	// For now, return a basic response
	return &pb.MatchReceiptItemResponse{
		Item: &pb.ReceiptItem{
			Id:            req.ReceiptItemId,
			ItemVariantId: &req.ItemVariantId,
		},
	}, nil
}

func (h *ReceiptHandler) AutoMatchReceiptItems(ctx context.Context, req *pb.AutoMatchReceiptItemsRequest) (*pb.AutoMatchReceiptItemsResponse, error) {
	similarityThreshold := float64(req.SimilarityThreshold)
	if similarityThreshold <= 0 {
		similarityThreshold = 0.7 // Default threshold
	}

	cmd := &command.AutoMatchReceiptItemsCommand{
		BaseCommand:         cqrs.BaseCommand{},
		ReceiptID:           req.ReceiptId,
		SimilarityThreshold: similarityThreshold,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AutoMatchReceiptItemsResponse{
		MatchedCount:   int32(cmd.MatchedCount),
		UnmatchedCount: int32(cmd.UnmatchedCount),
	}, nil
}

func (h *ReceiptHandler) CreateInventoryFromReceipt(ctx context.Context, req *pb.CreateInventoryFromReceiptRequest) (*pb.CreateInventoryFromReceiptResponse, error) {
	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		expiry := req.ExpiryDate.AsTime()
		expiryDate = &expiry
	}

	cmd := &command.CreateInventoryFromReceiptCommand{
		BaseCommand:       cqrs.BaseCommand{},
		ReceiptID:         req.ReceiptId,
		LocationID:        req.LocationId,
		DefaultExpiryDate: expiryDate,
		OnlyMatched:       req.OnlyMatched,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateInventoryFromReceiptResponse{
		CreatedCount: int32(cmd.CreatedCount),
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

	qry := &query.GetStatisticsQuery{
		BaseQuery:   cqrs.BaseQuery{},
		WarehouseID: warehouseID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	stats := result.(map[string]interface{})

	return &pb.GetReceiptStatisticsResponse{
		TotalReceipts:          int32(stats["total_receipts"].(int64)),
		TotalAmount:            stats["total_amount"].(float64),
		AverageAmount:          stats["average_amount"].(float64),
		TotalItems:             int32(stats["total_items"].(int64)),
		AverageItemsPerReceipt: stats["average_items"].(float64),
	}, nil
}
