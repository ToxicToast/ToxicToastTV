package grpc

import (
	"context"

	"toxictoast/services/notification-service/internal/handler/mapper"
	"toxictoast/services/notification-service/internal/usecase"
	pb "toxictoast/services/notification-service/api/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChannelHandler struct {
	pb.UnimplementedChannelManagementServiceServer
	channelUC      *usecase.ChannelUseCase
	notificationUC *usecase.NotificationUseCase
}

func NewChannelHandler(channelUC *usecase.ChannelUseCase, notificationUC *usecase.NotificationUseCase) *ChannelHandler {
	return &ChannelHandler{
		channelUC:      channelUC,
		notificationUC: notificationUC,
	}
}

func (h *ChannelHandler) CreateChannel(ctx context.Context, req *pb.CreateChannelRequest) (*pb.ChannelResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.WebhookUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "webhook_url is required")
	}

	channel, err := h.channelUC.CreateChannel(ctx, req.Name, req.WebhookUrl, req.EventTypes, int(req.Color), req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create channel: %v", err)
	}

	return &pb.ChannelResponse{Channel: mapper.ToProtoChannel(channel)}, nil
}

func (h *ChannelHandler) GetChannel(ctx context.Context, req *pb.GetChannelRequest) (*pb.ChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	channel, err := h.channelUC.GetChannel(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "channel not found: %v", err)
	}

	return &pb.ChannelResponse{Channel: mapper.ToProtoChannel(channel)}, nil
}

func (h *ChannelHandler) ListChannels(ctx context.Context, req *pb.ListChannelsRequest) (*pb.ListChannelsResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	if limit > 1000 {
		return nil, status.Error(codes.InvalidArgument, "limit cannot exceed 1000")
	}

	channels, total, err := h.channelUC.ListChannels(ctx, limit, int(req.Offset), req.ActiveOnly)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list channels: %v", err)
	}

	return &pb.ListChannelsResponse{
		Channels: mapper.ToProtoChannels(channels),
		Total:    total,
		Limit:    int32(limit),
		Offset:   req.Offset,
	}, nil
}

func (h *ChannelHandler) UpdateChannel(ctx context.Context, req *pb.UpdateChannelRequest) (*pb.ChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	channel, err := h.channelUC.UpdateChannel(ctx, req.Id, req.Name, req.WebhookUrl, req.EventTypes, int(req.Color), req.Description, req.Active)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update channel: %v", err)
	}

	return &pb.ChannelResponse{Channel: mapper.ToProtoChannel(channel)}, nil
}

func (h *ChannelHandler) DeleteChannel(ctx context.Context, req *pb.DeleteChannelRequest) (*pb.DeleteChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.channelUC.DeleteChannel(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete channel: %v", err)
	}

	return &pb.DeleteChannelResponse{Success: true, Message: "Channel deleted successfully"}, nil
}

func (h *ChannelHandler) ToggleChannel(ctx context.Context, req *pb.ToggleChannelRequest) (*pb.ChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	channel, err := h.channelUC.ToggleChannel(ctx, req.Id, req.Active)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to toggle channel: %v", err)
	}

	return &pb.ChannelResponse{Channel: mapper.ToProtoChannel(channel)}, nil
}

func (h *ChannelHandler) TestChannel(ctx context.Context, req *pb.TestChannelRequest) (*pb.TestChannelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.notificationUC.TestChannel(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to test channel: %v", err)
	}

	return &pb.TestChannelResponse{Success: true, Message: "Test notification sent successfully"}, nil
}
