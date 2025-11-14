# Makefile for gateway-service
.PHONY: help build run test clean docker-build docker-run deps lint fmt vet

# Variables
APP_NAME := gateway-service
DOCKER_IMAGE := gateway-service:latest
HTTP_PORT := 3000

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Install dependencies
	go mod download
	go mod tidy

build: ## Build the application
	go build -o bin/$(APP_NAME) cmd/server/main.go

run: ## Run the application
	go run cmd/server/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker commands
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run application in Docker
	docker run -p $(HTTP_PORT):$(HTTP_PORT) --env-file .env $(DOCKER_IMAGE)

# Development helpers
fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

mod-upgrade: ## Upgrade dependencies
	go get -u ./...
	go mod tidy

# Production helpers
build-prod: ## Build for production
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/$(APP_NAME) cmd/server/main.go
