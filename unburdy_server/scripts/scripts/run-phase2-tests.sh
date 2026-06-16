#!/bin/bash

# Phase 2 Service Test Runner
# Runs core business service tests (Email, Settings, Client Management, Booking)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPORT_DIR="${PROJECT_ROOT}/base-server/test_results/phase2"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}          PHASE 2: CORE BUSINESS SERVICE TESTS${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Create report directory
mkdir -p "${REPORT_DIR}"

# Initialize counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Test results storage
EMAIL_RESULT=""
SETTINGS_RESULT=""
CLIENT_RESULT=""

# Function to run tests for a service
run_service_tests() {
    local service_name=$1
    local test_path=$2
    local test_file=$3
    
    echo -e "${YELLOW}📦 Testing ${service_name}...${NC}"
    
    cd "${PROJECT_ROOT}/${test_path}"
    
    if [ -f "${test_file}" ]; then
        if go test -v -count=1 "${test_file}" 2>&1 | tee "${REPORT_DIR}/${service_name}_${TIMESTAMP}.log"; then
            echo -e "${GREEN}✅ ${service_name} tests passed${NC}\n"
            ((PASSED_TESTS++))
            return 0
        else
            echo -e "${RED}❌ ${service_name} tests failed${NC}\n"
            ((FAILED_TESTS++))
            return 1
        fi
    else
        echo -e "${RED}⚠️  ${service_name} test file not found: ${test_file}${NC}\n"
        ((FAILED_TESTS++))
        return 1
    fi
    
    ((TOTAL_TESTS++))
}

echo -e "${BLUE}Running Service Tests...${NC}\n"

# Email Service Tests
run_service_tests "EmailService" "base-server/modules/email/services" "email_service_test.go"
EMAIL_RESULT=$?

# Settings Service Tests  
run_service_tests "SettingsService" "base-server/pkg/settings/services" "settings_service_test.go"
SETTINGS_RESULT=$?

# Client Management Service Tests
run_service_tests "ClientService" "unburdy_server/modules/client_management/services" "client_service_test.go"
CLIENT_RESULT=$?

# Booking Service Tests
run_service_tests "BookingService" "modules/booking/services" "booking_service_test.go"
BOOKING_RESULT=$?

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}              PHASE 2 TEST SUMMARY${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Display results
if [ ${EMAIL_RESULT} -eq 0 ]; then
    echo -e "  ${GREEN}✅ EmailService: PASSED${NC}"
else
    echo -e "  ${RED}❌ EmailService: FAILED${NC}"
fi

if [ ${SETTINGS_RESULT} -eq 0 ]; then
    echo -e "  ${GREEN}✅ SettingsService: PASSED${NC}"
else
    echo -e "  ${RED}❌ SettingsService: FAILED${NC}"
fi

if [ ${CLIENT_RESULT} -eq 0 ]; then
    echo -e "  ${GREEN}✅ ClientService: PASSED${NC}"
else
    echo -e "  ${RED}❌ ClientService: FAILED${NC}"
fi

if [ ${BOOKING_RESULT} -eq 0 ]; then
    echo -e "  ${GREEN}✅ BookingService: PASSED${NC}"
else
    echo -e "  ${RED}❌ BookingService: FAILED${NC}"
fi

echo ""
echo -e "${BLUE}Statistics:${NC}"
echo -e "  Total Services:  ${TOTAL_TESTS}"
echo -e "  Passed:          ${GREEN}${PASSED_TESTS}${NC}"
echo -e "  Failed:          ${RED}${FAILED_TESTS}${NC}"
echo ""
echo -e "Reports saved to: ${REPORT_DIR}"
echo ""

# Exit with appropriate code
if [ ${FAILED_TESTS} -eq 0 ]; then
    echo -e "${GREEN}✅ Phase 2 tests completed successfully!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some Phase 2 tests failed. Review logs for details.${NC}"
    exit 1
fi
