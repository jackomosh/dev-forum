
# FORUM PROJECT MAKEFILE - PART 1: Setup & Variables
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