# Gateway Service

API Gateway für das ToxicToastGo Microservices Ökosystem.

## Features

### Protokoll-Support
- **HTTP REST API** - REST-Endpunkte für Web- und Mobile-Clients
- **gRPC** - Native gRPC-Unterstützung für Service-to-Service-Kommunikation
- **Hybrid Mode** - Beide Protokolle gleichzeitig aktiv

### Middleware & Features
- **Authentication** - Keycloak JWT-Validierung (optional)
- **Rate Limiting** - Token-Bucket pro IP/Client
- **CORS** - Configurable Cross-Origin Resource Sharing
- **Request Logging** - Strukturiertes Logging aller Requests
- **Health Checks** - `/health` und `/ready` Endpunkte

### Routing
Path-based Routing zu Backend-Services:
- `/api/blog/*` → Blog Service
- `/api/links/*` → Link Shortener Service
- `/api/foodfolio/*` → Foodfolio Service
- `/api/notifications/*` → Notification Service
- `/api/events/*` → SSE Service
- `/api/twitch/*` → TwitchBot Service
- `/api/webhooks/*` → Webhook Service

## Architektur

```
Client Request
      ↓
[CORS Middleware]
      ↓
[Rate Limiter]
      ↓
[Logging]
      ↓
[Auth - Optional]
      ↓
[Router] → HTTP-to-gRPC Translation
      ↓
Backend gRPC Services
```

## Configuration

Siehe `.env.example` für alle Konfigurationsoptionen.

### Wichtige Environment Variables

```bash
# Ports
HTTP_PORT=8080
GRPC_PORT=9090

# Rate Limiting
RATE_LIMIT_RPS=100        # Requests per second
RATE_LIMIT_BURST=200      # Burst capacity

# Backend Services (Service Discovery)
BLOG_SERVICE_URL=blog-service:9090
LINK_SERVICE_URL=link-service:9090
# ... weitere Services
```

## Deployment

### Docker

```bash
# Build
docker build -f services/gateway-service/Dockerfile -t gateway-service .

# Run
docker run -p 8080:8080 -p 9090:9090 \
  -e BLOG_SERVICE_URL=blog-service:9090 \
  -e LINK_SERVICE_URL=link-service:9090 \
  gateway-service
```

### Docker Compose

```yaml
gateway-service:
  build:
    context: .
    dockerfile: services/gateway-service/Dockerfile
  ports:
    - "8080:8080"
    - "9090:9090"
  environment:
    - BLOG_SERVICE_URL=blog-service:9090
    - LINK_SERVICE_URL=link-service:9090
    - ENABLE_CORS=true
    - RATE_LIMIT_RPS=100
  depends_on:
    - blog-service
    - link-service
```

## API Endpoints

### Health & Status

```bash
# Health check
GET /health

# Readiness check (shows service connectivity)
GET /ready
```

### Service Proxying

Alle Backend-Services sind über `/api/{service}/` erreichbar:

```bash
# Blog Service
GET /api/blog/posts
POST /api/blog/posts

# Link Service
GET /api/links/{shortCode}
POST /api/links

# Foodfolio Service
GET /api/foodfolio/items
```

## Development

```bash
# Install dependencies
go mod download

# Run locally
go run cmd/server/main.go

# Test
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

## Next Steps

Die aktuellen Proxy-Handler (`handleProxy`) sind Platzhalter. Für vollständige Funktionalität müssen HTTP-Requests in gRPC-Calls übersetzt werden:

1. Import der proto-generierten Clients für jeden Service
2. Request-Body parsing (JSON → Protobuf)
3. gRPC-Call mit Metadata (User-Context aus JWT)
4. Response-Translation (Protobuf → JSON)
5. Error-Handling und Status-Code-Mapping

## Technologie-Stack

- **HTTP Router**: Gorilla Mux
- **gRPC**: google.golang.org/grpc
- **Rate Limiting**: golang.org/x/time/rate
- **Auth**: Keycloak JWT (shared/auth)
- **Logging**: Structured Logging (shared/logger)
