package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

// RegisterRoutes registers all twitchbot routes
func (h *TwitchBotHandler) RegisterRoutes(router *mux.Router) {
	// Stream routes
	router.HandleFunc("/streams", h.ListStreams).Methods("GET")
	router.HandleFunc("/streams", h.CreateStream).Methods("POST")
	router.HandleFunc("/streams/active", h.GetActiveStream).Methods("GET")
	router.HandleFunc("/streams/{id}", h.GetStream).Methods("GET")
	router.HandleFunc("/streams/{id}", h.UpdateStream).Methods("PUT")
	router.HandleFunc("/streams/{id}", h.DeleteStream).Methods("DELETE")
	router.HandleFunc("/streams/{id}/end", h.EndStream).Methods("POST")
	router.HandleFunc("/streams/{id}/stats", h.GetStreamStats).Methods("GET")

	// Message routes
	router.HandleFunc("/messages", h.ListMessages).Methods("GET")
	router.HandleFunc("/messages", h.CreateMessage).Methods("POST")
	router.HandleFunc("/messages/search", h.SearchMessages).Methods("GET")
	router.HandleFunc("/messages/stats", h.GetMessageStats).Methods("GET")
	router.HandleFunc("/messages/{id}", h.GetMessage).Methods("GET")
	router.HandleFunc("/messages/{id}", h.DeleteMessage).Methods("DELETE")

	// Viewer routes
	router.HandleFunc("/viewers", h.ListViewers).Methods("GET")
	router.HandleFunc("/viewers", h.CreateViewer).Methods("POST")
	router.HandleFunc("/viewers/twitch/{twitch_id}", h.GetViewerByTwitchId).Methods("GET")
	router.HandleFunc("/viewers/{id}", h.GetViewer).Methods("GET")
	router.HandleFunc("/viewers/{id}", h.UpdateViewer).Methods("PUT")
	router.HandleFunc("/viewers/{id}", h.DeleteViewer).Methods("DELETE")
	router.HandleFunc("/viewers/{id}/stats", h.GetViewerStats).Methods("GET")

	// Clip routes
	router.HandleFunc("/clips", h.ListClips).Methods("GET")
	router.HandleFunc("/clips", h.CreateClip).Methods("POST")
	router.HandleFunc("/clips/twitch/{twitch_clip_id}", h.GetClipByTwitchId).Methods("GET")
	router.HandleFunc("/clips/{id}", h.GetClip).Methods("GET")
	router.HandleFunc("/clips/{id}", h.UpdateClip).Methods("PUT")
	router.HandleFunc("/clips/{id}", h.DeleteClip).Methods("DELETE")

	// Command routes
	router.HandleFunc("/commands", h.ListCommands).Methods("GET")
	router.HandleFunc("/commands", h.CreateCommand).Methods("POST")
	router.HandleFunc("/commands/name/{name}", h.GetCommandByName).Methods("GET")
	router.HandleFunc("/commands/{id}", h.GetCommand).Methods("GET")
	router.HandleFunc("/commands/{id}", h.UpdateCommand).Methods("PUT")
	router.HandleFunc("/commands/{id}", h.DeleteCommand).Methods("DELETE")
	router.HandleFunc("/commands/{name}/execute", h.ExecuteCommand).Methods("POST")

	// Bot management routes
	router.HandleFunc("/bot/status", h.GetBotStatus).Methods("GET")
	router.HandleFunc("/bot/channels", h.ListChannels).Methods("GET")
	router.HandleFunc("/bot/channels/join", h.JoinChannel).Methods("POST")
	router.HandleFunc("/bot/channels/leave", h.LeaveChannel).Methods("POST")
	router.HandleFunc("/bot/send", h.SendMessage).Methods("POST")

	// Channel viewer routes
	router.HandleFunc("/channel-viewers", h.ListChannelViewers).Methods("GET")
	router.HandleFunc("/channel-viewers/count", h.CountChannelViewers).Methods("GET")
	router.HandleFunc("/channel-viewers/{channel}/{twitch_id}", h.GetChannelViewer).Methods("GET")
	router.HandleFunc("/channel-viewers/{channel}/{twitch_id}", h.RemoveChannelViewer).Methods("DELETE")
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

	resp, err := h.streamClient.ListStreams(context.Background(), req)
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

	resp, err := h.streamClient.CreateStream(context.Background(), &req)
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
	resp, err := h.streamClient.GetStream(context.Background(), req)
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

	resp, err := h.streamClient.UpdateStream(context.Background(), &req)
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
	resp, err := h.streamClient.DeleteStream(context.Background(), req)
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
	resp, err := h.streamClient.EndStream(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to end stream: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) GetActiveStream(w http.ResponseWriter, r *http.Request) {
	req := &pb.GetActiveStreamRequest{}
	resp, err := h.streamClient.GetActiveStream(context.Background(), req)
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
	resp, err := h.streamClient.GetStreamStats(context.Background(), req)
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

	resp, err := h.messageClient.ListMessages(context.Background(), req)
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

	resp, err := h.messageClient.CreateMessage(context.Background(), &req)
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
	resp, err := h.messageClient.GetMessage(context.Background(), req)
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
	resp, err := h.messageClient.DeleteMessage(context.Background(), req)
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

	resp, err := h.messageClient.SearchMessages(context.Background(), req)
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

	resp, err := h.messageClient.GetMessageStats(context.Background(), req)
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
	resp, err := h.botClient.GetBotStatus(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to get bot status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TwitchBotHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	req := &pb.ListChannelsRequest{}
	resp, err := h.botClient.ListChannels(context.Background(), req)
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
	resp, err := h.botClient.JoinChannel(context.Background(), req)
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
	resp, err := h.botClient.LeaveChannel(context.Background(), req)
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

	resp, err := h.botClient.SendMessage(context.Background(), &req)
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

	resp, err := h.viewerClient.ListViewers(context.Background(), req)
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

	resp, err := h.viewerClient.CreateViewer(context.Background(), &req)
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
	resp, err := h.viewerClient.GetViewer(context.Background(), req)
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

	resp, err := h.viewerClient.UpdateViewer(context.Background(), &req)
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
	resp, err := h.viewerClient.DeleteViewer(context.Background(), req)
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
	resp, err := h.viewerClient.GetViewerByTwitchId(context.Background(), req)
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
	resp, err := h.viewerClient.GetViewerStats(context.Background(), req)
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

	resp, err := h.clipClient.ListClips(context.Background(), req)
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

	resp, err := h.clipClient.CreateClip(context.Background(), &req)
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
	resp, err := h.clipClient.GetClip(context.Background(), req)
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

	resp, err := h.clipClient.UpdateClip(context.Background(), &req)
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
	resp, err := h.clipClient.DeleteClip(context.Background(), req)
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
	resp, err := h.clipClient.GetClipByTwitchId(context.Background(), req)
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

	resp, err := h.commandClient.ListCommands(context.Background(), req)
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

	resp, err := h.commandClient.CreateCommand(context.Background(), &req)
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
	resp, err := h.commandClient.GetCommand(context.Background(), req)
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

	resp, err := h.commandClient.UpdateCommand(context.Background(), &req)
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
	resp, err := h.commandClient.DeleteCommand(context.Background(), req)
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
	resp, err := h.commandClient.GetCommandByName(context.Background(), req)
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

	resp, err := h.commandClient.ExecuteCommand(context.Background(), req)
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
	resp, err := h.channelViewerClient.ListChannelViewers(context.Background(), req)
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
	resp, err := h.channelViewerClient.GetChannelViewer(context.Background(), req)
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
	resp, err := h.channelViewerClient.CountChannelViewers(context.Background(), req)
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
	resp, err := h.channelViewerClient.RemoveChannelViewer(context.Background(), req)
	if err != nil {
		http.Error(w, "Failed to remove channel viewer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
