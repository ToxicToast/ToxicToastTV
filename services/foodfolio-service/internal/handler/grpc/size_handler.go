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

type SizeHandler struct {
	pb.UnimplementedSizeServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewSizeHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *SizeHandler {
	return &SizeHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *SizeHandler) CreateSize(ctx context.Context, req *pb.CreateSizeRequest) (*pb.CreateSizeResponse, error) {
	cmd := &command.CreateSizeCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
		Value:       req.Value,
		Unit:        req.Unit,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateSizeResponse{
		Size: &pb.Size{
			Name:  req.Name,
			Value: req.Value,
			Unit:  req.Unit,
		},
	}, nil
}

func (h *SizeHandler) GetSize(ctx context.Context, req *pb.IdRequest) (*pb.GetSizeResponse, error) {
	qry := &query.GetSizeByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "size not found")
	}

	size := result.(*domain.Size)

	return &pb.GetSizeResponse{
		Size: mapper.SizeToProto(size),
	}, nil
}

func (h *SizeHandler) ListSizes(ctx context.Context, req *pb.ListSizesRequest) (*pb.ListSizesResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	unit := ""
	if req.Unit != nil {
		unit = *req.Unit
	}

	var minValue, maxValue *float64
	if req.MinValue != nil {
		minValue = req.MinValue
	}
	if req.MaxValue != nil {
		maxValue = req.MaxValue
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	qry := &query.ListSizesQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           int(page),
		PageSize:       int(pageSize),
		Unit:           unit,
		MinValue:       minValue,
		MaxValue:       maxValue,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListSizesResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListSizesResponse{
		Sizes:      mapper.SizesToProto(listResult.Sizes),
		Total:      int32(listResult.Total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (h *SizeHandler) UpdateSize(ctx context.Context, req *pb.UpdateSizeRequest) (*pb.UpdateSizeResponse, error) {
	cmd := &command.UpdateSizeCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
		Value:       req.Value,
		Unit:        req.Unit,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated size
	qry := &query.GetSizeByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "size not found")
	}

	size := result.(*domain.Size)

	return &pb.UpdateSizeResponse{
		Size: mapper.SizeToProto(size),
	}, nil
}

func (h *SizeHandler) DeleteSize(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteSizeCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Size deleted successfully",
	}, nil
}
