package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const helixBaseURL = "https://api.twitch.tv/helix"

// HelixClient is a client for the Twitch Helix API
type HelixClient struct {
	clientID     string
	clientSecret string
	accessToken  string
	httpClient   *http.Client
	tokenManager *TokenManager // Optional token manager for auto-refresh
}

// StreamData represents stream information from Helix API
type StreamData struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
	TagIDs       []string  `json:"tag_ids"`
	IsMature     bool      `json:"is_mature"`
}

// ClipData represents clip information from Helix API
type ClipData struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	EmbedURL        string    `json:"embed_url"`
	BroadcasterID   string    `json:"broadcaster_id"`
	BroadcasterName string    `json:"broadcaster_name"`
	CreatorID       string    `json:"creator_id"`
	CreatorName     string    `json:"creator_name"`
	VideoID         string    `json:"video_id"`
	GameID          string    `json:"game_id"`
	Language        string    `json:"language"`
	Title           string    `json:"title"`
	ViewCount       int       `json:"view_count"`
	CreatedAt       time.Time `json:"created_at"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	Duration        float64   `json:"duration"`
}

// UserData represents user information from Helix API
type UserData struct {
	ID              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"display_name"`
	Type            string    `json:"type"`
	BroadcasterType string    `json:"broadcaster_type"`
	Description     string    `json:"description"`
	ProfileImageURL string    `json:"profile_image_url"`
	OfflineImageURL string    `json:"offline_image_url"`
	ViewCount       int       `json:"view_count"`
	CreatedAt       time.Time `json:"created_at"`
}

// ChatterData represents a user in a channel's chat
type ChatterData struct {
	UserID      string `json:"user_id"`
	UserLogin   string `json:"user_login"`
	UserName    string `json:"user_name"`
	IsModerator bool   // Derived from moderator list
	IsVIP       bool   // Derived from VIP list
}

// NewHelixClient creates a new Twitch Helix API client
func NewHelixClient(clientID, clientSecret, accessToken string) *HelixClient {
	return &HelixClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		accessToken:  accessToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetTokenManager sets the token manager for automatic token refresh
func (c *HelixClient) SetTokenManager(tm *TokenManager) {
	c.tokenManager = tm
}

// GetStream gets current stream information for a user
func (c *HelixClient) GetStream(userLogin string) (*StreamData, error) {
	params := url.Values{}
	params.Add("user_login", userLogin)

	var response struct {
		Data []StreamData `json:"data"`
	}

	err := c.makeRequest("GET", "/streams", params, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, nil // Stream is offline
	}

	return &response.Data[0], nil
}

// GetClips gets clips for a broadcaster
func (c *HelixClient) GetClips(broadcasterID string, startedAt, endedAt time.Time, first int) ([]ClipData, error) {
	params := url.Values{}
	params.Add("broadcaster_id", broadcasterID)
	if !startedAt.IsZero() {
		params.Add("started_at", startedAt.Format(time.RFC3339))
	}
	if !endedAt.IsZero() {
		params.Add("ended_at", endedAt.Format(time.RFC3339))
	}
	if first > 0 {
		params.Add("first", fmt.Sprintf("%d", first))
	}

	var response struct {
		Data []ClipData `json:"data"`
	}

	err := c.makeRequest("GET", "/clips", params, &response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// GetUser gets user information
func (c *HelixClient) GetUser(login string) (*UserData, error) {
	params := url.Values{}
	params.Add("login", login)

	var response struct {
		Data []UserData `json:"data"`
	}

	err := c.makeRequest("GET", "/users", params, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &response.Data[0], nil
}

// GetUserByID gets user information by ID
func (c *HelixClient) GetUserByID(userID string) (*UserData, error) {
	params := url.Values{}
	params.Add("id", userID)

	var response struct {
		Data []UserData `json:"data"`
	}

	err := c.makeRequest("GET", "/users", params, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &response.Data[0], nil
}

// GetChatters gets all chatters (viewers) currently in a channel's chat
// Requires moderator:read:chatters scope
func (c *HelixClient) GetChatters(broadcasterID, moderatorID string) ([]ChatterData, error) {
	if broadcasterID == "" {
		return nil, fmt.Errorf("broadcaster_id is required")
	}
	if moderatorID == "" {
		return nil, fmt.Errorf("moderator_id is required")
	}

	allChatters := []ChatterData{}
	cursor := ""

	// Paginate through all chatters (max 1000 per request)
	for {
		params := url.Values{}
		params.Add("broadcaster_id", broadcasterID)
		params.Add("moderator_id", moderatorID)
		params.Add("first", "1000") // Max allowed

		if cursor != "" {
			params.Add("after", cursor)
		}

		var response struct {
			Data []struct {
				UserID    string `json:"user_id"`
				UserLogin string `json:"user_login"`
				UserName  string `json:"user_name"`
			} `json:"data"`
			Pagination struct {
				Cursor string `json:"cursor"`
			} `json:"pagination"`
		}

		err := c.makeRequest("GET", "/chat/chatters", params, &response)
		if err != nil {
			return nil, err
		}

		// Convert to ChatterData
		for _, chatter := range response.Data {
			allChatters = append(allChatters, ChatterData{
				UserID:    chatter.UserID,
				UserLogin: chatter.UserLogin,
				UserName:  chatter.UserName,
			})
		}

		// Check if there are more pages
		cursor = response.Pagination.Cursor
		if cursor == "" {
			break
		}
	}

	return allChatters, nil
}

// makeRequest makes an HTTP request to the Helix API
func (c *HelixClient) makeRequest(method, endpoint string, params url.Values, result interface{}) error {
	// Get current token (will auto-refresh if needed)
	token := c.accessToken
	if c.tokenManager != nil {
		var err error
		token, err = c.tokenManager.GetAccessToken()
		if err != nil {
			return fmt.Errorf("failed to get access token: %w", err)
		}
	}

	reqURL := helixBaseURL + endpoint
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Client-ID", c.clientID)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Handle 401 Unauthorized - token might be expired
	if resp.StatusCode == http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)

		// If we have a token manager, try to refresh and retry once
		if c.tokenManager != nil {
			log.Println("⚠️  API returned 401, attempting token refresh...")
			newToken, err := c.tokenManager.GetAccessToken()
			if err != nil {
				return fmt.Errorf("token refresh failed: %w", err)
			}

			// Retry request with new token
			req.Header.Set("Authorization", "Bearer "+newToken)
			resp2, err := c.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("retry request failed: %w", err)
			}
			defer resp2.Body.Close()

			if resp2.StatusCode != http.StatusOK {
				body2, _ := io.ReadAll(resp2.Body)
				return fmt.Errorf("API retry failed with status %d: %s", resp2.StatusCode, string(body2))
			}

			if err := json.NewDecoder(resp2.Body).Decode(result); err != nil {
				return fmt.Errorf("failed to decode retry response: %w", err)
			}

			return nil
		}

		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
