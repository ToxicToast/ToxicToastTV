package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/toxictoast/toxictoastgo/shared/cqrs"

	"toxictoast/services/warcraft-service/internal/domain"
	"toxictoast/services/warcraft-service/internal/repository"
	"toxictoast/services/warcraft-service/pkg/blizzard"
)

// ============================================================================
// Queries
// ============================================================================

// GetGuildQuery retrieves a guild by ID
type GetGuildQuery struct {
	cqrs.BaseQuery
	GuildID string `json:"guild_id"`
}

func (q *GetGuildQuery) QueryName() string {
	return "get_guild"
}

func (q *GetGuildQuery) Validate() error {
	if q.GuildID == "" {
		return errors.New("guild_id is required")
	}
	return nil
}

// ListGuildsQuery retrieves a list of guilds with filters
type ListGuildsQuery struct {
	cqrs.BaseQuery
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
	Region   *string `json:"region"`
	Realm    *string `json:"realm"`
	Faction  *string `json:"faction"`
}

func (q *ListGuildsQuery) QueryName() string {
	return "list_guilds"
}

func (q *ListGuildsQuery) Validate() error {
	// No strict validation required for filtering
	return nil
}

// GetGuildRosterQuery retrieves guild roster from Blizzard API
// Note: This fetches fresh data from Blizzard API (not cached)
type GetGuildRosterQuery struct {
	cqrs.BaseQuery
	GuildID  string `json:"guild_id"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

func (q *GetGuildRosterQuery) QueryName() string {
	return "get_guild_roster"
}

func (q *GetGuildRosterQuery) Validate() error {
	if q.GuildID == "" {
		return errors.New("guild_id is required")
	}
	return nil
}

// ============================================================================
// Query Results
// ============================================================================

// ListGuildsResult contains the result of listing guilds
type ListGuildsResult struct {
	Guilds []*domain.Guild
	Total  int
}

// GetGuildRosterResult contains the guild roster members
type GetGuildRosterResult struct {
	Members []domain.GuildMember
	Total   int
}

// ============================================================================
// Query Handlers
// ============================================================================

// GetGuildHandler handles guild retrieval by ID
type GetGuildHandler struct {
	repo repository.GuildRepository
}

func NewGetGuildHandler(repo repository.GuildRepository) *GetGuildHandler {
	return &GetGuildHandler{
		repo: repo,
	}
}

func (h *GetGuildHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetGuildQuery)

	guild, err := h.repo.FindByID(ctx, q.GuildID)
	if err != nil {
		return nil, fmt.Errorf("guild not found: %w", err)
	}

	return guild, nil
}

// ListGuildsHandler handles guild listing with filters
type ListGuildsHandler struct {
	repo repository.GuildRepository
}

func NewListGuildsHandler(repo repository.GuildRepository) *ListGuildsHandler {
	return &ListGuildsHandler{
		repo: repo,
	}
}

func (h *ListGuildsHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*ListGuildsQuery)

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

	guilds, total, err := h.repo.List(ctx, page, pageSize, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list guilds: %w", err)
	}

	return &ListGuildsResult{
		Guilds: guilds,
		Total:  total,
	}, nil
}

// GetGuildRosterHandler handles guild roster retrieval
// Note: This fetches fresh data from Blizzard API (not cached)
type GetGuildRosterHandler struct {
	guildRepo      repository.GuildRepository
	blizzardClient *blizzard.Client
}

func NewGetGuildRosterHandler(
	guildRepo repository.GuildRepository,
	blizzardClient *blizzard.Client,
) *GetGuildRosterHandler {
	return &GetGuildRosterHandler{
		guildRepo:      guildRepo,
		blizzardClient: blizzardClient,
	}
}

func (h *GetGuildRosterHandler) Handle(ctx context.Context, query cqrs.Query) (interface{}, error) {
	q := query.(*GetGuildRosterQuery)

	guild, err := h.guildRepo.FindByID(ctx, q.GuildID)
	if err != nil {
		return nil, fmt.Errorf("guild not found: %w", err)
	}

	// Fetch roster from Blizzard API
	members, err := h.blizzardClient.GetGuildRoster(ctx, guild.Name, guild.Realm, guild.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roster from Blizzard API: %w", err)
	}

	// Default pagination
	page := q.Page
	pageSize := q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// Apply pagination
	total := len(members)
	start := (page - 1) * pageSize
	if start >= total {
		return &GetGuildRosterResult{
			Members: []domain.GuildMember{},
			Total:   total,
		}, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return &GetGuildRosterResult{
		Members: members[start:end],
		Total:   total,
	}, nil
}
