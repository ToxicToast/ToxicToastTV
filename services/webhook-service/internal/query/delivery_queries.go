package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/webhook-service/internal/delivery"
	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/interfaces"
)

type GetDeliveryByIDQuery struct {
	cqrs.BaseQuery
	ID string `json:"id"`
}

func (q *GetDeliveryByIDQuery) QueryName() string { return "get_delivery_by_id" }
func (q *GetDeliveryByIDQuery) Validate() error {
	if q.ID == "" {
		return errors.New("delivery ID is required")
	}
	return nil
}

type ListDeliveriesQuery struct {
	cqrs.BaseQuery
	WebhookID string                `json:"webhook_id"`
	Status    domain.DeliveryStatus `json:"status"`
	Limit     int                   `json:"limit"`
	Offset    int                   `json:"offset"`
}

func (q *ListDeliveriesQuery) QueryName() string { return "list_deliveries" }
func (q *ListDeliveriesQuery) Validate() error {
	if q.Limit < 0 {
		return errors.New("limit cannot be negative")
	}
	if q.Offset < 0 {
		return errors.New("offset cannot be negative")
	}
	return nil
}

type GetQueueStatusQuery struct {
	cqrs.BaseQuery
}

func (q *GetQueueStatusQuery) QueryName() string { return "get_queue_status" }
func (q *GetQueueStatusQuery) Validate() error {
	return nil
}

// Results

type GetDeliveryResult struct {
	Delivery *domain.Delivery
	Attempts []*domain.DeliveryAttempt
}

type ListDeliveriesResult struct {
	Deliveries []*domain.Delivery
	Total      int64
}

type GetQueueStatusResult struct {
	DeliveryQueueSize int
	RetryQueueSize    int
}

// Handlers

type GetDeliveryByIDHandler struct {
	deliveryRepo interfaces.DeliveryRepository
}

func NewGetDeliveryByIDHandler(deliveryRepo interfaces.DeliveryRepository) *GetDeliveryByIDHandler {
	return &GetDeliveryByIDHandler{deliveryRepo: deliveryRepo}
}

func (h *GetDeliveryByIDHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*GetDeliveryByIDQuery)

	delivery, err := h.deliveryRepo.GetByID(ctx, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	attempts, err := h.deliveryRepo.GetAttempts(ctx, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery attempts: %w", err)
	}

	return &GetDeliveryResult{
		Delivery: delivery,
		Attempts: attempts,
	}, nil
}

type ListDeliveriesHandler struct {
	deliveryRepo interfaces.DeliveryRepository
}

func NewListDeliveriesHandler(deliveryRepo interfaces.DeliveryRepository) *ListDeliveriesHandler {
	return &ListDeliveriesHandler{deliveryRepo: deliveryRepo}
}

func (h *ListDeliveriesHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	qry := query.(*ListDeliveriesQuery)

	deliveries, total, err := h.deliveryRepo.List(ctx, qry.WebhookID, qry.Status, qry.Limit, qry.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}

	return &ListDeliveriesResult{
		Deliveries: deliveries,
		Total:      total,
	}, nil
}

type GetQueueStatusHandler struct {
	deliveryPool *delivery.Pool
}

func NewGetQueueStatusHandler(deliveryPool *delivery.Pool) *GetQueueStatusHandler {
	return &GetQueueStatusHandler{deliveryPool: deliveryPool}
}

func (h *GetQueueStatusHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	deliveryQueueSize, retryQueueSize := h.deliveryPool.GetQueueStatus()

	return &GetQueueStatusResult{
		DeliveryQueueSize: deliveryQueueSize,
		RetryQueueSize:    retryQueueSize,
	}, nil
}
