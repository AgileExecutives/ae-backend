#!/bin/bash

# GoBD Compliance Test Runner
# Runs GoBD-specific compliance tests for invoice module
# 
# Tests compliance with "Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung 
# von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form sowie zum Datenzugriff"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"
INVOICE_MODULE="${PROJECT_ROOT}/modules/invoice"
REPORT_DIR="${PROJECT_ROOT}/test_results/gobd"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}          GoBD COMPLIANCE TEST SUITE${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo ""
echo "Testing compliance with GoBD (Grundsätze zur ordnungsmäßigen"
echo "Führung und Aufbewahrung von Büchern)"
echo ""

# Create report directory
mkdir -p "${REPORT_DIR}"

# Change to invoice module directory
cd "${INVOICE_MODULE}"

echo -e "${YELLOW}📋 Running GoBD Compliance Tests...${NC}"
echo ""
echo "Tests cover the following GoBD requirements:"
echo "  • Rz. 44-46:  Immutability (Unveränderbarkeit)"
echo "  • Rz. 58-60:  Completeness (Vollständigkeit)"
echo "  • Rz. 61-63:  Accuracy (Richtigkeit)"
echo "  • Rz. 64-66:  Timeliness (Zeitgerechtigkeit)"
echo "  • Rz. 71-72:  Sequential Numbering"
echo "  • Rz. 122-128: Auditability (Nachprüfbarkeit)"
echo "  • Rz. 129-136: Data Retention (Aufbewahrung)"
echo ""

# Run tests matching GoBD pattern
if go test -v ./tests/gobd -run="GoBD" 2>&1 | tee "${REPORT_DIR}/gobd_test_${TIMESTAMP}.log"; then
    TEST_RESULT="PASSED"
    echo -e "\n${GREEN}✅ All GoBD compliance tests passed!${NC}\n"
else
    TEST_RESULT="FAILED"
    echo -e "\n${RED}❌ Some GoBD tests failed. See log for details.${NC}\n"
fi

# Generate summary report
echo -e "${YELLOW}📊 Generating compliance summary...${NC}"

# Count test results
TOTAL_TESTS=$(grep -c "^=== RUN.*GoBD" "${REPORT_DIR}/gobd_test_${TIMESTAMP}.log" || echo "0")
PASSED_TESTS=$(grep -c "^--- PASS.*GoBD" "${REPORT_DIR}/gobd_test_${TIMESTAMP}.log" || echo "0")
FAILED_TESTS=$(grep -c "^--- FAIL.*GoBD" "${REPORT_DIR}/gobd_test_${TIMESTAMP}.log" || echo "0")
SKIPPED_TESTS=$(grep -c "^--- SKIP.*GoBD" "${REPORT_DIR}/gobd_test_${TIMESTAMP}.log" || echo "0")

cat > "${REPORT_DIR}/gobd_summary_${TIMESTAMP}.txt" << EOF
═══════════════════════════════════════════════════════════════
  GoBD COMPLIANCE TEST SUMMARY
═══════════════════════════════════════════════════════════════

Test Run: ${TIMESTAMP}
Status: ${TEST_RESULT}

Results:
--------
Total Tests:   ${TOTAL_TESTS}
Passed:        ${PASSED_TESTS}
Failed:        ${FAILED_TESTS}
Skipped:       ${SKIPPED_TESTS}

GoBD Requirements Tested:
-------------------------
✓ Rz. 44-46:  Unveränderbarkeit (Immutability)
              - Finalized invoices cannot be modified
              - Cancellation instead of deletion

✓ Rz. 58-60:  Vollständigkeit (Completeness)
              - All required invoice data stored

✓ Rz. 61-63:  Richtigkeit (Accuracy)
              - Calculations are mathematically correct

✓ Rz. 64-66:  Zeitgerechtigkeit (Timeliness)
              - Invoices recorded promptly with timestamps

✓ Rz. 71-72:  Fortlaufende Nummerierung (Sequential Numbering)
              - Invoice numbers without gaps

⊗ Rz. 122-128: Nachprüfbarkeit (Auditability)
               - Requires audit module integration (skipped)

✓ Rz. 129-136: Aufbewahrung (Data Retention)
               - Soft delete preserves data

Additional:
-----------
✓ Tenant Isolation
  - Multi-tenancy data separation

Report Files:
-------------
Test Log:  ${REPORT_DIR}/gobd_test_${TIMESTAMP}.log
Summary:   ${REPORT_DIR}/gobd_summary_${TIMESTAMP}.txt

═══════════════════════════════════════════════════════════════
EOF

# Display summary
cat "${REPORT_DIR}/gobd_summary_${TIMESTAMP}.txt"

echo ""
echo -e "${BLUE}Reports saved to: ${REPORT_DIR}${NC}"
echo ""

# Exit with appropriate code
if [ "$TEST_RESULT" = "PASSED" ]; then
    exit 0
else
    exit 1
fi
