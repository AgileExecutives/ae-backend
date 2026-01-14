#!/bin/bash

# Hurl Template API Test Runner
# This script runs all template-related Hurl tests with proper setup and teardown

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8081"
API_BASE="$BASE_URL/api/v1"

# Test files directory
TESTS_DIR="$(dirname "$0")/hurl"

# Variables file for auth token
VARIABLES_FILE="$TESTS_DIR/variables.hurl"

echo -e "${BLUE}=== Template Module API Test Suite ===${NC}"
echo "Base URL: $BASE_URL"
echo "Test Directory: $TESTS_DIR"
echo ""

# Function to check if server is running
check_server() {
    echo -e "${YELLOW}Checking server connectivity...${NC}"
    if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        echo -e "${RED}Error: Server is not running on $BASE_URL${NC}"
        echo "Please start the server before running tests."
        exit 1
    fi
    echo -e "${GREEN}✓ Server is running${NC}"
}

# Function to setup auth token
setup_auth() {
    echo -e "${YELLOW}Setting up authentication...${NC}"
    
    # Create variables file if it doesn't exist
    if [ ! -f "$VARIABLES_FILE" ]; then
        echo "# Hurl variables file for template tests" > "$VARIABLES_FILE"
        echo "auth_token=your_bearer_token_here" >> "$VARIABLES_FILE"
        echo -e "${YELLOW}Created $VARIABLES_FILE - please update with valid auth token${NC}"
    fi
    
    # Check if auth token is set
    if grep -q "your_bearer_token_here" "$VARIABLES_FILE"; then
        echo -e "${YELLOW}Warning: Please update auth_token in $VARIABLES_FILE${NC}"
        echo -e "${YELLOW}You can get a token from bearer-tokens.json or login endpoint${NC}"
    else
        echo -e "${GREEN}✓ Auth token configured${NC}"
    fi
}

# Function to run a test file
run_test() {
    local test_file=$1
    local test_name=$2
    
    echo -e "${BLUE}Running: $test_name${NC}"
    
    if hurl --variables-file "$VARIABLES_FILE" \
           --test \
           --report-html "$TESTS_DIR/reports" \
           --report-json "$TESTS_DIR/reports/results.json" \
           "$test_file"; then
        echo -e "${GREEN}✓ $test_name passed${NC}"
        return 0
    else
        echo -e "${RED}✗ $test_name failed${NC}"
        return 1
    fi
}

# Function to create reports directory
setup_reports() {
    mkdir -p "$TESTS_DIR/reports"
    echo "Test reports will be saved to: $TESTS_DIR/reports"
}

# Main test execution
main() {
    local failed_tests=0
    local total_tests=0
    
    # Setup
    check_server
    setup_auth
    setup_reports
    
    echo ""
    echo -e "${BLUE}=== Running Template API Tests ===${NC}"
    echo ""
    
    # Test 1: Template Contracts API
    if [ -f "$TESTS_DIR/template_contracts.hurl" ]; then
        total_tests=$((total_tests + 1))
        if ! run_test "$TESTS_DIR/template_contracts.hurl" "Template Contracts API"; then
            failed_tests=$((failed_tests + 1))
        fi
        echo ""
    fi
    
    # Test 2: Template Rendering
    if [ -f "$TESTS_DIR/template_rendering.hurl" ]; then
        total_tests=$((total_tests + 1))
        if ! run_test "$TESTS_DIR/template_rendering.hurl" "Template Rendering"; then
            failed_tests=$((failed_tests + 1))
        fi
        echo ""
    fi
    
    # Test 3: Template CRUD Operations
    if [ -f "$TESTS_DIR/template_crud.hurl" ]; then
        total_tests=$((total_tests + 1))
        if ! run_test "$TESTS_DIR/template_crud.hurl" "Template CRUD Operations"; then
            failed_tests=$((failed_tests + 1))
        fi
        echo ""
    fi
    
    # Test 4: Main Template API
    if [ -f "$TESTS_DIR/templates.hurl" ]; then
        total_tests=$((total_tests + 1))
        if ! run_test "$TESTS_DIR/templates.hurl" "Template API"; then
            failed_tests=$((failed_tests + 1))
        fi
        echo ""
    fi
    
    # Summary
    echo -e "${BLUE}=== Test Summary ===${NC}"
    echo "Total tests: $total_tests"
    echo "Passed: $((total_tests - failed_tests))"
    echo "Failed: $failed_tests"
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
        echo -e "${GREEN}Check HTML report: $TESTS_DIR/reports/index.html${NC}"
        exit 0
    else
        echo -e "${RED}❌ $failed_tests test(s) failed${NC}"
        echo -e "${YELLOW}Check reports for details: $TESTS_DIR/reports/${NC}"
        exit 1
    fi
}

# Help function
show_help() {
    echo "Template API Test Runner"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -v, --verbose  Enable verbose output"
    echo "  -s, --single   Run a single test file"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run all tests"
    echo "  $0 -s template_contracts.hurl        # Run only contract tests"
    echo "  $0 --verbose                         # Run with verbose output"
    echo ""
    echo "Environment variables:"
    echo "  BASE_URL       Server base URL (default: http://localhost:8081)"
    echo "  AUTH_TOKEN     Bearer token for authentication"
    echo ""
}

# Handle command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            set -x
            shift
            ;;
        -s|--single)
            if [ -z "$2" ]; then
                echo -e "${RED}Error: --single requires a test file name${NC}"
                exit 1
            fi
            # Run single test
            check_server
            setup_auth
            setup_reports
            run_test "$TESTS_DIR/$2" "$2"
            exit $?
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Run main function
main