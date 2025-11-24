package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	sharedgrpc "github.com/toxictoast/toxictoastgo/shared/grpc"
	"github.com/toxictoast/toxictoastgo/shared/middleware"
	pb "toxictoast/services/link-service/api/proto"
	"google.golang.org/grpc"
)

// LinkHandler handles HTTP-to-gRPC translation for link service
type LinkHandler struct {
	client pb.LinkServiceClient
}

// NewLinkHandler creates a new link handler
func NewLinkHandler(conn *grpc.ClientConn) *LinkHandler {
	return &LinkHandler{
		client: pb.NewLinkServiceClient(conn),
	}
}

// getContextWithAuth extracts JWT claims from HTTP request and injects them into gRPC metadata
func (h *LinkHandler) getContextWithAuth(r *http.Request) context.Context {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims != nil {
		ctx = sharedgrpc.InjectClaimsIntoMetadata(ctx, claims)
	}
	return ctx
}

// RegisterRoutes registers all link routes
func (h *LinkHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Link CRUD routes
	router.HandleFunc("/links", h.ListLinks).Methods("GET")
	router.Handle("/links", authMiddleware.Authenticate(http.HandlerFunc(h.CreateLink))).Methods("POST")
	router.HandleFunc("/links/{id}", h.GetLink).Methods("GET")
	router.Handle("/links/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.UpdateLink))).Methods("PUT")
	router.Handle("/links/{id}", authMiddleware.Authenticate(http.HandlerFunc(h.DeleteLink))).Methods("DELETE")

	// Short code routes (public)
	router.HandleFunc("/s/{short_code}", h.GetLinkByShortCode).Methods("GET")
	router.HandleFunc("/s/{short_code}/click", h.IncrementClick).Methods("POST")

	// Analytics routes (read-only, public)
	router.HandleFunc("/links/{id}/stats", h.GetLinkStats).Methods("GET")
	router.HandleFunc("/links/{id}/clicks", h.GetLinkClicks).Methods("GET")
	router.HandleFunc("/links/{id}/clicks-by-date", h.GetClicksByDate).Methods("GET")
	router.HandleFunc("/links/{id}/record-click", h.RecordClick).Methods("POST")
}

// ListLinks handles GET /links
func (h *LinkHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	req := &pb.ListLinksRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	// Optional filters
	if isActive := r.URL.Query().Get("is_active"); isActive != "" {
		active := isActive == "true"
		req.IsActive = &active
	}
	if includeExpired := r.URL.Query().Get("include_expired"); includeExpired != "" {
		expired := includeExpired == "true"
		req.IncludeExpired = &expired
	}
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	resp, err := h.client.ListLinks(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to list links: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateLink handles POST /links
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateLink(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to create link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetLink handles GET /links/{id}
func (h *LinkHandler) GetLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.GetLinkRequest{Id: id}
	resp, err := h.client.GetLink(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get link: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateLink handles PUT /links/{id}
func (h *LinkHandler) UpdateLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req pb.UpdateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	resp, err := h.client.UpdateLink(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to update link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DeleteLink handles DELETE /links/{id}
func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := &pb.DeleteLinkRequest{Id: id}
	resp, err := h.client.DeleteLink(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to delete link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetLinkByShortCode handles GET /s/{short_code}
func (h *LinkHandler) GetLinkByShortCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["short_code"]

	req := &pb.GetLinkByShortCodeRequest{ShortCode: shortCode}
	resp, err := h.client.GetLinkByShortCode(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Short link not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Auto-redirect if requested
	if r.URL.Query().Get("redirect") == "true" {
		// Record click before redirect
		h.recordClickFromRequest(resp.Link.Id, r)
		http.Redirect(w, r, resp.Link.OriginalUrl, http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// IncrementClick handles POST /s/{short_code}/click
func (h *LinkHandler) IncrementClick(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["short_code"]

	req := &pb.IncrementClickRequest{ShortCode: shortCode}
	resp, err := h.client.IncrementClick(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to increment click: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetLinkStats handles GET /links/{id}/stats
func (h *LinkHandler) GetLinkStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkID := vars["id"]

	req := &pb.GetLinkStatsRequest{LinkId: linkID}
	resp, err := h.client.GetLinkStats(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get link stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RecordClick handles POST /links/{id}/record-click
func (h *LinkHandler) RecordClick(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkID := vars["id"]

	var reqBody struct {
		IPAddress  string  `json:"ip_address"`
		UserAgent  string  `json:"user_agent"`
		Referer    *string `json:"referer,omitempty"`
		Country    *string `json:"country,omitempty"`
		City       *string `json:"city,omitempty"`
		DeviceType *string `json:"device_type,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &pb.RecordClickRequest{
		LinkId:     linkID,
		IpAddress:  reqBody.IPAddress,
		UserAgent:  reqBody.UserAgent,
		Referer:    reqBody.Referer,
		Country:    reqBody.Country,
		City:       reqBody.City,
		DeviceType: reqBody.DeviceType,
	}

	resp, err := h.client.RecordClick(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to record click: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetLinkClicks handles GET /links/{id}/clicks
func (h *LinkHandler) GetLinkClicks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkID := vars["id"]

	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 50
	}

	req := &pb.GetLinkClicksRequest{
		LinkId:   linkID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.client.GetLinkClicks(h.getContextWithAuth(r), req)
	if err != nil {
		http.Error(w, "Failed to get link clicks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetClicksByDate handles GET /links/{id}/clicks-by-date
func (h *LinkHandler) GetClicksByDate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkID := vars["id"]

	// Parse start_date and end_date from query parameters
	// For simplicity, accepting RFC3339 format
	var req pb.GetClicksByDateRequest
	req.LinkId = linkID

	// Note: In production, you'd parse timestamps from query params
	// For now, this is a placeholder implementation
	resp, err := h.client.GetClicksByDate(h.getContextWithAuth(r), &req)
	if err != nil {
		http.Error(w, "Failed to get clicks by date: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Helper function to record click from HTTP request
func (h *LinkHandler) recordClickFromRequest(linkID string, r *http.Request) {
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}

	userAgent := r.Header.Get("User-Agent")
	referer := r.Header.Get("Referer")

	req := &pb.RecordClickRequest{
		LinkId:    linkID,
		IpAddress: ip,
		UserAgent: userAgent,
	}

	if referer != "" {
		req.Referer = &referer
	}

	// Fire and forget - don't wait for response
	go h.client.RecordClick(h.getContextWithAuth(r), req)
}
