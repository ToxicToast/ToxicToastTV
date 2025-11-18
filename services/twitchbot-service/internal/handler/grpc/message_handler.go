package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/twitchbot-service/internal/handler/mapper"
	"toxictoast/services/twitchbot-service/internal/usecase"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

type MessageHandler struct {
	pb.UnimplementedMessageServiceServer
	messageUC usecase.MessageUseCase
}

func NewMessageHandler(messageUC usecase.MessageUseCase) *MessageHandler {
	return &MessageHandler{
		messageUC: messageUC,
	}
}

func (h *MessageHandler) CreateMessage(ctx context.Context, req *pb.CreateMessageRequest) (*pb.CreateMessageResponse, error) {
	message, err := h.messageUC.CreateMessage(
		ctx,
		req.StreamId,
		req.UserId,
		req.Username,
		req.DisplayName,
		req.Message,
		req.IsModerator,
		req.IsSubscriber,
		req.IsVip,
		req.IsBroadcaster,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateMessageResponse{
		Message: mapper.MessageToProto(message),
	}, nil
}

func (h *MessageHandler) GetMessage(ctx context.Context, req *pb.IdRequest) (*pb.GetMessageResponse, error) {
	message, err := h.messageUC.GetMessageByID(ctx, req.Id)
	if err != nil {
		if err == usecase.ErrMessageNotFound {
			return nil, status.Error(codes.NotFound, "message not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetMessageResponse{
		Message: mapper.MessageToProto(message),
	}, nil
}

func (h *MessageHandler) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
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

	messages, total, err := h.messageUC.ListMessages(ctx, int(page), int(pageSize), req.StreamId, req.UserId, includeDeleted)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ListMessagesResponse{
		Messages: mapper.MessagesToProto(messages),
		Total:    int32(total),
	}, nil
}

func (h *MessageHandler) DeleteMessage(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	if err := h.messageUC.DeleteMessage(ctx, req.Id); err != nil {
		if err == usecase.ErrMessageNotFound {
			return nil, status.Error(codes.NotFound, "message not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Message deleted successfully",
	}, nil
}

func (h *MessageHandler) SearchMessages(ctx context.Context, req *pb.SearchMessagesRequest) (*pb.SearchMessagesResponse, error) {
	page := req.Offset
	if page <= 0 {
		page = 1
	}
	pageSize := req.Limit
	if pageSize <= 0 {
		pageSize = 10
	}

	messages, total, err := h.messageUC.SearchMessages(ctx, req.Query, req.StreamId, req.UserId, int(page), int(pageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SearchMessagesResponse{
		Messages: mapper.MessagesToProto(messages),
		Total:    int32(total),
	}, nil
}

func (h *MessageHandler) GetMessageStats(ctx context.Context, req *pb.GetMessageStatsRequest) (*pb.GetMessageStatsResponse, error) {
	totalMessages, uniqueUsers, mostActiveUser, mostActiveUserCount, err := h.messageUC.GetMessageStats(ctx, req.StreamId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetMessageStatsResponse{
		TotalMessages:        int32(totalMessages),
		UniqueUsers:          int32(uniqueUsers),
		MostActiveUser:       mostActiveUser,
		MostActiveUserCount:  int32(mostActiveUserCount),
	}, nil
}
