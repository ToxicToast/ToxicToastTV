# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Install dependencies
RUN apk add --no-cache git

# Copy shared module first
COPY shared/ ./shared/

# Copy all services for consistency
COPY services/ ./services/

# Build the service
WORKDIR /workspace/services/user-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o /workspace/bin/user-service ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /workspace/bin/user-service .

EXPOSE 8080 9090

CMD ["./user-service"]
