.PHONY: proto proto-gen proto-clean run build test clean

# Variables
PROTO_DIR := api/proto
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

# Proto generation
proto-gen:
	@echo "Generating protobuf files..."
	protoc \
		--go_out=$(PROTO_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) \
		--go-grpc_opt=paths=source_relative \
		--proto_path=$(PROTO_DIR) \
		$(PROTO_FILES)
	@echo "Protobuf generation complete!"

proto-clean:
	@echo "Cleaning generated protobuf files..."
	@rm -f $(PROTO_DIR)/*.pb.go
	@echo "Clean complete!"

proto: proto-clean proto-gen

# Build
build:
	@echo "Building foodfolio-service..."
	@go build -o bin/foodfolio-service cmd/server/main.go
	@echo "Build complete!"

# Run
run:
	@echo "Running foodfolio-service..."
	@go run cmd/server/main.go

# Test
test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Development
dev:
	@echo "Running in development mode..."
	@air

# Dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Clean
clean: proto-clean
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Docker
docker-build:
	@echo "Building Docker image..."
	@docker build -t foodfolio-service:latest .

docker-run:
	@echo "Running Docker container..."
	@docker run -p 8889:8889 -p 9091:9091 foodfolio-service:latest

# Docker Compose
docker-up:
	@echo "Starting all services with Docker Compose..."
	@docker-compose up -d

docker-down:
	@echo "Stopping all services..."
	@docker-compose down

docker-logs:
	@echo "Showing logs..."
	@docker-compose logs -f

docker-restart:
	@echo "Restarting services..."
	@docker-compose restart

docker-clean:
	@echo "Cleaning up Docker resources..."
	@docker-compose down -v
	@docker system prune -f

# Development with Docker
docker-dev:
	@echo "Starting dependencies only (PostgreSQL + Kafka)..."
	@docker-compose up -d postgres redpanda redpanda-console

# Database
db-migrate:
	@echo "Running database migrations..."
	@go run cmd/server/main.go

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Proto:"
	@echo "  proto-gen         - Generate protobuf files"
	@echo "  proto-clean       - Clean generated protobuf files"
	@echo "  proto             - Clean and regenerate protobuf files"
	@echo ""
	@echo "Build & Run:"
	@echo "  build             - Build the service binary"
	@echo "  run               - Run the service locally"
	@echo "  dev               - Run in development mode (requires air)"
	@echo ""
	@echo "Testing:"
	@echo "  test              - Run tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps              - Install and tidy dependencies"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-run        - Run Docker container"
	@echo "  docker-up         - Start all services (including dependencies)"
	@echo "  docker-down       - Stop all services"
	@echo "  docker-logs       - Show logs"
	@echo "  docker-restart    - Restart all services"
	@echo "  docker-clean      - Clean up Docker resources"
	@echo "  docker-dev        - Start dependencies only (DB + Kafka)"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate        - Run database migrations"
	@echo ""
	@echo "Utilities:"
	@echo "  clean             - Clean all build artifacts"
	@echo "  help              - Show this help message"
