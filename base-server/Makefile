# AE SaaS Basic Makefile

.PHONY: help build run test clean docker-build docker-run setup

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
build: ## Build the application
	go build -v -o bin/ae-saas-basic .

run: ## Run the application
	go run main.go

test: ## Run Go unit tests
	go test -v ./...

test-hurl: ## Run HURL API tests
	./run-hurl-tests.sh

test-all: test test-hurl ## Run all tests (unit + API)

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

setup: ## Set up the development environment
	cp .env.example .env
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Setup complete! Edit .env file with your configuration."

# Database commands
db-create: ## Create database (requires psql)
	createdb ae_saas_basic

db-drop: ## Drop database (requires psql)
	dropdb ae_saas_basic

db-reset: db-drop db-create ## Reset database

# Docker commands
docker-build: ## Build Docker image
	docker build -t ae-saas-basic .

docker-run: ## Run with Docker Compose
	docker-compose up -d

docker-stop: ## Stop Docker containers
	docker-compose down

# Code quality
fmt: ## Format code
	go fmt ./...

vet: ## Vet code
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run

# Documentation
swag-install: ## Install Swagger CLI tool
	go install github.com/swaggo/swag/cmd/swag@latest

swag: ## Generate Swagger documentation
	swag init -g main.go -o ./docs

docs: swag ## Alias for swag target

build-with-docs: swag build ## Build application with fresh documentation

# Dependencies
deps: ## Download dependencies
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy