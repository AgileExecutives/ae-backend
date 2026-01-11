# Unburdy Extended API Makefile

# Default goal
.DEFAULT_GOAL := help

# Variables
APP_NAME := unburdy-server-api
GO_VERSION := 1.24
MAIN_FILE := main.go
BASE_API_PATH := ../base-server

# Colors for output
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

## help: Display this help message
.PHONY: help
help:
	@echo "$(CYAN)Unburdy Extended API Commands:$(RESET)"
	@grep -E '^## .*:' $(MAKEFILE_LIST) | sed 's/## \(.*\): \(.*\)/  $(GREEN)\1$(RESET): \2/' | column -t -s ':'

## install: Install all dependencies
.PHONY: install
install:
	@echo "$(YELLOW)Installing dependencies...$(RESET)"
	go mod tidy
	go mod download

## run: Run the extended API server
.PHONY: run
run:
	@echo "$(YELLOW)Starting Unburdy Extended API server...$(RESET)"
	@echo "$(CYAN)Includes all AE SaaS Basic endpoints plus client management$(RESET)"
	go run $(MAIN_FILE)

## dev: Run in development mode with hot reload
.PHONY: dev
dev:
	@echo "$(YELLOW)Starting development server with hot reload...$(RESET)"
	go run $(MAIN_FILE)

## build: Build the application
.PHONY: build
build:
	@echo "$(YELLOW)Building $(APP_NAME)...$(RESET)"
	go build -o bin/$(APP_NAME) $(MAIN_FILE)
	@echo "$(GREEN)Build completed: bin/$(APP_NAME)$(RESET)"

## test: Run all tests
.PHONY: test
test:
	@echo "$(YELLOW)Running tests...$(RESET)"
	go test ./... -v

## test-hurl: Run HURL API tests (uses base-server tests)
.PHONY: test-hurl
test-hurl:
	@echo "$(YELLOW)Running HURL API tests against unburdy server...$(RESET)"
	@echo "$(CYAN)Using base-server test suite to verify API compatibility$(RESET)"
	cd $(BASE_API_PATH) && ./run-hurl-tests.sh

## test-all: Run all tests (unit + API)
.PHONY: test-all
test-all: test test-hurl

## clean: Clean build artifacts
.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	rm -rf bin/
	go clean

## fmt: Format Go code
.PHONY: fmt
fmt:
	@echo "$(YELLOW)Formatting Go code...$(RESET)"
	go fmt ./...

## lint: Run linter
.PHONY: lint
lint:
	@echo "$(YELLOW)Running linter...$(RESET)"
	golangci-lint run

## swagger: Generate Swagger documentation (includes all modules)
.PHONY: swagger
swagger:
	@echo "$(YELLOW)Generating Swagger documentation...$(RESET)"
	@echo "$(CYAN)Scanning all modules: unburdy handlers, base API, documents, invoice, templates, audit, client_management, calendar, and booking...$(RESET)"
	swag init -g $(MAIN_FILE) -o ./docs --parseDependency --parseInternal --parseDependencyLevel 3 --parseGoList --dir .,$(BASE_API_PATH),../modules/calendar,../modules/booking,../modules/documents,../modules/invoice,$(BASE_API_PATH)/modules/templates,./modules/audit,./modules/client_management
	@echo "$(GREEN)✓ Swagger docs generated with all module endpoints$(RESET)"

## docker-build: Build Docker image
.PHONY: docker-build
docker-build:
	@echo "$(YELLOW)Building Docker image...$(RESET)"
	docker build -t $(APP_NAME):latest .

## check-base: Check if base API is available
.PHONY: check-base
check-base:
	@echo "$(YELLOW)Checking base AE SaaS API...$(RESET)"
	@if [ -d "$(BASE_API_PATH)" ]; then \
		echo "$(GREEN)✓ Base AE SaaS API found at $(BASE_API_PATH)$(RESET)"; \
	else \
		echo "$(RED)✗ Base AE SaaS API not found at $(BASE_API_PATH)$(RESET)"; \
		echo "$(YELLOW)Please ensure the ae-saas/server-api directory exists$(RESET)"; \
	fi

## endpoints: List all available API endpoints
.PHONY: endpoints
endpoints:
	@echo "$(CYAN)Available API Endpoints:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Base Endpoints (from AE SaaS Basic):$(RESET)"
	@echo "  POST   /api/v1/auth/login"
	@echo "  POST   /api/v1/auth/register"
	@echo "  POST   /api/v1/auth/refresh"
	@echo "  GET    /api/v1/users"
	@echo "  GET    /api/v1/customers"
	@echo "  POST   /api/v1/customers"
	@echo "  GET    /api/v1/contacts"
	@echo "  POST   /api/v1/contacts"
	@echo "  GET    /api/v1/emails"
	@echo "  POST   /api/v1/emails"
	@echo ""
	@echo "$(YELLOW)Extended Endpoints (Client Management):$(RESET)"
	@echo "  GET    /api/v1/clients          - Get all clients"
	@echo "  POST   /api/v1/clients          - Create new client"
	@echo "  GET    /api/v1/clients/search   - Search clients"
	@echo "  GET    /api/v1/clients/:id      - Get client by ID"
	@echo "  PUT    /api/v1/clients/:id      - Update client"
	@echo "  DELETE /api/v1/clients/:id      - Delete client"
	@echo ""
	@echo "$(YELLOW)Calendar Module Endpoints:$(RESET)"
	@echo "  GET    /api/v1/calendar         - Get all calendars"
	@echo "  POST   /api/v1/calendar         - Create new calendar"
	@echo "  GET    /api/v1/calendar-entries - Get all calendar entries"
	@echo "  POST   /api/v1/calendar-entries - Create new calendar entry"
	@echo "  GET    /api/v1/calendar-series  - Get all calendar series"
	@echo "  POST   /api/v1/calendar-series  - Create new calendar series"
	@echo ""
	@echo "$(CYAN)Documentation: http://localhost:8080/swagger/index.html$(RESET)"

## status: Show project status
.PHONY: status
status:
	@echo "$(CYAN)Unburdy Extended API Status:$(RESET)"
	@echo "$(YELLOW)Go version:$(RESET) $(shell go version)"
	@echo "$(YELLOW)Module:$(RESET) $(shell grep '^module' go.mod | cut -d' ' -f2)"
	@echo "$(YELLOW)Base API:$(RESET) $(BASE_API_PATH)"
	@make check-base

## init: Initialize project (run after cloning)
.PHONY: init
init: check-base install
	@echo "$(GREEN)✓ Project initialized successfully!$(RESET)"
	@echo "$(YELLOW)Next steps:$(RESET)"
	@echo "  1. Run 'make run' to start the server"
	@echo "  2. Visit http://localhost:8080/swagger/index.html for API docs"
	@echo "  3. Use 'make endpoints' to see all available endpoints"