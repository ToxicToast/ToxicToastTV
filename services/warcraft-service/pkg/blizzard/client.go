package blizzard

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"toxictoast/services/warcraft-service/internal/domain"
)

// Client is a Blizzard API client
type Client struct {
	tokenManager *TokenManager
	httpClient   *http.Client
}

// NewClient creates a new Blizzard API client
func NewClient(clientID, clientSecret, region string) *Client {
	return &Client{
		tokenManager: NewTokenManager(clientID, clientSecret, region),
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// makeRequest is a helper function to make authenticated API requests
func (c *Client) makeRequest(ctx context.Context, method, endpoint string) ([]byte, error) {
	// Get valid access token
	token, err := c.tokenManager.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Build full URL
	baseURL := c.tokenManager.getAPIBaseURL()
	fullURL := baseURL + endpoint

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetCharacter fetches character data from Blizzard API
func (c *Client) GetCharacter(ctx context.Context, name, realm, region string) (*CharacterProfile, error) {
	// Build endpoint
	endpoint := fmt.Sprintf("/profile/wow/character/%s/%s", slugify(realm), strings.ToLower(name))

	// Add namespace query parameter
	endpoint += "?namespace=profile-" + region + "&locale=en_US"

	// Make request
	body, err := c.makeRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	// Parse response
	var apiResp struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Realm struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"realm"`
		Level      int `json:"level"`
		ItemLevel  int `json:"equipped_item_level"`
		Class      struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"character_class"`
		Race struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"race"`
		Faction struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"faction"`
		Guild *struct {
			Name string `json:"name"`
			Realm struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"realm"`
		} `json:"guild"`
		AchievementPoints int    `json:"achievement_points"`
		AvatarURL         string `json:"avatar_url"`
		LastLoginTimestamp int64 `json:"last_login_timestamp"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse character data: %w", err)
	}

	// Map to CharacterProfile
	profile := &CharacterProfile{
		Name:              strings.ToLower(apiResp.Name),
		Realm:             apiResp.Realm.Slug,
		Region:            region,
		DisplayName:       apiResp.Name,
		DisplayRealm:      apiResp.Realm.Name,
		Level:             apiResp.Level,
		ItemLevel:         apiResp.ItemLevel,
		ClassName:         apiResp.Class.Name,
		ClassID:           apiResp.Class.ID,
		RaceName:          apiResp.Race.Name,
		RaceID:            apiResp.Race.ID,
		FactionType:       apiResp.Faction.Type,
		AchievementPoints: apiResp.AchievementPoints,
	}

	if apiResp.AvatarURL != "" {
		profile.ThumbnailURL = &apiResp.AvatarURL
	}

	if apiResp.Guild != nil {
		profile.GuildName = &apiResp.Guild.Name
		profile.GuildRealm = &apiResp.Guild.Realm.Slug
	}

	return profile, nil
}

// GetCharacterEquipment fetches character equipment from Blizzard API
func (c *Client) GetCharacterEquipment(ctx context.Context, name, realm, region string) (*domain.CharacterEquipment, error) {
	// Build endpoint
	endpoint := fmt.Sprintf("/profile/wow/character/%s/%s/equipment", slugify(realm), strings.ToLower(name))
	endpoint += "?namespace=profile-" + region + "&locale=en_US"

	// Make request
	body, err := c.makeRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	// Parse response
	var apiResp struct {
		EquippedItems []struct {
			Slot struct {
				Type string `json:"type"`
				Name string `json:"name"`
			} `json:"slot"`
			Item struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"item"`
			Quality struct {
				Type string `json:"type"`
				Name string `json:"name"`
			} `json:"quality"`
			Level struct {
				Value        int `json:"value"`
				DisplayValue int `json:"display_string"`
			} `json:"level"`
		} `json:"equipped_items"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse equipment data: %w", err)
	}

	// Convert to JSON for storage
	equipmentJSON, err := json.Marshal(apiResp.EquippedItems)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize equipment: %w", err)
	}

	equipment := &domain.CharacterEquipment{
		EquipmentJSON: equipmentJSON,
	}

	return equipment, nil
}

// GetCharacterStats fetches character stats from Blizzard API
func (c *Client) GetCharacterStats(ctx context.Context, name, realm, region string) (*domain.CharacterStats, error) {
	// Build endpoint
	endpoint := fmt.Sprintf("/profile/wow/character/%s/%s/statistics", slugify(realm), strings.ToLower(name))
	endpoint += "?namespace=profile-" + region + "&locale=en_US"

	// Make request
	body, err := c.makeRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	// Parse response
	var apiResp struct {
		Health              int     `json:"health"`
		Power               int     `json:"power"`
		PowerType           string  `json:"power_type"`
		Strength            int     `json:"strength"`
		Agility             int     `json:"agility"`
		Intellect           int     `json:"intellect"`
		Stamina             int     `json:"stamina"`
		MeleeCrit           float64 `json:"melee_crit"`
		MeleeHaste          float64 `json:"melee_haste"`
		Mastery             float64 `json:"mastery"`
		Versatility         float64 `json:"versatility"`
		SpellPower          int     `json:"spell_power"`
		SpellPenetration    int     `json:"spell_penetration"`
		SpellCrit           float64 `json:"spell_crit"`
		AttackPower         int     `json:"attack_power"`
		Armor               int     `json:"armor"`
		Dodge               float64 `json:"dodge"`
		Parry               float64 `json:"parry"`
		Block               float64 `json:"block"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse stats data: %w", err)
	}

	// Convert to JSON for storage
	statsJSON, err := json.Marshal(apiResp)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize stats: %w", err)
	}

	stats := &domain.CharacterStats{
		StatsJSON: statsJSON,
	}

	return stats, nil
}

// GetGuild fetches guild data from Blizzard API
func (c *Client) GetGuild(ctx context.Context, name, realm, region string) (*GuildProfile, error) {
	// Build endpoint - using data API for guilds
	endpoint := fmt.Sprintf("/data/wow/guild/%s/%s", slugify(realm), slugify(name))
	endpoint += "?namespace=profile-" + region + "&locale=en_US"

	// Make request
	body, err := c.makeRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	// Parse response
	var apiResp struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Realm struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"realm"`
		Faction struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"faction"`
		MemberCount       int    `json:"member_count"`
		AchievementPoints int    `json:"achievement_points"`
		Crest             string `json:"crest"`
		CreatedTimestamp  int64  `json:"created_timestamp"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse guild data: %w", err)
	}

	// Map to GuildProfile
	profile := &GuildProfile{
		Name:              apiResp.Name,
		Realm:             apiResp.Realm.Slug,
		Region:            region,
		FactionType:       apiResp.Faction.Type,
		MemberCount:       apiResp.MemberCount,
		AchievementPoints: apiResp.AchievementPoints,
		Crest:             apiResp.Crest,
	}

	return profile, nil
}

// GetGuildRoster fetches guild roster from Blizzard API
func (c *Client) GetGuildRoster(ctx context.Context, name, realm, region string) ([]domain.GuildMember, error) {
	// Build endpoint
	endpoint := fmt.Sprintf("/data/wow/guild/%s/%s/roster", slugify(realm), slugify(name))
	endpoint += "?namespace=profile-" + region + "&locale=en_US"

	// Make request
	body, err := c.makeRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	// Parse response
	var apiResp struct {
		Members []struct {
			Character struct {
				Name  string `json:"name"`
				ID    int    `json:"id"`
				Realm struct {
					Slug string `json:"slug"`
				} `json:"realm"`
				Level int `json:"level"`
				Class struct {
					ID int `json:"id"`
				} `json:"playable_class"`
				Race struct {
					ID int `json:"id"`
				} `json:"playable_race"`
			} `json:"character"`
			Rank int `json:"rank"`
		} `json:"members"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse roster data: %w", err)
	}

	// Map to domain model
	members := make([]domain.GuildMember, len(apiResp.Members))
	for i, member := range apiResp.Members {
		members[i] = domain.GuildMember{
			CharacterName: strings.ToLower(member.Character.Name),
			Rank:          member.Rank,
		}
	}

	return members, nil
}

// slugify converts a realm name to a slug (lowercase, no special chars)
func slugify(s string) string {
	s = strings.ToLower(s)
	// Remove apostrophes and spaces
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// URL encode any remaining special characters
	return url.PathEscape(s)
}
