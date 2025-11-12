# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go workspace files
COPY go.work go.work.sum ./

# Copy shared module
COPY shared/ ./shared/

# Copy service files
COPY services/notification-service/ ./services/notification-service/

# Download dependencies
WORKDIR /build/services/notification-service
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /build/bin/notification-service ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /build/bin/notification-service .

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./notification-service"]
