package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

type WarehouseHandler struct {
	pb.UnimplementedWarehouseServiceServer
	warehouseUC usecase.WarehouseUseCase
}

func NewWarehouseHandler(warehouseUC usecase.WarehouseUseCase) *WarehouseHandler {
	return &WarehouseHandler{
		warehouseUC: warehouseUC,
	}
}

func (h *WarehouseHandler) CreateWarehouse(ctx context.Context, req *pb.CreateWarehouseRequest) (*pb.CreateWarehouseResponse, error) {
	warehouse, err := h.warehouseUC.CreateWarehouse(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateWarehouseResponse{
		Warehouse: mapper.WarehouseToProto(warehouse),
	}, nil
}

func (h *WarehouseHandler) GetWarehouse(ctx context.Context, req *pb.IdRequest) (*pb.GetWarehouseResponse, error) {
	warehouse, err := h.warehouseUC.GetWarehouseByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrWarehouseNotFound {
			return nil, status.Error(codes.NotFound, "warehouse not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetWarehouseResponse{
		Warehouse: mapper.WarehouseToProto(warehouse),
	}, nil
}

func (h *WarehouseHandler) ListWarehouses(ctx context.Context, req *pb.ListWarehousesRequest) (*pb.ListWarehousesResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	search := ""
	if req.Search != nil {
		search = *req.Search
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	warehouses, total, err := h.warehouseUC.ListWarehouses(ctx, page, pageSize, search, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListWarehousesResponse{
		Warehouses: mapper.WarehousesToProto(warehouses),
		Pagination: mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}

func (h *WarehouseHandler) UpdateWarehouse(ctx context.Context, req *pb.UpdateWarehouseRequest) (*pb.UpdateWarehouseResponse, error) {
	warehouse, err := h.warehouseUC.UpdateWarehouse(ctx, req.Id, req.Name)
	if err != nil {
		if err == usecase.ErrWarehouseNotFound {
			return nil, status.Error(codes.NotFound, "warehouse not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateWarehouseResponse{
		Warehouse: mapper.WarehouseToProto(warehouse),
	}, nil
}

func (h *WarehouseHandler) DeleteWarehouse(ctx context.Context, req *pb.IdRequest) (*pb.SuccessResponse, error) {
	err := h.warehouseUC.DeleteWarehouse(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrWarehouseNotFound {
			return nil, status.Error(codes.NotFound, "warehouse not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SuccessResponse{
		Success: true,
		Message: "Warehouse deleted successfully",
	}, nil
}
