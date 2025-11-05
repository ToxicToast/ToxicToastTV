package mapper

import (
	"toxictoast/services/foodfolio-service/internal/domain"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

// ShoppinglistToProto converts domain Shoppinglist to protobuf
func ShoppinglistToProto(list *domain.Shoppinglist) *pb.Shoppinglist {
	if list == nil {
		return nil
	}

	sl := &pb.Shoppinglist{
		Id:         list.ID,
		Name:       list.Name,
		Timestamps: ToTimestamps(list.CreatedAt, list.UpdatedAt, nil),
		ItemCount:  int32(len(list.Items)),
	}

	// Include items if loaded
	if list.Items != nil && len(list.Items) > 0 {
		sl.Items = ShoppinglistItemsToProto(list.Items)
	}

	return sl
}

// ShoppinglistsToProto converts slice of Shoppinglists
func ShoppinglistsToProto(lists []*domain.Shoppinglist) []*pb.Shoppinglist {
	result := make([]*pb.Shoppinglist, len(lists))
	for i, list := range lists {
		result[i] = ShoppinglistToProto(list)
	}
	return result
}

// ShoppinglistItemToProto converts domain ShoppinglistItem to protobuf
func ShoppinglistItemToProto(item *domain.ShoppinglistItem) *pb.ShoppinglistItem {
	if item == nil {
		return nil
	}

	si := &pb.ShoppinglistItem{
		Id:             item.ID,
		ShoppinglistId: item.ShoppinglistID,
		ItemVariantId:  item.ItemVariantID,
		Quantity:       int32(item.Quantity),
		IsPurchased:    item.IsPurchased,
		Timestamps:     ToTimestamps(item.CreatedAt, item.UpdatedAt, nil),
	}

	// Include relations if loaded
	if item.ItemVariant != nil {
		si.ItemVariant = ItemVariantToProto(item.ItemVariant)
	}

	return si
}

// ShoppinglistItemsToProto converts slice of ShoppinglistItems
func ShoppinglistItemsToProto(items []*domain.ShoppinglistItem) []*pb.ShoppinglistItem {
	result := make([]*pb.ShoppinglistItem, len(items))
	for i, item := range items {
		result[i] = ShoppinglistItemToProto(item)
	}
	return result
}
