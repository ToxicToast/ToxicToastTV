package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	sharedgrpc "github.com/toxictoast/toxictoastgo/shared/grpc"
	"github.com/toxictoast/toxictoastgo/shared/middleware"
	pb "toxictoast/services/sse-service/api/proto"
	"google.golang.org/grpc"
)

// SSEHandler handles HTTP-to-gRPC translation for SSE service
type SSEHandler struct {
	client pb.SSEManagementServiceClient
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(conn *grpc.ClientConn) *SSEHandler {
	return &SSEHandler{
		client: pb.NewSSEManagementServiceClient(conn),
	}
}

// getContextWithAuth extracts JWT claims from HTTP request and injects them into gRPC metadata
func (h *SSEHandler) getContextWithAuth(r *http.Request) context.Context {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims != nil {
		ctx = sharedgrpc.InjectClaimsIntoMetadata(ctx, claims)
	}
	return ctx
}

// RegisterRoutes registers all SSE routes
func (h *SSEHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Management routes (public read)
	router.HandleFunc("/stats", h.GetStats).Methods("GET")
	router.HandleFunc("/health", h.GetHealth).Methods("GET")
	router.HandleFunc("/clients", h.GetClients).Methods("GET")

	// Protected operations (admin only)
	router.Handle("/clients/{id}/disconnect", authMiddleware.Authenticate(http.HandlerFunc(h.DisconnectClient))).Methods("POST")
}

// Management Handlers

// GetStats handles GET /api/events/stats
func (h *SSEHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	req := &pb.GetStatsRequest{}
	resp, err := h.client.GetStats(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetHealth handles GET /api/events/health
func (h *SSEHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	req := &pb.GetHealthRequest{}
	resp, err := h.client.GetHealth(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get health: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetClients handles GET /api/events/clients
func (h *SSEHandler) GetClients(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)

	if limit == 0 {
		limit = 50
	}

	req := &pb.GetClientsRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	resp, err := h.client.GetClients(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get clients: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DisconnectClient handles POST /api/events/clients/{id}/disconnect
func (h *SSEHandler) DisconnectClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	req := &pb.DisconnectClientRequest{
		ClientId: clientID,
	}

	resp, err := h.client.DisconnectClient(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to disconnect client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !resp.Success {
		http.Error(w, resp.Message, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
