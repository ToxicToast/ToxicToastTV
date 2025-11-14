package blizzard

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenResponse represents the OAuth2 token response from Blizzard
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// TokenManager handles OAuth2 token lifecycle
type TokenManager struct {
	clientID     string
	clientSecret string
	region       string

	mu           sync.RWMutex
	accessToken  string
	expiresAt    time.Time
}

// NewTokenManager creates a new token manager
func NewTokenManager(clientID, clientSecret, region string) *TokenManager {
	return &TokenManager{
		clientID:     clientID,
		clientSecret: clientSecret,
		region:       region,
	}
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (tm *TokenManager) GetAccessToken(ctx context.Context) (string, error) {
	tm.mu.RLock()
	// Check if we have a valid token (with 5 minute buffer before expiration)
	if tm.accessToken != "" && time.Now().Add(5*time.Minute).Before(tm.expiresAt) {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	// Need to refresh token
	return tm.refreshToken(ctx)
}

// refreshToken fetches a new access token from Blizzard
func (tm *TokenManager) refreshToken(ctx context.Context) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring lock
	if tm.accessToken != "" && time.Now().Add(5*time.Minute).Before(tm.expiresAt) {
		return tm.accessToken, nil
	}

	// Determine token endpoint based on region
	tokenURL := tm.getTokenURL()

	// Prepare request
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(tm.clientID, tm.clientSecret)

	// Execute request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Store token
	tm.accessToken = tokenResp.AccessToken
	tm.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return tm.accessToken, nil
}

// getTokenURL returns the appropriate OAuth token URL for the region
func (tm *TokenManager) getTokenURL() string {
	// China uses different domain
	if tm.region == "cn" {
		return "https://oauth.battlenet.com.cn/token"
	}
	// All other regions use oauth.battle.net
	return "https://oauth.battle.net/token"
}

// getAPIBaseURL returns the API base URL for the region
func (tm *TokenManager) getAPIBaseURL() string {
	// China uses different domain
	if tm.region == "cn" {
		return "https://gateway.battlenet.com.cn"
	}
	// All other regions use regional subdomains
	return fmt.Sprintf("https://%s.api.blizzard.com", tm.region)
}
