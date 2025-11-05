# ToxicToastGo

A collection of microservices built with Go, gRPC, and Clean Architecture for the ToxicToast ecosystem.

## 🏗️ Architecture

This is a **Go monorepo** using Go workspaces, containing multiple microservices that communicate via gRPC and Kafka/Redpanda.

### Design Principles
- **Microservices Architecture** - Independent, scalable services
- **Clean Architecture** - Domain-driven design with clear separation of concerns
- **Event-Driven** - Kafka/Redpanda for asynchronous communication
- **API-First** - gRPC with Protocol Buffers for high-performance RPC
- **Shared Modules** - Common functionality extracted into reusable packages

## 📁 Project Structure

```
ToxicToastGo/
├── services/                   # Microservices
│   ├── blog-service/          # Blog CMS backend
│   ├── [future-service]/      # Additional services coming soon
│   └── ...
├── shared/                     # Shared modules
│   ├── auth/                  # Authentication (Keycloak JWT)
│   ├── kafka/                 # Event producer/consumer
│   ├── database/              # PostgreSQL connection
│   ├── logger/                # Structured logging
│   └── config/                # Configuration utilities
├── go.work                     # Go workspace configuration
└── LICENSE                     # Proprietary license
```

## 🚀 Services

### Blog Service
**Status:** ✅ Production Ready

A full-featured blog CMS backend with support for posts, categories, tags, comments, and media management.

**Features:**
- Markdown posts with SEO metadata
- Hierarchical categories & tags
- Nested comments with moderation
- Media upload with automatic thumbnails
- gRPC API with streaming support
- Optional Keycloak authentication

👉 [View Blog Service Documentation](./services/blog-service/README.md)

### [Future Services]
Additional microservices will be added here as the ecosystem grows.

## 🛠️ Tech Stack

### Core Technologies
- **Language:** Go 1.24+
- **API:** gRPC with Protocol Buffers
- **Database:** PostgreSQL with GORM
- **Messaging:** Kafka/Redpanda
- **Authentication:** Keycloak (JWT)

### Shared Infrastructure
- **Monorepo:** Go Workspaces
- **Containerization:** Docker & Docker Compose
- **CI/CD:** GitHub Actions (planned)
- **Observability:** Structured logging (planned: Prometheus, Grafana)

## 🏃 Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 14+
- Kafka/Redpanda (optional)
- Docker & Docker Compose (optional)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/ToxicToast/ToxicToastTV.git
   cd ToxicToastGo
   ```

2. **Install dependencies**
   ```bash
   # Download all workspace dependencies
   go work sync
   ```

3. **Set up services**
   ```bash
   # Navigate to a service directory
   cd services/blog-service

   # Copy and configure environment
   cp .env.example .env

   # Run the service
   go run cmd/server/main.go
   ```

## 📦 Shared Modules

### Authentication (`shared/auth`)
Keycloak JWT authentication with gRPC interceptors.
- Token validation via JWKS
- Role-based access control
- User context extraction

### Kafka (`shared/kafka`)
Event producer/consumer for asynchronous messaging.
- Type-safe event definitions
- Automatic serialization
- Error handling and retries

### Database (`shared/database`)
PostgreSQL connection management.
- Connection pooling
- Automatic retry logic
- GORM integration
- Migration support

### Logger (`shared/logger`)
Structured logging utilities.
- JSON formatting
- Log levels
- Context propagation

### Config (`shared/config`)
Environment-based configuration.
- .env file support
- Type-safe config structs
- Validation helpers

## 🔐 Security

- **Authentication:** JWT-based auth with Keycloak
- **Authorization:** Role-based access control (RBAC)
- **Data Validation:** Input sanitization and validation
- **Secure Defaults:** All services start with security best practices

## 🐳 Docker Support

Each service includes Docker support:

```bash
# Build and run a service
cd services/blog-service
docker-compose up -d
```

Or run the entire stack:

```bash
# From root directory
docker-compose up -d
```

## 🏗️ Development

### Go Workspace Commands

```bash
# Sync all workspace dependencies
go work sync

# Build all services
go build ./...

# Test all services
go test ./...

# Update all dependencies
go get -u ./...
```

### Adding a New Service

1. Create service directory in `services/`
2. Initialize Go module
3. Add to `go.work`:
   ```bash
   go work use ./services/your-service
   ```
4. Import shared modules as needed

## 📊 Monitoring & Observability

**Planned:**
- Prometheus metrics
- Grafana dashboards
- Distributed tracing (OpenTelemetry)
- Centralized logging (ELK/Loki)

## 🗺️ Roadmap

### Planned Services
- [ ] User Service - User management and authentication
- [ ] Notification Service - Email, SMS, push notifications
- [ ] Analytics Service - Usage analytics and reporting
- [ ] Search Service - Full-text search with Elasticsearch
- [ ] Gateway Service - API Gateway with rate limiting

### Infrastructure
- [ ] Service mesh (Istio/Linkerd)
- [ ] CI/CD pipelines
- [ ] Kubernetes deployment
- [ ] Infrastructure as Code (Terraform)

## 📄 License

This project is proprietary software. See [LICENSE](LICENSE) for details.

**IMPORTANT:** This software is for PRIVATE USE ONLY by the author (ToxicToast). Any unauthorized use, reproduction, or distribution is strictly prohibited.

## 👤 Author

**ToxicToast**

- GitHub: [@ToxicToast](https://github.com/ToxicToast)
- Repository: [ToxicToastTV](https://github.com/ToxicToast/ToxicToastTV)

## 🤝 Contributing

This is a private project and not open for external contributions.

---

**Current Services:** 1 (Blog Service)
**In Development:** 0
**Planned:** 5+

Built with ❤️ using Go and gRPC
