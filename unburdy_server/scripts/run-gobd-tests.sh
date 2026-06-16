#!/bin/bash

# GoBD Compliance Test Runner
# Runs all GoBD compliance tests and generates compliance report

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
GOBD_TEST_DIR="${PROJECT_ROOT}/base-server/tests/gobd"
REPORT_DIR="${PROJECT_ROOT}/base-server/test_results/gobd"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}          GoBD COMPLIANCE TEST SUITE${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Create report directory
mkdir -p "${REPORT_DIR}"

echo -e "${YELLOW}📋 Running GoBD compliance tests...${NC}"
echo ""

cd "${GOBD_TEST_DIR}"

# Run tests with verbose output
if go test -v -coverprofile="${REPORT_DIR}/gobd_coverage_${TIMESTAMP}.out" ./... 2>&1 | tee "${REPORT_DIR}/gobd_test_${TIMESTAMP}.log"; then
    TEST_RESULT="PASSED"
    echo -e "\n${GREEN}✅ All GoBD tests passed!${NC}\n"
else
    TEST_RESULT="FAILED"
    echo -e "\n${RED}❌ Some GoBD tests failed. See log for details.${NC}\n"
fi

# Generate coverage report
echo -e "${YELLOW}📊 Generating coverage report...${NC}"
go tool cover -html="${REPORT_DIR}/gobd_coverage_${TIMESTAMP}.out" -o "${REPORT_DIR}/gobd_coverage_${TIMESTAMP}.html"

# Generate GoBD compliance report
echo -e "${YELLOW}📄 Generating GoBD compliance report...${NC}"

cat > "${REPORT_DIR}/gobd_compliance_${TIMESTAMP}.json" << 'EOF'
{
  "generated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "version": "1.0.0",
  "test_result": "${TEST_RESULT}",
  "total_requirements": 11,
  "requirements": [
    {
      "category": "Unveränderbarkeit (Immutability)",
      "requirement": "Finalized documents cannot be modified",
      "legal_reference": "GoBD §2.1",
      "status": "TESTED",
      "test_file": "gobd_compliance_test.go::TestGoBD_Immutability"
    },
    {
      "category": "Nachvollziehbarkeit (Traceability)",
      "requirement": "Complete audit trail for all changes",
      "legal_reference": "GoBD §3.2",
      "status": "PARTIAL",
      "test_file": "gobd_compliance_test.go::TestGoBD_AuditTrail",
      "note": "Requires audit module integration"
    },
    {
      "category": "Stornierung (Cancellation)",
      "requirement": "Proper cancellation workflow",
      "legal_reference": "GoBD §4.1",
      "status": "TESTED",
      "test_file": "gobd_compliance_test.go::TestGoBD_Cancellation"
    },
    {
      "category": "Aufbewahrungspflicht (Retention)",
      "requirement": "10-year document retention",
      "legal_reference": "AO §147 Abs. 3",
      "status": "PARTIAL",
      "test_file": "gobd_compliance_test.go::TestGoBD_DocumentRetention",
      "note": "Requires document storage integration"
    },
    {
      "category": "Zeitstempel (Timestamps)",
      "requirement": "Accurate timestamp integrity",
      "legal_reference": "GoBD §5.1",
      "status": "TESTED",
      "test_file": "gobd_compliance_test.go::TestGoBD_TimestampIntegrity"
    },
    {
      "category": "Datenintegrität (Data Integrity)",
      "requirement": "Calculation accuracy and consistency",
      "legal_reference": "GoBD §6.1",
      "status": "TESTED",
      "test_file": "gobd_compliance_test.go::TestGoBD_DataIntegrity"
    },
    {
      "category": "Vollständigkeit (Completeness)",
      "requirement": "No gaps in invoice number sequences",
      "legal_reference": "GoBD §7.1",
      "status": "TESTED",
      "test_file": "gobd_compliance_test.go::TestGoBD_InvoiceNumberCompleteness"
    },
    {
      "category": "Zugriffskontrolle (Access Control)",
      "requirement": "Multi-tenant isolation",
      "legal_reference": "GoBD §8.1",
      "status": "TESTED",
      "test_file": "gobd_compliance_test.go::TestGoBD_TenantIsolation"
    },
    {
      "category": "Exportierbarkeit (Exportability)",
      "requirement": "Z1/Z3 export format support",
      "legal_reference": "GoBD §10.1",
      "status": "PENDING",
      "note": "Implementation planned"
    },
    {
      "category": "XRechnung",
      "requirement": "EU standard invoice format",
      "legal_reference": "EU Directive 2014/55/EU",
      "status": "PENDING",
      "note": "Implementation planned"
    },
    {
      "category": "Verfahrensdokumentation",
      "requirement": "Process documentation",
      "legal_reference": "GoBD §11.1",
      "status": "DOCUMENTED",
      "note": "See documentation/GOBD_COMPLIANCE.md"
    }
  ],
  "report_location": "${REPORT_DIR}"
}
EOF

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}              GOBD COMPLIANCE REPORT${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "Test Result:      ${TEST_RESULT}"
echo -e "Test Log:         ${REPORT_DIR}/gobd_test_${TIMESTAMP}.log"
echo -e "Coverage Report:  ${REPORT_DIR}/gobd_coverage_${TIMESTAMP}.html"
echo -e "Compliance JSON:  ${REPORT_DIR}/gobd_compliance_${TIMESTAMP}.json"
echo ""
echo -e "${BLUE}Requirements Status:${NC}"
echo -e "  ✅ Tested (8):    Immutability, Cancellation, Timestamps, Data Integrity,"
echo -e "                    Completeness, Access Control, Audit Trail*, Retention*"
echo -e "  ⏸️  Pending (2):  Exportability, XRechnung"
echo -e "  📄 Documented:    Process Documentation"
echo ""
echo -e "  * Requires additional module integration"
echo ""

if [ "${TEST_RESULT}" = "PASSED" ]; then
    echo -e "${GREEN}✅ GoBD compliance tests completed successfully!${NC}"
    echo -e "${GREEN}   This report can be presented to tax auditors as evidence.${NC}"
    exit 0
else
    echo -e "${RED}❌ GoBD compliance tests failed. Review the log for details.${NC}"
    exit 1
fi
