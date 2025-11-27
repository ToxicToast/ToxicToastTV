package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/interfaces"
)

type GetWebhookByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetWebhookByIDQuery) QueryName() string { return "get_webhook_by_id" }
func (q *GetWebhookByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("webhook ID is required")
	}
	return nil
}

type ListWebhooksQuery struct {
	cqrs.BaseQuery
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	ActiveOnly bool `json:"active_only"`
}

func (q *ListWebhooksQuery) QueryName() string { return "list_webhooks" }
func (q *ListWebhooksQuery) Validate() error {
	if q.Limit < 0 {
		return errors.New("limit cannot be negative")
	}
	if q.Offset < 0 {
		return errors.New("offset cannot be negative")
	}
	return nil
}

type GetActiveWebhooksForEventQuery struct {
	cqrs.BaseQuery
	EventType string `json:"event_type"`
}

func (q *GetActiveWebhooksForEventQuery) QueryName() string {
	return "get_active_webhooks_for_event"
}
func (q *GetActiveWebhooksForEventQuery) Validate() error {
	if q.EventType == "" {
		return errors.New("event type is required")
	}
	return nil
}

// Results

type GetWebhookResult struct {
	Webhook *domain.Webhook
}

type ListWebhooksResult struct {
	Webhooks []*domain.Webhook
	Total    int64
}

type GetActiveWebhooksResult struct {
	Webhooks []*domain.Webhook
}

// Handlers

type GetWebhookByIDHandler struct {
	webhookRepo interfaces.WebhookRepository
}

func NewGetWebhookByIDHandler(webhookRepo interfaces.WebhookRepository) *GetWebhookByIDHandler {
	return &GetWebhookByIDHandler{webhookRepo: webhookRepo}
}

func (h *GetWebhookByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetWebhookByIDQuery)

	webhook, err := h.webhookRepo.GetByID(ctx, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	return &GetWebhookResult{Webhook: webhook}, nil
}

type ListWebhooksHandler struct {
	webhookRepo interfaces.WebhookRepository
}

func NewListWebhooksHandler(webhookRepo interfaces.WebhookRepository) *ListWebhooksHandler {
	return &ListWebhooksHandler{webhookRepo: webhookRepo}
}

func (h *ListWebhooksHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*ListWebhooksQuery)

	webhooks, total, err := h.webhookRepo.List(ctx, qry.Limit, qry.Offset, qry.ActiveOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	return &ListWebhooksResult{
		Webhooks: webhooks,
		Total:    total,
	}, nil
}

type GetActiveWebhooksForEventHandler struct {
	webhookRepo interfaces.WebhookRepository
}

func NewGetActiveWebhooksForEventHandler(webhookRepo interfaces.WebhookRepository) *GetActiveWebhooksForEventHandler {
	return &GetActiveWebhooksForEventHandler{webhookRepo: webhookRepo}
}

func (h *GetActiveWebhooksForEventHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetActiveWebhooksForEventQuery)

	webhooks, err := h.webhookRepo.GetActiveWebhooksForEvent(ctx, qry.EventType)
	if err != nil {
		return nil, fmt.Errorf("failed to get active webhooks for event: %w", err)
	}

	return &GetActiveWebhooksResult{Webhooks: webhooks}, nil
}
