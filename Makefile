# LocalS3 Makefile
# Comprehensive build and test automation

.PHONY: help build clean run stop test test-unit test-all test-basic test-aws test-quick setup dev docker docker-build docker-run install deps

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := locals3
PORT := 3000
ENDPOINT_URL := http://localhost:$(PORT)
SCRIPTS_DIR := scripts
GO_FILES := $(shell find . -name "*.go" -type f)

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
CYAN := \033[36m
RESET := \033[0m

# Help target
help: ## Show this help message
	@echo "$(CYAN)LocalS3 - S3 Compatible Server$(RESET)"
	@echo "$(CYAN)==============================$(RESET)"
	@echo ""
	@echo "$(BLUE)Build Commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(build|clean|install|deps)"
	@echo ""
	@echo "$(BLUE)Server Commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(run|stop|dev)"
	@echo ""
	@echo "$(BLUE)Test Commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(test|setup)"
	@echo ""
	@echo "$(BLUE)Docker Commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(docker)"
	@echo ""
	@echo "$(YELLOW)Examples:$(RESET)"
	@echo "  make build          # Build the binary"
	@echo "  make run            # Start the server"
	@echo "  make test-all       # Run all tests"
	@echo "  make setup          # Setup AWS CLI"
	@echo "  make docker-run     # Run with Docker"

# Build targets
deps: ## Install Go dependencies
	@echo "$(BLUE)Installing dependencies...$(RESET)"
	go mod download
	go mod tidy

build: deps ## Build the LocalS3 binary
	@echo "$(BLUE)Building $(BINARY_NAME)...$(RESET)"
	go build -o $(BINARY_NAME) .
	@echo "$(GREEN)✓ Build complete$(RESET)"

install: build ## Install binary to system PATH
	@echo "$(BLUE)Installing $(BINARY_NAME) to system...$(RESET)"
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✓ Installed to /usr/local/bin/$(BINARY_NAME)$(RESET)"

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	rm -f $(BINARY_NAME)
	rm -rf ./data/*
	@echo "$(GREEN)✓ Clean complete$(RESET)"

# Server targets
run: build ## Start the LocalS3 server
	@echo "$(BLUE)Starting LocalS3 server on port $(PORT)...$(RESET)"
	@echo "$(CYAN)Press Ctrl+C to stop$(RESET)"
	PORT=$(PORT) ./$(BINARY_NAME)

dev: ## Start server in development mode with auto-restart
	@echo "$(BLUE)Starting LocalS3 in development mode...$(RESET)"
	@echo "$(YELLOW)Note: Install 'air' for auto-restart: go install github.com/cosmtrek/air@latest$(RESET)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		make run; \
	fi

stop: ## Stop the LocalS3 server
	@echo "$(BLUE)Stopping LocalS3 server...$(RESET)"
	@pkill -f "./$(BINARY_NAME)" || true
	@pkill -f "go run main.go" || true
	@echo "$(GREEN)✓ Server stopped$(RESET)"

# Test targets
test: test-basic ## Run basic tests (alias for test-basic)

test-unit: ## Run Go unit tests
	@echo "$(CYAN)Running unit tests...$(RESET)"
	@go test ./... -v

test-basic: ## Run basic HTTP API tests
	@echo "$(BLUE)Running basic tests...$(RESET)"
	@chmod +x $(SCRIPTS_DIR)/test.sh
	@cd $(SCRIPTS_DIR) && ./test.sh

test-sample: ## Run sample data tests
	@echo "$(BLUE)Running sample data tests...$(RESET)"
	@chmod +x $(SCRIPTS_DIR)/test_sample.sh
	@cd $(SCRIPTS_DIR) && ./test_sample.sh

test-http: ## Run direct HTTP tests
	@echo "$(BLUE)Running direct HTTP tests...$(RESET)"
	@chmod +x $(SCRIPTS_DIR)/test_direct_http.sh
	@cd $(SCRIPTS_DIR) && ./test_direct_http.sh

test-aws-quick: ## Run quick AWS CLI tests
	@echo "$(BLUE)Running quick AWS CLI tests...$(RESET)"
	@chmod +x $(SCRIPTS_DIR)/test_aws_quick.sh
	@cd $(SCRIPTS_DIR) && ./test_aws_quick.sh

test-aws: ## Run comprehensive AWS CLI tests
	@echo "$(BLUE)Running comprehensive AWS CLI tests...$(RESET)"
	@chmod +x $(SCRIPTS_DIR)/test_aws_cli_complete.sh
	@cd $(SCRIPTS_DIR) && ./test_aws_cli_complete.sh

test-all: ## Run all tests
	@echo "$(CYAN)Running all LocalS3 tests...$(RESET)"
	@echo "$(CYAN)=========================$(RESET)"
	@make --no-print-directory test-basic
	@echo ""
	@make --no-print-directory test-sample
	@echo ""
	@make --no-print-directory test-http
	@echo ""
	@make --no-print-directory test-aws-quick
	@echo ""
	@echo "$(GREEN)✓ All tests completed$(RESET)"

test-menu: ## Run interactive test menu
	@echo "$(BLUE)Starting interactive test menu...$(RESET)"
	@./test-menu

# Setup targets
setup: ## Setup AWS CLI for LocalS3
	@echo "$(BLUE)Setting up AWS CLI...$(RESET)"
	@chmod +x $(SCRIPTS_DIR)/setup_aws_cli.sh
	@cd $(SCRIPTS_DIR) && ./setup_aws_cli.sh

# Docker targets
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(RESET)"
	docker build -t locals3:latest .
	@echo "$(GREEN)✓ Docker image built$(RESET)"

docker-run: docker-build ## Run LocalS3 in Docker container
	@echo "$(BLUE)Running LocalS3 in Docker...$(RESET)"
	docker run -d --name locals3 -p $(PORT):3000 \
		-v locals3_data:/data \
		-e PORT=3000 \
		-e LOG_LEVEL=info \
		locals3:latest
	@echo "$(GREEN)✓ LocalS3 running in Docker on port $(PORT)$(RESET)"
	@echo "$(CYAN)Use 'make docker-stop' to stop$(RESET)"

docker-stop: ## Stop Docker container
	@echo "$(BLUE)Stopping Docker container...$(RESET)"
	@docker stop locals3 || true
	@docker rm locals3 || true
	@echo "$(GREEN)✓ Docker container stopped$(RESET)"

docker-logs: ## Show Docker container logs
	@docker logs -f locals3

docker-compose-up: ## Start with docker-compose
	@echo "$(BLUE)Starting LocalS3 with docker-compose...$(RESET)"
	docker-compose up -d
	@echo "$(GREEN)✓ LocalS3 started with docker-compose$(RESET)"

docker-compose-down: ## Stop docker-compose services
	@echo "$(BLUE)Stopping docker-compose services...$(RESET)"
	docker-compose down
	@echo "$(GREEN)✓ Services stopped$(RESET)"

# Development targets
fmt: ## Format Go code
	@echo "$(BLUE)Formatting Go code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(RESET)"

lint: ## Run Go linter
	@echo "$(BLUE)Running Go linter...$(RESET)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)"; \
		go vet ./...; \
	fi

vet: ## Run Go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	go vet ./...
	@echo "$(GREEN)✓ Go vet passed$(RESET)"

# Health check targets
health: ## Check if server is running
	@echo "$(BLUE)Checking server health...$(RESET)"
	@if curl -s $(ENDPOINT_URL)/health > /dev/null; then \
		echo "$(GREEN)✓ Server is running on $(ENDPOINT_URL)$(RESET)"; \
	else \
		echo "$(RED)✗ Server is not running$(RESET)"; \
		exit 1; \
	fi

status: health ## Alias for health check

# Quick start targets
quick-start: build run ## Build and start server quickly

scripts: ## List all available scripts
	@echo "$(BLUE)Available scripts in $(SCRIPTS_DIR):$(RESET)"
	@for script in $(SCRIPTS_DIR)/*.sh; do \
		basename "$$script" .sh | sed 's/^/  - /'; \
	done
	@echo ""
	@echo "$(YELLOW)Usage:$(RESET)"
	@echo "  make test-basic      # Run specific test"
	@echo "  ./run-script test    # Run script directly"
	@echo "  make test-menu       # Interactive menu"

demo: build ## Run a complete demo
	@echo "$(CYAN)LocalS3 Demo$(RESET)"
	@echo "$(CYAN)============$(RESET)"
	@echo ""
	@echo "$(BLUE)1. Starting server...$(RESET)"
	@make --no-print-directory run &
	@sleep 3
	@echo ""
	@echo "$(BLUE)2. Running basic tests...$(RESET)"
	@make --no-print-directory test-basic
	@echo ""
	@echo "$(BLUE)3. Running AWS CLI tests...$(RESET)"
	@make --no-print-directory test-aws-quick
	@echo ""
	@echo "$(GREEN)✓ Demo completed successfully!$(RESET)"
	@make --no-print-directory stop

# Release targets
version: ## Show version information
	@echo "LocalS3 Version: 1.0.0"
	@echo "Go Version: $(shell go version)"
	@echo "Build Date: $(shell date)"

# Maintenance targets
update-deps: ## Update Go dependencies
	@echo "$(BLUE)Updating dependencies...$(RESET)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(RESET)"

backup-data: ## Backup data directory
	@echo "$(BLUE)Backing up data directory...$(RESET)"
	@if [ -d "./data" ]; then \
		tar -czf "backup_$(shell date +%Y%m%d_%H%M%S).tar.gz" ./data; \
		echo "$(GREEN)✓ Data backed up$(RESET)"; \
	else \
		echo "$(YELLOW)No data directory found$(RESET)"; \
	fi

# Check dependencies
check-deps: ## Check if required tools are installed
	@echo "$(BLUE)Checking dependencies...$(RESET)"
	@echo -n "Go: "
	@if command -v go > /dev/null; then echo "$(GREEN)✓$(RESET)"; else echo "$(RED)✗$(RESET)"; fi
	@echo -n "Docker: "
	@if command -v docker > /dev/null; then echo "$(GREEN)✓$(RESET)"; else echo "$(RED)✗$(RESET)"; fi
	@echo -n "AWS CLI: "
	@if command -v aws > /dev/null; then echo "$(GREEN)✓$(RESET)"; else echo "$(RED)✗$(RESET)"; fi
	@echo -n "curl: "
	@if command -v curl > /dev/null; then echo "$(GREEN)✓$(RESET)"; else echo "$(RED)✗$(RESET)"; fi
	@echo -n "jq: "
	@if command -v jq > /dev/null; then echo "$(GREEN)✓$(RESET)"; else echo "$(RED)✗$(RESET)"; fi
