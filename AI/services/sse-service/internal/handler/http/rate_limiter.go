package http

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter tracks rate limits per IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	r        rate.Limit
	b        int
	enabled  bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMin int, burstSize int, enabled bool) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        rate.Limit(float64(requestsPerMin) / 60.0), // Convert per-minute to per-second
		b:        burstSize,
		enabled:  enabled,
	}
}

// GetLimiter gets or creates a rate limiter for an IP
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.limiters[ip] = limiter

		// Clean up old limiters (after 5 minutes of inactivity)
		go func() {
			time.Sleep(5 * time.Minute)
			rl.mu.Lock()
			delete(rl.limiters, ip)
			rl.mu.Unlock()
		}()
	}

	return limiter
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(r *http.Request) bool {
	if !rl.enabled {
		return true
	}

	ip := getIP(r)
	limiter := rl.GetLimiter(ip)
	return limiter.Allow()
}

// Middleware creates an HTTP middleware for rate limiting
func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow(r) {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// getIP extracts the real IP from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if host, _, err := net.SplitHostPort(forwarded); err == nil {
			return host
		}
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}
