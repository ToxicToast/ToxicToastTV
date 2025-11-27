package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/repository"
	"toxictoast/services/warcraft-service/pkg/blizzard"
)

// ============================================================================
// Queries
// ============================================================================

// GetCharacterQuery retrieves a character by ID
type GetCharacterQuery struct {
	cqrs.BaseQuery
	CharacterID string `json:"character_id"`
}

func (q *GetCharacterQuery) QueryName() string {
	return "get_character"
}

func (q *GetCharacterQuery) Validate() error {
	if q.CharacterID == "" {
		return errors.New("character_id is required")
	}
	return nil
}

// ListCharactersQuery retrieves a list of characters with filters
type ListCharactersQuery struct {
	cqrs.BaseQuery
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
	Region   *string `json:"region"`
	Realm    *string `json:"realm"`
	Faction  *string `json:"faction"`
}

func (q *ListCharactersQuery) QueryName() string {
	return "list_characters"
}

func (q *ListCharactersQuery) Validate() error {
	// No strict validation required for filtering
	return nil
}

// GetCharacterEquipmentQuery retrieves character equipment
// Note: This fetches fresh data from Blizzard API and caches it
type GetCharacterEquipmentQuery struct {
	cqrs.BaseQuery
	CharacterID string `json:"character_id"`
}

func (q *GetCharacterEquipmentQuery) QueryName() string {
	return "get_character_equipment"
}

func (q *GetCharacterEquipmentQuery) Validate() error {
	if q.CharacterID == "" {
		return errors.New("character_id is required")
	}
	return nil
}

// GetCharacterStatsQuery retrieves character stats
// Note: This fetches fresh data from Blizzard API and caches it
type GetCharacterStatsQuery struct {
	cqrs.BaseQuery
	CharacterID string `json:"character_id"`
}

func (q *GetCharacterStatsQuery) QueryName() string {
	return "get_character_stats"
}

func (q *GetCharacterStatsQuery) Validate() error {
	if q.CharacterID == "" {
		return errors.New("character_id is required")
	}
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// ListCharactersResult contains the result of listing characters
type ListCharactersResult struct {
	Characters []*domain.Character
	Total      int
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetCharacterHandler handles character retrieval by ID
type GetCharacterHandler struct {
	repo repository.CharacterRepository
}

func NewGetCharacterHandler(repo repository.CharacterRepository) *GetCharacterHandler {
	return &GetCharacterHandler{
		repo: repo,
	}
}

func (h *GetCharacterHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCharacterQuery)

	character, err := h.repo.FindByID(ctx, q.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}

	return character, nil
}

// ListCharactersHandler handles character listing with filters
type ListCharactersHandler struct {
	repo repository.CharacterRepository
}

func NewListCharactersHandler(repo repository.CharacterRepository) *ListCharactersHandler {
	return &ListCharactersHandler{
		repo: repo,
	}
}

func (h *ListCharactersHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListCharactersQuery)

	// Default pagination
	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// Build filters
	filters := make(map[string]interface{})
	if q.Region != nil {
		filters["region"] = *q.Region
	}
	if q.Realm != nil {
		filters["realm"] = *q.Realm
	}
	if q.Faction != nil {
		filters["faction"] = *q.Faction
	}

	characters, total, err := h.repo.List(ctx, page, pageSize, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}

	return &ListCharactersResult{
		Characters: characters,
		Total:      total,
	}, nil
}

// GetCharacterEquipmentHandler handles equipment retrieval
// Note: This is a hybrid query that fetches from Blizzard API and caches the result
type GetCharacterEquipmentHandler struct {
	characterRepo repository.CharacterRepository
	equipmentRepo repository.CharacterEquipmentRepository
	blizzardClient *blizzard.Client
	kafkaProducer *kafka.Producer
}

func NewGetCharacterEquipmentHandler(
	characterRepo repository.CharacterRepository,
	equipmentRepo repository.CharacterEquipmentRepository,
	blizzardClient *blizzard.Client,
	kafkaProducer *kafka.Producer,
) *GetCharacterEquipmentHandler {
	return &GetCharacterEquipmentHandler{
		characterRepo: characterRepo,
		equipmentRepo: equipmentRepo,
		blizzardClient: blizzardClient,
		kafkaProducer: kafkaProducer,
	}
}

func (h *GetCharacterEquipmentHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCharacterEquipmentQuery)

	character, err := h.characterRepo.FindByID(ctx, q.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}

	// Fetch equipment from Blizzard API
	equipment, err := h.blizzardClient.GetCharacterEquipment(ctx, character.Name, character.Realm, character.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch equipment from Blizzard API: %w", err)
	}

	// Set character ID and save
	equipment.ID = uuid.New().String()
	equipment.CharacterID = q.CharacterID
	equipment.CreatedAt = time.Now()
	equipment.UpdatedAt = time.Now()

	// CreateOrUpdate equipment in database
	equipment, err = h.equipmentRepo.CreateOrUpdate(ctx, equipment)
	if err != nil {
		return nil, fmt.Errorf("failed to save equipment: %w", err)
	}

	// Publish equipment updated event
	if h.kafkaProducer != nil {
		event := kafka.WarcraftCharacterEquipmentUpdatedEvent{
			CharacterID: q.CharacterID,
			Name:        character.Name,
			UpdatedAt:   equipment.UpdatedAt,
		}
		if err := h.kafkaProducer.PublishWarcraftCharacterEquipmentUpdated("warcraft.character.equipment.updated", event); err != nil {
			fmt.Printf("Warning: Failed to publish equipment updated event: %v\n", err)
		}
	}

	return equipment, nil
}

// GetCharacterStatsHandler handles stats retrieval
// Note: This is a hybrid query that fetches from Blizzard API and caches the result
type GetCharacterStatsHandler struct {
	characterRepo repository.CharacterRepository
	statsRepo     repository.CharacterStatsRepository
	blizzardClient *blizzard.Client
	kafkaProducer *kafka.Producer
}

func NewGetCharacterStatsHandler(
	characterRepo repository.CharacterRepository,
	statsRepo repository.CharacterStatsRepository,
	blizzardClient *blizzard.Client,
	kafkaProducer *kafka.Producer,
) *GetCharacterStatsHandler {
	return &GetCharacterStatsHandler{
		characterRepo: characterRepo,
		statsRepo:     statsRepo,
		blizzardClient: blizzardClient,
		kafkaProducer: kafkaProducer,
	}
}

func (h *GetCharacterStatsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetCharacterStatsQuery)

	character, err := h.characterRepo.FindByID(ctx, q.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}

	// Fetch stats from Blizzard API
	stats, err := h.blizzardClient.GetCharacterStats(ctx, character.Name, character.Realm, character.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats from Blizzard API: %w", err)
	}

	// Set character ID and save
	stats.ID = uuid.New().String()
	stats.CharacterID = q.CharacterID
	stats.CreatedAt = time.Now()
	stats.UpdatedAt = time.Now()

	// CreateOrUpdate stats in database
	stats, err = h.statsRepo.CreateOrUpdate(ctx, stats)
	if err != nil {
		return nil, fmt.Errorf("failed to save stats: %w", err)
	}

	// Publish stats updated event
	if h.kafkaProducer != nil {
		event := kafka.WarcraftCharacterStatsUpdatedEvent{
			CharacterID: q.CharacterID,
			Name:        character.Name,
			UpdatedAt:   stats.UpdatedAt,
		}
		if err := h.kafkaProducer.PublishWarcraftCharacterStatsUpdated("warcraft.character.stats.updated", event); err != nil {
			fmt.Printf("Warning: Failed to publish stats updated event: %v\n", err)
		}
	}

	return stats, nil
}
