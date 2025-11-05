.PHONY: proto proto-gen proto-clean run build test clean

# Variables
PROTO_DIR := api/proto
PROTO_OUT_DIR := api/proto/foodfolio/v1
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

# Proto generation
proto-gen:
	@echo "Generating protobuf files..."
	@mkdir -p $(PROTO_OUT_DIR)
	protoc \
		--go_out=$(PROTO_OUT_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT_DIR) \
		--go-grpc_opt=paths=source_relative \
		--proto_path=$(PROTO_DIR) \
		$(PROTO_FILES)
	@echo "Protobuf generation complete!"

proto-clean:
	@echo "Cleaning generated protobuf files..."
	@rm -rf $(PROTO_OUT_DIR)
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
	@docker run -p 8081:8081 -p 9091:9091 foodfolio-service:latest

# Database
db-migrate:
	@echo "Running database migrations..."
	@go run cmd/server/main.go

# Help
help:
	@echo "Available targets:"
	@echo "  proto-gen         - Generate protobuf files"
	@echo "  proto-clean       - Clean generated protobuf files"
	@echo "  proto             - Clean and regenerate protobuf files"
	@echo "  build             - Build the service binary"
	@echo "  run               - Run the service"
	@echo "  test              - Run tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  dev               - Run in development mode (requires air)"
	@echo "  deps              - Install and tidy dependencies"
	@echo "  clean             - Clean all build artifacts"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-run        - Run Docker container"
	@echo "  db-migrate        - Run database migrations"
	@echo "  help              - Show this help message"
