package grpc

import (
	"context"

	"toxictoast/services/webhook-service/internal/handler/mapper"
	"toxictoast/services/webhook-service/internal/usecase"
	pb "toxictoast/services/webhook-service/api/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeliveryHandler struct {
	pb.UnimplementedDeliveryServiceServer
	deliveryUC *usecase.DeliveryUseCase
}

func NewDeliveryHandler(deliveryUC *usecase.DeliveryUseCase) *DeliveryHandler {
	return &DeliveryHandler{
		deliveryUC: deliveryUC,
	}
}

// GetDelivery gets a delivery by ID with all attempts
func (h *DeliveryHandler) GetDelivery(ctx context.Context, req *pb.GetDeliveryRequest) (*pb.DeliveryResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	delivery, attempts, err := h.deliveryUC.GetDelivery(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "delivery not found: %v", err)
	}

	return &pb.DeliveryResponse{
		Delivery: mapper.ToProtoDelivery(delivery),
		Attempts: mapper.ToProtoDeliveryAttempts(attempts),
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

	deliveries, total, err := h.deliveryUC.ListDeliveries(ctx, req.WebhookId, domainStatus, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list deliveries: %v", err)
	}

	return &pb.ListDeliveriesResponse{
		Deliveries: mapper.ToProtoDeliveries(deliveries),
		Total:      total,
		Limit:      int32(limit),
		Offset:     int32(offset),
	}, nil
}

// RetryDelivery manually retries a failed delivery
func (h *DeliveryHandler) RetryDelivery(ctx context.Context, req *pb.RetryDeliveryRequest) (*pb.RetryDeliveryResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.deliveryUC.RetryDelivery(ctx, req.Id); err != nil {
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

	if err := h.deliveryUC.DeleteDelivery(ctx, req.Id); err != nil {
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

	deleted, err := h.deliveryUC.CleanupOldDeliveries(ctx, int(req.OlderThanDays))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cleanup deliveries: %v", err)
	}

	return &pb.CleanupOldDeliveriesResponse{
		DeletedCount: int32(deleted),
		Message:      "Cleanup completed successfully",
	}, nil
}

// GetQueueStatus returns the current delivery queue status
func (h *DeliveryHandler) GetQueueStatus(ctx context.Context, req *pb.GetQueueStatusRequest) (*pb.GetQueueStatusResponse, error) {
	deliveryQueueSize, retryQueueSize := h.deliveryUC.GetQueueStatus()

	return &pb.GetQueueStatusResponse{
		DeliveryQueueSize: int32(deliveryQueueSize),
		RetryQueueSize:    int32(retryQueueSize),
	}, nil
}
