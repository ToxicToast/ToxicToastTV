# SSE Service - Quick Start Guide

## 🚀 Quick Start (5 minutes)

### 1. Prerequisites

Ensure you have:
- Kafka/Redpanda running on `localhost:9092`
- Other services (blog, twitchbot, link) publishing events

### 2. Configure

```bash
cd services/sse-service
cp .env.example .env
```

Edit `.env` if needed:
```env
KAFKA_BROKERS=localhost:19092
# Important: Use explicit topic names, NOT wildcards like blog.*
KAFKA_TOPICS=blog.posts,blog.comments,blog.categories,blog.tags,twitchbot.streams,twitchbot.messages,link.links
PORT=8084
GRPC_PORT=9094
```

⚠️ **Important**: Kafka subscriptions require explicit topic names. Wildcards (`blog.*`) only work for SSE event filtering, not Kafka topics!

### 3. Run

```bash
go run cmd/server/main.go
```

Output:
```
Starting SSE Service v0.1.0
🚀 SSE Broker started
🎧 Kafka consumer started
   Brokers: [localhost:19092]
   Topics: [blog.* twitchbot.* link.*]
🌐 HTTP server starting on port 8084
   SSE Endpoint: http://localhost:8084/events
🔧 gRPC server starting on port 9094
```

### 4. Connect & Test

**Terminal 1 - Subscribe to all events:**
```bash
curl -N http://localhost:8084/events
```

**Terminal 2 - Subscribe to specific events:**
```bash
curl -N "http://localhost:8084/events?event_types=blog.*,twitchbot.message.*"
```

**Browser - Open DevTools Console:**
```javascript
const es = new EventSource('http://localhost:8084/events?event_types=blog.*');
es.onmessage = (e) => console.log('Event:', JSON.parse(e.data));
```

### 5. Test with gRPC

**Check stats:**
```bash
grpcurl -plaintext localhost:9094 sse.SSEManagementService/GetStats
```

**List connected clients:**
```bash
grpcurl -plaintext -d '{"limit": 10}' \
  localhost:9094 sse.SSEManagementService/GetClients
```

## 📊 Monitoring

**Health check:**
```bash
curl http://localhost:8084/health
```

**Statistics:**
```bash
curl http://localhost:8084/stats
```

## 🔧 Common Filters

```bash
# All blog events
curl -N "http://localhost:8084/events?event_types=blog.*"

# Only post creations
curl -N "http://localhost:8084/events?event_types=blog.post.created"

# All twitchbot events
curl -N "http://localhost:8084/events?event_types=twitchbot.*"

# Multiple specific events
curl -N "http://localhost:8084/events?event_types=blog.post.created,twitchbot.message.created"

# Filter by source
curl -N "http://localhost:8084/events?sources=blog-service"
```

## 🐛 Troubleshooting

**"Invalid topic" error?**
```bash
# List available Kafka topics
kafka-topics --list --bootstrap-server localhost:19092

# Copy exact topic names to .env
# KAFKA_TOPICS=blog.posts,twitchbot.messages,...
```

See [KAFKA_TOPICS_FIX.md](./KAFKA_TOPICS_FIX.md) for detailed fix.

**No events appearing?**
1. Check Kafka is running: `docker ps | grep kafka`
2. Check topics exist: `kafka-topics --list --bootstrap-server localhost:9092`
3. Verify other services are publishing events
4. Check logs for Kafka consumer errors

**Connection drops?**
- Default heartbeat is 30 seconds
- Increase `SSE_HEARTBEAT_SECONDS` in `.env`
- Implement reconnection logic in client

## 📖 Full Documentation

See [README.md](./README.md) for complete documentation.

## 🎯 Next Steps

1. **Integrate with Frontend**: Use EventSource API
2. **Add Authentication**: Enable `AUTH_ENABLED=true`
3. **Scale**: Run multiple instances with same `KAFKA_GROUP_ID`
4. **Monitor**: Use `/stats` endpoint with Prometheus/Grafana
