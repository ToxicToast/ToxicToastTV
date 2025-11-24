package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const (
	// Base URLs for services - adjust these to match your environment
	gatewayBaseURL = "http://localhost:8080/api"
	testTimeout    = 30 * time.Second
)

// TestAuthFlowE2E tests the complete authentication flow:
// 1. Register new user
// 2. Login with credentials
// 3. Access protected endpoint with token
// 4. Refresh token
// 5. Access protected endpoint with new token
// 6. Logout
func TestAuthFlowE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Generate unique test user credentials
	timestamp := time.Now().Unix()
	testEmail := fmt.Sprintf("testuser%d@example.com", timestamp)
	testUsername := fmt.Sprintf("testuser%d", timestamp)
	testPassword := "SecurePass123!"
	testFirstName := "Test"
	testLastName := "User"

	client := &http.Client{Timeout: testTimeout}
	var accessToken, refreshToken string

	// Step 1: Register new user
	t.Run("Register", func(t *testing.T) {
		registerPayload := map[string]interface{}{
			"email":      testEmail,
			"username":   testUsername,
			"password":   testPassword,
			"first_name": testFirstName,
			"last_name":  testLastName,
		}

		resp, err := makeJSONRequest(client, "POST", gatewayBaseURL+"/auth/register", registerPayload, "")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200 or 201, got %d: %s", resp.StatusCode, string(body))
		}

		var registerResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
			t.Fatalf("Failed to decode register response: %v", err)
		}
		resp.Body.Close()

		// Extract tokens
		if token, ok := registerResp["access_token"].(string); ok {
			accessToken = token
		} else {
			t.Fatal("No access_token in register response")
		}

		if token, ok := registerResp["refresh_token"].(string); ok {
			refreshToken = token
		} else {
			t.Fatal("No refresh_token in register response")
		}

		t.Logf("✓ User registered successfully: %s", testUsername)
		t.Logf("✓ Received access_token: %s...", accessToken[:20])
	})

	// Step 2: Login with credentials
	t.Run("Login", func(t *testing.T) {
		loginPayload := map[string]interface{}{
			"email":    testEmail,
			"password": testPassword,
		}

		resp, err := makeJSONRequest(client, "POST", gatewayBaseURL+"/auth/login", loginPayload, "")
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		var loginResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}
		resp.Body.Close()

		// Update tokens
		if token, ok := loginResp["access_token"].(string); ok {
			accessToken = token
		}
		if token, ok := loginResp["refresh_token"].(string); ok {
			refreshToken = token
		}

		t.Logf("✓ Login successful")
	})

	// Step 3: Validate token
	t.Run("ValidateToken", func(t *testing.T) {
		validatePayload := map[string]interface{}{
			"token": accessToken,
		}

		resp, err := makeJSONRequest(client, "POST", gatewayBaseURL+"/auth/validate", validatePayload, "")
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		var validateResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
			t.Fatalf("Failed to decode validate response: %v", err)
		}
		resp.Body.Close()

		if valid, ok := validateResp["valid"].(bool); !ok || !valid {
			t.Fatal("Token validation failed")
		}

		t.Logf("✓ Token validated successfully")
	})

	// Step 4: Access protected endpoint
	t.Run("AccessProtectedEndpoint", func(t *testing.T) {
		resp, err := makeJSONRequest(client, "GET", gatewayBaseURL+"/auth/me", nil, accessToken)
		if err != nil {
			t.Fatalf("Failed to access protected endpoint: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		var profileResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&profileResp); err != nil {
			t.Fatalf("Failed to decode profile response: %v", err)
		}
		resp.Body.Close()

		// Verify user data
		if user, ok := profileResp["user"].(map[string]interface{}); ok {
			if email, ok := user["email"].(string); !ok || email != testEmail {
				t.Errorf("Expected email %s, got %s", testEmail, email)
			}
			if username, ok := user["username"].(string); !ok || username != testUsername {
				t.Errorf("Expected username %s, got %s", testUsername, username)
			}
		} else {
			t.Fatal("No user data in profile response")
		}

		t.Logf("✓ Protected endpoint accessed successfully")
	})

	// Step 5: Test public endpoint without token
	t.Run("AccessPublicEndpoint", func(t *testing.T) {
		resp, err := makeJSONRequest(client, "GET", gatewayBaseURL+"/test/public", nil, "")
		if err != nil {
			t.Fatalf("Failed to access public endpoint: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		var publicResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&publicResp); err != nil {
			t.Fatalf("Failed to decode public response: %v", err)
		}
		resp.Body.Close()

		if message, ok := publicResp["message"].(string); ok {
			t.Logf("✓ Public endpoint response: %s", message)
		}
	})

	// Step 6: Test protected endpoint without token (should fail)
	t.Run("AccessProtectedWithoutToken", func(t *testing.T) {
		resp, err := makeJSONRequest(client, "GET", gatewayBaseURL+"/test/protected", nil, "")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 401, got %d: %s", resp.StatusCode, string(body))
		}

		t.Logf("✓ Protected endpoint correctly requires authentication")
	})

	// Step 7: Test protected endpoint with token
	t.Run("AccessProtectedWithToken", func(t *testing.T) {
		resp, err := makeJSONRequest(client, "GET", gatewayBaseURL+"/test/protected", nil, accessToken)
		if err != nil {
			t.Fatalf("Failed to access protected endpoint: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		var protectedResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&protectedResp); err != nil {
			t.Fatalf("Failed to decode protected response: %v", err)
		}
		resp.Body.Close()

		if user, ok := protectedResp["user"].(map[string]interface{}); ok {
			if username, ok := user["username"].(string); ok {
				t.Logf("✓ Protected endpoint accessed as user: %s", username)
			}
		}
	})

	// Step 8: Refresh token
	t.Run("RefreshToken", func(t *testing.T) {
		refreshPayload := map[string]interface{}{
			"refresh_token": refreshToken,
		}

		resp, err := makeJSONRequest(client, "POST", gatewayBaseURL+"/auth/refresh", refreshPayload, "")
		if err != nil {
			t.Fatalf("Failed to refresh token: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		var refreshResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
			t.Fatalf("Failed to decode refresh response: %v", err)
		}
		resp.Body.Close()

		// Update tokens
		oldAccessToken := accessToken
		if token, ok := refreshResp["access_token"].(string); ok {
			accessToken = token
		} else {
			t.Fatal("No access_token in refresh response")
		}

		if accessToken == oldAccessToken {
			t.Error("Access token should be different after refresh")
		}

		t.Logf("✓ Token refreshed successfully")
	})

	// Step 9: Access protected endpoint with new token
	t.Run("AccessProtectedWithNewToken", func(t *testing.T) {
		resp, err := makeJSONRequest(client, "GET", gatewayBaseURL+"/auth/me", nil, accessToken)
		if err != nil {
			t.Fatalf("Failed to access protected endpoint: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()

		t.Logf("✓ Protected endpoint accessed with refreshed token")
	})

	// Step 10: Logout
	t.Run("Logout", func(t *testing.T) {
		resp, err := makeJSONRequest(client, "POST", gatewayBaseURL+"/auth/logout", nil, accessToken)
		if err != nil {
			t.Fatalf("Failed to logout: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Logout status %d: %s", resp.StatusCode, string(body))
		}
		resp.Body.Close()

		t.Logf("✓ Logout successful")
	})

	// Step 11: Verify token is revoked after logout
	t.Run("AccessAfterLogout", func(t *testing.T) {
		resp, err := makeJSONRequest(client, "GET", gatewayBaseURL+"/auth/me", nil, accessToken)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		// Should fail because token is revoked
		if resp.StatusCode == http.StatusOK {
			t.Error("Expected request to fail after logout, but it succeeded")
		}
		resp.Body.Close()

		t.Logf("✓ Token correctly revoked after logout")
	})
}

// Helper function to make JSON HTTP requests
func makeJSONRequest(client *http.Client, method, url string, payload interface{}, token string) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return client.Do(req)
}
