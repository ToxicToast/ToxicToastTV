package mapper

import (
	"toxictoast/services/foodfolio-service/internal/domain"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

// ItemToProto converts domain Item to protobuf
func ItemToProto(item *domain.Item) *pb.Item {
	if item == nil {
		return nil
	}

	i := &pb.Item{
		Id:           item.ID,
		Name:         item.Name,
		Slug:         item.Slug,
		CategoryId:   item.CategoryID,
		CompanyId:    item.CompanyID,
		TypeId:       item.TypeID,
		Timestamps:   ToTimestamps(item.CreatedAt, item.UpdatedAt, nil),
		VariantCount: int32(len(item.ItemVariants)),
	}

	// Include relations if loaded
	if item.Category != nil {
		i.Category = CategoryToProto(item.Category)
	}

	if item.Company != nil {
		i.Company = CompanyToProto(item.Company)
	}

	if item.Type != nil {
		i.Type = TypeToProto(item.Type)
	}

	return i
}

// ItemsToProto converts slice of Items
func ItemsToProto(items []*domain.Item) []*pb.Item {
	result := make([]*pb.Item, len(items))
	for i, item := range items {
		result[i] = ItemToProto(item)
	}
	return result
}

// ItemVariantToProto converts domain ItemVariant to protobuf
func ItemVariantToProto(variant *domain.ItemVariant) *pb.ItemVariant {
	if variant == nil {
		return nil
	}

	v := &pb.ItemVariant{
		Id:               variant.ID,
		ItemId:           variant.ItemID,
		SizeId:           variant.SizeID,
		VariantName:      variant.VariantName,
		MinSku:           int32(variant.MinSKU),
		MaxSku:           int32(variant.MaxSKU),
		IsNormallyFrozen: variant.IsNormallyFrozen,
		Timestamps:       ToTimestamps(variant.CreatedAt, variant.UpdatedAt, nil),
		CurrentStock:     int32(variant.CurrentStock()),
		DetailCount:      int32(len(variant.ItemDetails)),
	}

	if variant.Barcode != nil {
		v.Barcode = variant.Barcode
	}

	// Include relations if loaded
	if variant.Item != nil {
		v.Item = ItemToProto(variant.Item)
	}

	if variant.Size != nil {
		v.Size = SizeToProto(variant.Size)
	}

	return v
}

// ItemVariantsToProto converts slice of ItemVariants
func ItemVariantsToProto(variants []*domain.ItemVariant) []*pb.ItemVariant {
	result := make([]*pb.ItemVariant, len(variants))
	for i, variant := range variants {
		result[i] = ItemVariantToProto(variant)
	}
	return result
}

// ItemDetailToProto converts domain ItemDetail to protobuf
func ItemDetailToProto(detail *domain.ItemDetail) *pb.ItemDetail {
	if detail == nil {
		return nil
	}

	d := &pb.ItemDetail{
		Id:            detail.ID,
		ItemVariantId: detail.ItemVariantID,
		WarehouseId:   detail.WarehouseID,
		LocationId:    detail.LocationID,
		PurchasePrice: detail.PurchasePrice,
		PurchaseDate:  TimeToProto(detail.PurchaseDate),
		IsOpened:      detail.IsOpened,
		HasDeposit:    detail.HasDeposit,
		IsFrozen:      detail.IsFrozen,
		Timestamps:    ToTimestamps(detail.CreatedAt, detail.UpdatedAt, nil),
		IsExpired:     detail.IsExpired(),
		IsExpiringSoon: detail.IsExpiringSoon(7),
		IsConsumed:    detail.IsConsumed(),
	}

	if detail.ArticleNumber != nil {
		d.ArticleNumber = detail.ArticleNumber
	}

	if detail.ExpiryDate != nil {
		d.ExpiryDate = TimeToProto(*detail.ExpiryDate)
	}

	if detail.OpenedDate != nil {
		d.OpenedDate = TimeToProto(*detail.OpenedDate)
	}

	// Include relations if loaded
	if detail.ItemVariant != nil {
		d.ItemVariant = ItemVariantToProto(detail.ItemVariant)
	}

	if detail.Warehouse != nil {
		d.Warehouse = WarehouseToProto(detail.Warehouse)
	}

	if detail.Location != nil {
		d.Location = LocationToProto(detail.Location)
	}

	return d
}

// ItemDetailsToProto converts slice of ItemDetails
func ItemDetailsToProto(details []*domain.ItemDetail) []*pb.ItemDetail {
	result := make([]*pb.ItemDetail, len(details))
	for i, detail := range details {
		result[i] = ItemDetailToProto(detail)
	}
	return result
}
