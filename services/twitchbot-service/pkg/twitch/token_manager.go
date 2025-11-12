package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenResponse represents the response from Twitch OAuth token endpoint
type TokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	ExpiresIn    int      `json:"expires_in"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

// TokenValidationResponse represents the response from token validation
type TokenValidationResponse struct {
	ClientID  string   `json:"client_id"`
	Login     string   `json:"login"`
	Scopes    []string `json:"scopes"`
	UserID    string   `json:"user_id"`
	ExpiresIn int      `json:"expires_in"`
}

// TokenManager handles automatic token refresh
type TokenManager struct {
	clientID       string
	clientSecret   string
	accessToken    string
	refreshToken   string
	expiresAt      time.Time
	mu             sync.RWMutex
	onTokenRefresh func(newToken string) // Callback when token is refreshed
}

// NewTokenManager creates a new token manager
func NewTokenManager(clientID, clientSecret, initialToken, refreshToken string) *TokenManager {
	return &TokenManager{
		clientID:     clientID,
		clientSecret: clientSecret,
		accessToken:  initialToken,
		refreshToken: refreshToken,
	}
}

// SetTokenRefreshCallback sets a callback that's called when token is refreshed
func (tm *TokenManager) SetTokenRefreshCallback(callback func(newToken string)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTokenRefresh = callback
}

// GetAccessToken returns the current access token, refreshing if needed
func (tm *TokenManager) GetAccessToken() (string, error) {
	tm.mu.RLock()

	// Check if token is still valid (with 5 minute buffer)
	if !tm.expiresAt.IsZero() && time.Now().Add(5*time.Minute).Before(tm.expiresAt) {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}

	tm.mu.RUnlock()

	// Token expired or about to expire, refresh it
	return tm.refreshAccessToken()
}

// ValidateToken validates the current token and gets expiration info
// Returns the validated user login name
func (tm *TokenManager) ValidateToken() (string, error) {
	tm.mu.RLock()
	token := tm.accessToken
	tm.mu.RUnlock()

	if token == "" {
		return "", fmt.Errorf("no access token set")
	}

	// Remove oauth: prefix if present
	token = strings.TrimPrefix(token, "oauth:")

	req, err := http.NewRequest("GET", "https://id.twitch.tv/oauth2/validate", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		log.Println("⚠️  Access token is invalid or expired")
		// Try to refresh the token
		_, err := tm.refreshAccessToken()
		return "", err
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token validation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var validation TokenValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validation); err != nil {
		return "", fmt.Errorf("failed to decode validation response: %w", err)
	}

	// Update expiration time
	tm.mu.Lock()
	tm.expiresAt = time.Now().Add(time.Duration(validation.ExpiresIn) * time.Second)
	tm.mu.Unlock()

	log.Printf("✅ Token validated successfully (expires in %d seconds)", validation.ExpiresIn)
	log.Printf("   User: %s (ID: %s)", validation.Login, validation.UserID)
	log.Printf("   Scopes: %v", validation.Scopes)

	// Check if we have required scopes for IRC chat
	requiredScopes := []string{"chat:read", "chat:edit"}
	missingScopes := []string{}
	for _, required := range requiredScopes {
		found := false
		for _, scope := range validation.Scopes {
			if scope == required {
				found = true
				break
			}
		}
		if !found {
			missingScopes = append(missingScopes, required)
		}
	}

	if len(missingScopes) > 0 {
		log.Printf("⚠️  WARNING: Token is missing required scopes: %v", missingScopes)
		log.Println("   The bot may not be able to read or send chat messages!")
		log.Println("   Generate a new token with chat:read and chat:edit scopes")
	}

	return validation.Login, nil
}

// refreshAccessToken refreshes the access token using refresh token or client credentials
func (tm *TokenManager) refreshAccessToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	log.Println("🔄 Refreshing access token...")

	var newToken *TokenResponse
	var err error

	// Try refresh token first if available
	if tm.refreshToken != "" {
		newToken, err = tm.refreshWithRefreshToken()
		if err == nil {
			log.Println("✅ Token refreshed using refresh token")
		} else {
			log.Printf("⚠️  Refresh token failed: %v, trying client credentials...", err)
		}
	}

	// If refresh token failed or not available, try client credentials
	if newToken == nil && tm.clientID != "" && tm.clientSecret != "" {
		newToken, err = tm.refreshWithClientCredentials()
		if err != nil {
			return "", fmt.Errorf("failed to refresh token with client credentials: %w", err)
		}
		log.Println("✅ Token refreshed using client credentials")
	}

	if newToken == nil {
		return "", fmt.Errorf("unable to refresh token: no refresh token or client credentials available")
	}

	// Update stored token
	tm.accessToken = newToken.AccessToken
	if newToken.RefreshToken != "" {
		tm.refreshToken = newToken.RefreshToken
	}
	tm.expiresAt = time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)

	log.Printf("✅ New token expires in %d seconds", newToken.ExpiresIn)

	// Call callback if set
	if tm.onTokenRefresh != nil {
		go tm.onTokenRefresh(tm.accessToken)
	}

	return tm.accessToken, nil
}

// refreshWithRefreshToken refreshes using OAuth refresh token
func (tm *TokenManager) refreshWithRefreshToken() (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", tm.clientID)
	data.Set("client_secret", tm.clientSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", tm.refreshToken)

	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// refreshWithClientCredentials gets a new app access token using client credentials
func (tm *TokenManager) refreshWithClientCredentials() (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", tm.clientID)
	data.Set("client_secret", tm.clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("credentials grant failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// UpdateAccessToken manually updates the access token
func (tm *TokenManager) UpdateAccessToken(token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.accessToken = strings.TrimPrefix(token, "oauth:")
}

// UpdateRefreshToken manually updates the refresh token
func (tm *TokenManager) UpdateRefreshToken(token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.refreshToken = token
}
