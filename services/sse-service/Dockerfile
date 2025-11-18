# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy shared module first
COPY shared/ ./shared/

# Copy service module
COPY services/sse-service/ ./services/sse-service/

# Set working directory to service
WORKDIR /workspace/services/sse-service

# Download dependencies (go.mod has replace directive pointing to ../../shared)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o /workspace/bin/sse-service ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /workspace/bin/sse-service .

# Expose ports (HTTP for SSE, gRPC for management)
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./sse-service"]
