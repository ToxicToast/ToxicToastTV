# Background Jobs Overview

This document provides an overview of all automated background job schedulers across ToxicToastGo services.

## Summary

| Service | Schedulers | Purpose |
|---------|-----------|---------|
| **blog-service** | 1 | Automated content publishing |
| **foodfolio-service** | 2 | Inventory monitoring and alerts |
| **link-service** | 1 | Link lifecycle management |
| **notification-service** | 2 | Delivery retry and cleanup |
| **sse-service** | 1 | Connection management |
| **twitchbot-service** | 2 | Data cleanup and session management |
| **warcraft-service** | 2 | External API synchronization |
| **webhook-service** | 2 | Delivery retry and cleanup |
| **gateway-service** | 0 | N/A (routing only) |

**Total: 13 schedulers across 8 services**

---

## Blog Service

### Post Publisher Scheduler
- **Interval:** 5 minutes (default)
- **Purpose:** Automatically publishes scheduled blog posts
- **Configuration:**
  - `POST_PUBLISHER_ENABLED=true`
  - `POST_PUBLISHER_INTERVAL=5m`
- **Behavior:**
  - Finds draft posts with `published_at` in the past
  - Changes status from `draft` to `published`
  - Publishes `blog.post.scheduled.published` Kafka event

---

## Foodfolio Service

### 1. Item Expiration Scheduler
- **Interval:** 24 hours (default)
- **Purpose:** Monitors food item expiration dates
- **Configuration:**
  - `ITEM_EXPIRATION_ENABLED=true`
  - `ITEM_EXPIRATION_INTERVAL=24h`
- **Behavior:**
  - Checks all item details for expiration
  - Publishes `foodfolio.detail.expired` for expired items
  - Publishes `foodfolio.detail.expiring.soon` for items expiring within 7 days

### 2. Stock Level Scheduler
- **Interval:** 6 hours (default)
- **Purpose:** Monitors inventory stock levels
- **Configuration:**
  - `STOCK_LEVEL_ENABLED=true`
  - `STOCK_LEVEL_INTERVAL=6h`
- **Behavior:**
  - Checks all item variant stock levels
  - Publishes `foodfolio.variant.stock.empty` when stock = 0
  - Publishes `foodfolio.variant.stock.low` when stock < min_sku

---

## Link Service

### Link Expiration Scheduler
- **Interval:** 1 hour (default)
- **Purpose:** Deactivates expired short links
- **Configuration:**
  - `LINK_EXPIRATION_ENABLED=true`
  - `LINK_EXPIRATION_INTERVAL=1h`
- **Behavior:**
  - Finds active links with `expiry_date` in the past
  - Sets `is_active=false`
  - Publishes `link.expired` Kafka event

---

## Notification Service

### 1. Notification Retry Scheduler
- **Interval:** 5 minutes (default)
- **Purpose:** Retries failed Discord notification deliveries
- **Configuration:**
  - `NOTIFICATION_RETRY_ENABLED=true`
  - `NOTIFICATION_RETRY_INTERVAL=5m`
  - `NOTIFICATION_RETRY_MAX_RETRIES=3`
- **Behavior:**
  - Finds failed notifications with attempts < max_retries
  - Reconstructs event from stored payload
  - Attempts redelivery to Discord webhook
  - Increments attempt counter
  - Updates status on success/failure

### 2. Notification Cleanup Scheduler
- **Interval:** 24 hours (default)
- **Purpose:** Removes old successful notifications
- **Configuration:**
  - `NOTIFICATION_CLEANUP_ENABLED=true`
  - `NOTIFICATION_CLEANUP_INTERVAL=24h`
  - `NOTIFICATION_CLEANUP_RETENTION_DAYS=30`
- **Behavior:**
  - Deletes successful notifications older than retention period
  - Maintains database performance

---

## SSE Service

### Client Cleanup Scheduler
- **Interval:** 5 minutes (default)
- **Purpose:** Disconnects inactive SSE client connections
- **Configuration:**
  - `SSE_CLIENT_CLEANUP_ENABLED=true`
  - `SSE_CLIENT_CLEANUP_INTERVAL=5m`
  - `SSE_CLIENT_CLEANUP_INACTIVE_TIMEOUT=30m`
- **Behavior:**
  - Checks last event time for all connected clients
  - Disconnects clients inactive for 30+ minutes
  - Frees server resources
  - Prevents zombie connections

---

## Twitchbot Service

### 1. Message Cleanup Scheduler
- **Interval:** 24 hours (default)
- **Purpose:** Permanently deletes old chat messages
- **Configuration:**
  - `MESSAGE_CLEANUP_ENABLED=true`
  - `MESSAGE_CLEANUP_INTERVAL=24h`
  - `MESSAGE_CLEANUP_RETENTION_DAYS=90`
- **Behavior:**
  - Hard deletes chat messages older than 90 days
  - Maintains database performance
  - Reduces storage costs

### 2. Stream Session Closer Scheduler
- **Interval:** 1 hour (default)
- **Purpose:** Closes orphaned active stream sessions
- **Configuration:**
  - `STREAM_CLOSER_ENABLED=true`
  - `STREAM_CLOSER_INTERVAL=1h`
  - `STREAM_CLOSER_INACTIVE_TIMEOUT=24h`
- **Behavior:**
  - Finds active streams not updated in 24+ hours
  - Calls `EndStream()` to mark as ended
  - Handles edge cases where bot didn't gracefully disconnect

---

## Warcraft Service

### 1. Character Sync Scheduler
- **Interval:** Configurable
- **Purpose:** Syncs character data from Battle.net API
- **Configuration:**
  - See `warcraft-service` documentation
- **Behavior:**
  - Fetches latest character data from Blizzard API
  - Updates character stats, gear, achievements
  - Publishes update events

### 2. Guild Sync Scheduler
- **Interval:** Configurable
- **Purpose:** Syncs guild data from Battle.net API
- **Configuration:**
  - See `warcraft-service` documentation
- **Behavior:**
  - Fetches latest guild roster and data
  - Updates member information
  - Publishes update events

---

## Webhook Service

### 1. Webhook Retry Scheduler
- **Interval:** 5 minutes (default)
- **Purpose:** Schedules failed webhook deliveries for retry
- **Configuration:**
  - `WEBHOOK_RETRY_SCHEDULER_ENABLED=true`
  - `WEBHOOK_RETRY_SCHEDULER_INTERVAL=5m`
  - `WEBHOOK_MAX_RETRIES=3`
- **Behavior:**
  - Finds failed deliveries with attempts < max_retries
  - Calculates next retry time with exponential backoff (2^attempt minutes)
  - Sets status to `retrying` with `next_retry_at` timestamp
  - Delivery pool automatically picks up retries when ready

### 2. Webhook Cleanup Scheduler
- **Interval:** 24 hours (default)
- **Purpose:** Removes old webhook delivery records
- **Configuration:**
  - `WEBHOOK_CLEANUP_ENABLED=true`
  - `WEBHOOK_CLEANUP_INTERVAL=24h`
  - `WEBHOOK_CLEANUP_RETENTION_DAYS=30`
- **Behavior:**
  - Deletes both successful and failed deliveries older than retention period
  - Maintains database performance

---

## Architecture Pattern

All schedulers follow a consistent pattern:

### Structure
```go
type XxxScheduler struct {
    repo        Repository
    interval    time.Duration
    enabled     bool
    stopChan    chan struct{}
}
```

### Lifecycle
1. **Initialization** - Created in `main.go` with config
2. **Start** - Launches goroutine with ticker
3. **Execution** - Runs immediately, then on interval
4. **Stop** - Graceful shutdown via stopChan

### Configuration
- All schedulers support enable/disable flag
- All intervals configurable via environment variables
- Duration format: `5m`, `1h`, `24h`, etc.

### Integration
```go
// In main.go
scheduler := NewXxxScheduler(repo, interval, enabled)
scheduler.Start()
defer scheduler.Stop()
```

---

## Monitoring Recommendations

### Logs
All schedulers log:
- Start/stop events
- Execution cycles
- Item counts processed
- Errors encountered

### Metrics (Future)
Consider adding:
- Execution duration
- Items processed per run
- Error rates
- Last successful run timestamp

### Alerts (Future)
Monitor for:
- Scheduler failures
- Excessive retry attempts
- Cleanup backlog growth
- Resource exhaustion

---

## Best Practices

1. **Disable in Development** - Set `*_ENABLED=false` for faster startup
2. **Adjust Intervals** - Tune based on data volume and load
3. **Monitor Logs** - Watch for errors and performance issues
4. **Retention Policies** - Balance storage costs vs. data needs
5. **Graceful Degradation** - Services continue if schedulers fail

---

## Related Documentation

- [KAFKA_TOPICS.md](./KAFKA_TOPICS.md) - Event types published by schedulers
- [KEYCLOAK_SETUP.md](./KEYCLOAK_SETUP.md) - Authentication configuration
- Service-specific READMEs - Detailed scheduler documentation
