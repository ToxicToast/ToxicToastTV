package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/foodfolio-service/internal/handler/mapper"
	"toxictoast/services/foodfolio-service/internal/usecase"
	pb "toxictoast/services/foodfolio-service/api/proto/foodfolio"
)

type SizeHandler struct {
	pb.UnimplementedSizeServiceServer
	sizeUC usecase.SizeUseCase
}

func NewSizeHandler(sizeUC usecase.SizeUseCase) *SizeHandler {
	return &SizeHandler{
		sizeUC: sizeUC,
	}
}

func (h *SizeHandler) CreateSize(ctx context.Context, req *pb.CreateSizeRequest) (*pb.CreateSizeResponse, error) {
	size, err := h.sizeUC.CreateSize(ctx, req.Name, req.Value, req.Unit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateSizeResponse{
		Size: mapper.SizeToProto(size),
	}, nil
}

func (h *SizeHandler) GetSize(ctx context.Context, req *pb.IdRequest) (*pb.GetSizeResponse, error) {
	size, err := h.sizeUC.GetSizeByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrSizeNotFound {
			return nil, status.Error(codes.NotFound, "size not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

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

	sizes, total, err := h.sizeUC.ListSizes(ctx, int(page), int(pageSize), unit, minValue, maxValue, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPages := (int(total) + int(pageSize) - 1) / int(pageSize)

	return &pb.ListSizesResponse{
		Sizes:      mapper.SizesToProto(sizes),
		Total:      int32(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (h *SizeHandler) UpdateSize(ctx context.Context, req *pb.UpdateSizeRequest) (*pb.UpdateSizeResponse, error) {
	name := req.Name
	if req.Name == nil || *req.Name == "" {
		// Get existing to keep name
		size, err := h.sizeUC.GetSizeByID(ctx, req.Id)
		if err != nil {
			if err == usecase.ErrSizeNotFound {
				return nil, status.Error(codes.NotFound, "size not found")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		nameVal := size.Name
		name = &nameVal
	}

	value := req.Value
	if req.Value == nil {
		size, err := h.sizeUC.GetSizeByID(ctx, req.Id)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		valueVal := size.Value
		value = &valueVal
	}

	unit := req.Unit
	if req.Unit == nil || *req.Unit == "" {
		size, err := h.sizeUC.GetSizeByID(ctx, req.Id)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		unitVal := size.Unit
		unit = &unitVal
	}

	size, err := h.sizeUC.UpdateSize(ctx, req.Id, *name, *value, *unit)
	if err != nil {
		if err == usecase.ErrSizeNotFound {
			return nil, status.Error(codes.NotFound, "size not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateSizeResponse{
		Size: mapper.SizeToProto(size),
	}, nil
}

func (h *SizeHandler) DeleteSize(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	err := h.sizeUC.DeleteSize(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrSizeNotFound {
			return nil, status.Error(codes.NotFound, "size not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Size deleted successfully",
	}, nil
}
