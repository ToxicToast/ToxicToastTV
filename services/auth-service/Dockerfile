# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Install dependencies
RUN apk add --no-cache git

# Copy shared module first
COPY shared/ ./shared/

# Copy all services so auth-service can import user-service proto
COPY services/ ./services/

# Build the service
WORKDIR /workspace/services/auth-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o /workspace/bin/auth-service ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /workspace/bin/auth-service .

EXPOSE 8080 9090

CMD ["./auth-service"]
