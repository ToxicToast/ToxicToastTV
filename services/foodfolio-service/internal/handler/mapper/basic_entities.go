package mapper

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"toxictoast/services/foodfolio-service/internal/domain"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio"
)

// TypeToProto converts domain Type to protobuf
func TypeToProto(t *domain.Type) *pb.Type {
	if t == nil {
		return nil
	}

	return &pb.Type{
		Id:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
		DeletedAt: timestampOrNil(nil),
		ItemCount: 0,
	}
}

// TypesToProto converts slice of Types
func TypesToProto(types []*domain.Type) []*pb.Type {
	result := make([]*pb.Type, len(types))
	for i, t := range types {
		result[i] = TypeToProto(t)
	}
	return result
}

// SizeToProto converts domain Size to protobuf
func SizeToProto(size *domain.Size) *pb.Size {
	if size == nil {
		return nil
	}

	return &pb.Size{
		Id:           size.ID,
		Name:         size.Name,
		Value:        size.Value,
		Unit:         size.Unit,
		CreatedAt:    timestamppb.New(size.CreatedAt),
		UpdatedAt:    timestamppb.New(size.UpdatedAt),
		DeletedAt:    timestampOrNil(nil),
		VariantCount: 0,
	}
}

// SizesToProto converts slice of Sizes
func SizesToProto(sizes []*domain.Size) []*pb.Size {
	result := make([]*pb.Size, len(sizes))
	for i, size := range sizes {
		result[i] = SizeToProto(size)
	}
	return result
}

// WarehouseToProto converts domain Warehouse to protobuf
func WarehouseToProto(warehouse *domain.Warehouse) *pb.Warehouse {
	if warehouse == nil {
		return nil
	}

	return &pb.Warehouse{
		Id:            warehouse.ID,
		Name:          warehouse.Name,
		Slug:          warehouse.Slug,
		CreatedAt:     timestamppb.New(warehouse.CreatedAt),
		UpdatedAt:     timestamppb.New(warehouse.UpdatedAt),
		DeletedAt:     timestampOrNil(nil),
		PurchaseCount: 0,
	}
}

// WarehousesToProto converts slice of Warehouses
func WarehousesToProto(warehouses []*domain.Warehouse) []*pb.Warehouse {
	result := make([]*pb.Warehouse, len(warehouses))
	for i, warehouse := range warehouses {
		result[i] = WarehouseToProto(warehouse)
	}
	return result
}

// CategoryToProto converts domain Category to protobuf
func CategoryToProto(category *domain.Category) *pb.Category {
	if category == nil {
		return nil
	}

	cat := &pb.Category{
		Id:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		CreatedAt: timestamppb.New(category.CreatedAt),
		UpdatedAt: timestamppb.New(category.UpdatedAt),
		DeletedAt: timestampOrNil(nil),
		ItemCount: 0,
	}

	if category.ParentID != nil {
		parentID := *category.ParentID
		cat.ParentId = &parentID
	}

	// Include parent if loaded
	if category.Parent != nil {
		cat.Parent = CategoryToProto(category.Parent)
	}

	// Include children if loaded
	if len(category.Children) > 0 {
		// Convert []Category to []*Category
		childrenPtrs := make([]*domain.Category, len(category.Children))
		for i := range category.Children {
			childrenPtrs[i] = &category.Children[i]
		}
		cat.Children = CategoriesToProto(childrenPtrs)
	}

	return cat
}

// CategoriesToProto converts slice of Categories
func CategoriesToProto(categories []*domain.Category) []*pb.Category {
	result := make([]*pb.Category, len(categories))
	for i, category := range categories {
		result[i] = CategoryToProto(category)
	}
	return result
}

// LocationToProto converts domain Location to protobuf
func LocationToProto(location *domain.Location) *pb.Location {
	if location == nil {
		return nil
	}

	loc := &pb.Location{
		Id:        location.ID,
		Name:      location.Name,
		Slug:      location.Slug,
		CreatedAt: timestamppb.New(location.CreatedAt),
		UpdatedAt: timestamppb.New(location.UpdatedAt),
		DeletedAt: timestampOrNil(nil),
		ItemCount: 0,
	}

	if location.ParentID != nil {
		parentID := *location.ParentID
		loc.ParentId = &parentID
	}

	// Include parent if loaded
	if location.Parent != nil {
		loc.Parent = LocationToProto(location.Parent)
	}

	// Include children if loaded
	if len(location.Children) > 0 {
		// Convert []Location to []*Location
		childrenPtrs := make([]*domain.Location, len(location.Children))
		for i := range location.Children {
			childrenPtrs[i] = &location.Children[i]
		}
		loc.Children = LocationsToProto(childrenPtrs)
	}

	return loc
}

// LocationsToProto converts slice of Locations
func LocationsToProto(locations []*domain.Location) []*pb.Location {
	result := make([]*pb.Location, len(locations))
	for i, location := range locations {
		result[i] = LocationToProto(location)
	}
	return result
}
