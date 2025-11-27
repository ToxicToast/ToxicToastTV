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

type TypeHandler struct {
	pb.UnimplementedTypeServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewTypeHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *TypeHandler {
	return &TypeHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *TypeHandler) CreateType(ctx context.Context, req *pb.CreateTypeRequest) (*pb.CreateTypeResponse, error) {
	cmd := &command.CreateTypeCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateTypeResponse{
		Type: &pb.Type{
			Name: req.Name,
		},
	}, nil
}

func (h *TypeHandler) GetType(ctx context.Context, req *pb.IdRequest) (*pb.GetTypeResponse, error) {
	qry := &query.GetTypeByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "type not found")
	}

	typeEntity := result.(*domain.Type)

	return &pb.GetTypeResponse{
		Type: mapper.TypeToProto(typeEntity),
	}, nil
}

func (h *TypeHandler) ListTypes(ctx context.Context, req *pb.ListTypesRequest) (*pb.ListTypesResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	search := ""
	if req.Search != nil {
		search = *req.Search
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	qry := &query.ListTypesQuery{
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

	listResult := result.(*query.ListTypesResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListTypesResponse{
		Types:      mapper.TypesToProto(listResult.Types),
		Total:      int32(listResult.Total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (h *TypeHandler) UpdateType(ctx context.Context, req *pb.UpdateTypeRequest) (*pb.UpdateTypeResponse, error) {
	cmd := &command.UpdateTypeCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated type
	qry := &query.GetTypeByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "type not found")
	}

	typeEntity := result.(*domain.Type)

	return &pb.UpdateTypeResponse{
		Type: mapper.TypeToProto(typeEntity),
	}, nil
}

func (h *TypeHandler) DeleteType(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteTypeCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Type deleted successfully",
	}, nil
}
