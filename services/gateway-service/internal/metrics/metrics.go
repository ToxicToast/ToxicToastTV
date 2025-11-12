package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the gateway
type Metrics struct {
	// HTTP Request metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestsInFlight  prometheus.Gauge
	HTTPResponseSizeBytes *prometheus.HistogramVec

	// gRPC Backend metrics
	GRPCRequestsTotal    *prometheus.CounterVec
	GRPCRequestDuration  *prometheus.HistogramVec
	GRPCRequestsInFlight *prometheus.GaugeVec

	// Rate Limiting metrics
	RateLimitHitsTotal    *prometheus.CounterVec
	RateLimitAllowedTotal *prometheus.CounterVec

	// Backend health metrics
	BackendConnectionsActive *prometheus.GaugeVec
	BackendHealthStatus      *prometheus.GaugeVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP Request metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "gateway_http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),
		HTTPResponseSizeBytes: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "path"},
		),

		// gRPC Backend metrics
		GRPCRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_grpc_requests_total",
				Help: "Total number of gRPC requests to backend services",
			},
			[]string{"service", "method", "status"},
		),
		GRPCRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_grpc_request_duration_seconds",
				Help:    "gRPC request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "status"},
		),
		GRPCRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gateway_grpc_requests_in_flight",
				Help: "Current number of gRPC requests being processed",
			},
			[]string{"service"},
		),

		// Rate Limiting metrics
		RateLimitHitsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_rate_limit_hits_total",
				Help: "Total number of requests that hit rate limit",
			},
			[]string{"ip"},
		),
		RateLimitAllowedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_rate_limit_allowed_total",
				Help: "Total number of requests allowed through rate limit",
			},
			[]string{"ip"},
		),

		// Backend health metrics
		BackendConnectionsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gateway_backend_connections_active",
				Help: "Number of active connections to backend services",
			},
			[]string{"service"},
		),
		BackendHealthStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gateway_backend_health_status",
				Help: "Health status of backend services (1 = healthy, 0 = unhealthy)",
			},
			[]string{"service"},
		),
	}
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration float64, responseSize int) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
	m.HTTPResponseSizeBytes.WithLabelValues(method, path).Observe(float64(responseSize))
}

// RecordGRPCRequest records a gRPC request metric
func (m *Metrics) RecordGRPCRequest(service, method, status string, duration float64) {
	m.GRPCRequestsTotal.WithLabelValues(service, method, status).Inc()
	m.GRPCRequestDuration.WithLabelValues(service, method, status).Observe(duration)
}

// IncHTTPInFlight increments the in-flight HTTP requests counter
func (m *Metrics) IncHTTPInFlight() {
	m.HTTPRequestsInFlight.Inc()
}

// DecHTTPInFlight decrements the in-flight HTTP requests counter
func (m *Metrics) DecHTTPInFlight() {
	m.HTTPRequestsInFlight.Dec()
}

// IncGRPCInFlight increments the in-flight gRPC requests counter for a service
func (m *Metrics) IncGRPCInFlight(service string) {
	m.GRPCRequestsInFlight.WithLabelValues(service).Inc()
}

// DecGRPCInFlight decrements the in-flight gRPC requests counter for a service
func (m *Metrics) DecGRPCInFlight(service string) {
	m.GRPCRequestsInFlight.WithLabelValues(service).Dec()
}

// RecordRateLimitHit records a rate limit hit
func (m *Metrics) RecordRateLimitHit(ip string) {
	m.RateLimitHitsTotal.WithLabelValues(ip).Inc()
}

// RecordRateLimitAllowed records a rate limit allowed request
func (m *Metrics) RecordRateLimitAllowed(ip string) {
	m.RateLimitAllowedTotal.WithLabelValues(ip).Inc()
}

// SetBackendConnectionsActive sets the number of active connections for a service
func (m *Metrics) SetBackendConnectionsActive(service string, count int) {
	m.BackendConnectionsActive.WithLabelValues(service).Set(float64(count))
}

// SetBackendHealthStatus sets the health status for a service
func (m *Metrics) SetBackendHealthStatus(service string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	m.BackendHealthStatus.WithLabelValues(service).Set(status)
}
