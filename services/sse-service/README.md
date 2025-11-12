# SSE Service

A Server-Sent Events (SSE) service that consumes events from Kafka and streams them to connected clients in real-time.

## Features

- **Real-Time Event Streaming** - Stream events from all services via SSE
- **Kafka Integration** - Consumes events from multiple Kafka topics
- **Subscription Filtering** - Clients can subscribe to specific event types or sources
- **📜 Event History/Replay** - New clients receive last 100 events (configurable)
- **🔄 Last-Event-ID Support** - Proper SSE reconnection with no data loss
- **🚦 Rate Limiting** - Per-IP rate limiting to prevent abuse (60 req/min default)
- **🌐 Configurable CORS** - Production-ready CORS with multiple origin support
- **Client Management** - Track and manage connected clients
- **gRPC Management API** - Monitor and control the service via gRPC
- **Scalable Architecture** - Handle up to 1000+ concurrent connections
- **Automatic Heartbeats** - Keep connections alive with periodic heartbeats
- **Clean Architecture** - Maintainable and testable codebase

## Architecture

```
Kafka Topics → Kafka Consumer → SSE Broker → SSE Clients
                                     ↓
                              gRPC Management API
```

The service follows Clean Architecture principles:

```
sse-service/
├── api/proto/              # gRPC proto definitions
├── cmd/server/             # Service entry point
├── internal/
│   ├── domain/             # Domain models (Event, Client, Subscription)
│   ├── broker/             # SSE broker for client management
│   ├── consumer/           # Kafka consumer
│   └── handler/
│       ├── http/           # HTTP SSE endpoint
│       ├── grpc/           # gRPC management handlers
│       └── mapper/         # Proto mappers
└── pkg/
    └── config/             # Service configuration
```

## Supported Event Types

The service streams events from all ToxicToastGo services:

### Blog Service Events
- `blog.post.created` - New blog post created
- `blog.post.updated` - Blog post updated
- `blog.post.deleted` - Blog post deleted
- `blog.comment.created` - New comment added
- `blog.category.created` - New category created
- `blog.tag.created` - New tag created
- `blog.media.uploaded` - Media file uploaded

### Twitchbot Service Events
- `twitchbot.stream.started` - Stream started
- `twitchbot.stream.ended` - Stream ended
- `twitchbot.message.created` - Chat message received
- `twitchbot.viewer.joined` - Viewer joined channel
- `twitchbot.clip.created` - Clip created
- `twitchbot.command.executed` - Command executed

### Link Service Events
- `link.created` - Short link created
- `link.deleted` - Link deleted
- `link.clicked` - Link was clicked

## ✨ New in v0.2.0

- **Event History** - Clients get last 100 events on connect
- **Last-Event-ID** - Perfect reconnection support (SSE standard)
- **Rate Limiting** - Protection against abuse (60 req/min per IP)
- **CORS Configuration** - Production-ready cross-origin support

See [FEATURES.md](./FEATURES.md) for detailed documentation and examples!

## Getting Started

### Prerequisites

- Go 1.24 or higher
- Kafka/Redpanda running
- Other services publishing events to Kafka

### Installation

1. **Clone the repository**
   ```bash
   cd services/sse-service
   ```

2. **Copy environment file**
   ```bash
   cp .env.example .env
   ```

3. **Configure Kafka**
   Edit `.env`:
   ```env
   KAFKA_BROKERS=localhost:19092
   # List all explicit topic names (Kafka doesn't support wildcards in subscriptions)
   KAFKA_TOPICS=blog.posts,blog.comments,twitchbot.streams,twitchbot.messages,link.links,link.clicks
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Generate proto files** (if needed)
   ```bash
   protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. api/proto/sse.proto
   ```

6. **Run the service**
   ```bash
   go run cmd/server/main.go
   ```

The service will start:
- HTTP SSE server on port `8084`
- gRPC server on port `9094`

## Usage

### Connecting to SSE Stream

**Basic Connection (all events):**
```bash
curl -N http://localhost:8084/events
```

**Filter by Event Types:**
```bash
# Only blog events
curl -N "http://localhost:8084/events?event_types=blog.*"

# Specific event types
curl -N "http://localhost:8084/events?event_types=blog.post.created,twitchbot.message.created"

# Wildcard matching
curl -N "http://localhost:8084/events?event_types=twitchbot.*"
```

**Filter by Source Services:**
```bash
# Only events from blog-service
curl -N "http://localhost:8084/events?sources=blog-service"

# Multiple sources
curl -N "http://localhost:8084/events?sources=blog-service,twitchbot-service"
```

**Combined Filters:**
```bash
curl -N "http://localhost:8084/events?event_types=*.created&sources=blog-service"
```

### JavaScript/Browser Example

```javascript
const eventSource = new EventSource(
  'http://localhost:8084/events?event_types=blog.*,twitchbot.message.*'
);

// Listen for specific event types
eventSource.addEventListener('blog.post.created', (event) => {
  const data = JSON.parse(event.data);
  console.log('New blog post:', data);
});

eventSource.addEventListener('twitchbot.message.created', (event) => {
  const data = JSON.parse(event.data);
  console.log('New Twitch message:', data);
});

// Listen for all events
eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event received:', data);
};

// Handle errors
eventSource.onerror = (error) => {
  console.error('SSE error:', error);
};

// Close connection
// eventSource.close();
```

### React Example

```typescript
import { useEffect, useState } from 'react';

interface Event {
  id: string;
  type: string;
  source: string;
  timestamp: string;
  data: any;
}

function useSSE(eventTypes?: string[], sources?: string[]) {
  const [events, setEvents] = useState<Event[]>([]);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams();
    if (eventTypes) params.set('event_types', eventTypes.join(','));
    if (sources) params.set('sources', sources.join(','));

    const url = `http://localhost:8084/events?${params}`;
    const eventSource = new EventSource(url);

    eventSource.onopen = () => setConnected(true);

    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setEvents(prev => [...prev, data]);
    };

    eventSource.onerror = () => {
      setConnected(false);
    };

    return () => eventSource.close();
  }, []);

  return { events, connected };
}

// Usage
function App() {
  const { events, connected } = useSSE(['blog.*', 'twitchbot.*']);

  return (
    <div>
      <p>Status: {connected ? 'Connected' : 'Disconnected'}</p>
      {events.map(event => (
        <div key={event.id}>
          <strong>{event.type}</strong>: {JSON.stringify(event.data)}
        </div>
      ))}
    </div>
  );
}
```

## Management API (gRPC)

### Get Statistics

```bash
grpcurl -plaintext localhost:9094 sse.SSEManagementService/GetStats
```

### List Connected Clients

```bash
grpcurl -plaintext -d '{"limit": 50, "offset": 0}' \
  localhost:9094 sse.SSEManagementService/GetClients
```

### Disconnect Client

```bash
grpcurl -plaintext -d '{"client_id": "abc-123"}' \
  localhost:9094 sse.SSEManagementService/DisconnectClient
```

### Health Check

```bash
grpcurl -plaintext localhost:9094 sse.SSEManagementService/GetHealth
```

## HTTP Endpoints

- `GET /events` - SSE stream endpoint
  - Query params:
    - `event_types` - Comma-separated event types (supports wildcards)
    - `sources` - Comma-separated source services
- `GET /health` - Health check
- `GET /stats` - Broker statistics (JSON)

## Configuration

See `.env.example` for all available configuration options.

**Key Variables:**

**Kafka:**
- `KAFKA_BROKERS` - Kafka broker addresses
- `KAFKA_TOPICS` - Topics to consume (**explicit names only**, comma-separated)
  - ⚠️ **Note**: Kafka subscriptions don't support wildcards like `blog.*`
  - Use exact topic names: `blog.posts,blog.comments,twitchbot.streams,...`
  - Wildcards work for **event filtering** on the SSE endpoint, not Kafka subscriptions

**SSE:**
- `SSE_MAX_CLIENTS` - Maximum concurrent connections (default: 1000)
- `SSE_HEARTBEAT_SECONDS` - Heartbeat interval (default: 30)
- `SSE_EVENT_BUFFER_SIZE` - Event buffer per client (default: 100)
- `SSE_HISTORY_SIZE` - Number of events to keep for replay (default: 100)

**CORS:**
- `CORS_ALLOWED_ORIGINS` - Allowed origins (default: `*`)
  - Production: `https://app.example.com,https://www.example.com`
  - Development: `*`
- `CORS_ALLOWED_HEADERS` - Allowed headers (default: `Content-Type,Last-Event-ID`)

**Rate Limiting:**
- `RATE_LIMIT_ENABLED` - Enable rate limiting (default: true)
- `RATE_LIMIT_REQUESTS_PER_MIN` - Max requests per IP per minute (default: 60)
- `RATE_LIMIT_BURST_SIZE` - Burst size (default: 10)

## Event Format

All events follow this structure:

```json
{
  "id": "unique-event-id",
  "type": "blog.post.created",
  "source": "blog-service",
  "timestamp": "2025-01-06T12:00:00Z",
  "data": {
    "post_id": "123",
    "title": "My Blog Post",
    ...
  }
}
```

## Performance

- **Concurrent Connections**: Up to 1000+ clients (configurable)
- **Event Throughput**: Limited by Kafka consumer performance
- **Memory**: ~1-5 MB per connected client (depends on buffer size)
- **Latency**: Sub-second event delivery

## Monitoring

The service provides:
- Real-time client count via gRPC
- Per-client statistics (events received, connection time)
- HTTP `/stats` endpoint for monitoring tools
- Structured logging for all events

## Development

### Build

```bash
go build ./...
```

### Test

```bash
go test ./...
```

### Regenerate Proto Files

```bash
protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. api/proto/sse.proto
```

## Troubleshooting

**No events received:**
- Check if Kafka is running and accessible
- **Verify topics exist:** `kafka-topics --list --bootstrap-server localhost:9092`
  - You should see topics like: `blog.posts`, `twitchbot.streams`, etc.
  - If topics don't exist, the other services haven't published events yet
- **Update KAFKA_TOPICS in .env** with the exact topic names from the list command
- Check if other services are publishing events
- Review logs for Kafka consumer errors

**Invalid topic error:**
- Error: `kafka server: The request attempted to perform an operation on an invalid topic`
- Solution: Use explicit topic names in `KAFKA_TOPICS`, not wildcards
- List available topics: `kafka-topics --list --bootstrap-server localhost:9092`
- Update `.env` with exact names: `KAFKA_TOPICS=blog.posts,twitchbot.streams,...`

**Connection drops:**
- Increase `SSE_HEARTBEAT_SECONDS`
- Check firewall/proxy timeouts
- Verify client implements reconnection logic

**High memory usage:**
- Reduce `SSE_EVENT_BUFFER_SIZE`
- Lower `SSE_MAX_CLIENTS`
- Implement client authentication to prevent abuse

## License

Proprietary - ToxicToast
