package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	sharedgrpc "github.com/toxictoast/toxictoastgo/shared/grpc"
	"github.com/toxictoast/toxictoastgo/shared/middleware"
	pb "toxictoast/services/twitchbot-service/api/proto"
	"google.golang.org/grpc"
)

// TwitchBotHandler handles HTTP-to-gRPC translation for twitchbot service
type TwitchBotHandler struct {
	streamClient        pb.StreamServiceClient
	messageClient       pb.MessageServiceClient
	viewerClient        pb.ViewerServiceClient
	clipClient          pb.ClipServiceClient
	commandClient       pb.CommandServiceClient
	botClient           pb.BotServiceClient
	channelViewerClient pb.ChannelViewerServiceClient
}

// NewTwitchBotHandler creates a new twitchbot handler
func NewTwitchBotHandler(conn *grpc.ClientConn) *TwitchBotHandler {
	return &TwitchBotHandler{
		streamClient:        pb.NewStreamServiceClient(conn),
		messageClient:       pb.NewMessageServiceClient(conn),
		viewerClient:        pb.NewViewerServiceClient(conn),
		clipClient:          pb.NewClipServiceClient(conn),
		commandClient:       pb.NewCommandServiceClient(conn),
		botClient:           pb.NewBotServiceClient(conn),
		channelViewerClient: pb.NewChannelViewerServiceClient(conn),
	}
}

// getContextWithAuth extracts JWT claims from HTTP request and injects them into gRPC metadata
func (h *TwitchBotHandler) getContextWithAuth(r *http.Request) context.Context {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims != nil {
		ctx = sharedgrpc.InjectClaimsIntoMetadata(ctx, claims)
	}
	return ctx
}

// RegisterRoutes registers all twitchbot routes
func (h *TwitchBotHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Stream routes (read-public, write-protected)
	router.HandleFunc("/streams", h.ListStreams).Methods("GET")
	router.Handle("/streams", authMiddleware.Authenticate(http.HandlerFunc(h.CreateStream))).Methods("POST")
	router.HandleFunc("/streams/active", h.GetActiveStream).Methods("GET")
	router.HandleFunc("/streams/{id}", h.GetStream).Methods("GET")
	router.Handle("/streams/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.UpdateStream))).Methods("PUT")
	router.Handle("/streams/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.DeleteStream))).Methods("DELETE")
	router.Handle("/streams/{id}/end", authMiddleware.Authenticate(http.HandlerFunc(h.EndStream))).Methods("POST")
	router.HandleFunc("/streams/{id}/stats", h.GetStreamStats).Methods("GET")

	// Message routes (read-public, write-protected)
	router.HandleFunc("/messages", h.ListMessages).Methods("GET")
	router.Handle("/messages", authMiddleware.Authenticate(http.HandlerFunc(h.CreateMessage))).Methods("POST")
	router.HandleFunc("/messages/search", h.SearchMessages).Methods("GET")
	router.HandleFunc("/messages/stats", h.GetMessageStats).Methods("GET")
	router.HandleFunc("/messages/{id}", h.GetMessage).Methods("GET")
	router.Handle("/messages/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.DeleteMessage))).Methods("DELETE")

	// Viewer routes (read-public, write-protected)
	router.HandleFunc("/viewers", h.ListViewers).Methods("GET")
	router.Handle("/viewers", authMiddleware.Authenticate(http.HandlerFunc(h.CreateViewer))).Methods("POST")
	router.HandleFunc("/viewers/twitch/{twitch_id}", h.GetViewerByTwitchId).Methods("GET")
	router.HandleFunc("/viewers/{id}", h.GetViewer).Methods("GET")
	router.Handle("/viewers/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.UpdateViewer))).Methods("PUT")
	router.Handle("/viewers/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.DeleteViewer))).Methods("DELETE")
	router.HandleFunc("/viewers/{id}/stats", h.GetViewerStats).Methods("GET")

	// Clip routes (read-public, write-protected)
	router.HandleFunc("/clips", h.ListClips).Methods("GET")
	router.Handle("/clips", authMiddleware.Authenticate(http.HandlerFunc(h.CreateClip))).Methods("POST")
	router.HandleFunc("/clips/twitch/{twitch_clip_id}", h.GetClipByTwitchId).Methods("GET")
	router.HandleFunc("/clips/{id}", h.GetClip).Methods("GET")
	router.Handle("/clips/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.UpdateClip))).Methods("PUT")
	router.Handle("/clips/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.DeleteClip))).Methods("DELETE")

	// Command routes (read-public, write-protected)
	router.HandleFunc("/commands", h.ListCommands).Methods("GET")
	router.Handle("/commands", authMiddleware.Authenticate(http.HandlerFunc(h.CreateCommand))).Methods("POST")
	router.HandleFunc("/commands/name/{name}", h.GetCommandByName).Methods("GET")
	router.HandleFunc("/commands/{id}", h.GetCommand).Methods("GET")
	router.Handle("/commands/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.UpdateCommand))).Methods("PUT")
	router.Handle("/commands/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.DeleteCommand))).Methods("DELETE")
	router.Handle("/commands/{name}/execute", authMiddleware.Authenticate(http.HandlerFunc(h.ExecuteCommand))).Methods("POST")

	// Bot management routes (read-public, write-protected)
	router.HandleFunc("/bot/status", h.GetBotStatus).Methods("GET")
	router.HandleFunc("/bot/channels", h.ListChannels).Methods("GET")
	router.Handle("/bot/channels/join", authMiddleware.Authenticate(http.HandlerFunc(h.JoinChannel))).Methods("POST")
	router.Handle("/bot/channels/leave", authMiddleware.Authenticate(http.HandlerFunc(h.LeaveChannel))).Methods("POST")
	router.Handle("/bot/send", authMiddleware.Authenticate(http.HandlerFunc(h.SendMessage))).Methods("POST")

	// Channel viewer routes (read-public, write-protected)
	router.HandleFunc("/channel-viewers", h.ListChannelViewers).Methods("GET")
	router.HandleFunc("/channel-viewers/count", h.CountChannelViewers).Methods("GET")
	router.HandleFunc("/channel-viewers/{channel}/{twitch_id}", h.GetChannelViewer).Methods("GET")
	router.Handle("/channel-viewers/{channel}/{twitch_id}", authMiddleware.Authenticate(http.HandlerFunc(h.RemoveChannelViewer))).Methods("DELETE")
}

// Stream handlers
func (h *TwitchBotHandler) ListStreams(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit == 0 {
		limit = 20
	}

	req := &pb.ListStreamsRequest{
		Limit:      int32(limit),
		Offset:     int32(offset),
		OnlyActive: r.URL.Query().Get("only_active") == "true",
	}

	if gameName := r.URL.Query().Get("game_name"); gameName != "" {
		req.GameName = gameName
	}

	resp, err := h.streamClient.ListStreams(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to list streams: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) CreateStream(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.streamClient.CreateStream(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to create stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.streamClient.GetStream(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get stream: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) UpdateStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.streamClient.UpdateStream(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to update stream: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) DeleteStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.streamClient.DeleteStream(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to delete stream: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) EndStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.EndStreamRequest{Id: vars["id"]}
	resp, err := h.streamClient.EndStream(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to end stream: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetActiveStream(w http.ResponseWriter, r *http.Request) {
	req := &pb.GetActiveStreamRequest{}
	resp, err := h.streamClient.GetActiveStream(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get active stream: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetStreamStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.streamClient.GetStreamStats(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get stream stats: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Message handlers
func (h *TwitchBotHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit == 0 {
		limit = 50
	}

	req := &pb.ListMessagesRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	if streamID := r.URL.Query().Get("stream_id"); streamID != "" {
		req.StreamId = streamID
	}
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		req.UserId = userID
	}

	resp, err := h.messageClient.ListMessages(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to list messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.messageClient.CreateMessage(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to create message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.messageClient.GetMessage(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get message: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.messageClient.DeleteMessage(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to delete message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit == 0 {
		limit = 50
	}

	req := &pb.SearchMessagesRequest{
		Query:  r.URL.Query().Get("query"),
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	if streamID := r.URL.Query().Get("stream_id"); streamID != "" {
		req.StreamId = streamID
	}
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		req.UserId = userID
	}

	resp, err := h.messageClient.SearchMessages(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to search messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetMessageStats(w http.ResponseWriter, r *http.Request) {
	req := &pb.GetMessageStatsRequest{}
	if streamID := r.URL.Query().Get("stream_id"); streamID != "" {
		req.StreamId = streamID
	}

	resp, err := h.messageClient.GetMessageStats(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get message stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Bot management handlers
func (h *TwitchBotHandler) GetBotStatus(w http.ResponseWriter, r *http.Request) {
	req := &pb.GetBotStatusRequest{}
	resp, err := h.botClient.GetBotStatus(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get bot status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	req := &pb.ListChannelsRequest{}
	resp, err := h.botClient.ListChannels(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to list channels: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) JoinChannel(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Channel string `json:"channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &pb.JoinChannelRequest{Channel: reqBody.Channel}
	resp, err := h.botClient.JoinChannel(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to join channel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) LeaveChannel(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Channel string `json:"channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &pb.LeaveChannelRequest{Channel: reqBody.Channel}
	resp, err := h.botClient.LeaveChannel(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to leave channel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req pb.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.botClient.SendMessage(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to send message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Viewer handlers
func (h *TwitchBotHandler) ListViewers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit == 0 {
		limit = 50
	}

	req := &pb.ListViewersRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	resp, err := h.viewerClient.ListViewers(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to list viewers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) CreateViewer(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateViewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.viewerClient.CreateViewer(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to create viewer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetViewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.viewerClient.GetViewer(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get viewer: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) UpdateViewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateViewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.viewerClient.UpdateViewer(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to update viewer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) DeleteViewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.viewerClient.DeleteViewer(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to delete viewer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetViewerByTwitchId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetViewerByTwitchIdRequest{TwitchId: vars["twitch_id"]}
	resp, err := h.viewerClient.GetViewerByTwitchId(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get viewer by twitch id: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetViewerStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.viewerClient.GetViewerStats(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get viewer stats: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
// Clip handlers
func (h *TwitchBotHandler) ListClips(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit == 0 {
		limit = 20
	}

	req := &pb.ListClipsRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	if streamID := r.URL.Query().Get("stream_id"); streamID != "" {
		req.StreamId = streamID
	}

	resp, err := h.clipClient.ListClips(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to list clips: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) CreateClip(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateClipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.clipClient.CreateClip(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to create clip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetClip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.clipClient.GetClip(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get clip: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) UpdateClip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateClipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.clipClient.UpdateClip(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to update clip: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) DeleteClip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.clipClient.DeleteClip(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to delete clip: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetClipByTwitchId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetClipByTwitchIdRequest{TwitchClipId: vars["twitch_clip_id"]}
	resp, err := h.clipClient.GetClipByTwitchId(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get clip by twitch id: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
// Command handlers
func (h *TwitchBotHandler) ListCommands(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit == 0 {
		limit = 50
	}

	req := &pb.ListCommandsRequest{
		Limit:      int32(limit),
		Offset:     int32(offset),
		OnlyActive: r.URL.Query().Get("only_active") == "true",
	}

	resp, err := h.commandClient.ListCommands(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to list commands: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) CreateCommand(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.commandClient.CreateCommand(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to create command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.commandClient.GetCommand(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get command: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) UpdateCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.commandClient.UpdateCommand(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to update command: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) DeleteCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.IdRequest{Id: vars["id"]}
	resp, err := h.commandClient.DeleteCommand(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to delete command: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetCommandByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetCommandByNameRequest{Name: vars["name"]}
	resp, err := h.commandClient.GetCommandByName(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get command by name: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var reqBody struct {
		UserId        string `json:"user_id"`
		Username      string `json:"username"`
		IsModerator   bool   `json:"is_moderator"`
		IsSubscriber  bool   `json:"is_subscriber"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &pb.ExecuteCommandRequest{
		CommandName:  vars["name"],
		UserId:       reqBody.UserId,
		Username:     reqBody.Username,
		IsModerator:  reqBody.IsModerator,
		IsSubscriber: reqBody.IsSubscriber,
	}

	resp, err := h.commandClient.ExecuteCommand(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to execute command: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
// ChannelViewer handlers
func (h *TwitchBotHandler) ListChannelViewers(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "Missing channel parameter", http.StatusBadRequest)
		return
	}

	req := &pb.ListChannelViewersRequest{Channel: channel}
	resp, err := h.channelViewerClient.ListChannelViewers(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to list channel viewers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetChannelViewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetChannelViewerRequest{
		Channel:  vars["channel"],
		TwitchId: vars["twitch_id"],
	}
	resp, err := h.channelViewerClient.GetChannelViewer(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get channel viewer: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) CountChannelViewers(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "Missing channel parameter", http.StatusBadRequest)
		return
	}

	req := &pb.CountChannelViewersRequest{Channel: channel}
	resp, err := h.channelViewerClient.CountChannelViewers(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to count channel viewers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) RemoveChannelViewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.RemoveChannelViewerRequest{
		Channel:  vars["channel"],
		TwitchId: vars["twitch_id"],
	}
	resp, err := h.channelViewerClient.RemoveChannelViewer(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to remove channel viewer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
