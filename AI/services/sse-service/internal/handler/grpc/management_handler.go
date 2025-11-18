package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"toxictoast/services/sse-service/internal/broker"
	"toxictoast/services/sse-service/internal/domain"
	"toxictoast/services/sse-service/internal/handler/mapper"
	pb "toxictoast/services/sse-service/api/proto"
)

// ManagementHandler handles gRPC management requests
type ManagementHandler struct {
	pb.UnimplementedSSEManagementServiceServer
	broker *broker.Broker
}

// NewManagementHandler creates a new management handler
func NewManagementHandler(b *broker.Broker) *ManagementHandler {
	return &ManagementHandler{
		broker: b,
	}
}

// GetStats returns broker statistics
func (h *ManagementHandler) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	stats := h.broker.GetStats()

	return &pb.GetStatsResponse{
		TotalClients: int32(stats.TotalClients),
		MaxClients:   int32(stats.MaxClients),
	}, nil
}

// GetClients returns list of connected clients
func (h *ManagementHandler) GetClients(ctx context.Context, req *pb.GetClientsRequest) (*pb.GetClientsResponse, error) {
	stats := h.broker.GetStats()

	// Apply pagination
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	total := len(stats.ConnectedClients)
	end := offset + limit
	if end > total {
		end = total
	}

	clients := stats.ConnectedClients
	if offset < total {
		clients = clients[offset:end]
	} else {
		clients = []domain.ClientStats{}
	}

	return &pb.GetClientsResponse{
		Clients: mapper.ClientStatsListToProto(clients),
		Total:   int32(total),
	}, nil
}

// DisconnectClient disconnects a specific client
func (h *ManagementHandler) DisconnectClient(ctx context.Context, req *pb.DisconnectClientRequest) (*pb.DisconnectClientResponse, error) {
	if req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "client_id is required")
	}

	h.broker.UnregisterClient(req.ClientId)

	return &pb.DisconnectClientResponse{
		Success: true,
		Message: "Client disconnected successfully",
	}, nil
}

// GetHealth returns health status
func (h *ManagementHandler) GetHealth(ctx context.Context, req *pb.GetHealthRequest) (*pb.GetHealthResponse, error) {
	clientCount := h.broker.GetClientCount()

	return &pb.GetHealthResponse{
		Healthy:          true,
		Message:          "SSE service is healthy",
		ConnectedClients: int32(clientCount),
	}, nil
}
