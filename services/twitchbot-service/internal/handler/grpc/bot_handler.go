package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "toxictoast/services/twitchbot-service/api/proto"
	"toxictoast/services/twitchbot-service/pkg/bot"
)

type BotHandler struct {
	pb.UnimplementedBotServiceServer
	botManager *bot.Manager
}

func NewBotHandler(botManager *bot.Manager) *BotHandler {
	return &BotHandler{
		botManager: botManager,
	}
}

func (h *BotHandler) JoinChannel(ctx context.Context, req *pb.JoinChannelRequest) (*pb.JoinChannelResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}

	if err := h.botManager.JoinChannel(req.Channel); err != nil {
		return &pb.JoinChannelResponse{
			Success: false,
			Message: err.Error(),
			Channel: req.Channel,
		}, nil
	}

	return &pb.JoinChannelResponse{
		Success: true,
		Message: "Successfully joined channel",
		Channel: req.Channel,
	}, nil
}

func (h *BotHandler) LeaveChannel(ctx context.Context, req *pb.LeaveChannelRequest) (*pb.LeaveChannelResponse, error) {
	if req.Channel == "" {
		return nil, status.Error(codes.InvalidArgument, "channel is required")
	}

	if err := h.botManager.LeaveChannel(req.Channel); err != nil {
		return &pb.LeaveChannelResponse{
			Success: false,
			Message: err.Error(),
			Channel: req.Channel,
		}, nil
	}

	return &pb.LeaveChannelResponse{
		Success: true,
		Message: "Successfully left channel",
		Channel: req.Channel,
	}, nil
}

func (h *BotHandler) ListChannels(ctx context.Context, req *pb.ListChannelsRequest) (*pb.ListChannelsResponse, error) {
	channels := h.botManager.GetJoinedChannels()
	primaryChannel := h.botManager.GetPrimaryChannel()

	return &pb.ListChannelsResponse{
		Channels:       channels,
		PrimaryChannel: primaryChannel,
	}, nil
}

func (h *BotHandler) GetBotStatus(ctx context.Context, req *pb.GetBotStatusRequest) (*pb.GetBotStatusResponse, error) {
	status := h.botManager.GetStatus()

	var connectedSince *timestamppb.Timestamp
	if !status.ConnectedSince.IsZero() {
		connectedSince = timestamppb.New(status.ConnectedSince)
	}

	return &pb.GetBotStatusResponse{
		Connected:       status.Connected,
		Authenticated:   status.Authenticated,
		ChannelsJoined:  int32(status.ChannelsJoined),
		ActiveChannels:  status.ActiveChannels,
		BotUsername:     status.BotUsername,
		ConnectedSince:  connectedSince,
	}, nil
}

func (h *BotHandler) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	if req.Message == "" {
		return nil, status.Error(codes.InvalidArgument, "message is required")
	}

	if err := h.botManager.SendMessage(req.Channel, req.Message); err != nil {
		return &pb.SendMessageResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SendMessageResponse{
		Success: true,
	}, nil
}
