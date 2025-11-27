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

type LocationHandler struct {
	pb.UnimplementedLocationServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewLocationHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *LocationHandler {
	return &LocationHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *LocationHandler) CreateLocation(ctx context.Context, req *pb.CreateLocationRequest) (*pb.CreateLocationResponse, error) {
	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	cmd := &command.CreateLocationCommand{
		BaseCommand: cqrs.BaseCommand{},
		Name:        req.Name,
		ParentID:    parentID,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateLocationResponse{
		Location: &pb.Location{
			Name:     req.Name,
			ParentId: parentID,
		},
	}, nil
}

func (h *LocationHandler) GetLocation(ctx context.Context, req *pb.IdRequest) (*pb.GetLocationResponse, error) {
	qry := &query.GetLocationByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "location not found")
	}

	location := result.(*domain.Location)

	return &pb.GetLocationResponse{
		Location: mapper.LocationToProto(location),
	}, nil
}

func (h *LocationHandler) ListLocations(ctx context.Context, req *pb.ListLocationsRequest) (*pb.ListLocationsResponse, error) {
	page, pageSize := mapper.GetDefaultPagination(req.Page, req.PageSize)

	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	includeChildren := req.IncludeChildren

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	qry := &query.ListLocationsQuery{
		BaseQuery:       cqrs.BaseQuery{},
		Page:            int(page),
		PageSize:        int(pageSize),
		ParentID:        parentID,
		IncludeChildren: includeChildren,
		IncludeDeleted:  includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListLocationsResult)
	totalPages := (int(listResult.Total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListLocationsResponse{
		Locations:  mapper.LocationsToProto(listResult.Locations),
		Total:      int32(listResult.Total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (h *LocationHandler) UpdateLocation(ctx context.Context, req *pb.UpdateLocationRequest) (*pb.UpdateLocationResponse, error) {
	cmd := &command.UpdateLocationCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Name:        req.Name,
		ParentID:    req.ParentId,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Query the updated location
	qry := &query.GetLocationByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "location not found")
	}

	location := result.(*domain.Location)

	return &pb.UpdateLocationResponse{
		Location: mapper.LocationToProto(location),
	}, nil
}

func (h *LocationHandler) DeleteLocation(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteLocationCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Location deleted successfully",
	}, nil
}

func (h *LocationHandler) GetLocationTree(ctx context.Context, req *pb.GetLocationTreeRequest) (*pb.GetLocationTreeResponse, error) {
	var rootID *string
	if req.RootId != nil {
		rootID = req.RootId
	}

	maxDepth := int(req.MaxDepth)

	qry := &query.GetLocationTreeQuery{
		BaseQuery: cqrs.BaseQuery{},
		RootID:    rootID,
		MaxDepth:  maxDepth,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	locations := result.([]*domain.Location)

	return &pb.GetLocationTreeResponse{
		Locations: mapper.LocationsToProto(locations),
	}, nil
}
