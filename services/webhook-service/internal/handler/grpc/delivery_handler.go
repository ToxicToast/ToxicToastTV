package grpc

import (
	"context"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	pb "toxictoast/services/webhook-service/api/proto"
	"toxictoast/services/webhook-service/internal/command"
	"toxictoast/services/webhook-service/internal/handler/mapper"
	"toxictoast/services/webhook-service/internal/query"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeliveryHandler struct {
	pb.UnimplementedDeliveryServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewDeliveryHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *DeliveryHandler {
	return &DeliveryHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

// GetDelivery gets a delivery by ID with all attempts
func (h *DeliveryHandler) GetDelivery(ctx context.Context, req *pb.GetDeliveryRequest) (*pb.DeliveryResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	qry := &query.GetDeliveryByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "delivery not found: %v", err)
	}

	deliveryResult := result.(*query.GetDeliveryResult)

	return &pb.DeliveryResponse{
		Delivery: mapper.ToProtoDelivery(deliveryResult.Delivery),
		Attempts: mapper.ToProtoDeliveryAttempts(deliveryResult.Attempts),
	}, nil
}

// ListDeliveries lists deliveries with filters
func (h *DeliveryHandler) ListDeliveries(ctx context.Context, req *pb.ListDeliveriesRequest) (*pb.ListDeliveriesResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	if limit > 1000 {
		return nil, status.Error(codes.InvalidArgument, "limit cannot exceed 1000")
	}

	offset := int(req.Offset)

	domainStatus := mapper.FromProtoDeliveryStatus(req.Status)

	qry := &query.ListDeliveriesQuery{
		BaseQuery: cqrs.BaseQuery{},
		WebhookID: req.WebhookId,
		Status:    domainStatus,
		Limit:     limit,
		Offset:    offset,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list deliveries: %v", err)
	}

	listResult := result.(*query.ListDeliveriesResult)

	return &pb.ListDeliveriesResponse{
		Deliveries: mapper.ToProtoDeliveries(listResult.Deliveries),
		Total:      listResult.Total,
		Limit:      int32(limit),
		Offset:     int32(offset),
	}, nil
}

// RetryDelivery manually retries a failed delivery
func (h *DeliveryHandler) RetryDelivery(ctx context.Context, req *pb.RetryDeliveryRequest) (*pb.RetryDeliveryResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.RetryDeliveryCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retry delivery: %v", err)
	}

	return &pb.RetryDeliveryResponse{
		Success: true,
		Message: "Delivery queued for retry",
	}, nil
}

// DeleteDelivery soft deletes a delivery
func (h *DeliveryHandler) DeleteDelivery(ctx context.Context, req *pb.DeleteDeliveryRequest) (*pb.DeleteDeliveryResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.DeleteDeliveryCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete delivery: %v", err)
	}

	return &pb.DeleteDeliveryResponse{
		Success: true,
		Message: "Delivery deleted successfully",
	}, nil
}

// CleanupOldDeliveries removes old completed/failed deliveries
func (h *DeliveryHandler) CleanupOldDeliveries(ctx context.Context, req *pb.CleanupOldDeliveriesRequest) (*pb.CleanupOldDeliveriesResponse, error) {
	if req.OlderThanDays <= 0 {
		return nil, status.Error(codes.InvalidArgument, "older_than_days must be positive")
	}

	cmd := &command.CleanupOldDeliveriesCommand{
		BaseCommand:   cqrs.BaseCommand{},
		OlderThanDays: int(req.OlderThanDays),
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cleanup deliveries: %v", err)
	}

	return &pb.CleanupOldDeliveriesResponse{
		DeletedCount: 0, // Note: we don't return count from command anymore
		Message:      "Cleanup completed successfully",
	}, nil
}

// GetQueueStatus returns the current delivery queue status
func (h *DeliveryHandler) GetQueueStatus(ctx context.Context, req *pb.GetQueueStatusRequest) (*pb.GetQueueStatusResponse, error) {
	qry := &query.GetQueueStatusQuery{
		BaseQuery: cqrs.BaseQuery{},
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get queue status: %v", err)
	}

	statusResult := result.(*query.GetQueueStatusResult)

	return &pb.GetQueueStatusResponse{
		DeliveryQueueSize: int32(statusResult.DeliveryQueueSize),
		RetryQueueSize:    int32(statusResult.RetryQueueSize),
	}, nil
}
