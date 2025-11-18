# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make protobuf-dev

WORKDIR /workspace

# Copy shared module first
COPY shared/ ./shared/

# Copy service module
COPY services/weather-service/ ./services/weather-service/

# Set working directory to service
WORKDIR /workspace/services/weather-service

# Download dependencies (go.mod has replace directive pointing to ../../shared)
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o /workspace/bin/weather-service ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /workspace/bin/weather-service .

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./weather-service"]
