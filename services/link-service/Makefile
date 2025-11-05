# Makefile for link-service
.PHONY: help build run test clean docker-build docker-run deps lint proto

# Variables
APP_NAME := link-service
DOCKER_IMAGE := link-service:latest
DATABASE_URL := postgres://postgres:@localhost/link-service?sslmode=disable

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

# Proto generation
proto-install: ## Install protoc dependencies
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto-gen: ## Generate gRPC code from proto files
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/link.proto

proto-clean: ## Clean generated proto files
	rm -f api/proto/link/*.pb.go

proto: proto-clean proto-gen ## Regenerate proto files

# Docker commands
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run application in Docker
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

docker-up: ## Start services with Docker Compose
	docker-compose up -d

docker-down: ## Stop services with Docker Compose
	docker-compose down

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

# Development helpers
dev-db: ## Start development database
	docker run --name link-service-postgres -e POSTGRES_DB=link-service -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:16

dev-db-stop: ## Stop development database
	docker stop link-service-postgres && docker rm link-service-postgres

dev-redis: ## Start development Redis
	docker run --name link-service-redis -p 6379:6379 -d redis:7-alpine

dev-redis-stop: ## Stop development Redis
	docker stop link-service-redis && docker rm link-service-redis

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

# Security
sec-scan: ## Run security scan
	gosec ./...
