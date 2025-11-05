package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

type LocationHandler struct {
	pb.UnimplementedLocationServiceServer
	locationUC usecase.LocationUseCase
}

func NewLocationHandler(locationUC usecase.LocationUseCase) *LocationHandler {
	return &LocationHandler{
		locationUC: locationUC,
	}
}

func (h *LocationHandler) CreateLocation(ctx context.Context, req *pb.CreateLocationRequest) (*pb.CreateLocationResponse, error) {
	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	location, err := h.locationUC.CreateLocation(ctx, req.Name, parentID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateLocationResponse{
		Location: mapper.LocationToProto(location),
	}, nil
}

func (h *LocationHandler) GetLocation(ctx context.Context, req *pb.IdRequest) (*pb.GetLocationResponse, error) {
	location, err := h.locationUC.GetLocationByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrLocationNotFound {
			return nil, status.Error(codes.NotFound, "location not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetLocationResponse{
		Location: mapper.LocationToProto(location),
	}, nil
}

func (h *LocationHandler) ListLocations(ctx context.Context, req *pb.ListLocationsRequest) (*pb.ListLocationsResponse, error) {
	page := int(req.Pagination.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.Pagination.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	includeChildren := req.IncludeChildren

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	locations, total, err := h.locationUC.ListLocations(ctx, page, pageSize, parentID, includeChildren, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListLocationsResponse{
		Locations:  mapper.LocationsToProto(locations),
		Pagination: mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}

func (h *LocationHandler) UpdateLocation(ctx context.Context, req *pb.UpdateLocationRequest) (*pb.UpdateLocationResponse, error) {
	var name string
	if req.Name != nil {
		name = *req.Name
	} else {
		// Get existing to keep name
		loc, err := h.locationUC.GetLocationByID(ctx, req.Id)
		if err != nil {
			if err == usecase.ErrLocationNotFound {
				return nil, status.Error(codes.NotFound, "location not found")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		name = loc.Name
	}

	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	location, err := h.locationUC.UpdateLocation(ctx, req.Id, name, parentID)
	if err != nil {
		if err == usecase.ErrLocationNotFound {
			return nil, status.Error(codes.NotFound, "location not found")
		}
		if err == usecase.ErrCircularReference {
			return nil, status.Error(codes.InvalidArgument, "circular reference detected")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateLocationResponse{
		Location: mapper.LocationToProto(location),
	}, nil
}

func (h *LocationHandler) DeleteLocation(ctx context.Context, req *pb.IdRequest) (*pb.SuccessResponse, error) {
	err := h.locationUC.DeleteLocation(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrLocationNotFound {
			return nil, status.Error(codes.NotFound, "location not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SuccessResponse{
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

	locations, err := h.locationUC.GetLocationTree(ctx, rootID, maxDepth)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetLocationTreeResponse{
		Locations: mapper.LocationsToProto(locations),
	}, nil
}
