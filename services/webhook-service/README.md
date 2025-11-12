# Webhook Service

The Webhook Service receives events from Kafka and delivers them to registered HTTP webhook endpoints with retry logic, HMAC signatures, and comprehensive delivery tracking.

## Features

- **Event-Driven Architecture**: Consumes events from Kafka topics
- **Webhook Management**: Register, update, delete, and test webhooks via gRPC
- **Event Filtering**: Webhooks can subscribe to specific event types (supports wildcards like `blog.*`)
- **Reliable Delivery**: Exponential backoff retry logic with configurable attempts
- **Security**: HMAC-SHA256 signatures for webhook payload verification
- **Delivery Tracking**: Complete audit trail of all delivery attempts
- **Worker Pool**: Concurrent delivery processing with configurable workers
- **Automatic Retries**: Background retry processor for failed deliveries
- **Statistics**: Track success/failure rates per webhook
- **Testing**: Send test events to webhooks

## Architecture

```
Kafka Topics → Kafka Consumer → Event Processing
                                       ↓
                               Webhook Matching
                                       ↓
                               Delivery Queue → Worker Pool → HTTP POST
                                       ↓                          ↓
                                   Database ← Attempt Tracking ←┘
                                       ↓
                               Retry Checker (periodic)
```

## Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL database
- Kafka/Redpanda broker
- Protocol Buffers compiler (`protoc`)

### Installation

1. **Clone and navigate to service:**
   ```bash
   cd services/webhook-service
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Generate proto files:**
   ```bash
   protoc --go_out=. --go-grpc_out=. api/proto/webhook/v1/*.proto
   ```

4. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your database and Kafka settings
   ```

5. **Run the service:**
   ```bash
   go run cmd/server/main.go
   ```

## Configuration

See `.env.example` for all configuration options.

### Key Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `GRPC_PORT` | 9095 | gRPC server port |
| `KAFKA_BROKERS` | localhost:9092 | Kafka broker addresses |
| `KAFKA_TOPICS` | See .env.example | Comma-separated list of topics to consume |
| `WEBHOOK_MAX_RETRIES` | 5 | Maximum delivery attempts |
| `WEBHOOK_WORKER_COUNT` | 10 | Number of concurrent workers |
| `WEBHOOK_TIMEOUT_SECONDS` | 30 | HTTP request timeout |

### Retry Configuration

The service uses exponential backoff for retries:

```
Attempt 1: Immediate
Attempt 2: After 5 seconds
Attempt 3: After 10 seconds
Attempt 4: After 20 seconds
Attempt 5: After 40 seconds
```

Max retry delay is capped at 5 minutes (configurable).

## gRPC API

### Webhook Management Service

**Create Webhook**
```protobuf
rpc CreateWebhook(CreateWebhookRequest) returns (WebhookResponse)
```
Register a new webhook endpoint.

**Get Webhook**
```protobuf
rpc GetWebhook(GetWebhookRequest) returns (WebhookResponse)
```
Get webhook details by ID.

**List Webhooks**
```protobuf
rpc ListWebhooks(ListWebhooksRequest) returns (ListWebhooksResponse)
```
List all webhooks with pagination. Filter by active status.

**Update Webhook**
```protobuf
rpc UpdateWebhook(UpdateWebhookRequest) returns (WebhookResponse)
```
Update webhook configuration.

**Delete Webhook**
```protobuf
rpc DeleteWebhook(DeleteWebhookRequest) returns (DeleteWebhookResponse)
```
Soft delete a webhook.

**Toggle Webhook**
```protobuf
rpc ToggleWebhook(ToggleWebhookRequest) returns (WebhookResponse)
```
Enable or disable a webhook.

**Regenerate Secret**
```protobuf
rpc RegenerateSecret(RegenerateSecretRequest) returns (WebhookResponse)
```
Generate a new HMAC secret for a webhook.

**Test Webhook**
```protobuf
rpc TestWebhook(TestWebhookRequest) returns (TestWebhookResponse)
```
Send a test event to a webhook.

### Delivery Service

**Get Delivery**
```protobuf
rpc GetDelivery(GetDeliveryRequest) returns (DeliveryResponse)
```
Get delivery details with all attempts.

**List Deliveries**
```protobuf
rpc ListDeliveries(ListDeliveriesRequest) returns (ListDeliveriesResponse)
```
List deliveries with filters (webhook ID, status, pagination).

**Retry Delivery**
```protobuf
rpc RetryDelivery(RetryDeliveryRequest) returns (RetryDeliveryResponse)
```
Manually retry a failed delivery.

**Cleanup Old Deliveries**
```protobuf
rpc CleanupOldDeliveries(CleanupOldDeliveriesRequest) returns (CleanupOldDeliveriesResponse)
```
Remove old completed/failed deliveries.

**Get Queue Status**
```protobuf
rpc GetQueueStatus(GetQueueStatusRequest) returns (GetQueueStatusResponse)
```
Get current delivery and retry queue sizes.

## Webhook Payload Format

When an event is delivered, the webhook receives a POST request with:

### Headers

```
Content-Type: application/json
User-Agent: ToxicToastGo-Webhook/1.0
X-Webhook-Event: <event_type>
X-Webhook-Delivery: <delivery_id>
X-Webhook-Attempt: <attempt_number>
X-Webhook-Timestamp: <unix_timestamp>
X-Webhook-Signature: <hmac_sha256_signature>
X-Webhook-Signature-256: sha256=<hmac_sha256_signature>
```

### Body

The event payload in JSON format:

```json
{
  "id": "event-uuid",
  "type": "blog.post.created",
  "source": "blog-service",
  "timestamp": "2025-01-07T12:00:00Z",
  "data": {
    "post_id": "123",
    "title": "My Post",
    ...
  }
}
```

## Signature Verification

Webhooks include HMAC-SHA256 signatures for payload verification.

### Verify Signature (Go)

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func verifySignature(payload []byte, secret string, signature string) bool {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(payload)
    expected := hex.EncodeToString(h.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Verify Signature (Node.js)

```javascript
const crypto = require('crypto');

function verifySignature(payload, secret, signature) {
    const hmac = crypto.createHmac('sha256', secret);
    hmac.update(payload);
    const expected = hmac.digest('hex');
    return crypto.timingSafeEqual(
        Buffer.from(expected),
        Buffer.from(signature)
    );
}
```

## Event Type Filtering

Webhooks can subscribe to specific event types using exact matches or wildcards:

- **Exact**: `blog.post.created` - Only matches this specific event
- **Wildcard**: `blog.*` - Matches `blog.post.created`, `blog.comment.added`, etc.
- **All**: `*` - Matches all events

## Database Schema

### Webhooks Table

- `id` - UUID primary key
- `url` - Webhook endpoint URL (unique)
- `secret` - HMAC secret
- `event_types` - Comma-separated event type patterns
- `description` - Optional description
- `active` - Active status (boolean)
- `total_deliveries` - Total delivery count
- `success_deliveries` - Successful delivery count
- `failed_deliveries` - Failed delivery count
- `last_delivery_at` - Last delivery timestamp
- `last_success_at` - Last successful delivery timestamp
- `last_failure_at` - Last failed delivery timestamp
- Timestamps: `created_at`, `updated_at`, `deleted_at`

### Deliveries Table

- `id` - UUID primary key
- `webhook_id` - Foreign key to webhooks
- `event_id` - Original event ID
- `event_type` - Event type
- `event_payload` - JSON payload
- `status` - pending | success | failed | retrying
- `attempt_count` - Number of attempts
- `next_retry_at` - Next retry timestamp
- `last_attempt_at` - Last attempt timestamp
- `last_error` - Last error message
- Timestamps: `created_at`, `updated_at`, `deleted_at`

### Delivery Attempts Table

- `id` - UUID primary key
- `delivery_id` - Foreign key to deliveries
- `attempt_number` - Attempt number (1-based)
- `request_url` - Webhook URL at time of attempt
- `response_status` - HTTP status code
- `response_body` - Response body (truncated to 10KB)
- `success` - Success flag
- `error_message` - Error message if failed
- `duration_ms` - Request duration in milliseconds
- `created_at` - Attempt timestamp

## Usage Examples

### Using grpcurl

**Create a webhook:**
```bash
grpcurl -plaintext -d '{
  "url": "https://example.com/webhook",
  "event_types": ["blog.*", "twitchbot.streams"],
  "description": "My webhook"
}' localhost:9095 webhook.v1.WebhookManagementService/CreateWebhook
```

**List webhooks:**
```bash
grpcurl -plaintext -d '{
  "limit": 10,
  "active_only": true
}' localhost:9095 webhook.v1.WebhookManagementService/ListWebhooks
```

**Test a webhook:**
```bash
grpcurl -plaintext -d '{
  "id": "webhook-uuid"
}' localhost:9095 webhook.v1.WebhookManagementService/TestWebhook
```

**Get delivery status:**
```bash
grpcurl -plaintext -d '{
  "id": "delivery-uuid"
}' localhost:9095 webhook.v1.DeliveryService/GetDelivery
```

**Get queue status:**
```bash
grpcurl -plaintext \
  localhost:9095 webhook.v1.DeliveryService/GetQueueStatus
```

## Monitoring

### Key Metrics to Monitor

- **Queue Sizes**: Check `GetQueueStatus` for queue buildup
- **Delivery Success Rate**: Monitor webhook statistics
- **Retry Count**: High retry counts indicate endpoint issues
- **Processing Time**: Check `duration_ms` in delivery attempts

### Health Checks

- Kafka consumer is connected and consuming messages
- Database connection is active
- Worker pool is processing deliveries
- No excessive queue buildup

## Troubleshooting

### Deliveries Not Being Sent

1. Check webhook is active: `active = true`
2. Verify event type matches webhook filter
3. Check queue status - may be full
4. Review Kafka consumer logs

### High Failure Rates

1. Verify webhook endpoint is accessible
2. Check webhook timeout settings
3. Review delivery attempt error messages
4. Verify signature verification on webhook endpoint

### Queue Buildup

1. Increase worker count: `WEBHOOK_WORKER_COUNT`
2. Increase queue size: `WEBHOOK_QUEUE_SIZE`
3. Check for slow/failing webhook endpoints
4. Consider cleaning up old deliveries

## Development

### Project Structure

```
webhook-service/
├── api/proto/webhook/v1/     # Proto definitions
├── cmd/server/               # Application entry point
├── internal/
│   ├── consumer/            # Kafka consumer
│   ├── delivery/            # Worker pool & delivery logic
│   ├── domain/              # Domain models
│   ├── handler/             # gRPC handlers & mappers
│   ├── repository/          # Data access layer
│   └── usecase/             # Business logic
└── pkg/config/              # Configuration
```

### Generate Proto Files

```bash
protoc --go_out=. --go-grpc_out=. api/proto/webhook/v1/*.proto
```

### Run Tests

```bash
go test ./...
```

### Build Binary

```bash
go build -o bin/webhook-service cmd/server/main.go
```

## Contributing

Follow Clean Architecture principles:
- Domain models should not depend on external packages
- Business logic goes in use cases
- Infrastructure code goes in repositories, consumers, and delivery workers

## License

See main repository LICENSE file.
