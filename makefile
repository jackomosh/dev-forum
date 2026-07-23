
# FORUM PROJECT MAKEFILE - Setup & Variables
# Variables
APP_NAME := forum
BINARY_NAME := forum
DOCKER_IMAGE := forum
DOCKER_TAG := latest
DOCKER_CONTAINER := forum-container
DOCKER_PORT := 8089

# Go variables
GO := go
GOFLAGS := -v
GOTEST := $(GO) test
GOTESTFLAGS := -v -race -coverprofile=coverage.out
GOMOD := $(GO) mod
GOFMT := $(GO) fmt

# Directories
CMD_DIR := ./cmd/forum
BUILD_DIR := ./build

# Colors
GREEN  := $(shell tput -Txterm setaf 2 2>/dev/null || echo "")
YELLOW := $(shell tput -Txterm setaf 3 2>/dev/null || echo "")
RED    := $(shell tput -Txterm setaf 1 2>/dev/null || echo "")
RESET  := $(shell tput -Txterm sgr0 2>/dev/null || echo "")

# HELP & UTILITY TARGETS

.PHONY: help
help:
	@echo '$(GREEN)Forum Project Makefile$(RESET)'
	@echo ''
	@echo '$(YELLOW)Usage:$(RESET)'
	@echo '  make $(GREEN)<target>$(RESET)'
	@echo ''
	@echo '$(YELLOW)Targets:$(RESET)'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ''

.PHONY: version
version: ## Show Go version
	$(GO) version

.PHONY: env
env: ## Show environment variables
	@echo "$(YELLOW)Environment:$(RESET)"
	@echo "  APP_NAME: $(APP_NAME)"
	@echo "  BINARY_NAME: $(BINARY_NAME)"
	@echo "  BUILD_DIR: $(BUILD_DIR)"
	@echo "  DOCKER_IMAGE: $(DOCKER_IMAGE)"
	@echo "  DOCKER_TAG: $(DOCKER_TAG)"
	@echo "  DOCKER_PORT: $(DOCKER_PORT)"

# DEPENDENCIES
.PHONY: deps
deps: ## Download dependencies
	@echo "$(YELLOW)Downloading dependencies...$(RESET)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies downloaded$(RESET)"

# CODE QUALITY
.PHONY: fmt
fmt: ## Format code
	@echo "$(YELLOW)Formatting code...$(RESET)"
	$(GOFMT) ./...
	@echo "$(GREEN)Code formatted$(RESET)"

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@echo "$(YELLOW)Running linter...$(RESET)"
	@which golangci-lint > /dev/null 2>&1 || (echo "$(RED)golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)" && exit 1)
	golangci-lint run ./...
	@echo "$(GREEN)Linting complete$(RESET)"

# TESTING
.PHONY: test
test: ## Run tests with coverage
	@echo "$(YELLOW)Running tests...$(RESET)"
	$(GOTEST) $(GOTESTFLAGS) ./...
	@echo "$(GREEN)Tests complete$(RESET)"

# BUILD
.PHONY: build
build: deps ## Build the application
	@echo "$(YELLOW)Building application...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

.PHONY: run
run: ## Run the application locally
	@echo "$(YELLOW)Running application...$(RESET)"
	$(GO) run $(CMD_DIR)/main.go $(CMD_DIR)/app.go

# RELEASE
.PHONY: release
release: clean deps test build ## Create a release build
	@echo "$(GREEN)Release build complete$(RESET)"

.PHONY: cross-build
cross-build: ## Build for multiple platforms
	@echo "$(YELLOW)Building for multiple platforms...$(RESET)"
	@mkdir -p $(BUILD_DIR)/releases
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/releases/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "$(GREEN)Cross-platform builds complete$(RESET)"
	@ls -la $(BUILD_DIR)/releases/

# CLEAN
.PHONY: clean
clean: ## Clean build artifacts and cache
	@echo "$(YELLOW)Cleaning...$(RESET)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out
	$(GO) clean -cache
	@echo "$(GREEN)Cleaned$(RESET)"

.PHONY: clean-db
clean-db: ## Remove database file
	@echo "$(YELLOW)Removing database...$(RESET)"
	rm -f forum.db
	@echo "$(GREEN)Database removed$(RESET)"

.PHONY: reset
reset: clean clean-db ## Reset everything
	@echo "$(GREEN)Reset complete$(RESET)"

# INSTALL TOOLS
.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(YELLOW)Installing development tools...$(RESET)"
	$(GO) install github.com/air-verse/air@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/go-delve/delve/cmd/dlv@latest
	@echo "$(GREEN)Development tools installed$(RESET)"

# DEV (hot reload)
.PHONY: dev
dev: ## Run in development mode with hot reload (requires air)
	@echo "$(YELLOW)Running in development mode...$(RESET)"
	@which air > /dev/null 2>&1 || (echo "$(RED)air not installed. Run: make install-tools$(RESET)" && exit 1)
	air

# ALL
.PHONY: all
all: fmt lint test build ## Run all targets (fmt, lint, test, build)

# FORUM PROJECT MAKEFILE - Docker Targets

# DOCKER TARGETS

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(RESET)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"

.PHONY: docker-run
docker-run: ## Run Docker container (does NOT rebuild)
	@echo "$(YELLOW)Running Docker container...$(RESET)"
	@docker rm -f $(DOCKER_CONTAINER) 2>/dev/null || true
	docker run -d --name $(DOCKER_CONTAINER) -p 8089:$(DOCKER_PORT) $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "$(GREEN)Docker container running: $(DOCKER_CONTAINER)$(RESET)"
	@echo "$(GREEN)Access at: http://localhost:8089$(RESET)"

.PHONY: docker-rebuild
docker-rebuild: docker-build docker-run ## Rebuild and run Docker container

.PHONY: docker-stop
docker-stop: ## Stop Docker container
	@echo "$(YELLOW)Stopping Docker container...$(RESET)"
	@docker stop $(DOCKER_CONTAINER) 2>/dev/null || echo "Container not running"
	@docker rm $(DOCKER_CONTAINER) 2>/dev/null || echo "Container not found"
	@echo "$(GREEN)Docker container stopped$(RESET)"

.PHONY: docker-logs
docker-logs: ## View Docker container logs
	docker logs -f $(DOCKER_CONTAINER)

.PHONY: docker-shell
docker-shell: ## Shell into Docker container
	docker exec -it $(DOCKER_CONTAINER) sh

.PHONY: docker-clean
docker-clean: docker-stop ## Remove Docker image
	@echo "$(YELLOW)Removing Docker image...$(RESET)"
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || echo "Image not found"
	@echo "$(GREEN)Docker image removed$(RESET)"

.PHONY: docker-ps
docker-ps: ## Show running containers
	docker ps --filter "name=$(DOCKER_CONTAINER)"

# DOCKER-COMPOSE TARGETS (optional)
.PHONY: compose-up
compose-up: ## Start with docker-compose
	docker-compose up -d

.PHONY: compose-down
compose-down: ## Stop docker-compose
	docker-compose down

.PHONY: compose-logs
compose-logs: ## View docker-compose logs
	docker-compose logs -f