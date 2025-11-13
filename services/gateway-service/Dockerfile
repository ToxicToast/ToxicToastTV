# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy shared module first
COPY shared/ ./shared/

# Copy all service modules that gateway depends on (for replace directives)
COPY services/blog-service/ ./services/blog-service/
COPY services/foodfolio-service/ ./services/foodfolio-service/
COPY services/link-service/ ./services/link-service/
COPY services/notification-service/ ./services/notification-service/
COPY services/sse-service/ ./services/sse-service/
COPY services/twitchbot-service/ ./services/twitchbot-service/
COPY services/warcraft-service/ ./services/warcraft-service/
COPY services/webhook-service/ ./services/webhook-service/

# Copy gateway service
COPY services/gateway-service/ ./services/gateway-service/

# Set working directory to service
WORKDIR /workspace/services/gateway-service

# Download dependencies (go.mod has replace directives pointing to other services)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o /workspace/bin/gateway-service ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /workspace/bin/gateway-service .

# Expose ports (HTTP and gRPC)
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./gateway-service"]
