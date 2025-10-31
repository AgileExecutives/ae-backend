#!/bin/bash

# Unburdy Server HURL Test Runner
# Runs comprehensive API tests for client management module

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
HOST="http://localhost:8080"
CONFIG_FILE="tests/hurl/hurl.config"
HURL_DIR="tests/hurl"
RESULTS_DIR="test_results"

# Generate unique identifiers for this test run
TIMESTAMP=$(date +%s)
NANO_PART=$(date +%N | cut -c1-6)  # Get microseconds
PROCESS_ID=$$  # Current process ID
RANDOM_PART=$RANDOM
RANDOM_ID=$(echo "${TIMESTAMP}${NANO_PART}${PROCESS_ID}${RANDOM_PART}" | shasum | cut -c1-8)
UNIQUE_ID="${TIMESTAMP}_${RANDOM_ID}"

# Read host from config if it exists
if [ -f "$CONFIG_FILE" ]; then
    HOST=$(grep "^host" "$CONFIG_FILE" | cut -d' ' -f3 2>/dev/null || echo "$HOST")
fi
HOST=${HOST:-${TEST_HOST:-"http://localhost:8080"}}

# Create directories
mkdir -p "$RESULTS_DIR"

echo -e "${BLUE}🚀 Starting Unburdy Server Client Management Tests${NC}"
echo -e "${BLUE}Host: ${HOST}${NC}"
echo -e "${BLUE}Results Directory: ${RESULTS_DIR}${NC}"
echo -e "${BLUE}Unique Test ID: ${UNIQUE_ID}${NC}"
echo ""

# Function to check server availability
check_server() {
    echo -e "${YELLOW}🔍 Checking server availability...${NC}"
    if curl -s --max-time 5 "${HOST}/api/v1/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is running${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not responding at ${HOST}${NC}"
        echo -e "${YELLOW}💡 Make sure the server is running with: ./tmp/test or air${NC}"
        return 1
    fi
}

# Function to run a single test
run_test() {
    local test_file="$1"
    local test_name=$(basename "$test_file" .hurl)
    
    echo -e "${BLUE}🧪 Running ${test_name}.hurl...${NC}"
    
    # Run the test
    if hurl "$test_file" \
        --variable "host=${HOST}" \
        --test \
        --json > "$RESULTS_DIR/${test_name}.json" 2>/dev/null; then
        echo -e "${GREEN}✅ ${test_name}.hurl passed${NC}"
        return 0
    else
        echo -e "${RED}❌ ${test_name}.hurl failed${NC}"
        
        # Show error details if available
        if [ -f "$RESULTS_DIR/${test_name}.json" ]; then
            local error_msg=$(jq -r '.entries[0].asserts[]? | select(.success == false) | .message' "$RESULTS_DIR/${test_name}.json" 2>/dev/null | head -1)
            if [ -n "$error_msg" ] && [ "$error_msg" != "null" ]; then
                echo -e "${RED}📝 Error details:${NC}"
                echo "\"$error_msg\""
            else
                echo -e "${RED}📝 Run with --verbose for more details${NC}"
                # Optionally show the full output for debugging
                if [ "$VERBOSE" = "1" ]; then
                    echo -e "${RED}Full output:${NC}"
                    cat "$RESULTS_DIR/${test_name}.json" 2>/dev/null || echo "No output file"
                fi
            fi
        fi
        return 1
    fi
}

# Check server availability first
if ! check_server; then
    exit 1
fi

echo ""

# Track test results
passed_tests=0
failed_tests=0

# Define test order (ensure dependencies are met)
test_files=(
    "01_auth_setup.hurl"
    "02_clients.hurl" 
    "03_cost_providers.hurl"
    "04_client_cost_provider_integration.hurl"
)

# Run tests in order
for test_name in "${test_files[@]}"; do
    test_file="$HURL_DIR/$test_name"
    if [ -f "$test_file" ]; then
        if run_test "$test_file"; then
            ((passed_tests++))
        else
            ((failed_tests++))
            
            # Stop on first failure for ordered tests
            if [ "$CONTINUE_ON_FAILURE" != "1" ]; then
                echo -e "${RED}⏹️  Stopping tests due to failure (set CONTINUE_ON_FAILURE=1 to continue)${NC}"
                break
            fi
        fi
    else
        echo -e "${YELLOW}⚠️  Test file not found: $test_file${NC}"
    fi
done

echo ""

# Print summary
echo -e "${BLUE}📊 Test Summary${NC}"
echo -e "${BLUE}===============${NC}"
echo -e "${BLUE}Total Tests: $((passed_tests + failed_tests))${NC}"
echo -e "${GREEN}Passed: ${passed_tests}${NC}"
echo -e "${RED}Failed: ${failed_tests}${NC}"

if [ $failed_tests -gt 0 ]; then
    echo -e "${RED}❌ ${failed_tests} test(s) failed${NC}"
    echo -e "${YELLOW}💡 Check individual result files in ${RESULTS_DIR}/ for details${NC}"
    echo -e "${YELLOW}💡 Run with VERBOSE=1 ./run-client-tests.sh for more details${NC}"
    exit 1
else
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
fi