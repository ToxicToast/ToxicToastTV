package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/twitchbot-service/internal/handler/mapper"
	"toxictoast/services/twitchbot-service/internal/usecase"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

type ViewerHandler struct {
	pb.UnimplementedViewerServiceServer
	viewerUC usecase.ViewerUseCase
}

func NewViewerHandler(viewerUC usecase.ViewerUseCase) *ViewerHandler {
	return &ViewerHandler{
		viewerUC: viewerUC,
	}
}

func (h *ViewerHandler) CreateViewer(ctx context.Context, req *pb.CreateViewerRequest) (*pb.CreateViewerResponse, error) {
	viewer, err := h.viewerUC.CreateViewer(ctx, req.TwitchId, req.Username, req.DisplayName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateViewerResponse{
		Viewer: mapper.ViewerToProto(viewer),
	}, nil
}

func (h *ViewerHandler) GetViewer(ctx context.Context, req *pb.IdRequest) (*pb.GetViewerResponse, error) {
	viewer, err := h.viewerUC.GetViewerByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrViewerNotFound {
			return nil, status.Error(codes.NotFound, "viewer not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetViewerResponse{
		Viewer: mapper.ViewerToProto(viewer),
	}, nil
}

func (h *ViewerHandler) ListViewers(ctx context.Context, req *pb.ListViewersRequest) (*pb.ListViewersResponse, error) {
	page := req.Offset
	if page <= 0 {
		page = 1
	}
	pageSize := req.Limit
	if pageSize <= 0 {
		pageSize = 10
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	viewers, total, err := h.viewerUC.ListViewers(ctx, int(page), int(pageSize), req.OrderBy, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListViewersResponse{
		Viewers: mapper.ViewersToProto(viewers),
		Total:   int32(total),
	}, nil
}

func (h *ViewerHandler) UpdateViewer(ctx context.Context, req *pb.UpdateViewerRequest) (*pb.UpdateViewerResponse, error) {
	var username, displayName *string
	var totalMessages, totalStreamsWatched *int

	if req.Username != nil {
		username = req.Username
	}
	if req.DisplayName != nil {
		displayName = req.DisplayName
	}
	if req.TotalMessages != nil {
		tm := int(*req.TotalMessages)
		totalMessages = &tm
	}
	if req.TotalStreamsWatched != nil {
		tsw := int(*req.TotalStreamsWatched)
		totalStreamsWatched = &tsw
	}

	viewer, err := h.viewerUC.UpdateViewer(ctx, req.Id, username, displayName, totalMessages, totalStreamsWatched)
	if err != nil {
		if err == usecase.ErrViewerNotFound {
			return nil, status.Error(codes.NotFound, "viewer not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateViewerResponse{
		Viewer: mapper.ViewerToProto(viewer),
	}, nil
}

func (h *ViewerHandler) DeleteViewer(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	if err := h.viewerUC.DeleteViewer(ctx, req.Id); err != nil {
		if err == usecase.ErrViewerNotFound {
			return nil, status.Error(codes.NotFound, "viewer not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Viewer deleted successfully",
	}, nil
}

func (h *ViewerHandler) GetViewerByTwitchId(ctx context.Context, req *pb.GetViewerByTwitchIdRequest) (*pb.GetViewerResponse, error) {
	viewer, err := h.viewerUC.GetViewerByTwitchID(ctx, req.TwitchId)
	if err != nil {
		if err == usecase.ErrViewerNotFound {
			return nil, status.Error(codes.NotFound, "viewer not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetViewerResponse{
		Viewer: mapper.ViewerToProto(viewer),
	}, nil
}

func (h *ViewerHandler) GetViewerStats(ctx context.Context, req *pb.IdRequest) (*pb.GetViewerStatsResponse, error) {
	totalMessages, totalStreamsWatched, daysSinceFirstSeen, daysSinceLastSeen, err := h.viewerUC.GetViewerStats(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrViewerNotFound {
			return nil, status.Error(codes.NotFound, "viewer not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetViewerStatsResponse{
		TotalMessages:       int32(totalMessages),
		TotalStreamsWatched: int32(totalStreamsWatched),
		DaysSinceFirstSeen:  int32(daysSinceFirstSeen),
		DaysSinceLastSeen:   int32(daysSinceLastSeen),
	}, nil
}
