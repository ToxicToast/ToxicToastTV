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

type WebhookHandler struct {
	pb.UnimplementedWebhookManagementServiceServer
	commandBus *cqrs.CommandBus
	queryBus   *cqrs.QueryBus
}

func NewWebhookHandler(commandBus *cqrs.CommandBus, queryBus *cqrs.QueryBus) *WebhookHandler {
	return &WebhookHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

// CreateWebhook creates a new webhook
func (h *WebhookHandler) CreateWebhook(ctx context.Context, req *pb.CreateWebhookRequest) (*pb.WebhookResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	cmd := &command.CreateWebhookCommand{
		BaseCommand: cqrs.BaseCommand{},
		URL:         req.Url,
		Secret:      req.Secret,
		EventTypes:  req.EventTypes,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create webhook: %v", err)
	}

	// Fetch the created webhook
	qry := &query.GetWebhookByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        cmd.AggregateID,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get created webhook: %v", err)
	}

	webhookResult := result.(*query.GetWebhookResult)

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhookResult.Webhook),
	}, nil
}

// GetWebhook gets a webhook by ID
func (h *WebhookHandler) GetWebhook(ctx context.Context, req *pb.GetWebhookRequest) (*pb.WebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	qry := &query.GetWebhookByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "webhook not found: %v", err)
	}

	webhookResult := result.(*query.GetWebhookResult)

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhookResult.Webhook),
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

	qry := &query.ListWebhooksQuery{
		BaseQuery:  cqrs.BaseQuery{},
		Limit:      limit,
		Offset:     offset,
		ActiveOnly: req.ActiveOnly,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list webhooks: %v", err)
	}

	listResult := result.(*query.ListWebhooksResult)

	return &pb.ListWebhooksResponse{
		Webhooks: mapper.ToProtoWebhooks(listResult.Webhooks),
		Total:    listResult.Total,
		Limit:    int32(limit),
		Offset:   int32(offset),
	}, nil
}

// UpdateWebhook updates a webhook
func (h *WebhookHandler) UpdateWebhook(ctx context.Context, req *pb.UpdateWebhookRequest) (*pb.WebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Convert to pointers for optional fields
	var url, secret, description *string
	var eventTypes *[]string
	var active *bool

	if req.Url != "" {
		url = &req.Url
	}
	if req.Secret != "" {
		secret = &req.Secret
	}
	if len(req.EventTypes) > 0 {
		eventTypes = &req.EventTypes
	}
	if req.Description != "" {
		description = &req.Description
	}
	activeVal := req.Active
	active = &activeVal

	cmd := &command.UpdateWebhookCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		URL:         url,
		Secret:      secret,
		EventTypes:  eventTypes,
		Description: description,
		Active:      active,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update webhook: %v", err)
	}

	// Fetch the updated webhook
	qry := &query.GetWebhookByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get updated webhook: %v", err)
	}

	webhookResult := result.(*query.GetWebhookResult)

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhookResult.Webhook),
	}, nil
}

// DeleteWebhook soft deletes a webhook
func (h *WebhookHandler) DeleteWebhook(ctx context.Context, req *pb.DeleteWebhookRequest) (*pb.DeleteWebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.DeleteWebhookCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
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

	cmd := &command.ToggleWebhookCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
		Active:      req.Active,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to toggle webhook: %v", err)
	}

	// Fetch the toggled webhook
	qry := &query.GetWebhookByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get toggled webhook: %v", err)
	}

	webhookResult := result.(*query.GetWebhookResult)

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhookResult.Webhook),
	}, nil
}

// RegenerateSecret generates a new secret for a webhook
func (h *WebhookHandler) RegenerateSecret(ctx context.Context, req *pb.RegenerateSecretRequest) (*pb.WebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.RegenerateSecretCommand{
		BaseCommand: cqrs.BaseCommand{AggregateID: req.Id},
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to regenerate secret: %v", err)
	}

	// Fetch the webhook with new secret
	qry := &query.GetWebhookByIDQuery{
		BaseQuery: cqrs.BaseQuery{},
		ID:        req.Id,
	}

	result, err := h.queryBus.Dispatch(ctx, qry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get updated webhook: %v", err)
	}

	webhookResult := result.(*query.GetWebhookResult)

	return &pb.WebhookResponse{
		Webhook: mapper.ToProtoWebhook(webhookResult.Webhook),
	}, nil
}

// TestWebhook sends a test event to a webhook
func (h *WebhookHandler) TestWebhook(ctx context.Context, req *pb.TestWebhookRequest) (*pb.TestWebhookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cmd := &command.TestWebhookCommand{
		BaseCommand: cqrs.BaseCommand{},
		WebhookID:   req.Id,
	}

	if err := h.commandBus.Dispatch(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to test webhook: %v", err)
	}

	return &pb.TestWebhookResponse{
		Success: true,
		Message: "Test webhook queued successfully",
	}, nil
}
