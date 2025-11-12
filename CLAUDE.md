# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ToxicToastGo** is a microservices monorepo built in Go, following Clean Architecture principles. The project uses a shared module approach for common functionality across services.

**Key Technologies:**
- **gRPC** for inter-service and client communication
- **Kafka/Redpanda** for event-driven architecture
- **PostgreSQL** with GORM for data persistence
- **Keycloak** for centralized JWT authentication
- **Go Workspaces** for monorepo management

## Monorepo Structure

\`\`\`
ToxicToastGo/
├── go.work                      # Go workspace configuration
├── CLAUDE.md                    # This file
│
├── shared/                      # Shared packages for all services
│   ├── go.mod
│   ├── auth/                    # Keycloak JWT authentication
│   ├── kafka/                   # Kafka producer/consumer
│   ├── database/                # PostgreSQL connection helpers
│   ├── logger/                  # Centralized logging
│   └── config/                  # Common configuration structs
│
└── services/
    ├── blog-service/            # Ghost-like blog management
    │   ├── go.mod
    │   ├── api/proto/           # gRPC proto definitions
    │   ├── cmd/server/          # Service entry point
    │   ├── internal/            # Service-specific business logic
    │   ├── pkg/                 # Service utilities & config
    │   └── QUICKSTART.md
    └── [other-services]/        # Future services
\`\`\`

## Shared Module (\`shared/\`)

Common functionality for all services.

### Import Pattern
\`\`\`go
import "github.com/toxictoast/toxictoastgo/shared/auth"
import "github.com/toxictoast/toxictoastgo/shared/kafka"
import "github.com/toxictoast/toxictoastgo/shared/database"
import "github.com/toxictoast/toxictoastgo/shared/config"
\`\`\`

See full API documentation in each shared package.

## Development Workflow

### Working in Monorepo
\`\`\`bash
# Edit shared code
cd shared/auth
# Make changes
cd ../..
go work sync

# Changes immediately available to all services
cd services/blog-service
go build ./...
\`\`\`

### Adding New Service
\`\`\`bash
mkdir -p services/my-service
cd services/my-service
go mod init toxictoast/services/my-service

# Link to shared
go mod edit -require=github.com/toxictoast/toxictoastgo/shared@v0.0.0
go mod edit -replace=github.com/toxictoast/toxictoastgo/shared=../../shared

# Add to workspace
cd ../..
go work use ./services/my-service
\`\`\`

## Service Pattern

All services follow Clean Architecture:

1. **Domain** (\`internal/domain/\`) - Pure entities
2. **Repository** (\`internal/repository/\`) - Data access
3. **Use Case** (\`internal/usecase/\`) - Business logic
4. **Handler** (\`internal/handler/grpc/\`) - gRPC endpoints
5. **Config** (\`pkg/config/\`) - Service-specific config (extends shared)
6. **Main** (\`cmd/server/main.go\`) - Bootstrap & wiring

## Common Commands

\`\`\`bash
# Build all
go build ./...

# Test all
go test ./...

# Proto generation (per service)
cd services/blog-service
protoc --go_out=. --go-grpc_out=. api/proto/*.proto

# Run service
go run services/blog-service/cmd/server/main.go
\`\`\`

## Environment Configuration

All services use these common variables:

\`\`\`env
# Servers
PORT=8080
GRPC_PORT=9090

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=service_db

# Kafka
KAFKA_BROKERS=localhost:19092

# Keycloak
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=my-realm
KEYCLOAK_CLIENT_ID=my-client
\`\`\`

## Service-Specific Documentation

- **Blog Service**: See \`services/blog-service/QUICKSTART.md\`
- Future services will have their own QUICKSTART.md

## Troubleshooting

**"cannot find package"**
\`\`\`bash
go work sync
go mod tidy
\`\`\`

**gRPC/Proto issues**
\`\`\`bash
# Regenerate proto files
protoc --go_out=. --go-grpc_out=. api/proto/*.proto
go mod tidy
\`\`\`

## Architecture Principles

- **Dependency Direction**: Always inward (Handler → Use Case → Domain)
- **Shared Code**: Common infra only (auth, kafka, db, logger)
- **Service Code**: Business logic stays in service
- **gRPC**: Synchronous client-service communication
- **Kafka**: Asynchronous event-driven communication

## Go Version

Requires **Go 1.21+**
