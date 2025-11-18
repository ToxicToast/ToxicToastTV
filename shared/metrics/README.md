# Metrics

Prometheus metrics wrapper for consistent metrics across all services.

## Features

- ✅ Standard HTTP metrics (requests, duration, size)
- ✅ Standard gRPC metrics (requests, duration)
- ✅ HTTP middleware for automatic tracking
- ✅ Custom metrics helpers
- ✅ Isolated registries per service

## Usage

### Basic Setup

```go
import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/toxictoast/toxictoastgo/shared/metrics"
)

// Create metrics for your service
m := metrics.New("my-service")

// Serve metrics endpoint
http.Handle("/metrics", promhttp.HandlerFor(m.Registry(), promhttp.HandlerOpts{}))
```

### HTTP Middleware

```go
// Wrap your HTTP server with metrics middleware
mux := http.NewServeMux()
mux.HandleFunc("/api/users", handleUsers)

// Apply metrics middleware
handler := m.HTTPMiddleware("my-service")(mux)

http.ListenAndServe(":8080", handler)
```

### Recording gRPC Metrics

```go
start := time.Now()
// ... handle gRPC request ...
duration := time.Since(start).Seconds()

m.GRPCRequestsTotal.WithLabelValues(
    "my-service",
    "GetUser",
    "OK",
).Inc()

m.GRPCRequestDuration.WithLabelValues(
    "my-service",
    "GetUser",
    "OK",
).Observe(duration)
```

### Custom Metrics

```go
// Counter
userCreated := m.NewCounter(
    "users_created_total",
    "Total number of users created",
    []string{"source"},
)
userCreated.WithLabelValues("api").Inc()

// Gauge
activeConnections := m.NewGauge(
    "active_connections",
    "Number of active connections",
    []string{"type"},
)
activeConnections.WithLabelValues("websocket").Set(42)

// Histogram
processingTime := m.NewHistogram(
    "processing_time_seconds",
    "Time spent processing requests",
    []string{"operation"},
    []float64{0.1, 0.5, 1, 2, 5, 10},
)
processingTime.WithLabelValues("encoding").Observe(1.23)
```

## Standard Metrics

### HTTP Metrics
- `http_requests_total` - Total HTTP requests (labels: service, method, path, status)
- `http_request_duration_seconds` - HTTP request duration (labels: service, method, path, status)
- `http_request_size_bytes` - HTTP request size (labels: service, method, path)
- `http_response_size_bytes` - HTTP response size (labels: service, method, path)
- `http_requests_in_flight` - Current in-flight HTTP requests

### gRPC Metrics
- `grpc_requests_total` - Total gRPC requests (labels: service, method, status)
- `grpc_request_duration_seconds` - gRPC request duration (labels: service, method, status)
