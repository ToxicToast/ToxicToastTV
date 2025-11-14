# Root Makefile for ToxicToastGo Monorepo
.PHONY: help build test clean docker-up docker-down docker-build docker-clean proto-gen proto-clean fmt vet lint deps

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Workspace commands
deps: ## Install dependencies for all services
	go work sync
	go mod download
	@echo "Installing dependencies for all services..."
	@for dir in services/*/; do \
		if [ -f $$dir/go.mod ]; then \
			echo "==> $$dir"; \
			cd $$dir && go mod download && go mod tidy && cd ../..; \
		fi \
	done

build: ## Build all services
	@echo "Building all services..."
	@for dir in services/*/; do \
		if [ -f $$dir/Makefile ]; then \
			echo "==> Building $$dir"; \
			$(MAKE) -C $$dir build; \
		fi \
	done

test: ## Run tests for all services
	@echo "Running tests for all services..."
	go test ./...

test-coverage: ## Run tests with coverage for all services
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Code quality
fmt: ## Format all code
	@echo "Formatting all services..."
	go fmt ./...

vet: ## Run go vet on all services
	@echo "Running go vet..."
	go vet ./...

lint: ## Run linter on all services
	@echo "Running linter..."
	golangci-lint run ./...

# Proto generation
proto-gen: ## Generate protobuf code for all services
	@echo "Generating proto files for all services..."
	@for dir in services/*/; do \
		if [ -f $$dir/Makefile ]; then \
			if grep -q "proto-gen:" $$dir/Makefile; then \
				echo "==> Generating protos for $$dir"; \
				$(MAKE) -C $$dir proto-gen; \
			fi \
		fi \
	done

proto-clean: ## Clean generated proto files
	@echo "Cleaning proto files..."
	@for dir in services/*/; do \
		if [ -f $$dir/Makefile ]; then \
			if grep -q "proto-clean:" $$dir/Makefile; then \
				echo "==> Cleaning protos for $$dir"; \
				$(MAKE) -C $$dir proto-clean; \
			fi \
		fi \
	done

proto-install: ## Install protoc dependencies
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Docker commands
docker-build: ## Build all Docker images
	docker-compose build

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-restart: ## Restart all services
	docker-compose restart

docker-logs: ## View logs from all services
	docker-compose logs -f

docker-ps: ## List running containers
	docker-compose ps

docker-clean: ## Remove all containers, volumes, and images
	docker-compose down -v --rmi all

# Infrastructure only
infra-up: ## Start only infrastructure (postgres, kafka, keycloak)
	docker-compose up -d postgres redpanda keycloak

infra-down: ## Stop infrastructure services
	docker-compose stop postgres redpanda keycloak

# Individual service commands
service-build: ## Build specific service (usage: make service-build SERVICE=blog-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "Error: SERVICE variable is required. Usage: make service-build SERVICE=blog-service"; \
		exit 1; \
	fi
	@if [ -f services/$(SERVICE)/Makefile ]; then \
		$(MAKE) -C services/$(SERVICE) build; \
	else \
		echo "Error: Service $(SERVICE) not found or has no Makefile"; \
		exit 1; \
	fi

service-test: ## Test specific service (usage: make service-test SERVICE=blog-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "Error: SERVICE variable is required. Usage: make service-test SERVICE=blog-service"; \
		exit 1; \
	fi
	@if [ -f services/$(SERVICE)/Makefile ]; then \
		$(MAKE) -C services/$(SERVICE) test; \
	else \
		echo "Error: Service $(SERVICE) not found or has no Makefile"; \
		exit 1; \
	fi

service-run: ## Run specific service (usage: make service-run SERVICE=blog-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "Error: SERVICE variable is required. Usage: make service-run SERVICE=blog-service"; \
		exit 1; \
	fi
	@if [ -f services/$(SERVICE)/Makefile ]; then \
		$(MAKE) -C services/$(SERVICE) run; \
	else \
		echo "Error: Service $(SERVICE) not found or has no Makefile"; \
		exit 1; \
	fi

# Cleanup
clean: ## Clean all build artifacts
	@echo "Cleaning all services..."
	@for dir in services/*/; do \
		if [ -f $$dir/Makefile ]; then \
			echo "==> Cleaning $$dir"; \
			$(MAKE) -C $$dir clean; \
		fi \
	done
	rm -f coverage.out coverage.html

# Database management
db-reset: ## Reset all databases (WARNING: destroys all data)
	docker-compose down -v
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5
	docker-compose restart blog-service foodfolio-service link-service notification-service webhook-service twitchbot-service

# Development helpers
dev-setup: deps proto-install ## Setup development environment
	@echo "Development environment ready!"
	@echo "Next steps:"
	@echo "  1. Copy .env.docker to .env and configure credentials"
	@echo "  2. Run 'make docker-up' to start all services"
	@echo "  3. Access gateway at http://localhost:3000"

dev-start: infra-up ## Start minimal development stack (infrastructure only)
	@echo "Infrastructure started. Start individual services with:"
	@echo "  make service-run SERVICE=blog-service"

# Git helpers
git-status: ## Show git status for all services
	@echo "=== Root ===" && git status -s
	@for dir in services/*/; do \
		if [ -d $$dir/.git ]; then \
			echo "=== $$dir ===" && cd $$dir && git status -s && cd ../..; \
		fi \
	done

# Info
info: ## Show workspace information
	@echo "ToxicToastGo Monorepo"
	@echo "====================="
	@echo "Go version: $$(go version)"
	@echo ""
	@echo "Services:"
	@for dir in services/*/; do \
		echo "  - $$(basename $$dir)"; \
	done
	@echo ""
	@echo "Infrastructure:"
	@echo "  - PostgreSQL (port 5432)"
	@echo "  - Redpanda/Kafka (port 19092)"
	@echo "  - Keycloak (port 8080)"
	@echo ""
	@echo "Run 'make help' for available commands"
