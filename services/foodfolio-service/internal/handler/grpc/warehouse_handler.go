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

type WarehouseHandler struct {
	pb.UnimplementedWarehouseServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewWarehouseHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *WarehouseHandler {
	return &WarehouseHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *WarehouseHandler) CreateWarehouse(ctx context.Context, req *pb.CreateWarehouseRequest) (*pb.CreateWarehouseResponse, error) {
	cmd := &command.CreateWarehouseCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateWarehouseResponse{
		Warehouse: &pb.Warehouse{
			Name: req.Name,
		},
	}, nil
}

func (h *WarehouseHandler) GetWarehouse(ctx context.Context, req *pb.IdRequest) (*pb.GetWarehouseResponse, error) {
	qry := &query.GetWarehouseByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "warehouse not found")
	}

	warehouse := result.(*domain.Warehouse)

	return &pb.GetWarehouseResponse{
		Warehouse: mapper.WarehouseToProto(warehouse),
	}, nil
}

func (h *WarehouseHandler) ListWarehouses(ctx context.Context, req *pb.ListWarehousesRequest) (*pb.ListWarehousesResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	search := ""
	if req.Search != nil {
		search = *req.Search
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	qry := &query.ListWarehousesQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           int(page),
		PageSize:       int(pageSize),
		Search:         search,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListWarehousesResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListWarehousesResponse{
		Warehouses: mapper.WarehousesToProto(listResult.Warehouses),
		Total:      int32(listResult.Total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (h *WarehouseHandler) UpdateWarehouse(ctx context.Context, req *pb.UpdateWarehouseRequest) (*pb.UpdateWarehouseResponse, error) {
	cmd := &command.UpdateWarehouseCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated warehouse
	qry := &query.GetWarehouseByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "warehouse not found")
	}

	warehouse := result.(*domain.Warehouse)

	return &pb.UpdateWarehouseResponse{
		Warehouse: mapper.WarehouseToProto(warehouse),
	}, nil
}

func (h *WarehouseHandler) DeleteWarehouse(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteWarehouseCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Warehouse deleted successfully",
	}, nil
}
