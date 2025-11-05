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
│   ├── blog-service/          # ✅ Blog CMS backend
│   ├── warcraft-service/      # 📋 WoW character & guild tracking
│   ├── foodfolio-service/     # 📋 Food inventory management
│   ├── twitchbot-service/     # 📋 Twitch stream analytics
│   ├── notification-service/  # 📋 Multi-channel notifications
│   ├── sse-service/           # 📋 Real-time events for frontends
│   └── gateway-service/       # 📋 API Gateway (last priority)
├── shared/                     # Shared modules
│   ├── auth/                  # Authentication (Keycloak JWT)
│   ├── kafka/                 # Event producer/consumer
│   ├── database/              # PostgreSQL connection
│   ├── logger/                # Structured logging
│   ├── config/                # Configuration utilities
│   └── [telemetry]/           # 📋 Metrics & tracing (planned)
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

### Warcraft Service
**Status:** 📋 Planned

Integration with Blizzard API for World of Warcraft data tracking.

**Planned Features:**
- Character management and tracking
- Guild information and roster
- Blizzard API integration
- Character progression tracking
- Achievement tracking
- Item and gear tracking

### Foodfolio Service
**Status:** 📋 Planned

A custom food inventory management system inspired by Grocy, focused on food items.

**Planned Features:**
- Food inventory tracking
- Expiration date management
- Purchase location tracking
- Receipt scanning and OCR
- Shopping list generation
- Nutritional information
- Barcode scanning

### Twitchbot Service
**Status:** 📋 Planned

A Twitch bot for tracking and managing stream data (personal channel only).

**Planned Features:**
- Viewer tracking and analytics
- Message logging and analysis
- Stream session tracking
- Clip archival and management
- Chat command system
- Stream notifications
- Custom alerts

### Notification Service
**Status:** 📋 Planned

Centralized notification system for all services.

**Planned Features:**
- Multi-channel notifications (Email, Discord, Telegram)
- Event-driven notifications via Kafka
- Template-based messages
- Notification history
- User preferences
- Rate limiting
- Delivery tracking

### SSE Service
**Status:** 📋 Planned

Server-Sent Events service for real-time frontend updates.

**Planned Features:**
- Real-time event streaming to frontends
- Connection management
- Event filtering per client
- Reconnection handling
- Message queueing
- Multi-tenant support
- WebSocket fallback

### Gateway Service
**Status:** 📋 Planned (Last Priority)

API Gateway for unified access to all microservices.

**Planned Features:**
- Request routing to services
- Authentication/Authorization
- Rate limiting
- Request/Response transformation
- Load balancing
- API versioning
- Monitoring and logging

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

**Telemetry** is integrated directly into each service for:
- Metrics collection (Prometheus-compatible)
- Distributed tracing (OpenTelemetry)
- Structured logging
- Health checks

**Planned Infrastructure:**
- Grafana dashboards for visualization
- Centralized logging (ELK/Loki)
- Alerting (Alertmanager)
- Service health monitoring

## 🗺️ Roadmap

### Service Implementation Priority

**High Priority:**
1. [ ] Warcraft Service - Blizzard API integration
2. [ ] Foodfolio Service - Food inventory management
3. [ ] Twitchbot Service - Stream tracking and analytics
4. [ ] Notification Service - Multi-channel notifications
5. [ ] SSE Service - Real-time frontend updates

**Low Priority:**
6. [ ] Gateway Service - API Gateway (implement last)

### Infrastructure Improvements
- [ ] Telemetry integration in all services (Prometheus, OpenTelemetry)
- [ ] CI/CD pipelines (GitHub Actions)
- [ ] Kubernetes deployment manifests
- [ ] Docker Compose orchestration for local development
- [ ] Infrastructure as Code (Terraform)
- [ ] Centralized logging (ELK/Loki)
- [ ] Grafana dashboards

### Shared Module Enhancements
- [ ] Telemetry module (metrics, tracing)
- [ ] SSE client module
- [ ] Notification client module
- [ ] Common gRPC interceptors
- [ ] Error handling utilities

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
**Planned Services:** 6 (Warcraft, Foodfolio, Twitchbot, Notification, SSE, Gateway)
**Total Ecosystem:** 7 Microservices

Built with ❤️ using Go and gRPC
