#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Running tests with coverage...${NC}"
echo ""

# Create test results directory
mkdir -p test_results

# Run tests with coverage
echo -e "${YELLOW}Running all tests...${NC}"
go test -v -race -coverprofile=test_results/coverage.out -covermode=atomic ./... 2>&1 | tee test_results/test_output.log

# Check if tests passed
if [ ${PIPESTATUS[0]} -ne 0 ]; then
    echo -e "${RED}❌ Tests failed!${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ All tests passed!${NC}"
echo ""

# Generate HTML coverage report
echo -e "${YELLOW}Generating coverage reports...${NC}"
go tool cover -html=test_results/coverage.out -o test_results/coverage.html

# Generate text coverage report
go tool cover -func=test_results/coverage.out > test_results/coverage.txt

# Extract total coverage
TOTAL_COVERAGE=$(go tool cover -func=test_results/coverage.out | grep total | awk '{print $3}')

echo ""
echo -e "${BLUE}📊 Coverage Summary${NC}"
echo -e "${GREEN}Total Coverage: ${TOTAL_COVERAGE}${NC}"
echo ""

# Show top 10 most tested files
echo -e "${BLUE}Top 10 Most Tested Files:${NC}"
go tool cover -func=test_results/coverage.out | grep -v "total" | sort -k3 -rn | head -10

echo ""
echo -e "${BLUE}Top 10 Least Tested Files:${NC}"
go tool cover -func=test_results/coverage.out | grep -v "total" | grep -v "100.0%" | sort -k3 -n | head -10

echo ""
echo -e "${GREEN}✅ Coverage reports generated:${NC}"
echo -e "  📄 HTML Report: ${BLUE}test_results/coverage.html${NC}"
echo -e "  📄 Text Report: ${BLUE}test_results/coverage.txt${NC}"
echo -e "  📄 Coverage Data: ${BLUE}test_results/coverage.out${NC}"
echo -e "  📄 Test Output: ${BLUE}test_results/test_output.log${NC}"
echo ""

# Open HTML report if on macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo -e "${YELLOW}Opening coverage report in browser...${NC}"
    open test_results/coverage.html
fi

echo -e "${GREEN}✅ Done!${NC}"
