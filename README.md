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
│   ├── user-service/          # ✅ User management & profiles
│   ├── auth-service/          # ✅ Authentication & authorization (JWT + RBAC)
│   ├── blog-service/          # ✅ Blog CMS backend
│   ├── foodfolio-service/     # ✅ Food inventory management
│   ├── link-service/          # ✅ URL shortener
│   ├── notification-service/  # ✅ Discord webhook notifications
│   ├── sse-service/           # ✅ Real-time events for frontends
│   ├── twitchbot-service/     # 📋 Twitch stream analytics
│   ├── warcraft-service/      # ✅ WoW character & guild tracking
│   ├── webhook-service/       # ✅ Webhook event management
│   └── gateway-service/       # ✅ API Gateway with HTTP-to-gRPC proxy
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

### User Service
**Status:** ✅ Production Ready

Centralized user management service handling user profiles, credentials, and lifecycle events.

**Features:**
- User CRUD operations (Create, Read, Update, Delete)
- User profiles with email, username, and metadata
- Password management with bcrypt hashing
- User status management (Active, Inactive)
- User search and pagination
- Last login tracking
- Kafka event publishing for all user operations
- gRPC API with Protocol Buffers

**Kafka Events:**
- `user.created` - New user registered
- `user.updated` - User profile modified
- `user.deleted` - User account deleted
- `user.activated` - User account activated
- `user.deactivated` - User account deactivated
- `user.password.changed` - User password updated

### Auth Service
**Status:** ✅ Production Ready

Authentication and authorization service providing JWT-based auth with Role-Based Access Control (RBAC).

**Features:**
- User registration and login
- JWT token generation (Access + Refresh tokens)
- Token validation and refresh
- Role-Based Access Control (RBAC)
- Permission management per role
- User role assignment
- Kafka event publishing for auth operations
- Integration with user-service via gRPC

**Kafka Events:**
- `auth.registered` - User completed registration
- `auth.login` - Successful user login
- `auth.token.refreshed` - JWT token refreshed

**RBAC Features:**
- Dynamic role creation and management
- Fine-grained permission system
- User-to-role assignment
- Role-to-permission mapping
- Permission validation in JWT claims

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
**Status:** ✅ Production Ready

Integration with Blizzard Battle.net API for World of Warcraft character and guild tracking.

**Features:**
- Character management (CRUD operations)
- Character details with class, race, faction tracking
- Equipment and stats tracking
- Guild management and roster
- Reference data (races, classes, factions) with normalization
- Blizzard API client stub (OAuth2 ready for future integration)
- gRPC API with Protocol Buffers
- Clean Architecture with separated concerns

**Data Model:**
- Character entity separation (API fields vs. game data)
- Normalized reference data with foreign keys
- One-to-one relationships for character details and equipment

👉 [View Warcraft Service Documentation](./services/warcraft-service/README.md)

### Foodfolio Service
**Status:** ✅ Production Ready

A custom food inventory management system inspired by Grocy, focused on food items.

**Features:**
- Food inventory tracking with categories and types
- Item management with variants and details
- Company and location tracking
- Warehouse and storage management
- Shopping list generation
- Receipt management
- Size management
- gRPC API with Protocol Buffers

### Link Service
**Status:** ✅ Production Ready

URL shortener service for creating and managing short links with click tracking.

**Features:**
- URL shortening with custom slugs
- Click tracking and analytics
- Link expiration management
- gRPC API with Protocol Buffers
- Clean Architecture

### Webhook Service
**Status:** ✅ Production Ready

Webhook event management and delivery system.

**Features:**
- Webhook registration and management
- Event delivery tracking
- Retry logic for failed deliveries
- gRPC API with Protocol Buffers
- Clean Architecture

### Notification Service
**Status:** ✅ Production Ready

Discord webhook notification system for centralized event notifications.

**Features:**
- Discord channel management via webhooks
- Notification history tracking
- Multi-channel support
- Notification delivery tracking
- Channel enable/disable toggle
- Test webhook functionality
- Cleanup of old notifications
- gRPC API with Protocol Buffers

### SSE Service
**Status:** ✅ Production Ready

Server-Sent Events service for real-time frontend updates.

**Features:**
- Real-time event streaming to frontends
- Connection management
- Client tracking and statistics
- Health monitoring
- Message broadcasting
- gRPC management API
- Clean Architecture

### Twitchbot Service
**Status:** ✅ Production Ready

Twitch bot service for tracking and managing stream data with comprehensive analytics.

**Features:**
- Stream session tracking with viewer and message statistics
- 24/7 message logging with full-text search (Chat-Only stream support)
- Multi-channel support - monitor multiple Twitch channels simultaneously
- Viewer tracking per channel with automatic fetching
- Clip archival and management
- Custom command system with permissions and cooldowns
- Bot management via gRPC (join/leave channels, send messages)
- Auto token refresh with reconnect
- Kafka event publishing for all activities
- gRPC API with 7 services and 43 endpoints
- Optional Twitch integration (API-only mode available)

### Gateway Service
**Status:** ✅ Production Ready

API Gateway providing unified HTTP REST API access to all gRPC microservices.

**Features:**
- HTTP-to-gRPC translation for all backend services
- Request routing with path-based routing (/api/{service}/*)
- CORS middleware support
- Rate limiting (configurable RPS and burst)
- Prometheus metrics collection
- Health and readiness endpoints
- Swagger UI (dev mode)
- Logging middleware
- Clean Architecture

**Supported Services:**
- Blog Service - `/api/blog/*`
- Foodfolio Service - `/api/foodfolio/*`
- Link Service - `/api/links/*`
- Notification Service - `/api/notifications/*`
- SSE Service - `/api/events/*`
- TwitchBot Service - `/api/twitch/*`
- Warcraft Service - `/api/warcraft/*`
- Webhook Service - `/api/webhooks/*`

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

### Service Implementation Status

**Completed Services:**
1. [x] Blog Service - Full-featured blog CMS
2. [x] Foodfolio Service - Food inventory management
3. [x] Link Service - URL shortener with click tracking
4. [x] Notification Service - Discord webhook notifications
5. [x] SSE Service - Real-time event streaming
6. [x] Twitchbot Service - Stream tracking and analytics
7. [x] Warcraft Service - WoW character & guild tracking
8. [x] Webhook Service - Webhook event management
9. [x] Gateway Service - API Gateway with HTTP-to-gRPC proxy

**All Services Implemented!** 🎉

### Service Enhancements
- [x] Warcraft: Implement full Blizzard OAuth2 integration
- [x] Warcraft: Add character equipment and stats endpoints
- [x] Warcraft: Add guild roster sync
- [ ] All Services: Add Kafka event publishing
- [ ] All Services: Add integration tests

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

**Implemented Services:** 9 (Blog, Foodfolio, Link, Notification, SSE, Twitchbot, Warcraft, Webhook, Gateway)
**Planned Services:** 0
**Total Ecosystem:** 9 Microservices - All Implemented! 🎉

Built with ❤️ using Go and gRPC
