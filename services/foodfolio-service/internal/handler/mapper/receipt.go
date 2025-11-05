package mapper

import (
	"toxictoast/services/foodfolio-service/internal/domain"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

// ReceiptToProto converts domain Receipt to protobuf
func ReceiptToProto(receipt *domain.Receipt) *pb.Receipt {
	if receipt == nil {
		return nil
	}

	r := &pb.Receipt{
		Id:                  receipt.ID,
		WarehouseId:         receipt.WarehouseID,
		PurchaseDate:        TimeToProto(receipt.PurchaseDate),
		TotalAmount:         receipt.TotalAmount,
		IsProcessed:         receipt.IsProcessed,
		IsInventoryCreated:  receipt.IsInventoryCreated,
		Timestamps:          ToTimestamps(receipt.CreatedAt, receipt.UpdatedAt, nil),
		ItemCount:           int32(len(receipt.Items)),
	}

	if receipt.ImagePath != nil {
		r.ImagePath = receipt.ImagePath
	}

	if receipt.OcrText != nil {
		r.OcrText = receipt.OcrText
	}

	// Include relations if loaded
	if receipt.Warehouse != nil {
		r.Warehouse = WarehouseToProto(receipt.Warehouse)
	}

	if receipt.Items != nil && len(receipt.Items) > 0 {
		r.Items = ReceiptItemsToProto(receipt.Items)
	}

	return r
}

// ReceiptsToProto converts slice of Receipts
func ReceiptsToProto(receipts []*domain.Receipt) []*pb.Receipt {
	result := make([]*pb.Receipt, len(receipts))
	for i, receipt := range receipts {
		result[i] = ReceiptToProto(receipt)
	}
	return result
}

// ReceiptItemToProto converts domain ReceiptItem to protobuf
func ReceiptItemToProto(item *domain.ReceiptItem) *pb.ReceiptItem {
	if item == nil {
		return nil
	}

	ri := &pb.ReceiptItem{
		Id:            item.ID,
		ReceiptId:     item.ReceiptID,
		ItemName:      item.ItemName,
		Quantity:      int32(item.Quantity),
		UnitPrice:     item.UnitPrice,
		TotalPrice:    item.TotalPrice,
		IsMatched:     item.IsMatched,
		Timestamps:    ToTimestamps(item.CreatedAt, item.UpdatedAt, nil),
	}

	if item.ItemVariantID != nil {
		ri.ItemVariantId = item.ItemVariantID
	}

	// Include relations if loaded
	if item.ItemVariant != nil {
		ri.ItemVariant = ItemVariantToProto(item.ItemVariant)
	}

	return ri
}

// ReceiptItemsToProto converts slice of ReceiptItems
func ReceiptItemsToProto(items []*domain.ReceiptItem) []*pb.ReceiptItem {
	result := make([]*pb.ReceiptItem, len(items))
	for i, item := range items {
		result[i] = ReceiptItemToProto(item)
	}
	return result
}
