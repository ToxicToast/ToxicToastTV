package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio/v1"
)

type TypeHandler struct {
	pb.UnimplementedTypeServiceServer
	typeUC usecase.TypeUseCase
}

func NewTypeHandler(typeUC usecase.TypeUseCase) *TypeHandler {
	return &TypeHandler{
		typeUC: typeUC,
	}
}

func (h *TypeHandler) CreateType(ctx context.Context, req *pb.CreateTypeRequest) (*pb.CreateTypeResponse, error) {
	typeEntity, err := h.typeUC.CreateType(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateTypeResponse{
		Type: mapper.TypeToProto(typeEntity),
	}, nil
}

func (h *TypeHandler) GetType(ctx context.Context, req *pb.IdRequest) (*pb.GetTypeResponse, error) {
	typeEntity, err := h.typeUC.GetTypeByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrTypeNotFound {
			return nil, status.Error(codes.NotFound, "type not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetTypeResponse{
		Type: mapper.TypeToProto(typeEntity),
	}, nil
}

func (h *TypeHandler) ListTypes(ctx context.Context, req *pb.ListTypesRequest) (*pb.ListTypesResponse, error) {
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

	types, total, err := h.typeUC.ListTypes(ctx, page, pageSize, search, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListTypesResponse{
		Types:      mapper.TypesToProto(types),
		Pagination: mapper.ToPaginationResponse(page, pageSize, total),
	}, nil
}

func (h *TypeHandler) UpdateType(ctx context.Context, req *pb.UpdateTypeRequest) (*pb.UpdateTypeResponse, error) {
	typeEntity, err := h.typeUC.UpdateType(ctx, req.Id, req.Name)
	if err != nil {
		if err == usecase.ErrTypeNotFound {
			return nil, status.Error(codes.NotFound, "type not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateTypeResponse{
		Type: mapper.TypeToProto(typeEntity),
	}, nil
}

func (h *TypeHandler) DeleteType(ctx context.Context, req *pb.IdRequest) (*pb.SuccessResponse, error) {
	err := h.typeUC.DeleteType(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrTypeNotFound {
			return nil, status.Error(codes.NotFound, "type not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SuccessResponse{
		Success: true,
		Message: "Type deleted successfully",
	}, nil
}
