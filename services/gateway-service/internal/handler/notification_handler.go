package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	pb "toxictoast/services/notification-service/api/proto"
	"google.golang.org/grpc"
)

// NotificationHandler handles HTTP-to-gRPC translation for notification service
type NotificationHandler struct {
	channelClient      pb.ChannelManagementServiceClient
	notificationClient pb.NotificationServiceClient
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(conn *grpc.ClientConn) *NotificationHandler {
	return &NotificationHandler{
		channelClient:      pb.NewChannelManagementServiceClient(conn),
		notificationClient: pb.NewNotificationServiceClient(conn),
	}
}

// RegisterRoutes registers all notification routes
func (h *NotificationHandler) RegisterRoutes(router *mux.Router) {
	// Channel Management routes
	router.HandleFunc("/channels", h.ListChannels).Methods("GET")
	router.HandleFunc("/channels", h.CreateChannel).Methods("POST")
	router.HandleFunc("/channels/{id}", h.GetChannel).Methods("GET")
	router.HandleFunc("/channels/{id}", h.UpdateChannel).Methods("PUT")
	router.HandleFunc("/channels/{id}", h.DeleteChannel).Methods("DELETE")
	router.HandleFunc("/channels/{id}/toggle", h.ToggleChannel).Methods("POST")
	router.HandleFunc("/channels/{id}/test", h.TestChannel).Methods("POST")

	// Notification History routes
	router.HandleFunc("/notifications", h.ListNotifications).Methods("GET")
	router.HandleFunc("/notifications/cleanup", h.CleanupOldNotifications).Methods("POST")
	router.HandleFunc("/notifications/{id}", h.GetNotification).Methods("GET")
	router.HandleFunc("/notifications/{id}", h.DeleteNotification).Methods("DELETE")
}

// Channel Management Handlers

// ListChannels handles GET /api/notifications/channels
func (h *NotificationHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)

	if limit == 0 {
		limit = 20
	}

	req := &pb.ListChannelsRequest{
		Limit:      int32(limit),
		Offset:     int32(offset),
		ActiveOnly: r.URL.Query().Get("active_only") == "true",
	}

	resp, err := h.channelClient.ListChannels(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list channels: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateChannel handles POST /api/notifications/channels
func (h *NotificationHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.channelClient.CreateChannel(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to create channel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetChannel handles GET /api/notifications/channels/{id}
func (h *NotificationHandler) GetChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.GetChannelRequest{Id: id}
	resp, err := h.channelClient.GetChannel(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get channel: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateChannel handles PUT /api/notifications/channels/{id}
func (h *NotificationHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pb.UpdateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	resp, err := h.channelClient.UpdateChannel(context.Background(), &req)
	if err != nil {
		http.Error(w, "Failed to update channel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteChannel handles DELETE /api/notifications/channels/{id}
func (h *NotificationHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeleteChannelRequest{Id: id}
	resp, err := h.channelClient.DeleteChannel(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete channel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ToggleChannel handles POST /api/notifications/channels/{id}/toggle
func (h *NotificationHandler) ToggleChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var reqBody struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &pb.ToggleChannelRequest{
		Id:     id,
		Active: reqBody.Active,
	}

	resp, err := h.channelClient.ToggleChannel(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to toggle channel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// TestChannel handles POST /api/notifications/channels/{id}/test
func (h *NotificationHandler) TestChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.TestChannelRequest{Id: id}
	resp, err := h.channelClient.TestChannel(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to test channel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Notification History Handlers

// ListNotifications handles GET /api/notifications/notifications
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)

	if limit == 0 {
		limit = 50
	}

	req := &pb.ListNotificationsRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	if channelID := r.URL.Query().Get("channel_id"); channelID != "" {
		req.ChannelId = channelID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = parseNotificationStatus(status)
	}

	resp, err := h.notificationClient.ListNotifications(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to list notifications: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetNotification handles GET /api/notifications/notifications/{id}
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.GetNotificationRequest{Id: id}
	resp, err := h.notificationClient.GetNotification(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get notification: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteNotification handles DELETE /api/notifications/notifications/{id}
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeleteNotificationRequest{Id: id}
	resp, err := h.notificationClient.DeleteNotification(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to delete notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CleanupOldNotifications handles POST /api/notifications/notifications/cleanup
func (h *NotificationHandler) CleanupOldNotifications(w http.ResponseWriter, r *http.Request) {
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

	req := &pb.CleanupOldNotificationsRequest{
		OlderThanDays: reqBody.OlderThanDays,
	}

	resp, err := h.notificationClient.CleanupOldNotifications(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to cleanup old notifications: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Helper functions

func parseNotificationStatus(status string) pb.NotificationStatus {
	switch status {
	case "pending":
		return pb.NotificationStatus_NOTIFICATION_STATUS_PENDING
	case "success":
		return pb.NotificationStatus_NOTIFICATION_STATUS_SUCCESS
	case "failed":
		return pb.NotificationStatus_NOTIFICATION_STATUS_FAILED
	default:
		return pb.NotificationStatus_NOTIFICATION_STATUS_UNSPECIFIED
	}
}
