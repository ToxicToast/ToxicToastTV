package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	pb "toxictoast/services/warcraft-service/api/proto"
	"google.golang.org/grpc"
)

// WarcraftHandler handles HTTP-to-gRPC translation for warcraft service
type WarcraftHandler struct {
	characterClient pb.CharacterServiceClient
	guildClient     pb.GuildServiceClient
	raceClient      pb.RaceServiceClient
	classClient     pb.ClassServiceClient
	factionClient   pb.FactionServiceClient
}

// NewWarcraftHandler creates a new warcraft handler
func NewWarcraftHandler(conn *grpc.ClientConn) *WarcraftHandler {
	return &WarcraftHandler{
		characterClient: pb.NewCharacterServiceClient(conn),
		guildClient:     pb.NewGuildServiceClient(conn),
		raceClient:      pb.NewRaceServiceClient(conn),
		classClient:     pb.NewClassServiceClient(conn),
		factionClient:   pb.NewFactionServiceClient(conn),
	}
}

// RegisterRoutes registers all warcraft routes
func (h *WarcraftHandler) RegisterRoutes(router *mux.Router) {
	// Character routes
	router.HandleFunc("/characters", h.ListCharacters).Methods("GET")
	router.HandleFunc("/characters", h.CreateCharacter).Methods("POST")
	router.HandleFunc("/characters/{id}", h.GetCharacter).Methods("GET")
	router.HandleFunc("/characters/{id}", h.UpdateCharacter).Methods("PUT")
	router.HandleFunc("/characters/{id}", h.DeleteCharacter).Methods("DELETE")
	router.HandleFunc("/characters/{id}/refresh", h.RefreshCharacter).Methods("POST")
	router.HandleFunc("/characters/{id}/equipment", h.GetCharacterEquipment).Methods("GET")
	router.HandleFunc("/characters/{id}/stats", h.GetCharacterStats).Methods("GET")

	// Guild routes
	router.HandleFunc("/guilds", h.ListGuilds).Methods("GET")
	router.HandleFunc("/guilds", h.CreateGuild).Methods("POST")
	router.HandleFunc("/guilds/{id}", h.GetGuild).Methods("GET")
	router.HandleFunc("/guilds/{id}", h.UpdateGuild).Methods("PUT")
	router.HandleFunc("/guilds/{id}", h.DeleteGuild).Methods("DELETE")
	router.HandleFunc("/guilds/{id}/refresh", h.RefreshGuild).Methods("POST")
	router.HandleFunc("/guilds/{id}/roster", h.GetGuildRoster).Methods("GET")

	// Reference data routes
	router.HandleFunc("/races", h.ListRaces).Methods("GET")
	router.HandleFunc("/races/{id}", h.GetRace).Methods("GET")
	router.HandleFunc("/classes", h.ListClasses).Methods("GET")
	router.HandleFunc("/classes/{id}", h.GetClass).Methods("GET")
	router.HandleFunc("/factions", h.ListFactions).Methods("GET")
	router.HandleFunc("/factions/{id}", h.GetFaction).Methods("GET")
}

// ============================================================================
// Character Handlers
// ============================================================================

func (h *WarcraftHandler) ListCharacters(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	req := &pb.ListCharactersRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if region := r.URL.Query().Get("region"); region != "" {
		req.Region = &region
	}
	if realm := r.URL.Query().Get("realm"); realm != "" {
		req.Realm = &realm
	}
	if faction := r.URL.Query().Get("faction"); faction != "" {
		req.Faction = &faction
	}

	resp, err := h.characterClient.ListCharacters(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.characterClient.CreateCharacter(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) GetCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetCharacterRequest{Id: vars["id"]}

	resp, err := h.characterClient.GetCharacter(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) UpdateCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.characterClient.UpdateCharacter(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.DeleteCharacterRequest{Id: vars["id"]}

	resp, err := h.characterClient.DeleteCharacter(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) RefreshCharacter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.RefreshCharacterRequest{Id: vars["id"]}

	resp, err := h.characterClient.RefreshCharacter(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) GetCharacterEquipment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetCharacterEquipmentRequest{CharacterId: vars["id"]}

	resp, err := h.characterClient.GetCharacterEquipment(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) GetCharacterStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetCharacterStatsRequest{CharacterId: vars["id"]}

	resp, err := h.characterClient.GetCharacterStats(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ============================================================================
// Guild Handlers
// ============================================================================

func (h *WarcraftHandler) ListGuilds(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	req := &pb.ListGuildsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	if region := r.URL.Query().Get("region"); region != "" {
		req.Region = &region
	}
	if realm := r.URL.Query().Get("realm"); realm != "" {
		req.Realm = &realm
	}
	if faction := r.URL.Query().Get("faction"); faction != "" {
		req.Faction = &faction
	}

	resp, err := h.guildClient.ListGuilds(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) CreateGuild(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateGuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.guildClient.CreateGuild(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) GetGuild(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetGuildRequest{Id: vars["id"]}

	resp, err := h.guildClient.GetGuild(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) UpdateGuild(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req pb.UpdateGuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = vars["id"]

	resp, err := h.guildClient.UpdateGuild(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) DeleteGuild(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.DeleteGuildRequest{Id: vars["id"]}

	resp, err := h.guildClient.DeleteGuild(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) RefreshGuild(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.RefreshGuildRequest{Id: vars["id"]}

	resp, err := h.guildClient.RefreshGuild(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) GetGuildRoster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	req := &pb.GetGuildRosterRequest{
		GuildId:  vars["id"],
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	resp, err := h.guildClient.GetGuildRoster(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ============================================================================
// Reference Data Handlers
// ============================================================================

func (h *WarcraftHandler) ListRaces(w http.ResponseWriter, r *http.Request) {
	req := &pb.ListRacesRequest{}

	resp, err := h.raceClient.ListRaces(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) GetRace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetRaceRequest{Id: vars["id"]}

	resp, err := h.raceClient.GetRace(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) ListClasses(w http.ResponseWriter, r *http.Request) {
	req := &pb.ListClassesRequest{}

	resp, err := h.classClient.ListClasses(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) GetClass(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetClassRequest{Id: vars["id"]}

	resp, err := h.classClient.GetClass(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) ListFactions(w http.ResponseWriter, r *http.Request) {
	req := &pb.ListFactionsRequest{}

	resp, err := h.factionClient.ListFactions(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WarcraftHandler) GetFaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req := &pb.GetFactionRequest{Id: vars["id"]}

	resp, err := h.factionClient.GetFaction(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
