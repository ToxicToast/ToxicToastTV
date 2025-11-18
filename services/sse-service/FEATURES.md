# SSE Service - New Features Documentation

## ✨ New Features (v0.2.0)

### 1. 📜 Event History / Replay Buffer

**What it does:**
- Keeps the last 100 events in memory (configurable)
- New clients receive recent history immediately upon connection
- No need to wait for new events to start receiving data

**Configuration:**
```env
SSE_HISTORY_SIZE=100  # Number of events to keep in buffer
```

**Usage:**
```javascript
// Client connects and immediately receives last 100 events
const es = new EventSource('http://localhost:8084/events');
es.onmessage = (e) => {
  const event = JSON.parse(e.data);
  console.log('Event:', event);
  // First 100 events are from history, then live events
};
```

**Benefits:**
- New clients see recent activity immediately
- Better UX - no "empty screen" when connecting
- Catch up on missed events

---

### 2. 🔄 Last-Event-ID Support (SSE Standard)

**What it does:**
- Implements the SSE reconnection standard
- Client can resume from last received event after disconnect
- No duplicate events, no missed events

**How it works:**
```javascript
const es = new EventSource('http://localhost:8084/events');

// Browser automatically sends Last-Event-ID header on reconnect
// EventSource handles this internally!

es.onerror = (e) => {
  console.log('Connection lost, will auto-reconnect with Last-Event-ID');
  // EventSource automatically reconnects and sends:
  // Header: Last-Event-ID: <last-event-id>
};
```

**Manual usage (if needed):**
```bash
curl -N -H "Last-Event-ID: some-event-id-123" \
  http://localhost:8084/events
```

**Benefits:**
- Perfect reconnection handling
- No data loss on temporary disconnects
- Standard SSE behavior
- Browser EventSource API handles it automatically

---

### 3. 🚦 Rate Limiting (Per IP)

**What it does:**
- Limits requests per IP address
- Prevents abuse and DoS attacks
- Configurable limits

**Configuration:**
```env
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MIN=60  # 60 requests per minute
RATE_LIMIT_BURST_SIZE=10        # Allow bursts of 10
```

**Behavior:**
- Tracks each IP separately
- Returns `429 Too Many Requests` when limit exceeded
- Automatically cleans up old limiters (5 min inactivity)

**Example Response (Rate Limited):**
```
HTTP/1.1 429 Too Many Requests
Content-Type: text/plain

Rate limit exceeded. Please try again later.
```

**Benefits:**
- Protects against abuse
- Fair resource allocation
- Automatic cleanup
- X-Forwarded-For support (reverse proxy friendly)

---

### 4. 🌐 Configurable CORS

**What it does:**
- Fine-grained CORS control
- Support for multiple origins
- Proper preflight handling

**Configuration:**

**Development (Allow All):**
```env
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Last-Event-ID
```

**Production (Specific Origins):**
```env
CORS_ALLOWED_ORIGINS=https://app.example.com,https://www.example.com
CORS_ALLOWED_METHODS=GET,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Last-Event-ID,Authorization
```

**Features:**
- Multiple origin support
- Wildcard support (`*`)
- OPTIONS preflight handling
- Credentials support
- Custom headers

**Benefits:**
- Production-ready CORS
- Security best practices
- Easy multi-domain setup

---

## 🎯 Complete Usage Example

### JavaScript with All Features

```javascript
class SSEClient {
  constructor(url, eventTypes = []) {
    this.url = url;
    this.eventTypes = eventTypes;
    this.es = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
  }

  connect() {
    // Build URL with filters
    const params = new URLSearchParams();
    if (this.eventTypes.length > 0) {
      params.set('event_types', this.eventTypes.join(','));
    }

    const fullUrl = `${this.url}?${params}`;
    console.log('Connecting to:', fullUrl);

    // EventSource automatically handles:
    // - Last-Event-ID header on reconnect
    // - CORS preflight
    // - Automatic reconnection
    this.es = new EventSource(fullUrl);

    // Connection established
    this.es.addEventListener('connection.established', (e) => {
      const data = JSON.parse(e.data);
      console.log('✅ Connected:', data.client_id);
      console.log('📜 Receiving event history...');
      this.reconnectAttempts = 0;
    });

    // Listen for specific event types
    this.es.addEventListener('blog.post.created', (e) => {
      const event = JSON.parse(e.data);
      console.log('📝 New blog post:', event.data);
    });

    this.es.addEventListener('twitchbot.message.created', (e) => {
      const event = JSON.parse(e.data);
      console.log('💬 New Twitch message:', event.data);
    });

    // Heartbeat (optional - just for monitoring)
    this.es.addEventListener('heartbeat', (e) => {
      console.log('💓 Heartbeat received');
    });

    // Error handling
    this.es.onerror = (error) => {
      console.error('❌ Connection error:', error);
      console.log('🔄 Browser will auto-reconnect with Last-Event-ID');

      this.reconnectAttempts++;
      if (this.reconnectAttempts >= this.maxReconnectAttempts) {
        console.error('Max reconnect attempts reached, closing...');
        this.close();
      }
    };

    // All events (fallback)
    this.es.onmessage = (e) => {
      const event = JSON.parse(e.data);
      console.log('Event:', event);
    };
  }

  close() {
    if (this.es) {
      this.es.close();
      this.es = null;
    }
  }
}

// Usage
const client = new SSEClient('http://localhost:8084/events', ['blog.*', 'twitchbot.message.*']);
client.connect();

// Later...
// client.close();
```

### React Hook with All Features

```typescript
import { useEffect, useState, useRef } from 'react';

interface SSEEvent {
  id: string;
  type: string;
  source: string;
  timestamp: string;
  data: any;
}

export function useSSE(eventTypes?: string[]) {
  const [events, setEvents] = useState<SSEEvent[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const esRef = useRef<EventSource | null>(null);
  const reconnectAttempts = useRef(0);

  useEffect(() => {
    const params = new URLSearchParams();
    if (eventTypes && eventTypes.length > 0) {
      params.set('event_types', eventTypes.join(','));
    }

    const url = `http://localhost:8084/events?${params}`;
    const es = new EventSource(url);
    esRef.current = es;

    es.onopen = () => {
      setConnected(true);
      setError(null);
      reconnectAttempts.current = 0;
      console.log('✅ SSE Connected');
    };

    es.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data);
        setEvents(prev => [...prev, event]);
      } catch (err) {
        console.error('Failed to parse event:', err);
      }
    };

    es.onerror = (err) => {
      setConnected(false);
      reconnectAttempts.current++;

      if (reconnectAttempts.current > 5) {
        setError('Connection failed after 5 attempts');
        es.close();
      } else {
        setError(`Connection error, attempt ${reconnectAttempts.current}/5`);
      }
    };

    return () => {
      es.close();
    };
  }, [eventTypes?.join(',')]);

  return { events, connected, error };
}

// Usage in component
function App() {
  const { events, connected, error } = useSSE(['blog.*', 'twitchbot.*']);

  return (
    <div>
      <div>
        Status: {connected ? '🟢 Connected' : '🔴 Disconnected'}
        {error && <span> - {error}</span>}
      </div>
      <div>Total Events: {events.length}</div>
      <div>
        {events.slice(-10).map(event => (
          <div key={event.id}>
            <strong>{event.type}</strong>: {JSON.stringify(event.data)}
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## 📊 Monitoring

### Check History Size

```bash
curl http://localhost:8084/stats
```

Response:
```json
{
  "total_clients": 5,
  "max_clients": 1000,
  "history_size": 87,
  "connected_clients": [...]
}
```

### gRPC Management

```bash
grpcurl -plaintext localhost:9094 sse.SSEManagementService/GetStats
```

---

## 🔧 Performance

- **History Buffer**: ~1-5 MB memory (100 events)
- **Rate Limiting**: ~1 KB per IP
- **CORS**: No performance impact
- **Last-Event-ID**: Instant reconnection

---

## 🚀 Migration Guide

### From v0.1.0 to v0.2.0

**No breaking changes!** All new features are opt-in or have sensible defaults.

1. **Update `.env`** (optional):
```env
SSE_HISTORY_SIZE=100
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MIN=60
CORS_ALLOWED_ORIGINS=*
```

2. **Rebuild**:
```bash
go build ./cmd/server
```

3. **Done!** All clients automatically benefit from:
   - Event history on connect
   - Last-Event-ID reconnection (EventSource handles this)
   - Rate limiting protection
   - Proper CORS

---

## 🎉 Summary

| Feature | Benefit | Client Code Changes |
|---------|---------|---------------------|
| Event History | Instant data on connect | ✅ None! |
| Last-Event-ID | Perfect reconnection | ✅ None! (EventSource handles it) |
| Rate Limiting | DDoS protection | ✅ None! (handled server-side) |
| CORS Config | Production-ready | ✅ None! (configured in .env) |

**All features work out-of-the-box with EventSource API!** 🎊
