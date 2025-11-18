package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// HTTPMiddleware returns an HTTP middleware that records metrics
func (m *Metrics) HTTPMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Track in-flight requests
			m.HTTPRequestsInFlight.Inc()
			defer m.HTTPRequestsInFlight.Dec()

			// Record request size
			if r.ContentLength > 0 {
				m.HTTPRequestSize.WithLabelValues(
					serviceName,
					r.Method,
					r.URL.Path,
				).Observe(float64(r.ContentLength))
			}

			// Wrap response writer
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Track duration
			start := time.Now()
			next.ServeHTTP(rw, r)
			duration := time.Since(start).Seconds()

			// Record metrics
			status := strconv.Itoa(rw.statusCode)

			m.HTTPRequestsTotal.WithLabelValues(
				serviceName,
				r.Method,
				r.URL.Path,
				status,
			).Inc()

			m.HTTPRequestDuration.WithLabelValues(
				serviceName,
				r.Method,
				r.URL.Path,
				status,
			).Observe(duration)

			m.HTTPResponseSize.WithLabelValues(
				serviceName,
				r.Method,
				r.URL.Path,
			).Observe(float64(rw.size))
		})
	}
}
