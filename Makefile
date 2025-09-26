# Enterprise License Management System - Build Automation
# ========================================================

# Variables
APP_NAME := license-server
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | cut -d' ' -f3)
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# Directories
BIN_DIR := bin
DIST_DIR := dist
COVERAGE_DIR := coverage
DOCS_DIR := docs

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Default target
.PHONY: help
help: ## Show this help message
	@echo "$(BLUE)Enterprise License Management System$(NC)"
	@echo "$(BLUE)====================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
.PHONY: dev-setup
dev-setup: ## Setup development environment
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@go mod download
	@go mod verify
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/air-verse/air@latest
	@echo "$(GREEN)Development environment ready!$(NC)"

.PHONY: dev
dev: ## Start development server with hot reload
	@echo "$(BLUE)Starting development server...$(NC)"
	@air

.PHONY: build
build: clean ## Build the application
	@echo "$(BLUE)Building $(APP_NAME)...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) .
	@echo "$(GREEN)Build completed: $(BIN_DIR)/$(APP_NAME)$(NC)"

.PHONY: build-linux
build-linux: clean ## Build for Linux
	@echo "$(BLUE)Building for Linux...$(NC)"
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux .
	@echo "$(GREEN)Linux build completed: $(BIN_DIR)/$(APP_NAME)-linux$(NC)"

.PHONY: build-windows
build-windows: clean ## Build for Windows
	@echo "$(BLUE)Building for Windows...$(NC)"
	@mkdir -p $(BIN_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-windows.exe .
	@echo "$(GREEN)Windows build completed: $(BIN_DIR)/$(APP_NAME)-windows.exe$(NC)"

.PHONY: build-darwin
build-darwin: clean ## Build for macOS
	@echo "$(BLUE)Building for macOS...$(NC)"
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin .
	@echo "$(GREEN)macOS build completed: $(BIN_DIR)/$(APP_NAME)-darwin$(NC)"

.PHONY: build-all
build-all: build-linux build-windows build-darwin ## Build for all platforms
	@echo "$(GREEN)All platform builds completed!$(NC)"

# Testing targets
.PHONY: test
test: ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -v ./...

.PHONY: test-unit
test-unit: ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(NC)"
	@go test -v -short ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	@go test -v -tags=integration ./...

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "$(BLUE)Running end-to-end tests...$(NC)"
	@go test -v -tags=e2e ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_DIR)/coverage.html$(NC)"

.PHONY: test-performance
test-performance: ## Run performance tests
	@echo "$(BLUE)Running performance tests...$(NC)"
	@go test -v -tags=performance -bench=. ./...

.PHONY: test-security
test-security: ## Run security tests
	@echo "$(BLUE)Running security tests...$(NC)"
	@go test -v -tags=security ./...

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...

# Code quality targets
.PHONY: fmt
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Code formatted!$(NC)"

.PHONY: lint
lint: ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	@golangci-lint run

.PHONY: lint-fix
lint-fix: ## Run linter with auto-fix
	@echo "$(BLUE)Running linter with auto-fix...$(NC)"
	@golangci-lint run --fix

.PHONY: vet
vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "$(BLUE)Tidying go modules...$(NC)"
	@go mod tidy
	@go mod verify

# Documentation targets
.PHONY: docs
docs: ## Generate API documentation
	@echo "$(BLUE)Generating API documentation...$(NC)"
	@mkdir -p $(DOCS_DIR)
	@swag init -g main.go -o $(DOCS_DIR)/swagger
	@echo "$(GREEN)API documentation generated: $(DOCS_DIR)/swagger$(NC)"

.PHONY: docs-serve
docs-serve: docs ## Serve API documentation
	@echo "$(BLUE)Serving API documentation at http://localhost:8080/swagger/index.html$(NC)"
	@swag serve -g main.go

# Database targets
.PHONY: migrate
migrate: ## Run database migrations
	@echo "$(BLUE)Running database migrations...$(NC)"
	@go run . migrate

.PHONY: migrate-create
migrate-create: ## Create new migration
	@echo "$(BLUE)Creating new migration...$(NC)"
	@read -p "Enter migration name: " name; \
	mkdir -p migrations; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	echo "-- Migration: $$name" > migrations/$${timestamp}_$${name}.sql; \
	echo "-- Created: $$(date)" >> migrations/$${timestamp}_$${name}.sql; \
	echo "" >> migrations/$${timestamp}_$${name}.sql; \
	echo "$(GREEN)Migration created: migrations/$${timestamp}_$${name}.sql$(NC)"

.PHONY: db-backup
db-backup: ## Backup database
	@echo "$(BLUE)Backing up database...$(NC)"
	@mkdir -p backups
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	mysqldump -h localhost -u root license_db > backups/backup_$${timestamp}.sql; \
	echo "$(GREEN)Database backup created: backups/backup_$${timestamp}.sql$(NC)"

.PHONY: db-restore
db-restore: ## Restore database from backup
	@echo "$(BLUE)Restoring database...$(NC)"
	@read -p "Enter backup file path: " file; \
	mysql -h localhost -u root license_db < $$file; \
	echo "$(GREEN)Database restored from: $$file$(NC)"

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "$(GREEN)Docker image built: $(APP_NAME):$(VERSION)$(NC)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	@docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

.PHONY: docker-compose-up
docker-compose-up: ## Start services with Docker Compose
	@echo "$(BLUE)Starting services with Docker Compose...$(NC)"
	@docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop services with Docker Compose
	@echo "$(BLUE)Stopping services with Docker Compose...$(NC)"
	@docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## View Docker Compose logs
	@docker-compose logs -f

# Kubernetes targets
.PHONY: k8s-deploy
k8s-deploy: ## Deploy to Kubernetes
	@echo "$(BLUE)Deploying to Kubernetes...$(NC)"
	@kubectl apply -f deployments/kubernetes/

.PHONY: k8s-delete
k8s-delete: ## Delete from Kubernetes
	@echo "$(BLUE)Deleting from Kubernetes...$(NC)"
	@kubectl delete -f deployments/kubernetes/

.PHONY: k8s-status
k8s-status: ## Check Kubernetes deployment status
	@kubectl get pods -n license-system
	@kubectl get services -n license-system

# Security targets
.PHONY: security-scan
security-scan: ## Run security scan
	@echo "$(BLUE)Running security scan...$(NC)"
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@gosec ./...

.PHONY: dependency-check
dependency-check: ## Check for vulnerable dependencies
	@echo "$(BLUE)Checking for vulnerable dependencies...$(NC)"
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

# Release targets
.PHONY: release
release: test lint security-scan build-all ## Create release
	@echo "$(BLUE)Creating release...$(NC)"
	@mkdir -p $(DIST_DIR)
	@cp $(BIN_DIR)/* $(DIST_DIR)/
	@echo "$(GREEN)Release created in $(DIST_DIR)/$(NC)"

.PHONY: release-docker
release-docker: test lint security-scan docker-build ## Create Docker release
	@echo "$(BLUE)Creating Docker release...$(NC)"
	@docker push $(APP_NAME):$(VERSION)
	@docker push $(APP_NAME):latest
	@echo "$(GREEN)Docker release pushed!$(NC)"

# Monitoring targets
.PHONY: metrics
metrics: ## Start metrics collection
	@echo "$(BLUE)Starting metrics collection...$(NC)"
	@go run . serve --metrics

.PHONY: health-check
health-check: ## Check application health
	@echo "$(BLUE)Checking application health...$(NC)"
	@curl -f http://localhost:8080/health || echo "$(RED)Health check failed!$(NC)"

# Utility targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BIN_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@go clean
	@echo "$(GREEN)Clean completed!$(NC)"

.PHONY: install
install: build ## Install the application
	@echo "$(BLUE)Installing $(APP_NAME)...$(NC)"
	@sudo cp $(BIN_DIR)/$(APP_NAME) /usr/local/bin/
	@echo "$(GREEN)$(APP_NAME) installed to /usr/local/bin/$(NC)"

.PHONY: uninstall
uninstall: ## Uninstall the application
	@echo "$(BLUE)Uninstalling $(APP_NAME)...$(NC)"
	@sudo rm -f /usr/local/bin/$(APP_NAME)
	@echo "$(GREEN)$(APP_NAME) uninstalled!$(NC)"

.PHONY: version
version: ## Show version information
	@echo "$(BLUE)Version Information:$(NC)"
	@echo "  App: $(APP_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(GO_VERSION)"

# Development workflow
.PHONY: dev-workflow
dev-workflow: dev-setup fmt lint test ## Complete development workflow
	@echo "$(GREEN)Development workflow completed!$(NC)"

.PHONY: ci-workflow
ci-workflow: mod-tidy fmt lint test test-coverage security-scan build ## CI workflow
	@echo "$(GREEN)CI workflow completed!$(NC)"

.PHONY: pre-commit
pre-commit: fmt lint test ## Pre-commit checks
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

# Show help by default
.DEFAULT_GOAL := help