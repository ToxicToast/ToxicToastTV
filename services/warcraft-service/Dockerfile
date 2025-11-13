# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy shared module
COPY shared/ ./shared/

# Copy service
COPY services/warcraft-service/ ./services/warcraft-service/

# Set working directory to service
WORKDIR /workspace/services/warcraft-service

# Download dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o /workspace/bin/warcraft-service ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /workspace/bin/warcraft-service .

EXPOSE 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 9090 || exit 1

CMD ["./warcraft-service"]
