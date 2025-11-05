# ToxicToastGo

A modern, scalable blog CMS backend built with Go, gRPC, and Clean Architecture.

## 🚀 Features

### Core Functionality
- **Posts Management** - Full-featured blog posts with Markdown support, SEO metadata, and reading time calculation
- **Categories & Tags** - Hierarchical categories and simple tagging system with slug-based URLs
- **Comments System** - Nested comments with moderation (pending, approved, spam, trash)
- **Media Management** - File upload with streaming, automatic thumbnail generation, and image resizing
- **Authentication** - Optional Keycloak JWT authentication with role-based access control
- **Event Publishing** - Kafka/Redpanda integration for event-driven architecture

### Technical Highlights
- **gRPC API** - High-performance RPC with Protocol Buffers
- **Clean Architecture** - Domain-driven design with clear separation of concerns
- **Go Workspace** - Monorepo structure with shared modules
- **PostgreSQL** - Robust data persistence with GORM
- **Image Processing** - Automatic thumbnail generation (150x150, 300x300, 600x600) and smart resizing
- **Streaming Upload** - Efficient file upload via gRPC streaming

## 📁 Project Structure

```
ToxicToastGo/
├── services/
│   └── blog-service/          # Blog CMS microservice
│       ├── api/proto/          # gRPC Protocol Buffers
│       ├── cmd/server/         # Application entry point
│       ├── internal/
│       │   ├── domain/         # Business entities
│       │   ├── repository/     # Data access layer
│       │   ├── usecase/        # Business logic
│       │   └── handler/grpc/   # gRPC handlers
│       ├── pkg/
│       │   ├── config/         # Configuration
│       │   ├── image/          # Image processing
│       │   ├── storage/        # File storage
│       │   └── utils/          # Utilities (markdown, slug, etc.)
│       └── migrations/         # Database migrations
└── shared/                     # Shared modules
    ├── auth/                   # Keycloak authentication
    ├── kafka/                  # Event producer
    ├── database/               # PostgreSQL connection
    ├── logger/                 # Structured logging
    └── config/                 # Shared configuration

```

## 🛠️ Tech Stack

- **Language:** Go 1.24
- **API:** gRPC with Protocol Buffers
- **Database:** PostgreSQL with GORM ORM
- **Messaging:** Kafka/Redpanda
- **Authentication:** Keycloak (optional)
- **Image Processing:** disintegration/imaging
- **HTTP Router:** Gorilla Mux
- **Markdown:** Blackfriday v2 with Bluemonday sanitization

## 🚦 Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 14+
- Kafka/Redpanda (optional)
- Keycloak (optional)
- Protocol Buffers compiler (for development)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/ToxicToast/ToxicToastTV.git
   cd ToxicToastGo
   ```

2. **Set up environment variables**
   ```bash
   cd services/blog-service
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Run database migrations**
   ```bash
   # Create database first
   createdb blog

   # Migrations run automatically on startup
   ```

5. **Build and run**
   ```bash
   # Development
   go run cmd/server/main.go

   # Production
   make build
   ./bin/blog-service
   ```

## ⚙️ Configuration

Create a `.env` file in `services/blog-service/`:

```bash
# Server
PORT=8080
GRPC_PORT=9090
ENVIRONMENT=development
LOG_LEVEL=info

# Authentication (optional)
AUTH_ENABLED=false

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=blog
DB_PASSWORD=your_password
DB_NAME=blog
DB_SSL_MODE=disable

# Kafka (optional)
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=blog-service
KAFKA_TOPIC_POST_EVENTS=blog.events.post
KAFKA_TOPIC_COMMENT_EVENTS=blog.events.comment
KAFKA_TOPIC_MEDIA_EVENTS=blog.events.media

# Media Storage
MEDIA_STORAGE_PATH=./uploads
MEDIA_MAX_SIZE=10485760                                    # 10MB
MEDIA_ALLOWED_TYPES=image/jpeg,image/png,image/gif,image/webp
MEDIA_GENERATE_THUMBNAILS=true
MEDIA_THUMBNAIL_SIZES=small,medium,large
MEDIA_AUTO_RESIZE_LARGE=true
MEDIA_MAX_IMAGE_WIDTH=3840                                 # 4K
MEDIA_MAX_IMAGE_HEIGHT=2160

# Keycloak (if AUTH_ENABLED=true)
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=your-realm
KEYCLOAK_CLIENT_ID=blog-service
KEYCLOAK_CLIENT_SECRET=your-secret
```

## 🔌 API Endpoints

### Posts
- `CreatePost` - Create new blog post (auth required)
- `GetPost` - Get post by ID or slug (public)
- `UpdatePost` - Update existing post (auth required)
- `DeletePost` - Delete post (auth required)
- `ListPosts` - List posts with filters (public)
- `PublishPost` - Publish draft post (auth required)

### Categories
- `CreateCategory` - Create category (auth required)
- `GetCategory` - Get category by ID or slug (public)
- `UpdateCategory` - Update category (auth required)
- `DeleteCategory` - Delete category (auth required)
- `ListCategories` - List categories (public)

### Tags
- `CreateTag` - Create tag (auth required)
- `GetTag` - Get tag by ID or slug (public)
- `UpdateTag` - Update tag (auth required)
- `DeleteTag` - Delete tag (auth required)
- `ListTags` - List tags with search (public)

### Comments
- `CreateComment` - Create comment (public)
- `GetComment` - Get comment (public)
- `UpdateComment` - Update comment (public)
- `DeleteComment` - Delete comment (auth required)
- `ListComments` - List comments (public)
- `ModerateComment` - Change comment status (auth required)

### Media
- `UploadMedia` - Upload file via streaming (auth required)
- `GetMedia` - Get media metadata (public)
- `DeleteMedia` - Delete media (auth required)
- `ListMedia` - List media files (public)

## 📸 Image Processing

### Automatic Thumbnails
When uploading images, thumbnails are automatically generated:
- **Small:** 150x150 (square, center-cropped)
- **Medium:** 300x300 (square, center-cropped)
- **Large:** 600x600 (square, center-cropped)

### Smart Resizing
Large images can be automatically resized to save storage and bandwidth:
- Max dimensions: 3840x2160 (4K)
- Maintains aspect ratio
- JPEG quality: 85
- Configurable via `MEDIA_AUTO_RESIZE_LARGE`

## 🏗️ Architecture

### Clean Architecture Layers

1. **Domain** - Business entities and rules
2. **Repository** - Data access interfaces and implementations
3. **Use Case** - Application business logic
4. **Handler** - Delivery mechanism (gRPC)

### Design Patterns
- Repository Pattern
- Dependency Injection
- Interface Segregation
- Single Responsibility Principle

## 🔐 Security Features

- JWT authentication with Keycloak
- Role-based access control
- MIME type validation
- File size limits
- SQL injection prevention (GORM)
- XSS prevention (Bluemonday sanitization)
- Optional authentication mode for development

## 🐳 Docker Support

```bash
# Build image
docker build -t toxictoast/blog-service .

# Run with docker-compose
docker-compose up -d
```

## 📝 Development

### Generate Protocol Buffers
```bash
cd services/blog-service
make proto-gen
```

### Run Tests
```bash
go test ./...
```

### Build Binary
```bash
make build
```

## 📊 Database Schema

The service uses PostgreSQL with the following main tables:
- `posts` - Blog posts
- `categories` - Hierarchical categories
- `tags` - Simple tags
- `comments` - Nested comments
- `media` - Uploaded files

Auto-migrations run on startup using GORM.

## 🎯 Roadmap

- [ ] GraphQL API support
- [ ] Full-text search (Elasticsearch)
- [ ] CDN integration
- [ ] MinIO object storage
- [ ] Admin UI (React)
- [ ] Multi-language support
- [ ] Post scheduling
- [ ] Analytics integration

## 📄 License

This project is proprietary software. See [LICENSE](LICENSE) for details.

## 👤 Author

**ToxicToast**

- GitHub: [@ToxicToast](https://github.com/ToxicToast)
- Repository: [ToxicToastTV](https://github.com/ToxicToast/ToxicToastTV)

## 🤝 Contributing

This is a private project and not open for external contributions.

---

Built with ❤️ using Go and gRPC
