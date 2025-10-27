#!/bin/bash

# AE SaaS Basic HURL Test Runner with Template Support
# Runs comprehensive API tests using HURL with unique identifiers per run

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
TEMPLATES_DIR="tests/hurl/templates"
PROCESSED_DIR="tests/hurl/processed"
RESULTS_DIR="test_results"

# Generate unique identifiers for this test run
TIMESTAMP=$(date +%s)
NANO_PART=$(date +%N | cut -c1-6)  # Get microseconds
PROCESS_ID=$$  # Current process ID
RANDOM_PART=$RANDOM
RANDOM_ID=$(echo "${TIMESTAMP}${NANO_PART}${PROCESS_ID}${RANDOM_PART}" | shasum | cut -c1-8)
UNIQUE_ID="${TIMESTAMP}_${RANDOM_ID}"
UNIQUE_USERNAME="testuser_${UNIQUE_ID}"
UNIQUE_EMAIL="test_${UNIQUE_ID}@example.com"
UNIQUE_CUSTOMER="customer_${UNIQUE_ID}"
UNIQUE_ORG="org_${UNIQUE_ID}"
UNIQUE_PASSWORD="Pass123_${RANDOM_ID}"

# Read host from config if it exists
if [ -f "$CONFIG_FILE" ]; then
    HOST=$(grep "^host" "$CONFIG_FILE" | cut -d' ' -f3 2>/dev/null || echo "$HOST")
fi
HOST=${HOST:-${TEST_HOST:-"http://localhost:8080"}}

# Create directories
mkdir -p "$RESULTS_DIR"
mkdir -p "$PROCESSED_DIR"
mkdir -p "$TEMPLATES_DIR"

echo -e "${BLUE}🚀 Starting AE SaaS Basic HURL Tests with Templating${NC}"
echo -e "${BLUE}Host: ${HOST}${NC}"
echo -e "${BLUE}Results Directory: ${RESULTS_DIR}${NC}"
echo -e "${BLUE}Unique Test ID: ${UNIQUE_ID}${NC}"
echo -e "${BLUE}Test Username: ${UNIQUE_USERNAME}${NC}"
echo -e "${BLUE}Test Email: ${UNIQUE_EMAIL}${NC}"
echo ""

# Function to process template files
process_template() {
    local template_file="$1"
    local output_file="$2"
    
    sed -e "s|{{UNIQUE_ID}}|${UNIQUE_ID}|g" \
        -e "s|{{UNIQUE_USERNAME}}|${UNIQUE_USERNAME}|g" \
        -e "s|{{UNIQUE_EMAIL}}|${UNIQUE_EMAIL}|g" \
        -e "s|{{UNIQUE_CUSTOMER}}|${UNIQUE_CUSTOMER}|g" \
        -e "s|{{UNIQUE_ORG}}|${UNIQUE_ORG}|g" \
        -e "s|{{UNIQUE_PASSWORD}}|${UNIQUE_PASSWORD}|g" \
        -e "s|{{HOST}}|${HOST}|g" \
        -e "s|{{host}}|${HOST}|g" \
        "$template_file" > "$output_file"
}

# Function to check server availability
check_server() {
    echo -e "${YELLOW}🔍 Checking server availability...${NC}"
    if curl -s --max-time 5 "${HOST}/api/v1/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is running${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not responding at ${HOST}${NC}"
        echo -e "${YELLOW}💡 Make sure the server is running with: make run${NC}"
        return 1
    fi
}

# Function to run a single test
run_test() {
    local test_file="$1"
    local test_name=$(basename "$test_file" .hurl)
    
    echo -e "${BLUE}🧪 Running ${test_name}.hurl...${NC}"
    
    # Check if this is a template file or regular file
    local source_file="$PROCESSED_DIR/${test_name}.hurl"
    
    if [ -f "$TEMPLATES_DIR/${test_name}.hurl" ]; then
        # Process template
        process_template "$TEMPLATES_DIR/${test_name}.hurl" "$source_file"
    else
        # Use regular file and process it for variables
        process_template "$test_file" "$source_file"
    fi
    
    # Run the test
    if hurl "$source_file" \
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
                echo -e "${RED}📝 Error details:${NC}"
                echo "\"No response\""
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

# Run tests - only use templates, skip tests without templates
for template in "$TEMPLATES_DIR"/*.hurl; do
    if [ -f "$template" ]; then
        if run_test "$template"; then
            ((passed_tests++))
        else
            ((failed_tests++))
        fi
    fi
done

# Print summary
echo -e "${BLUE}📊 Test Summary${NC}"
echo -e "${BLUE}===============${NC}"
echo -e "${BLUE}Total Tests: $((passed_tests + failed_tests))${NC}"
echo -e "${GREEN}Passed: ${passed_tests}${NC}"
echo -e "${RED}Failed: ${failed_tests}${NC}"

if [ $failed_tests -gt 0 ]; then
    echo -e "${RED}❌ ${failed_tests} test(s) failed${NC}"
    echo -e "${YELLOW}💡 Check individual result files in ${RESULTS_DIR}/ for details${NC}"
    exit 1
else
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
fi