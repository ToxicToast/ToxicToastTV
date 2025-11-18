# Quick Start - Blog Service

## Voraussetzungen

- Go 1.21+
- PostgreSQL 15+
- Redpanda/Kafka (läuft auf localhost:19092)
- (Optional) Keycloak für Authentication

## 1. Datenbank Setup

### Option A: Docker (Empfohlen)
```bash
docker run --name blog-postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=blog-service \
  -p 5432:5432 \
  -d postgres:15
```

### Option B: Lokale PostgreSQL
```bash
createdb blog-service
```

## 2. Redpanda/Kafka Setup

Wenn du Redpanda bereits in Docker laufen hast, ist dieser Schritt fertig!

Andernfalls:
```bash
docker run -d --name redpanda \
  -p 9092:9092 \
  -p 9644:9644 \
  vectorized/redpanda:latest \
  redpanda start --smp 1 --memory 1G
```

## 3. Environment Configuration

Kopiere die `.env.example` und passe sie an:

```bash
cp .env.example .env
```

Minimale `.env` Konfiguration:
```env
# Server
PORT=8080
GRPC_PORT=9090

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=blog-service

# Kafka (Redpanda)
KAFKA_BROKERS=localhost:19092

# Keycloak (optional - wenn nicht konfiguriert, läuft Service ohne Auth)
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=blog-service
KEYCLOAK_CLIENT_ID=blog-service
```

## 4. Dependencies installieren

```bash
go mod download
```

## 5. Service starten

```bash
go run cmd/server/main.go
```

Du solltest folgende Ausgabe sehen:
```
Starting Blog Service v dev (built: unknown)
Environment: development
Database connected successfully
Migrated entity: *domain.Post
Migrated entity: *domain.Category
Migrated entity: *domain.Tag
Migrated entity: *domain.Media
Migrated entity: *domain.Comment
Database schema is up to date
Kafka producer connected successfully
gRPC server starting on port 9090
HTTP server starting on port 8080
```

## 6. Service testen

### Health Check (HTTP)
```bash
curl http://localhost:8080/health
```

Erwartete Antwort:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-04T...",
  "services": {
    "database": "healthy"
  },
  "version": "dev"
}
```

### gRPC Test mit grpcurl

Installiere grpcurl falls noch nicht vorhanden:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

#### Verfügbare Services auflisten:
```bash
grpcurl -plaintext localhost:9090 list
```

Output:
```
blog.BlogService
grpc.reflection.v1.ServerReflection
```

#### Post erstellen (ohne Auth für lokalen Test):

**Wichtig:** Für lokale Tests ohne Keycloak, kommentiere die Auth-Middleware temporär aus oder füge den Post-Endpoint zur Public-Liste hinzu.

```bash
grpcurl -plaintext -d '{
  "title": "Mein erster Blog Post",
  "content": "# Willkommen\n\nDas ist **markdown** Content!",
  "excerpt": "Ein toller erster Post",
  "featured": true
}' localhost:9090 blog.BlogService/CreatePost
```

#### Posts auflisten (public endpoint):
```bash
grpcurl -plaintext -d '{
  "page": 1,
  "page_size": 10
}' localhost:9090 blog.BlogService/ListPosts
```

#### Post via Slug abrufen (public endpoint):
```bash
grpcurl -plaintext -d '{
  "slug": "mein-erster-blog-post"
}' localhost:9090 blog.BlogService/GetPost
```

## 7. Kafka Events überprüfen

Du kannst Kafka-Events mit rpk (Redpanda CLI) überprüfen:

```bash
# Topics anzeigen
rpk topic list

# Events konsumieren
rpk topic consume blog.events.post.created --brokers localhost:9092
```

Oder mit kcat/kafkacat:
```bash
kcat -b localhost:19092 -t blog.events.post.created -C
```

## Mit Keycloak Authentication

### 1. Keycloak Setup

Erstelle einen Realm `blog-service` mit einem Client `blog-service`.

### 2. Token generieren

```bash
# Hole Access Token
export TOKEN=$(curl -X POST "http://localhost:8080/realms/blog-service/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin" \
  -d "password=admin" \
  -d "grant_type=password" \
  -d "client_id=blog-service" \
  -d "client_secret=YOUR_CLIENT_SECRET" | jq -r '.access_token')
```

### 3. Authenticated Request

```bash
grpcurl -plaintext \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Authenticated Post",
    "content": "This post was created with auth!"
  }' localhost:9090 blog.BlogService/CreatePost
```

## Nächste Schritte

### Posts mit Categories und Tags

```bash
# 1. Erstelle Category
grpcurl -plaintext -H "Authorization: Bearer $TOKEN" -d '{
  "name": "Technology",
  "description": "Tech articles"
}' localhost:9090 blog.BlogService/CreateCategory

# 2. Erstelle Tags
grpcurl -plaintext -H "Authorization: Bearer $TOKEN" -d '{
  "name": "golang"
}' localhost:9090 blog.BlogService/CreateTag

grpcurl -plaintext -H "Authorization: Bearer $TOKEN" -d '{
  "name": "grpc"
}' localhost:9090 blog.BlogService/CreateTag

# 3. Erstelle Post mit Category und Tags
grpcurl -plaintext -H "Authorization: Bearer $TOKEN" -d '{
  "title": "Building Microservices with Go",
  "content": "# Introduction\n\nLet me show you...",
  "category_ids": ["<category-uuid>"],
  "tag_ids": ["<tag1-uuid>", "<tag2-uuid>"],
  "seo": {
    "meta_title": "Building Microservices with Go - Complete Guide",
    "meta_description": "Learn how to build scalable microservices",
    "og_title": "Building Microservices with Go",
    "og_description": "Complete tutorial"
  }
}' localhost:9090 blog.BlogService/CreatePost
```

### Post Publishen

```bash
grpcurl -plaintext -H "Authorization: Bearer $TOKEN" -d '{
  "id": "<post-uuid>"
}' localhost:9090 blog.BlogService/PublishPost
```

## Troubleshooting

### Database Connection Failed
- Überprüfe ob PostgreSQL läuft: `psql -h localhost -U postgres -d blog-service`
- Überprüfe `.env` Credentials

### Kafka Connection Failed
- Überprüfe ob Redpanda läuft: `docker ps | grep redpanda`
- Service läuft auch ohne Kafka weiter (ohne Event-Publishing)

### Keycloak Auth Failed
- Service läuft auch ohne Keycloak (mit Warning-Logs)
- Überprüfe Keycloak URL und Realm in `.env`

### gRPC Fehler "method not found"
- Stelle sicher dass Proto-Files generiert sind: `protoc --go_out=. --go-grpc_out=. api/proto/blog.proto`
- Rebuild: `go build cmd/server/main.go`

## Development Workflow

```bash
# 1. Änderungen an Proto-Files
vim api/proto/blog.proto

# 2. Proto regenerieren
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/proto/blog.proto

# 3. Code ändern
vim internal/usecase/post_usecase.go

# 4. Testen
go test ./...

# 5. Starten
go run cmd/server/main.go
```

## Production Build

```bash
# Static binary erstellen
CGO_ENABLED=0 GOOS=linux go build \
  -a -installsuffix cgo \
  -ldflags '-extldflags "-static" -X main.Version=1.0.0 -X main.BuildTime='$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -o bin/blog-service \
  cmd/server/main.go

# Docker Build
docker build -t blog-service:latest .
```

## Weitere Implementierungen

Das ist nur die **Post-Management** Implementation. Für die anderen Features:

- **Categories & Tags**: Analog zu Posts implementiert (Use Cases fehlen noch)
- **Media Management**: Upload-Handler für File-Streaming implementieren
- **Comments**: Use Case + Handler implementieren
- **Search**: Full-text Search mit PostgreSQL oder Elasticsearch

Alle Repositories und Domain Entities sind bereits vorhanden!
