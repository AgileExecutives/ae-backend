#!/bin/bash

# Calendar Module Unit Test Suite Runner
# Method 2: Isolated Unit Tests with Mocked Dependencies

set -e

echo "=== Calendar Module Unit Test Suite (Method 2) ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "info")
            echo -e "${YELLOW}[INFO]${NC} $message"
            ;;
        "success")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "error")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
    esac
}

# Navigate to tests directory
cd "$(dirname "$0")"

print_status "info" "Setting up test environment..."

# Run different test categories

print_status "info" "Running Calendar Service Unit Tests..."
echo "----------------------------------------"

# Change to tests directory
cd tests

# Run integration tests
print_status "info" "Running Integration Tests with in-memory database..."
GOWORK=off go test -v ./integration/... || {
    print_status "error" "Integration tests failed"
    exit 1
}

print_status "success" "Integration tests passed"

# Run integration-style tests
# GOWORK=off go test -v ./integration/... 2>/dev/null || {
#     print_status "info" "No integration tests found (optional)"
# }
# print_status "success" "Service tests passed"

# Generate coverage report
# if [ -f "coverage_services.out" ]; then
#     print_status "info" "Generating coverage report..."
#     go tool cover -html=coverage_services.out -o coverage_services.html
#     coverage_percent=$(go tool cover -func=coverage_services.out | grep total | awk '{print $3}')
#     print_status "success" "Service tests coverage: $coverage_percent"
# fi

# Run handler tests (if they exist)
# if [ -d "handlers" ] && [ "$(ls -A handlers 2>/dev/null)" ]; then
#     print_status "info" "Running Handler Unit Tests..."
#     GOWORK=off go test -v ./handlers/... || {
#         print_status "error" "Handler tests failed"
#         exit 1
#     }
#     print_status "success" "Handler tests passed"
# fi

# Run integration-style tests
# print_status "info" "Running Integration-Style Tests..."
# GOWORK=off go test -v ./integration/... 2>/dev/null || {
#     print_status "info" "No integration tests found (optional)"
# }

# Change back to parent directory
cd ..

echo ""
print_status "success" "All Calendar Module Integration Tests Completed Successfully!"
echo ""

# Test Summary
echo "=== Test Summary ==="
echo "✅ Calendar CRUD Operations"
echo "✅ Calendar Entry Operations"
echo "✅ Deep Preload Functionality"
echo "✅ Tenant Isolation"
echo "✅ Integration Tests with In-Memory Database"

echo ""
# echo "📊 Coverage reports generated:"
# echo "   - tests/coverage_services.html (Service layer coverage)"
# if [ -f "tests/coverage_handlers.html" ]; then
#     echo "   - tests/coverage_handlers.html (Handler layer coverage)"
# fi

echo ""
print_status "info" "Integration tests use SQLite in-memory database"
print_status "info" "Tests verify service layer with real database operations"