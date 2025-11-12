package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/twitchbot-service/internal/handler/mapper"
	"toxictoast/services/twitchbot-service/internal/usecase"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

type ChannelViewerHandler struct {
	pb.UnimplementedChannelViewerServiceServer
	channelViewerUC usecase.ChannelViewerUseCase
}

func NewChannelViewerHandler(channelViewerUC usecase.ChannelViewerUseCase) *ChannelViewerHandler {
	return &ChannelViewerHandler{
		channelViewerUC: channelViewerUC,
	}
}

func (h *ChannelViewerHandler) GetChannelViewer(ctx context.Context, req *pb.GetChannelViewerRequest) (*pb.GetChannelViewerResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}
	if req.TwitchId == "" {
		return nil, status.Error(codes.InvalidArgument, "twitch_id is required")
	}

	viewer, err := h.channelViewerUC.GetViewer(ctx, req.Channel, req.TwitchId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if viewer == nil {
		return nil, status.Error(codes.NotFound, "viewer not found in this channel")
	}

	return &pb.GetChannelViewerResponse{
		Viewer: mapper.ChannelViewerToProto(viewer),
	}, nil
}

func (h *ChannelViewerHandler) ListChannelViewers(ctx context.Context, req *pb.ListChannelViewersRequest) (*pb.ListChannelViewersResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50 // Default
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	viewers, total, err := h.channelViewerUC.ListViewers(ctx, req.Channel, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListChannelViewersResponse{
		Viewers: mapper.ChannelViewersToProto(viewers),
		Total:   int32(total),
	}, nil
}

func (h *ChannelViewerHandler) CountChannelViewers(ctx context.Context, req *pb.CountChannelViewersRequest) (*pb.CountChannelViewersResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}

	count, err := h.channelViewerUC.CountViewers(ctx, req.Channel)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CountChannelViewersResponse{
		Count: int32(count),
	}, nil
}

func (h *ChannelViewerHandler) RemoveChannelViewer(ctx context.Context, req *pb.RemoveChannelViewerRequest) (*pb.DeleteResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}
	if req.TwitchId == "" {
		return nil, status.Error(codes.InvalidArgument, "twitch_id is required")
	}

	err := h.channelViewerUC.RemoveViewer(ctx, req.Channel, req.TwitchId)
	if err != nil {
		return &pb.DeleteResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Viewer removed from channel successfully",
	}, nil
}
