package middleware

import (
	"net/http"
	"strconv"
	"time"

	"toxictoast/services/gateway-service/internal/metrics"
)

// responseWriterWithMetrics wraps http.ResponseWriter to capture metrics
type responseWriterWithMetrics struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (rw *responseWriterWithMetrics) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWithMetrics) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// Metrics middleware collects Prometheus metrics for HTTP requests
func Metrics(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Increment in-flight requests
			m.IncHTTPInFlight()
			defer m.DecHTTPInFlight()

			// Start timer
			start := time.Now()

			// Wrap response writer
			wrapped := &responseWriterWithMetrics{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(wrapped.statusCode)

			m.RecordHTTPRequest(
				r.Method,
				normalizePath(r.URL.Path),
				status,
				duration,
				wrapped.written,
			)
		})
	}
}

// normalizePath normalizes URL paths for metrics to avoid cardinality explosion
// Converts /api/blog/posts/123 to /api/blog/posts/:id
func normalizePath(path string) string {
	// List of known path patterns
	patterns := map[string]string{
		"/api/blog/posts/":      "/api/blog/posts/:id",
		"/api/blog/categories/": "/api/blog/categories/:id",
		"/api/blog/tags/":       "/api/blog/tags/:id",
		"/api/blog/media/":      "/api/blog/media/:id",
		"/api/blog/comments/":   "/api/blog/comments/:id",
		"/api/links/":           "/api/links/:id",
		"/api/foodfolio/":       "/api/foodfolio/:path",
		"/api/notifications/":   "/api/notifications/:id",
		"/api/events/":          "/api/events/:id",
		"/api/twitch/":          "/api/twitch/:path",
		"/api/webhooks/":        "/api/webhooks/:id",
	}

	// Check for exact matches first (health, ready, swagger, etc.)
	exactMatches := []string{
		"/health",
		"/ready",
		"/swagger",
		"/swagger/doc.yaml",
		"/metrics",
		"/api/blog/posts",
		"/api/blog/categories",
		"/api/blog/tags",
		"/api/blog/media",
		"/api/blog/comments",
	}

	for _, match := range exactMatches {
		if path == match {
			return path
		}
	}

	// Try to match patterns
	for prefix, normalized := range patterns {
		if len(path) > len(prefix) && path[:len(prefix)] == prefix {
			// Check if there's a sub-path after the ID
			remainder := path[len(prefix):]
			if containsSlash(remainder) {
				// e.g., /api/blog/posts/123/publish
				parts := splitPath(remainder)
				if len(parts) > 1 {
					return normalized[:len(normalized)-3] + "/" + parts[1]
				}
			}
			return normalized
		}
	}

	return path
}

func containsSlash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return true
		}
	}
	return false
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}
