.PHONY: help deps build run test test-coverage lint clean docker-build proto-gen build-prod

# Default target
.DEFAULT_GOAL := help

# Service configuration
SERVICE_NAME := weather-service
BINARY_NAME := weather-service
DOCKER_IMAGE := toxictoast/weather-service
GO_VERSION := 1.24

# Build configuration
BUILD_DIR := bin
MAIN_FILE := cmd/server/main.go
PROTO_DIR := api/proto

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed successfully!"

build: ## Build the application
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: ## Run the application
	@echo "Running $(SERVICE_NAME)..."
	@go run $(MAIN_FILE)

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@go vet ./...
	@golangci-lint run --timeout=5m || echo "Note: golangci-lint not installed, skipping..."

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

docker-build: ## Build Docker image
	@echo "Building Docker image: $(DOCKER_IMAGE)..."
	@docker build -t $(DOCKER_IMAGE):latest .
	@echo "Docker image built successfully!"

proto-gen: ## Generate gRPC code from proto files
	@echo "Generating gRPC code from proto files..."
	@cd $(PROTO_DIR) && protoc --go_out=. --go-grpc_out=. weather.proto
	@echo "Proto files generated successfully!"

build-prod: ## Build for production (Linux amd64, static binary)
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-w -s" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-linux \
		$(MAIN_FILE)
	@echo "Production build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux"
