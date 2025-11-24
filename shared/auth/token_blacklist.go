package auth

import (
	"sync"
	"time"
)

// TokenBlacklist manages revoked JWT tokens
type TokenBlacklist struct {
	tokens map[string]time.Time // token -> expiration time
	mu     sync.RWMutex
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}

	// Start cleanup goroutine to remove expired tokens
	go bl.cleanupLoop()

	return bl
}

// Revoke adds a token to the blacklist until its expiration time
func (bl *TokenBlacklist) Revoke(token string, expiresAt time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.tokens[token] = expiresAt
}

// IsRevoked checks if a token is in the blacklist
func (bl *TokenBlacklist) IsRevoked(token string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	expiresAt, exists := bl.tokens[token]
	if !exists {
		return false
	}

	// If token has expired, it's no longer relevant
	if time.Now().After(expiresAt) {
		return false
	}

	return true
}

// cleanupLoop periodically removes expired tokens from the blacklist
func (bl *TokenBlacklist) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bl.cleanup()
	}
}

// cleanup removes expired tokens from the blacklist
func (bl *TokenBlacklist) cleanup() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	for token, expiresAt := range bl.tokens {
		if now.After(expiresAt) {
			delete(bl.tokens, token)
		}
	}
}

// Size returns the number of revoked tokens currently in the blacklist
func (bl *TokenBlacklist) Size() int {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return len(bl.tokens)
}
