# Monitoring and Observability

ToxicToastGo includes comprehensive monitoring and observability using Prometheus, Grafana, and OpenTelemetry.

## Quick Start

```bash
# Start monitoring stack
docker-compose up -d prometheus grafana postgres-exporter alertmanager

# Access dashboards
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3001 (admin/admin)
# Alertmanager: http://localhost:9093
```

## Architecture

### Components

| Component | Port | Purpose |
|-----------|------|---------|
| **Prometheus** | 9090 | Metrics collection and alerting |
| **Grafana** | 3001 | Visualization and dashboards |
| **Alertmanager** | 9093 | Alert routing and notifications |
| **Postgres Exporter** | 9187 | PostgreSQL metrics |

### Metrics Endpoints

All services expose metrics at `/metrics`:

- Blog Service: http://localhost:8082/metrics
- FoodFolio Service: http://localhost:8081/metrics
- Link Service: http://localhost:8083/metrics
- Notification Service: http://localhost:8084/metrics
- SSE Service: http://localhost:8085/metrics
- Webhook Service: http://localhost:8086/metrics
- Twitchbot Service: http://localhost:8087/metrics
- Warcraft Service: http://localhost:8088/metrics
- Gateway Service: http://localhost:3000/metrics

## Key Metrics

### Service Health

- `up` - Service availability (1 = up, 0 = down)
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency histogram
- `http_requests_in_flight` - Current active requests

### Application Metrics

- `database_connections_in_use` - Active database connections
- `database_connections_max` - Maximum database connections
- `kafka_messages_produced_total` - Kafka messages sent
- `kafka_messages_consumed_total` - Kafka messages received
- `kafka_consumer_lag` - Consumer lag by topic

### Background Jobs

- `background_job_executions_total` - Job execution count
- `background_job_failures_total` - Job failure count
- `background_job_duration_seconds` - Job execution time
- `notification_retry_queue_size` - Pending notification retries
- `webhook_delivery_failures_total` - Webhook delivery failures

### Go Runtime

- `go_goroutines` - Number of goroutines
- `go_memstats_alloc_bytes` - Allocated memory
- `go_memstats_heap_inuse_bytes` - Heap memory in use
- `process_cpu_seconds_total` - CPU time used
- `process_resident_memory_bytes` - Process memory

## Alerts

### Service Alerts

Located in `monitoring/prometheus/alerts/services.yml`:

- **ServiceDown** - Service unavailable for >2 minutes
- **HighErrorRate** - >5% error rate for >5 minutes
- **HighResponseTime** - p95 latency >1s for >5 minutes
- **DatabasePoolExhausted** - >90% connection pool usage
- **HighMemoryUsage** - >1GB memory usage for >10 minutes
- **KafkaConsumerLag** - >1000 messages lag for >10 minutes

### Background Job Alerts

- **BackgroundJobFailures** - >10% failure rate for >15 minutes
- **NotificationRetryQueueGrowing** - >100 pending retries for >30 minutes
- **WebhookDeliveryFailures** - >20% failure rate for >15 minutes

## Grafana Dashboards

Pre-configured dashboards are auto-imported on startup:

### Services Overview

- Service health status matrix
- Request rate by service
- Error rate trends
- Response time percentiles (p50, p95, p99)
- Active connections and goroutines

### Database Performance

- Connection pool usage
- Query duration percentiles
- Active transactions
- Table sizes and row counts

### Kafka Metrics

- Producer throughput by service
- Consumer lag by topic
- Message processing rate
- Topic partition distribution

### Background Jobs

- Job execution rate by scheduler
- Job failure rate and reasons
- Job duration trends
- Queue sizes

## Configuration

### Prometheus

Edit `monitoring/prometheus/prometheus.yml` to:
- Adjust scrape intervals (default: 15s)
- Add custom scrape targets
- Configure remote storage

```yaml
scrape_configs:
  - job_name: 'my-service'
    static_configs:
      - targets: ['my-service:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s
```

### Grafana

Default credentials:
- Username: `admin`
- Password: `admin`

Datasources are auto-configured in `monitoring/grafana/provisioning/datasources/`.

Custom dashboards go in `monitoring/grafana/dashboards/` and are auto-imported.

### Alertmanager

Edit `monitoring/alertmanager/config.yml` to configure:
- Alert routing rules
- Notification channels (email, Slack, PagerDuty, etc.)
- Inhibition rules
- Grouping strategies

Example: Send critical alerts to webhook service

```yaml
receivers:
  - name: 'critical'
    webhook_configs:
      - url: 'http://webhook-service:8086/webhooks/alertmanager/critical'
        send_resolved: true
```

## Service Instrumentation

### Adding Metrics to Services

Services use the Prometheus client library:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request latency",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

// Expose metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

### Custom Metrics

Add service-specific metrics:

```go
// Blog Service - Post metrics
postsPublished := promauto.NewCounter(prometheus.CounterOpts{
    Name: "blog_posts_published_total",
    Help: "Total published posts",
})

// FoodFolio Service - Item metrics
itemsExpired := promauto.NewCounter(prometheus.CounterOpts{
    Name: "foodfolio_items_expired_total",
    Help: "Total expired items detected",
})

// Link Service - Click metrics
linkClicks := promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "link_clicks_total",
        Help: "Total link clicks",
    },
    []string{"short_code"},
)
```

## Querying Metrics

### Prometheus Queries (PromQL)

```promql
# Request rate (requests per second)
rate(http_requests_total[5m])

# Error percentage
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# p95 latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Memory usage per service
sum by (job) (process_resident_memory_bytes)

# Active database connections
sum by (service) (database_connections_in_use)

# Kafka consumer lag
kafka_consumer_lag{topic="blog.events.post"}

# Background job failure rate
rate(background_job_failures_total[1h])
```

### Grafana Queries

Use the Prometheus datasource with PromQL queries in panels.

## Troubleshooting

### Prometheus Not Scraping

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check service metrics endpoint
curl http://localhost:8082/metrics

# View Prometheus logs
docker-compose logs prometheus
```

### Missing Metrics

```bash
# Verify service is exposing metrics
docker-compose exec blog-service curl localhost:8082/metrics

# Check Prometheus config
docker-compose exec prometheus cat /etc/prometheus/prometheus.yml

# Reload Prometheus config
curl -X POST http://localhost:9090/-/reload
```

### Grafana Dashboard Issues

```bash
# Check Grafana logs
docker-compose logs grafana

# Verify datasource connection
# Go to: http://localhost:3001/datasources

# Manually import dashboard
# Go to: http://localhost:3001/dashboard/import
```

### High Cardinality Warnings

Avoid high-cardinality labels (user IDs, timestamps, etc.):

```go
// ❌ Bad - user_id creates too many time series
requestsTotal.WithLabelValues("GET", "/api/users", userId, "200")

// ✅ Good - use aggregated labels
requestsTotal.WithLabelValues("GET", "/api/users", "200")
```

## Best Practices

1. **Metric Naming**: Use `<namespace>_<subsystem>_<name>_<unit>` convention
2. **Label Cardinality**: Keep labels low-cardinality (< 100 unique values)
3. **Scrape Intervals**: Balance between resolution and storage
4. **Recording Rules**: Pre-compute expensive queries
5. **Alert Fatigue**: Set appropriate thresholds and durations
6. **Dashboard Organization**: Group related metrics, use variables
7. **Retention**: Configure based on storage capacity (default: 15 days)

## Production Considerations

1. **Remote Storage**: Use Thanos or Cortex for long-term storage
2. **High Availability**: Deploy Prometheus in HA mode
3. **Federation**: Aggregate metrics from multiple Prometheus instances
4. **Security**: Enable TLS and authentication
5. **Backup**: Regular Grafana dashboard and Prometheus data backups
6. **Alerting**: Configure multiple notification channels
7. **Documentation**: Document all custom metrics and dashboards

## Related Documentation

- **Docker Setup**: See `DOCKER.md`
- **Background Jobs**: See `BACKGROUND_JOBS.md`
- **Kafka Topics**: See `KAFKA_TOPICS.md`
