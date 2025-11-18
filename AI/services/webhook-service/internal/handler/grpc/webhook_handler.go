package grpc

import (
	"context"

	"toxictoast/services/webhook-service/internal/handler/mapper"
	"toxictoast/services/webhook-service/internal/usecase"
	pb "toxictoast/services/webhook-service/api/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WebhookHandler struct {
	pb.UnimplementedWebhookManagementServiceServer
	webhookUC  *usecase.WebhookUseCase
	deliveryUC *usecase.DeliveryUseCase
}

func NewWebhookHandler(webhookUC *usecase.WebhookUseCase, deliveryUC *usecase.DeliveryUseCase) *WebhookHandler {
	return &WebhookHandler{
		webhookUC:  webhookUC,
		deliveryUC: deliveryUC,
	}
}

// CreateWebhook creates a new webhook
func (h *WebhookHandler) CreateWebhook(ctx context.Context, req *pb.CreateWebhookRequest) (*pb.WebhookResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	webhook, err := h.webhookUC.CreateWebhook(ctx, req.Url, req.Secret, req.EventTypes, req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create webhook: %v", err)
	}

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhook),
	}, nil
}

// GetWebhook gets a webhook by ID
func (h *WebhookHandler) GetWebhook(ctx context.Context, req *pb.GetWebhookRequest) (*pb.WebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	webhook, err := h.webhookUC.GetWebhook(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "webhook not found: %v", err)
	}

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhook),
	}, nil
}

// ListWebhooks lists webhooks with pagination
func (h *WebhookHandler) ListWebhooks(ctx context.Context, req *pb.ListWebhooksRequest) (*pb.ListWebhooksResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	if limit > 1000 {
		return nil, status.Error(codes.InvalidArgument, "limit cannot exceed 1000")
	}

	offset := int(req.Offset)

	webhooks, total, err := h.webhookUC.ListWebhooks(ctx, limit, offset, req.ActiveOnly)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list webhooks: %v", err)
	}

	return &pb.ListWebhooksResponse{
		Webhooks: mapper.ToProtoWebhooks(webhooks),
		Total:    total,
		Limit:    int32(limit),
		Offset:   int32(offset),
	}, nil
}

// UpdateWebhook updates a webhook
func (h *WebhookHandler) UpdateWebhook(ctx context.Context, req *pb.UpdateWebhookRequest) (*pb.WebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	webhook, err := h.webhookUC.UpdateWebhook(ctx, req.Id, req.Url, req.Secret, req.EventTypes, req.Description, req.Active)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update webhook: %v", err)
	}

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhook),
	}, nil
}

// DeleteWebhook soft deletes a webhook
func (h *WebhookHandler) DeleteWebhook(ctx context.Context, req *pb.DeleteWebhookRequest) (*pb.DeleteWebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.webhookUC.DeleteWebhook(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete webhook: %v", err)
	}

	return &pb.DeleteWebhookResponse{
		Success: true,
		Message: "Webhook deleted successfully",
	}, nil
}

// ToggleWebhook toggles webhook active status
func (h *WebhookHandler) ToggleWebhook(ctx context.Context, req *pb.ToggleWebhookRequest) (*pb.WebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	webhook, err := h.webhookUC.ToggleWebhook(ctx, req.Id, req.Active)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to toggle webhook: %v", err)
	}

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhook),
	}, nil
}

// RegenerateSecret generates a new secret for a webhook
func (h *WebhookHandler) RegenerateSecret(ctx context.Context, req *pb.RegenerateSecretRequest) (*pb.WebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	webhook, err := h.webhookUC.RegenerateSecret(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to regenerate secret: %v", err)
	}

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhook),
	}, nil
}

// TestWebhook sends a test event to a webhook
func (h *WebhookHandler) TestWebhook(ctx context.Context, req *pb.TestWebhookRequest) (*pb.TestWebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.deliveryUC.TestWebhook(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to test webhook: %v", err)
	}

	return &pb.TestWebhookResponse{
		Success: true,
		Message: "Test webhook queued successfully",
	}, nil
}
