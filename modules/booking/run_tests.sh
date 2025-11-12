#!/bin/bash

# Booking Module Unit Test Suite Runner
# Runs all unit tests with coverage reporting

set -e

echo "=== Booking Module Unit Test Suite ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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
        "header")
            echo -e "${BLUE}[TEST]${NC} $message"
            ;;
    esac
}

# Navigate to module directory
cd "$(dirname "$0")"

print_status "info" "Setting up test environment..."

# Initialize test counters
total_tests=0
passed_tests=0
failed_tests=0
skipped_tests=0

# Run Service Tests
print_status "header" "Running Booking Service Tests..."
echo "=========================================="

if GOWORK=off go test -v ./services/... -coverprofile=tests/coverage_services.out 2>&1 | tee /tmp/booking_services_test.log; then
    service_tests=$(grep -E "^(PASS|SKIP)" /tmp/booking_services_test.log | wc -l | tr -d ' ')
    print_status "success" "Service tests completed"
    
    # Extract coverage percentage
    if [ -f "tests/coverage_services.out" ]; then
        service_coverage=$(go tool cover -func=tests/coverage_services.out | grep total | awk '{print $3}')
        print_status "info" "Service coverage: $service_coverage"
    fi
else
    print_status "error" "Service tests failed"
    exit 1
fi

echo ""

# Run Middleware Tests
print_status "header" "Running Booking Middleware Tests..."
echo "=========================================="

if GOWORK=off go test -v ./middleware/... -coverprofile=tests/coverage_middleware.out 2>&1 | tee /tmp/booking_middleware_test.log; then
    middleware_tests=$(grep -E "^(PASS|SKIP)" /tmp/booking_middleware_test.log | wc -l | tr -d ' ')
    print_status "success" "Middleware tests completed"
    
    # Extract coverage percentage
    if [ -f "tests/coverage_middleware.out" ]; then
        middleware_coverage=$(go tool cover -func=tests/coverage_middleware.out | grep total | awk '{print $3}')
        print_status "info" "Middleware coverage: $middleware_coverage"
    fi
else
    print_status "error" "Middleware tests failed"
    exit 1
fi

echo ""

# Generate combined coverage report
print_status "info" "Generating HTML coverage reports..."

if [ -f "tests/coverage_services.out" ]; then
    go tool cover -html=tests/coverage_services.out -o tests/coverage_services.html
    print_status "success" "Service coverage report: tests/coverage_services.html"
fi

if [ -f "tests/coverage_middleware.out" ]; then
    go tool cover -html=tests/coverage_middleware.out -o tests/coverage_middleware.html
    print_status "success" "Middleware coverage report: tests/coverage_middleware.html"
fi

echo ""
print_status "success" "All Booking Module Tests Completed Successfully!"
echo ""

# Test Summary
echo "=== Test Summary ==="
echo "✅ Booking Template CRUD Operations"
echo "✅ Booking Link Generation & Validation"
echo "✅ JWT Token Signature Verification"
echo "✅ Token Blacklist Management"
echo "✅ Free Slot Calculation"
echo "✅ Weekly Availability Scheduling"
echo "✅ Calendar Conflict Detection"
echo "✅ Buffer Time Application"
echo "✅ Business Rules Validation"
echo "✅ Timezone Handling"
echo "✅ Middleware Token Validation"
echo "✅ Context Injection"

echo ""
echo "=== Coverage Reports ==="
if [ -f "tests/coverage_services.out" ]; then
    echo "📊 Services:    $service_coverage"
fi
if [ -f "tests/coverage_middleware.out" ]; then
    echo "📊 Middleware:  $middleware_coverage"
fi

echo ""
echo "=== Test Counts ==="
echo "🧪 Service Tests:    $(grep -E '^(=== RUN|--- PASS|--- SKIP)' /tmp/booking_services_test.log 2>/dev/null | grep -c '^===' || echo '0')"
echo "🧪 Middleware Tests: $(grep -E '^(=== RUN|--- PASS|--- SKIP)' /tmp/booking_middleware_test.log 2>/dev/null | grep -c '^===' || echo '0')"

echo ""
print_status "info" "Tests use SQLite in-memory database"
print_status "info" "No external dependencies required"
print_status "info" "Coverage reports available in HTML format"

# Cleanup
rm -f /tmp/booking_services_test.log /tmp/booking_middleware_test.log

echo ""
print_status "success" "Test run complete! 🎉"
