# GitHub Actions Workflows

## Docker Build and Push

The `docker-build.yml` workflow automatically builds and publishes Docker images for all services in the monorepo.

### Features

- **Smart Change Detection**: Only builds services that have changed
- **Monorepo Support**: Properly handles Go workspace and shared modules
- **Multi-Service**: Supports all 7 services in the monorepo
- **Registry**: Publishes to GitHub Container Registry (ghcr.io)
- **Caching**: Uses GitHub Actions cache for faster builds
- **Tagging Strategy**:
  - `latest` for main branch
  - `<branch>-<sha>` for feature branches
  - Semantic versioning support
  - PR numbers for pull requests

### Services

The workflow builds Docker images for:

1. **blog-service** - Ghost-like blog management
2. **foodfolio-service** - Food inventory management
3. **link-service** - URL shortener
4. **twitchbot-service** - Twitch bot with commands and viewer tracking
5. **notification-service** - Discord notification system
6. **sse-service** - Server-Sent Events streaming
7. **webhook-service** - Webhook management and delivery

### Triggers

- **Push to main**: Builds and pushes images with `latest` tag
- **Pull Requests**: Builds images for validation (no push)
- **Manual**: Can be triggered manually via workflow_dispatch

### Usage

#### Pull Images

```bash
# Pull latest version of a service
docker pull ghcr.io/<username>/blog-service:latest

# Pull specific version
docker pull ghcr.io/<username>/blog-service:main-abc1234
```

#### Use in Docker Compose

```yaml
services:
  blog-service:
    image: ghcr.io/<username>/blog-service:latest
    ports:
      - "8080:8080"
      - "9090:9090"
```

### Local Testing

To test Docker builds locally:

```bash
# Build a single service
docker build -f services/blog-service/Dockerfile -t blog-service:local .

# Build all services
for service in blog foodfolio link twitchbot notification sse webhook; do
  docker build -f services/${service}-service/Dockerfile -t ${service}-service:local .
done
```

### Dockerfile Structure

All services use multi-stage builds:

1. **Builder Stage**: Compiles Go binaries with shared modules
2. **Runtime Stage**: Minimal Alpine image with only the binary

Each Dockerfile:
- Uses Go workspace for shared modules
- Produces static binaries (CGO_ENABLED=0)
- Includes health checks
- Exposes standard ports (8080 for HTTP, 9090 for gRPC)

### Customization

To add a new service:

1. Create `services/<service-name>/Dockerfile`
2. Add to `detect-changes` outputs in workflow
3. Add service to path filters
4. Create new job following existing pattern

### Permissions

The workflow requires:
- `contents: read` - Read repository content
- `packages: write` - Push to GitHub Container Registry

These are automatically provided by `GITHUB_TOKEN`.
