# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go workspace files
COPY go.work go.work.sum ./

# Copy shared module
COPY shared/ ./shared/

# Copy service files
COPY services/twitchbot-service/ ./services/twitchbot-service/

# Download dependencies
WORKDIR /build/services/twitchbot-service
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/bin/twitchbot-service ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/twitchbot-service .

# Expose ports
EXPOSE 8083 9093

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

# Run the application
CMD ["./twitchbot-service"]
