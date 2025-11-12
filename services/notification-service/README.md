# Notification Service

The Notification Service consumes events from Kafka and sends rich Discord notifications using webhooks and embeds.

## Features

- **Discord Integration**: Send rich embeds to Discord channels via webhooks
- **Event Filtering**: Configure which event types go to which Discord channels (supports wildcards like `blog.*`)
- **Multiple Events per Channel**: Multiple event types can be routed to the same Discord channel
- **Rich Embeds**: Beautiful Discord embeds with titles, descriptions, fields, colors, and timestamps
- **Channel Management**: CRUD operations for Discord channels via gRPC
- **Delivery Tracking**: Complete audit trail of all notifications and attempts
- **Statistics**: Track success/failure rates per channel
- **Test Notifications**: Send test messages to verify Discord webhooks

## Architecture

```
Kafka Topics → Consumer → Event Processing
                              ↓
                      Channel Matching
                              ↓
                      Discord Webhook
                              ↓
                      Rich Embed → Discord Channel
                              ↓
                      Database (tracking)
```

## Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL database
- Kafka/Redpanda broker
- Discord webhook URL(s)

### Installation

1. **Navigate to service:**
   ```bash
   cd services/notification-service
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Generate proto files:**
   ```bash
   protoc --go_out=. --go_opt=paths=source_relative \
          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
          api/proto/notification.proto
   ```

4. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

5. **Run the service:**
   ```bash
   go run cmd/server/main.go
   ```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GRPC_PORT` | 9096 | gRPC server port |
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_NAME` | notification_service | Database name |
| `KAFKA_BROKERS` | localhost:9092 | Kafka broker addresses |
| `KAFKA_GROUP_ID` | notification-service | Consumer group ID |
| `KAFKA_TOPICS` | See .env.example | Comma-separated topics |

### Discord Webhook Setup

1. Go to Discord Server Settings → Integrations → Webhooks
2. Create New Webhook
3. Choose the channel
4. Copy the Webhook URL
5. Use this URL when creating a Discord channel via gRPC

## gRPC API

### Channel Management Service

**Create Channel**
```bash
grpcurl -plaintext -d '{
  "name": "Blog Updates",
  "webhook_url": "https://discord.com/api/webhooks/...",
  "event_types": ["blog.*"],
  "color": 3447003,
  "description": "All blog-related notifications"
}' localhost:9096 notification.ChannelManagementService/CreateChannel
```

**List Channels**
```bash
grpcurl -plaintext -d '{
  "limit": 10,
  "active_only": true
}' localhost:9096 notification.ChannelManagementService/ListChannels
```

**Test Channel**
```bash
grpcurl -plaintext -d '{
  "id": "channel-uuid"
}' localhost:9096 notification.ChannelManagementService/TestChannel
```

### Notification Service

**List Notifications**
```bash
grpcurl -plaintext -d '{
  "channel_id": "channel-uuid",
  "limit": 20
}' localhost:9096 notification.NotificationService/ListNotifications
```

**Get Notification**
```bash
grpcurl -plaintext -d '{
  "id": "notification-uuid"
}' localhost:9096 notification.NotificationService/GetNotification
```

## Discord Embed Structure

Embeds sent to Discord include:

- **Title**: Event type (e.g., "📢 blog.post.created")
- **Description**: Event source
- **Color**: Configurable per channel or auto-detected by event type
  - Blue (3447003) - blog events
  - Purple (10181046) - twitchbot events
  - Green (3066993) - link events
- **Fields**: All event data as inline fields
- **Footer**: Event ID
- **Timestamp**: Event timestamp (ISO 8601)

### Example Embed

```json
{
  "title": "📢 blog.post.created",
  "description": "Event from blog-service",
  "color": 3447003,
  "fields": [
    {"name": "title", "value": "My New Blog Post", "inline": true},
    {"name": "author", "value": "John Doe", "inline": true},
    {"name": "slug", "value": "my-new-blog-post", "inline": true}
  ],
  "footer": {"text": "Event ID: abc-123"},
  "timestamp": "2025-01-07T14:30:00Z"
}
```

## Event Type Filtering

Channels support flexible event filtering:

- **Exact Match**: `blog.post.created` - Only this specific event
- **Wildcard**: `blog.*` - All blog events (posts, comments, categories, etc.)
- **Multiple Types**: `blog.posts,twitchbot.streams` - Multiple event types to one channel
- **All Events**: `*` - Send everything

## Database Schema

### discord_channels

- Channel configuration
- Webhook URL
- Event type filters
- Embed color customization
- Statistics (total, success, failed counts)

### notifications

- Notification records
- Event payload (JSON)
- Status (pending, success, failed)
- Discord message ID
- Attempt count

### notification_attempts

- Individual send attempts
- Response status and body
- Duration tracking
- Error messages

## Color Codes (Decimal)

Common Discord embed colors:

| Color | Decimal | Hex |
|-------|---------|-----|
| Green (Success) | 3066993 | #2ECC71 |
| Red (Error) | 15158332 | #E74C3C |
| Yellow (Warning) | 16776960 | #FFFF00 |
| Blue (Info) | 3447003 | #3498DB |
| Purple (Stream) | 10181046 | #9B59B6 |
| Gray (Default) | 5814783 | #58BDDF |

## Usage Examples

### Multiple Events to One Channel

```bash
# Create a channel for all activity
grpcurl -plaintext -d '{
  "name": "All Activity",
  "webhook_url": "https://discord.com/api/webhooks/...",
  "event_types": ["blog.*", "twitchbot.streams", "link.links"],
  "color": 5814783
}' localhost:9096 notification.ChannelManagementService/CreateChannel
```

### Separate Channels for Different Services

```bash
# Blog notifications
grpcurl -plaintext -d '{
  "name": "Blog Updates",
  "webhook_url": "https://discord.com/api/webhooks/.../blog",
  "event_types": ["blog.*"],
  "color": 3447003
}' localhost:9096 notification.ChannelManagementService/CreateChannel

# Stream alerts
grpcurl -plaintext -d '{
  "name": "Stream Alerts",
  "webhook_url": "https://discord.com/api/webhooks/.../streams",
  "event_types": ["twitchbot.streams"],
  "color": 10181046
}' localhost:9096 notification.ChannelManagementService/CreateChannel
```

## Monitoring

### Key Metrics

- **Success Rate**: `success_notifications / total_notifications`
- **Last Notification**: Check `last_notification_at` timestamp
- **Failed Deliveries**: Monitor `failed_notifications` count

### Health Checks

- Kafka consumer is consuming messages
- Database connection is active
- Discord webhooks are responding with 200/204

## Troubleshooting

### Notifications Not Sending

1. Verify Discord webhook URL is valid
2. Check channel is active (`active = true`)
3. Verify event types match (`blog.*` matches `blog.post.created`)
4. Check Kafka consumer logs
5. Test channel with `TestChannel` RPC

### Discord API Errors

- **401 Unauthorized**: Invalid webhook URL
- **404 Not Found**: Webhook deleted or channel removed
- **429 Rate Limited**: Too many messages (Discord limits: 30 requests per minute per webhook)

### Event Not Matching

Check event type pattern matching:
- `blog.*` matches `blog.posts`, `blog.comments`, etc.
- `twitchbot.streams` only matches exactly `twitchbot.streams`
- Use `*` to match all events

## Development

### Project Structure

```
notification-service/
├── api/proto/                # Proto definition
├── cmd/server/               # Application entry point
├── internal/
│   ├── discord/             # Discord client with embeds
│   ├── domain/              # Domain models
│   ├── repository/          # Data access layer
│   ├── usecase/             # Business logic
│   ├── consumer/            # Kafka consumer
│   └── handler/             # gRPC handlers & mappers
└── pkg/config/              # Configuration
```

### Build Binary

```bash
go build -o bin/notification-service cmd/server/main.go
```

### Run Tests

```bash
go test ./...
```

## License

See main repository LICENSE file.
