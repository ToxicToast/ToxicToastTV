package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

// visitor tracks rate limit state for a single IP
type visitor struct {
	tokens     int
	lastSeen   time.Time
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: maximum requests per window
// window: time window duration
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Start cleanup goroutine to remove old visitors
	go rl.cleanupLoop()

	return rl
}

// Limit is a middleware that enforces rate limits
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip := getClientIP(r)

		// Check if request is allowed
		if !rl.allow(ip) {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allow checks if a request from the given IP is allowed
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.RLock()
	v, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if !exists {
		// Create new visitor
		rl.mu.Lock()
		// Double-check after acquiring write lock
		v, exists = rl.visitors[ip]
		if !exists {
			v = &visitor{
				tokens:     rl.rate,
				lastSeen:   time.Now(),
				lastRefill: time.Now(),
			}
			rl.visitors[ip] = v
		}
		rl.mu.Unlock()
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	// Update last seen
	v.lastSeen = time.Now()

	// Refill tokens if window has passed
	now := time.Now()
	if now.Sub(v.lastRefill) >= rl.window {
		v.tokens = rl.rate
		v.lastRefill = now
	}

	// Check if tokens available
	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// cleanupLoop periodically removes inactive visitors
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes visitors that haven't been seen for a while
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-10 * time.Minute)
	for ip, v := range rl.visitors {
		v.mu.Lock()
		if v.lastSeen.Before(threshold) {
			delete(rl.visitors, ip)
		}
		v.mu.Unlock()
	}
}

// getClientIP extracts the client IP from the request
// Checks X-Forwarded-For, X-Real-IP headers before falling back to RemoteAddr
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		for idx := 0; idx < len(xff); idx++ {
			if xff[idx] == ',' {
				return xff[:idx]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Reset resets the rate limiter for a specific IP (useful for testing)
func (rl *RateLimiter) Reset(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.visitors, ip)
}

// Stats returns current rate limiter statistics
func (rl *RateLimiter) Stats() map[string]int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]int)
	stats["total_visitors"] = len(rl.visitors)

	return stats
}
