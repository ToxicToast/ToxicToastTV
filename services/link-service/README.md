# Link Shortener Service

A production-ready URL shortener service built with Go, gRPC, and Clean Architecture principles.

## Features

- **URL Shortening**: Generate short codes for long URLs
- **Custom Aliases**: Use custom short codes for your links
- **Link Management**: Full CRUD operations for links
- **Click Analytics**: Track clicks with detailed metadata (IP, User Agent, Referer, Country, City, Device Type)
- **Expiration Support**: Set expiration dates for links
- **Link Statistics**: Get comprehensive analytics for each link
- **Soft Deletes**: Links are soft-deleted and can be recovered
- **PostgreSQL Storage**: Reliable data persistence
- **Redis Caching**: Optional caching for improved performance
- **gRPC API**: High-performance API with Protocol Buffers
- **Health Checks**: Built-in health check endpoints
- **Docker Support**: Easy deployment with Docker and Docker Compose
- **Background Jobs**: Automatic expiration checking and link deactivation
- **Event Publishing**: Kafka/Redpanda events for all link operations

## Architecture

This service follows Clean Architecture principles with clear separation of concerns:

```
link-service/
├── api/proto/              # Protocol Buffer definitions
│   └── link.proto
├── cmd/server/             # Application entry point
│   └── main.go
├── internal/
│   ├── domain/             # Domain entities
│   │   ├── link.go
│   │   └── click.go
│   ├── repository/         # Data access interfaces
│   │   ├── repository.go
│   │   └── impl/
│   │       ├── link_repository_impl.go
│   │       └── click_repository_impl.go
│   ├── usecase/            # Business logic
│   │   ├── link_usecase.go
│   │   └── click_usecase.go
│   ├── scheduler/          # Background jobs
│   │   └── link_expiration.go
│   └── handler/            # gRPC handlers
│       ├── grpc/
│       │   └── link_handler.go
│       └── mapper/
│           ├── common.go
│           └── link.go
├── pkg/
│   └── config/             # Configuration
│       └── config.go
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 16
- Redis 7 (optional)
- Protocol Buffers compiler (protoc)
- Docker and Docker Compose (optional)

### Installation

1. Clone the repository:
```bash
cd services/link-service
```

2. Install dependencies:
```bash
make deps
```

3. Copy environment file:
```bash
cp .env.example .env
```

4. Update `.env` with your configuration:
```bash
# Minimal configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=link-service
BASE_URL=http://localhost:8080
```

5. Start database (using Docker):
```bash
make dev-db
```

6. Run the service:
```bash
make run
```

The service will start on:
- HTTP (Health checks): `http://localhost:8080`
- gRPC: `localhost:9090`

### Using Docker Compose

The easiest way to run the entire stack:

```bash
docker-compose up -d
```

This starts:
- Link Service (HTTP: 8080, gRPC: 9090)
- PostgreSQL (5432)
- Redis (6379)

## API Usage

### Using grpcurl

Install grpcurl:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

#### Create a Link

```bash
grpcurl -plaintext -d '{
  "original_url": "https://example.com/very/long/url/that/needs/shortening",
  "title": "Example Link",
  "description": "A test link for demonstration"
}' localhost:9090 link.LinkService/CreateLink
```

Response:
```json
{
  "link": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "original_url": "https://example.com/very/long/url/that/needs/shortening",
    "short_code": "aB3x9Q",
    "title": "Example Link",
    "description": "A test link for demonstration",
    "is_active": true,
    "click_count": 0,
    "created_at": "2025-11-05T10:00:00Z",
    "updated_at": "2025-11-05T10:00:00Z"
  },
  "short_url": "http://localhost:8080/aB3x9Q"
}
```

#### Create Link with Custom Alias

```bash
grpcurl -plaintext -d '{
  "original_url": "https://example.com",
  "custom_alias": "example"
}' localhost:9090 link.LinkService/CreateLink
```

#### Get Link by Short Code

```bash
grpcurl -plaintext -d '{
  "short_code": "aB3x9Q"
}' localhost:9090 link.LinkService/GetLinkByShortCode
```

#### List Links

```bash
grpcurl -plaintext -d '{
  "page": 1,
  "page_size": 10,
  "is_active": true
}' localhost:9090 link.LinkService/ListLinks
```

#### Update Link

```bash
grpcurl -plaintext -d '{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Updated Title",
  "is_active": false
}' localhost:9090 link.LinkService/UpdateLink
```

#### Delete Link

```bash
grpcurl -plaintext -d '{
  "id": "123e4567-e89b-12d3-a456-426614174000"
}' localhost:9090 link.LinkService/DeleteLink
```

#### Increment Click Counter

```bash
grpcurl -plaintext -d '{
  "short_code": "aB3x9Q"
}' localhost:9090 link.LinkService/IncrementClick
```

#### Record Click with Analytics

```bash
grpcurl -plaintext -d '{
  "link_id": "123e4567-e89b-12d3-a456-426614174000",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "referer": "https://google.com",
  "country": "United States",
  "city": "New York",
  "device_type": "desktop"
}' localhost:9090 link.LinkService/RecordClick
```

#### Get Link Statistics

```bash
grpcurl -plaintext -d '{
  "link_id": "123e4567-e89b-12d3-a456-426614174000"
}' localhost:9090 link.LinkService/GetLinkStats
```

Response:
```json
{
  "link_id": "123e4567-e89b-12d3-a456-426614174000",
  "total_clicks": 150,
  "unique_ips": 87,
  "clicks_today": 12,
  "clicks_this_week": 45,
  "clicks_this_month": 150,
  "clicks_by_country": {
    "United States": 75,
    "United Kingdom": 30,
    "Germany": 25,
    "France": 20
  },
  "clicks_by_device": {
    "desktop": 90,
    "mobile": 50,
    "tablet": 10
  },
  "top_referers": [
    "https://google.com",
    "https://facebook.com",
    "https://twitter.com"
  ]
}
```

#### Get Link Clicks

```bash
grpcurl -plaintext -d '{
  "link_id": "123e4567-e89b-12d3-a456-426614174000",
  "page": 1,
  "page_size": 20
}' localhost:9090 link.LinkService/GetLinkClicks
```

#### Get Clicks by Date Range

```bash
grpcurl -plaintext -d '{
  "link_id": "123e4567-e89b-12d3-a456-426614174000",
  "start_date": "2025-11-01T00:00:00Z",
  "end_date": "2025-11-05T23:59:59Z"
}' localhost:9090 link.LinkService/GetClicksByDate
```

### Health Check Endpoints

```bash
# General health check
curl http://localhost:8080/health

# Readiness check (for Kubernetes)
curl http://localhost:8080/health/ready

# Liveness check (for Kubernetes)
curl http://localhost:8080/health/live
```

## Configuration

All configuration is done via environment variables. See `.env.example` for all available options.

### Key Configuration Options

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `GRPC_PORT` | gRPC server port | `9090` |
| `BASE_URL` | Base URL for short links | `http://localhost:8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | Database name | `link-service` |
| `REDIS_ENABLED` | Enable Redis caching | `false` |
| `AUTH_ENABLED` | Enable authentication | `false` |
| `LINK_EXPIRATION_ENABLED` | Enable link expiration checker | `true` |
| `LINK_EXPIRATION_INTERVAL` | Expiration check interval | `1h` |

### Background Jobs

The service includes a background job scheduler that automatically checks for and deactivates expired links:

- **Link Expiration Scheduler**: Runs periodically (default: every hour) to check for links that have passed their expiration date
- **Automatic Deactivation**: When an expired link is found, it is automatically deactivated and a `link.expired` event is published to Kafka
- **Configurable Interval**: The check interval can be adjusted via `LINK_EXPIRATION_INTERVAL` (supports formats like `30m`, `1h`, `2h`)
- **Graceful Operation**: The scheduler starts automatically with the service and stops gracefully during shutdown

To disable the expiration checker, set `LINK_EXPIRATION_ENABLED=false` in your environment configuration.

## Development

### Generate Proto Files

```bash
make proto
```

### Run Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make test-coverage
```

### Format Code

```bash
make fmt
```

### Run Linter

```bash
make lint
```

### Build Production Binary

```bash
make build-prod
```

## Deployment

### Docker

Build Docker image:
```bash
make docker-build
```

Run with Docker:
```bash
make docker-run
```

### Docker Compose

Start all services:
```bash
make docker-up
```

Stop all services:
```bash
make docker-down
```

View logs:
```bash
make docker-logs
```

## License

This project is part of the ToxicToastGo monorepo.

## Support

For issues and questions, please open an issue in the main repository.
