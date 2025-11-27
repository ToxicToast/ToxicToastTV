package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"toxictoast/services/twitchbot-service/internal/command"
	"toxictoast/services/twitchbot-service/internal/handler/mapper"
	"toxictoast/services/twitchbot-service/internal/query"
	pb "toxictoast/services/twitchbot-service/api/proto"
)

type MessageHandler struct {
	pb.UnimplementedMessageServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewMessageHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *MessageHandler {
	return &MessageHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *MessageHandler) CreateMessage(ctx context.Context, req *pb.CreateMessageRequest) (*pb.CreateMessageResponse, error) {
	cmd := &command.CreateMessageCommand{
		BaseCommand:   cqrs.BaseCommand{},
		StreamID:      req.StreamId,
		UserID:        req.UserId,
		Username:      req.Username,
		DisplayName:   req.DisplayName,
		Message:       req.Message,
		IsModerator:   req.IsModerator,
		IsSubscriber:  req.IsSubscriber,
		IsVIP:         req.IsVip,
		IsBroadcaster: req.IsBroadcaster,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch the created message
	qry := &query.GetMessageByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	messageResult := result.(*query.GetMessageResult)

	return &pb.CreateMessageResponse{
		Message: mapper.MessageToProto(messageResult.Message),
	}, nil
}

func (h *MessageHandler) GetMessage(ctx context.Context, req *pb.IdRequest) (*pb.GetMessageResponse, error) {
	qry := &query.GetMessageByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.NotFound, "message not found")
	}

	messageResult := result.(*query.GetMessageResult)

	return &pb.GetMessageResponse{
		Message: mapper.MessageToProto(messageResult.Message),
	}, nil
}

func (h *MessageHandler) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	page := int(req.Offset)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.Limit)
	if pageSize <= 0 {
		pageSize = 10
	}

	includeDeleted := false
	if req.DeletedFilter != nil {
		includeDeleted = req.DeletedFilter.IncludeDeleted
	}

	qry := &query.ListMessagesQuery{
		BaseQuery:      cqrs.BaseQuery{},
		Page:           page,
		PageSize:       pageSize,
		StreamID:       req.StreamId,
		UserID:         req.UserId,
		IncludeDeleted: includeDeleted,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	listResult := result.(*query.ListMessagesResult)

	return &pb.ListMessagesResponse{
		Messages: mapper.MessagesToProto(listResult.Messages),
		Total:    int32(listResult.Total),
	}, nil
}

func (h *MessageHandler) DeleteMessage(ctx context.Context, req *pb.IdRequest) (*pb.DeleteResponse, error) {
	cmd := &command.DeleteMessageCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteResponse{
		Success: true,
		Message: "Message deleted successfully",
	}, nil
}

func (h *MessageHandler) SearchMessages(ctx context.Context, req *pb.SearchMessagesRequest) (*pb.SearchMessagesResponse, error) {
	page := int(req.Offset)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.Limit)
	if pageSize <= 0 {
		pageSize = 10
	}

	qry := &query.SearchMessagesQuery{
		BaseQuery: cqrs.BaseQuery{},
		Query:     req.Query,
		StreamID:  req.StreamId,
		UserID:    req.UserId,
		Page:      page,
		PageSize:  pageSize,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	searchResult := result.(*query.SearchMessagesResult)

	return &pb.SearchMessagesResponse{
		Messages: mapper.MessagesToProto(searchResult.Messages),
		Total:    int32(searchResult.Total),
	}, nil
}

func (h *MessageHandler) GetMessageStats(ctx context.Context, req *pb.GetMessageStatsRequest) (*pb.GetMessageStatsResponse, error) {
	qry := &query.GetMessageStatsQuery{
		BaseQuery: cqrs.BaseQuery{},
		StreamID:  req.StreamId,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	statsResult := result.(*query.GetMessageStatsResult)

	return &pb.GetMessageStatsResponse{
		TotalMessages:       int32(statsResult.TotalMessages),
		UniqueUsers:         int32(statsResult.UniqueUsers),
		MostActiveUser:      statsResult.MostActiveUser,
		MostActiveUserCount: int32(statsResult.MostActiveUserCount),
	}, nil
}
