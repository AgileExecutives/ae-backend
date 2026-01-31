# GoBD Compliance Testing

## Overview

This directory contains comprehensive compliance tests for **GoBD** (Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form sowie zum Datenzugriff) - the German principles for proper bookkeeping and record retention.

## Running the Tests

### Quick Run (from project root)
```bash
./scripts/run-gobd-compliance-tests.sh
```

### Manual Run (from invoice module)
```bash
cd modules/invoice
go test -v ./tests/gobd -run GoBD
```

### Run Specific Test
```bash
cd modules/invoice
go test -v ./tests/gobd -run TestGoBD_Rz44_Immutability
```

## Test Coverage

### ✅ Implemented Tests (8 Passing, 1 Skipped)

| Test | GoBD Rz. | Requirement | Status |
|------|----------|-------------|--------|
| TestGoBD_Rz44_Immutability_FinalizedInvoiceCannotBeModified | Rz. 44-46 | Unveränderbarkeit (Immutability) | ✅ PASS |
| TestGoBD_Rz44_Immutability_CancellationInsteadOfDeletion | Rz. 44 | Cancellation vs Deletion | ✅ PASS |
| TestGoBD_Rz58_Completeness_AllInvoiceDataIsStored | Rz. 58-60 | Vollständigkeit (Completeness) | ✅ PASS |
| TestGoBD_Rz61_Accuracy_CalculationsAreCorrect | Rz. 61-63 | Richtigkeit (Accuracy) | ✅ PASS |
| TestGoBD_Rz64_Timeliness_InvoicesAreRecordedPromptly | Rz. 64-66 | Zeitgerechtigkeit (Timeliness) | ✅ PASS |
| TestGoBD_Rz71_SequentialNumbering_InvoiceNumbersMustBeSequential | Rz. 71-72 | Sequential Numbering | ✅ PASS |
| TestGoBD_Rz122_Auditability_AllChangesAreLogged | Rz. 122-128 | Nachprüfbarkeit (Auditability) | ⊗ SKIPPED |
| TestGoBD_Rz129_DataRetention_DeletedInvoicesAreRetained | Rz. 129-136 | Aufbewahrung (Data Retention) | ✅ PASS |
| TestGoBD_TenantIsolation_InvoicesAreSeparatedByTenant | Multi-tenancy | Tenant Data Separation | ✅ PASS |

## GoBD Requirements Explained

### Rz. 44-46: Unveränderbarkeit (Immutability)
Once an invoice is finalized, it must remain immutable. Changes must be made through compensating transactions (storno/cancellation invoices).

**Tests:**
- Finalized invoices cannot be modified
- Finalized invoices cannot be deleted (only cancelled)

### Rz. 58-60: Vollständigkeit (Completeness)
All business transactions must be completely recorded with all necessary data to reconstruct the transaction.

**Tests:**
- All required invoice fields are stored (number, date, amounts, currency, client)

### Rz. 61-63: Richtigkeit (Accuracy)
All calculations and derived values must be mathematically correct and verifiable.

**Tests:**
- Invoice totals are calculated correctly
- Net + VAT = Gross
- Line item calculations are accurate

### Rz. 64-66: Zeitgerechtigkeit (Timeliness)
Business transactions must be recorded promptly with timestamps to establish chronological order.

**Tests:**
- Creation timestamps are set
- Timestamps are accurate and reasonable

### Rz. 71-72: Fortlaufende Nummerierung (Sequential Numbering)
Invoice numbers must be unique, sequential, and without gaps to ensure completeness.

**Tests:**
- Invoice numbers are assigned sequentially
- No gaps in numbering sequence

### Rz. 122-128: Nachprüfbarkeit (Auditability)
All transactions must be traceable and verifiable. An audit trail must record who made what changes, when, and why.

**Tests:**
- ⊗ SKIPPED - Requires integration with audit module
- Audit log entries for all changes
- Complete audit trail (User, Timestamp, Action, Old/New Values)

### Rz. 129-136: Aufbewahrung (Data Retention)
Business records must be retained for the legally required period (typically 10 years for invoices).

**Tests:**
- Soft delete preserves invoice data
- Deleted invoices remain queryable (for compliance)
- All invoice data retained after deletion

### Multi-Tenancy / Tenant Isolation
Different legal entities must have separate record-keeping (GoBD requirement for separate organizations).

**Tests:**
- Invoices are separated by tenant
- No cross-tenant data access
- Proper tenant filtering in queries

## Test Reports

Test reports are automatically generated in:
```
test_results/gobd/
  ├── gobd_test_YYYYMMDD_HHMMSS.log    # Detailed test output
  └── gobd_summary_YYYYMMDD_HHMMSS.txt # Summary report
```

## Implementation Notes

### ⊗ Skipped Tests

**TestGoBD_Rz122_Auditability_AllChangesAreLogged** is currently skipped because it requires:
- Integration with the audit module
- Audit log table structure
- Automatic audit logging on invoice changes

To enable this test:
1. Integrate the audit module
2. Add audit logging hooks to invoice service
3. Update the test to query audit logs
4. Remove the `t.Skip()` line

### Database

Tests use SQLite in-memory database (`:memory:`) for fast, isolated testing. This allows tests to run without external dependencies.

### Mock Entities

The tests use simplified mock entities (`Invoice`, `InvoiceLineItem`) that mirror the expected structure. These can be replaced with actual entities once the invoice module implementation is stable.

## Adding New Tests

When adding new GoBD compliance tests:

1. **Name Format:** `TestGoBD_Rz##_Category_Description`
   - Include the GoBD Rz. (paragraph) number
   - Use German category name (Unveränderbarkeit, etc.)
   - Descriptive test name

2. **Documentation:**
   - Add comprehensive comment explaining the GoBD requirement
   - Include the Rz. number reference
   - Document what the test validates

3. **Assertions:**
   - Use clear assertion messages referencing GoBD Rz.
   - Log compliance status with `t.Logf()`

4. **Update this README:**
   - Add test to coverage table
   - Document the requirement being tested

## Compliance Certification

These tests serve as:
- **Documentation** of GoBD compliance requirements
- **Validation** that the system meets GoBD standards
- **Regression prevention** to ensure compliance is maintained
- **Audit support** for German tax authority (Finanzamt) audits

## Related Documentation

- [GoBD Official Document](https://www.bundesfinanzministerium.de/Content/DE/Downloads/BMF_Schreiben/Weitere_Steuerthemen/Abgabenordnung/2019-11-28-GoBD.pdf)
- [Invoice Module Documentation](../README.md)
- [Audit Trail Documentation](../../../documentation/AUDIT_TRAIL_README.md)

## Contact

For questions about GoBD compliance or these tests, contact the development team.
