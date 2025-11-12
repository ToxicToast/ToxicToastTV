package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/toxictoast/toxictoastgo/shared/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// Logging middleware logs all HTTP requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request details
		duration := time.Since(start)
		logger.Info(fmt.Sprintf(
			"[%s] %s - Status: %d - Duration: %dms - Bytes: %d - IP: %s - UA: %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration.Milliseconds(),
			wrapped.written,
			r.RemoteAddr,
			r.UserAgent(),
		))
	})
}
