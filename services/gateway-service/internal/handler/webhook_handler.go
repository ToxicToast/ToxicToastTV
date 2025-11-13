package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	pb "toxictoast/services/webhook-service/api/proto"
	"google.golang.org/grpc"
)

// WebhookHandler handles HTTP-to-gRPC translation for webhook service
type WebhookHandler struct {
	managementClient pb.WebhookManagementServiceClient
	deliveryClient   pb.DeliveryServiceClient
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(conn *grpc.ClientConn) *WebhookHandler {
	return &WebhookHandler{
		managementClient: pb.NewWebhookManagementServiceClient(conn),
		deliveryClient:   pb.NewDeliveryServiceClient(conn),
	}
}

// RegisterRoutes registers all webhook routes
func (h *WebhookHandler) RegisterRoutes(router *mux.Router) {
	// Webhook Management routes
	router.HandleFunc("/webhooks", h.ListWebhooks).Methods("GET")
	router.HandleFunc("/webhooks", h.CreateWebhook).Methods("POST")
	router.HandleFunc("/webhooks/{id}", h.GetWebhook).Methods("GET")
	router.HandleFunc("/webhooks/{id}", h.UpdateWebhook).Methods("PUT")
	router.HandleFunc("/webhooks/{id}", h.DeleteWebhook).Methods("DELETE")
	router.HandleFunc("/webhooks/{id}/toggle", h.ToggleWebhook).Methods("POST")
	router.HandleFunc("/webhooks/{id}/regenerate-secret", h.RegenerateSecret).Methods("POST")
	router.HandleFunc("/webhooks/{id}/test", h.TestWebhook).Methods("POST")

	// Delivery routes
	router.HandleFunc("/deliveries", h.ListDeliveries).Methods("GET")
	router.HandleFunc("/deliveries/queue-status", h.GetQueueStatus).Methods("GET")
	router.HandleFunc("/deliveries/cleanup", h.CleanupOldDeliveries).Methods("POST")
	router.HandleFunc("/deliveries/{id}", h.GetDelivery).Methods("GET")
	router.HandleFunc("/deliveries/{id}", h.DeleteDelivery).Methods("DELETE")
	router.HandleFunc("/deliveries/{id}/retry", h.RetryDelivery).Methods("POST")
}

// Webhook Management Handlers

func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)

	if limit == 0 {
		limit = 20
	}

	req := &pb.ListWebhooksRequest{
		Limit:      int32(limit),
		Offset:     int32(offset),
		ActiveOnly: r.URL.Query().Get("active_only") == "true",
	}

	resp, err := h.managementClient.ListWebhooks(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list webhooks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.managementClient.CreateWebhook(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create webhook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.GetWebhookRequest{Id: id}
	resp, err := h.managementClient.GetWebhook(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get webhook: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pb.UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	resp, err := h.managementClient.UpdateWebhook(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update webhook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeleteWebhookRequest{Id: id}
	resp, err := h.managementClient.DeleteWebhook(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete webhook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) ToggleWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var reqBody struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &pb.ToggleWebhookRequest{
		Id:     id,
		Active: reqBody.Active,
	}

	resp, err := h.managementClient.ToggleWebhook(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to toggle webhook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) RegenerateSecret(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.RegenerateSecretRequest{Id: id}
	resp, err := h.managementClient.RegenerateSecret(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to regenerate secret: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.TestWebhookRequest{Id: id}
	resp, err := h.managementClient.TestWebhook(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to test webhook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Delivery Handlers

func (h *WebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)

	if limit == 0 {
		limit = 50
	}

	req := &pb.ListDeliveriesRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	if webhookID := r.URL.Query().Get("webhook_id"); webhookID != "" {
		req.WebhookId = webhookID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = parseDeliveryStatus(status)
	}

	resp, err := h.deliveryClient.ListDeliveries(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list deliveries: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.GetDeliveryRequest{Id: id}
	resp, err := h.deliveryClient.GetDelivery(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get delivery: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) RetryDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.RetryDeliveryRequest{Id: id}
	resp, err := h.deliveryClient.RetryDelivery(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to retry delivery: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) DeleteDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeleteDeliveryRequest{Id: id}
	resp, err := h.deliveryClient.DeleteDelivery(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete delivery: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) CleanupOldDeliveries(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		OlderThanDays int32 `json:"older_than_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if reqBody.OlderThanDays == 0 {
		reqBody.OlderThanDays = 30 // Default to 30 days
	}

	req := &pb.CleanupOldDeliveriesRequest{
		OlderThanDays: reqBody.OlderThanDays,
	}

	resp, err := h.deliveryClient.CleanupOldDeliveries(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to cleanup old deliveries: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	req := &pb.GetQueueStatusRequest{}
	resp, err := h.deliveryClient.GetQueueStatus(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get queue status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Helper functions

func parseDeliveryStatus(status string) pb.DeliveryStatus {
	switch status {
	case "pending":
		return pb.DeliveryStatus_DELIVERY_STATUS_PENDING
	case "success":
		return pb.DeliveryStatus_DELIVERY_STATUS_SUCCESS
	case "failed":
		return pb.DeliveryStatus_DELIVERY_STATUS_FAILED
	case "retrying":
		return pb.DeliveryStatus_DELIVERY_STATUS_RETRYING
	default:
		return pb.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED
	}
}
