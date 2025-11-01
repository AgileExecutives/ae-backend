# Minimal Server Makefile

.PHONY: help run build clean test setup

# Default target
help:
	@echo "Available commands:"
	@echo "  setup      - Setup environment file"
	@echo "  run        - Run the server in development mode"
	@echo "  build      - Build the server binary"
	@echo "  clean      - Remove built files"
	@echo "  test       - Run API tests (requires running server)"
	@echo "  test-help  - Show detailed testing instructions"

# Setup environment
setup:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		echo "Please edit .env with your database configuration"; \
	else \
		echo ".env file already exists"; \
	fi

# Run in development
run: setup
	@echo "Starting minimal server..."

	go run .

# Build binary
build:
	@echo "Building minimal server..."
	@mkdir -p bin
	go build -o bin/minimal-server .
	@echo "Binary created at bin/minimal-server"

# Clean built files
clean:
	@echo "Cleaning..."
	rm -rf bin/
	@echo "Clean complete"

# Run tests
test:
	@echo "Testing API endpoints..."
	@echo "Health check:"
	@curl -s http://localhost:8080/api/v1/health | jq . || curl -s http://localhost:8080/api/v1/health
	@echo ""
	@echo "Ping module test (public):"
	@curl -s http://localhost:8080/api/v1/ping/ping | jq . || curl -s http://localhost:8080/api/v1/ping/ping
	@echo ""
	@echo "API ready for testing!"

# Show test instructions
test-help:
	@./test_ping.sh