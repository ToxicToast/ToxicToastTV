# Docker Compose Setup

Complete Docker Compose configuration for running the entire ToxicToastGo monorepo locally.

## Quick Start

```bash
# Copy environment template
cp .env.docker .env

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v
```

## Architecture

### Infrastructure Services

| Service | Port | Description |
|---------|------|-------------|
| **postgres** | 5432 | PostgreSQL 16 database |
| **redpanda** | 19092 | Kafka-compatible message broker |
| **keycloak** | 8080 | Authentication and identity management |

### Application Services

| Service | HTTP Port | gRPC Port | Database | Description |
|---------|-----------|-----------|----------|-------------|
| **blog-service** | 8082 | 9092 | blog_service | Blog CMS with posts, categories, tags |
| **foodfolio-service** | 8081 | 9091 | foodfolio_service | Food inventory and receipt scanning |
| **link-service** | 8083 | 9093 | link_service | URL shortener |
| **notification-service** | 8084 | 9096 | notification_service | Discord notifications |
| **sse-service** | 8085 | 9094 | - | Server-Sent Events streaming |
| **webhook-service** | 8086 | 9095 | webhook_service | Webhook delivery |
| **twitchbot-service** | 8087 | 9097 | twitchbot_service | Twitch chat bot |
| **warcraft-service** | 8088 | 9098 | - | Blizzard Battle.net API |
| **gateway-service** | 3000 | 9099 | - | API Gateway and routing |

## Service Configuration

### Database Initialization

A single shared database `toxictoast` is created via `scripts/init-databases.sql`.
All services use the same database with table prefixes for separation:
- Blog tables: `posts`, `categories`, `tags`, `comments`, `media`
- FoodFolio tables: `items`, `categories`, `warehouses`, `receipts`, etc.
- Link tables: `links`, `clicks`
- Notification tables: `channels`, `notifications`, `notification_attempts`
- Webhook tables: `subscriptions`, `deliveries`
- Twitchbot tables: `streams`, `messages`, `commands`

GORM handles automatic table creation and migrations on service startup.
Keycloak uses a separate `keycloak` database.

### Kafka Topics

Services automatically create their required Kafka topics. See `KAFKA_TOPICS.md` for full list.

### Authentication

Keycloak is optional for development. Set `AUTH_ENABLED=false` in service configs to disable.

**Keycloak Admin Console**: http://localhost:8080
- Username: `admin`
- Password: `admin`

## Development Workflow

### Build and Run All Services

```bash
# Build all Docker images
docker-compose build

# Start infrastructure only
docker-compose up -d postgres redpanda keycloak

# Start specific service
docker-compose up -d blog-service

# Rebuild and restart service
docker-compose up -d --build blog-service
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f blog-service

# Infrastructure only
docker-compose logs -f postgres redpanda keycloak
```

### Health Checks

All services include health checks. Check status:

```bash
# List running containers with health status
docker-compose ps

# Check specific service health
docker inspect --format='{{.State.Health.Status}}' blog-service
```

### Accessing Services

**API Gateway (main entry point)**:
- HTTP: http://localhost:3000
- Swagger UI: http://localhost:3000/swagger (when DEV_MODE=true)

**Direct Service Access**:
- Blog Service: http://localhost:8082
- FoodFolio Service: http://localhost:8081
- Link Service: http://localhost:8083
- Notification Service: http://localhost:8084
- SSE Service: http://localhost:8085
- Webhook Service: http://localhost:8086
- Twitchbot Service: http://localhost:8087
- Warcraft Service: http://localhost:8088

**Infrastructure**:
- PostgreSQL: localhost:5432
- Redpanda Kafka: localhost:19092
- Keycloak: http://localhost:8080

### Redpanda Console

```bash
# Access Redpanda admin UI
docker-compose exec redpanda rpk cluster info

# List topics
docker-compose exec redpanda rpk topic list

# Consume topic messages
docker-compose exec redpanda rpk topic consume blog.events.post -f '%v\n'
```

## Environment Variables

All services support environment variable configuration. Key variables:

```env
# Server
PORT=8080
GRPC_PORT=9090
ENVIRONMENT=docker
LOG_LEVEL=info

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=service_db

# Kafka
KAFKA_BROKERS=redpanda:29092
KAFKA_GROUP_ID=service-name

# Authentication (optional)
AUTH_ENABLED=false
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=your-realm
```

See individual service `.env.example` files for service-specific configuration.

## Networking

All services communicate via `toxictoast-network` bridge network:

- **Internal URLs**: Use service names (e.g., `postgres:5432`, `redpanda:29092`)
- **External URLs**: Use localhost with mapped ports

Gateway routes requests to backend services using internal gRPC URLs.

## Volumes

Persistent data is stored in Docker volumes:

- `postgres_data` - PostgreSQL databases
- `redpanda_data` - Kafka topics and messages
- `blog_uploads` - Blog media uploads
- `foodfolio_receipts` - Receipt images

**Clean all data**:
```bash
docker-compose down -v
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose logs service-name

# Check dependencies
docker-compose ps

# Restart service
docker-compose restart service-name
```

### Database Connection Errors

```bash
# Verify PostgreSQL is healthy
docker-compose ps postgres

# Check databases exist
docker-compose exec postgres psql -U postgres -l

# Connect to shared database
docker-compose exec postgres psql -U postgres -d toxictoast

# List all tables (shows all services' tables)
docker-compose exec postgres psql -U postgres -d toxictoast -c "\dt"
```

### Kafka Connection Errors

```bash
# Check Redpanda health
docker-compose ps redpanda

# Verify Kafka is accepting connections
docker-compose exec redpanda rpk cluster health
```

### Port Conflicts

If ports are already in use:

```bash
# Check what's using a port
netstat -ano | findstr :8080  # Windows
lsof -i :8080                  # Linux/Mac

# Update docker-compose.yml to use different ports
ports:
  - "8082:8080"  # Map external 8082 to internal 8080
```

### Rebuilding Services

```bash
# Rebuild all services
docker-compose build --no-cache

# Rebuild specific service
docker-compose build --no-cache blog-service

# Remove all images and rebuild
docker-compose down --rmi all
docker-compose build
```

## Production Considerations

This docker-compose.yml is for **local development only**. For production:

1. **Use secrets management** - Don't hardcode credentials
2. **Configure resource limits** - Add memory/CPU constraints
3. **Enable TLS** - Use HTTPS and secure gRPC
4. **Scale services** - Use Kubernetes or Docker Swarm
5. **External databases** - Use managed PostgreSQL
6. **External Kafka** - Use managed Kafka/Redpanda
7. **Monitoring** - Add Prometheus, Grafana, Jaeger
8. **Backup strategy** - Regular database backups
9. **Load balancing** - Use proper load balancers
10. **Security hardening** - Follow security best practices

## Useful Commands

```bash
# Start only infrastructure
docker-compose up -d postgres redpanda

# Start without Keycloak (optional auth)
docker-compose up -d postgres redpanda blog-service foodfolio-service

# View resource usage
docker stats

# Clean up stopped containers
docker-compose rm -f

# View service configuration
docker-compose config

# Execute command in service
docker-compose exec blog-service sh

# Follow logs from specific time
docker-compose logs --since 30m -f blog-service
```

## Related Documentation

- **Background Jobs**: See `BACKGROUND_JOBS.md`
- **Kafka Topics**: See `KAFKA_TOPICS.md`
- **Keycloak Setup**: See `KEYCLOAK_SETUP.md`
- **Service READMEs**: Check `services/*/README.md`
