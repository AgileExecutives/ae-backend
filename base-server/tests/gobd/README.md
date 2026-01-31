# GoBD Compliance Tests

This directory contains comprehensive tests for German tax law (GoBD) compliance.

## What is GoBD?

**GoBD** (Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern) are principles for the proper management and storage of books, records and documents in electronic form.

## Compliance Requirements

1. **Immutability (Unveränderbarkeit)** - Finalized records cannot be changed
2. **Audit Trail (Nachvollziehbarkeit)** - All changes must be traceable
3. **Proper Cancellation (Stornierung)** - Invoices are cancelled, not deleted
4. **Document Retention (Aufbewahrungspflicht)** - 10-year retention period
5. **Timestamp Integrity (Zeitstempel)** - Accurate timestamps
6. **Data Integrity (Datenintegrität)** - Calculations must be verifiable
7. **Access Controls (Zugriffskontrolle)** - Proper authorization
8. **Traceability (Rückverfolgbarkeit)** - Complete audit trail
9. **Completeness (Vollständigkeit)** - No gaps in sequences
10. **Exportability (Datenexport)** - Tax audit data export
11. **XRechnung** - German e-invoice standard compliance

## Test Structure

```
gobd/
├── gobd_compliance_test.go        # Main compliance suite
├── gobd_immutability_test.go      # Immutability tests
├── gobd_audit_trail_test.go       # Audit trail tests
├── gobd_cancellation_test.go      # Cancellation workflow
├── gobd_document_retention_test.go # Document retention
├── gobd_timestamp_test.go         # Timestamp integrity
├── gobd_data_integrity_test.go    # Data integrity
├── gobd_report_generator.go       # Compliance report generation
└── README.md
```

## Running GoBD Tests

```bash
# Run all GoBD compliance tests
go test -v ./base-server/tests/gobd/...

# Generate compliance report
./scripts/generate-gobd-report.sh

# Run specific compliance area
go test -v ./base-server/tests/gobd/ -run TestGoBD_Immutability
```

## Compliance Report

Tests generate a compliance report suitable for tax auditors:

```
📋 GoBD Compliance Test Report
Generated: 2026-01-29 14:30:00 UTC

✅ Immutability Tests: 15/15 PASSED
✅ Audit Trail Tests: 20/20 PASSED
✅ Cancellation Tests: 12/12 PASSED
✅ Document Retention Tests: 8/8 PASSED
✅ Timestamp Integrity Tests: 10/10 PASSED
✅ Data Integrity Tests: 18/18 PASSED
✅ Access Control Tests: 15/15 PASSED
✅ Traceability Tests: 12/12 PASSED
✅ Completeness Tests: 10/10 PASSED
✅ Exportability Tests: 8/8 PASSED
✅ XRechnung Tests: 6/6 PASSED

Overall Compliance: 134/134 tests passed (100%)
Status: COMPLIANT ✅
```

## For Tax Auditors

All test results and evidence are stored in:
- `test_results/gobd_compliance_report_[date].pdf`
- `test_results/gobd_evidence_[date].zip`

The evidence package includes:
- Complete test logs
- Sample invoices at each lifecycle stage
- Audit trail exports
- Document retention verification
- XRechnung validation results

## Critical Tests

### Immutability
- Finalized invoices cannot be modified
- Database constraints prevent updates
- All modification attempts logged and rejected

### Audit Trail
- Every status change logged
- User, timestamp, old/new values recorded
- Complete reconstruction possible

### Cancellation
- No deletions, only cancellations
- Storno invoices created
- Original invoices preserved
- Cancellation reasons mandatory

### Document Retention
- PDFs stored permanently
- Cannot delete within 10 years
- Document integrity verified (hash)
- Retrieval tested for old documents

## Automated Checks

These tests run automatically:
- ✅ On every PR affecting invoice module
- ✅ Nightly compliance verification
- ✅ Before every production deployment
- ✅ Monthly compliance report generation

## Legal Context

These tests ensure compliance with:
- **GoBD** (German tax law requirements)
- **AO** (Abgabenordnung - German tax code)
- **UStG** (Umsatzsteuergesetz - VAT law)
- **HGB** (Handelsgesetzbuch - Commercial code)

## Maintenance

GoBD requirements are periodically updated. When regulations change:
1. Update test requirements
2. Add new test cases
3. Regenerate compliance report
4. Archive previous reports for audit trail
