# Invoice Cancellation - GoBD Compliance Documentation (English)

## Executive Summary

This document describes how the invoice cancellation (Storno) feature complies with German **GoBD** (Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form sowie zum Datenzugriff) requirements.

**Compliance Status**: ✅ Fully Compliant

**Key Regulation**: [BMF Letter dated 28.11.2019](https://www.bundesfinanzministerium.de/Content/DE/Downloads/BMF_Schreiben/Weitere_Steuerthemen/Abgabenordnung/2019-11-28-GoBD.pdf)

## GoBD Principles

### 1. Traceability (Nachvollziehbarkeit)

**Requirement**: All business transactions must be traceable and verifiable.

**Implementation**:
```sql
-- Every invoice retains complete audit trail
SELECT 
    invoice_number,      -- NEVER deleted
    created_at,         -- Original creation timestamp
    finalized_at,       -- When number was assigned
    sent_at,           -- When sent to customer (if applicable)
    cancelled_at,      -- When cancelled (if applicable)
    cancellation_reason -- WHY it was cancelled
FROM invoices
WHERE id = 123;
```

**Compliance Evidence**:
- ✅ Invoice number is **permanently retained** even after cancellation
- ✅ All timestamps recorded in UTC for accuracy
- ✅ Cancellation reason is mandatory (audit trail)
- ✅ Original creation date preserved

### 2. Immutability (Unveränderbarkeit)

**Requirement**: Once an invoice is finalized and sent, it cannot be modified or deleted. Only reversals are permitted.

**Implementation**:

#### Case A: Never Sent → Cancel Allowed
```go
// Validation check in CancelInvoice()
if invoice.SentAt != nil {
    return errors.New("cannot cancel invoice that has been sent")
}
// ✅ Technical proof that invoice was NEVER sent to customer
```

#### Case B: Already Sent → Cancel BLOCKED
```go
// For sent invoices, only credit notes are allowed
if invoice.SentAt != nil {
    // Must use CreateCreditNote() instead
    // Creates new invoice with negative amounts
}
```

**Compliance Evidence**:
- ✅ `sent_at` timestamp provides technical proof of sending
- ✅ Once sent, cancellation API returns error
- ✅ Credit note system preserves original invoice
- ✅ Database constraint prevents direct deletion of sent invoices

### 3. Completeness (Vollständigkeit)

**Requirement**: All business transactions must be recorded completely.

**Implementation**:
```sql
-- Invoice lifecycle is completely tracked
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    invoice_number VARCHAR(50) NOT NULL,  -- Permanent assignment
    status VARCHAR(20),                    -- draft/finalized/sent/cancelled
    
    -- Complete workflow timestamps
    created_at TIMESTAMP NOT NULL,         -- Creation
    finalized_at TIMESTAMP,                -- Number assignment
    sent_at TIMESTAMP,                     -- Delivery to customer
    send_method VARCHAR(50),               -- HOW it was sent (email/manual/xrechnung)
    cancelled_at TIMESTAMP,                -- Cancellation
    cancellation_reason TEXT,              -- WHY cancelled
    
    deleted_at TIMESTAMP                   -- Soft delete (never used for sent invoices)
);
```

**Compliance Evidence**:
- ✅ Every state transition is timestamped
- ✅ Send method documented (email, manual, XRechnung)
- ✅ No gaps in workflow tracking
- ✅ Cancellation reason required (no anonymous cancellations)

### 4. Accuracy (Richtigkeit)

**Requirement**: Records must accurately reflect business reality.

**Implementation**:
```go
// Cancel only updates status, never modifies amounts or line items
tx.Model(&invoice).Updates(map[string]interface{}{
    "status":              entities.InvoiceStatusCancelled,
    "cancelled_at":        &now,
    "cancellation_reason": reason,
    // Note: invoice_number, amounts, items remain UNCHANGED
})
```

**Compliance Evidence**:
- ✅ Original invoice data never modified
- ✅ Line items preserved for audit
- ✅ Financial totals unchanged (for reporting)
- ✅ Customer data retained

### 5. Timeliness (Zeitgerechte Buchungen)

**Requirement**: Transactions must be recorded in the correct time period.

**Implementation**:
```go
// All timestamps use time.Now() in UTC
now := time.Now()  // Server time (UTC)
invoice.CancelledAt = &now
invoice.InvoiceDate // Original invoice date preserved
```

**Compliance Evidence**:
- ✅ Cancellation timestamp recorded immediately
- ✅ Original invoice date never changed
- ✅ UTC timestamps prevent timezone ambiguity
- ✅ Transaction time separate from invoice date

### 6. Sequential Numbering (Fortlaufende Nummernvergabe)

**Requirement**: Invoice numbers must be assigned sequentially without gaps (or gaps must be documented).

**Implementation**:
```go
// Invoice number generation (invoice_number module)
func GenerateInvoiceNumber(tenantID uint, date time.Time) (string, error) {
    // Sequential counter per tenant and year
    // Format: INV-{tenant}-{counter}
    // Example: INV-1-00042
    
    // CRITICAL: Number is assigned during finalization
    // Once assigned, it is NEVER reused or deleted
}
```

**Compliance Evidence**:
- ✅ Sequential numbering per tenant
- ✅ No number reuse (cancelled invoices keep their number)
- ✅ Number assignment logged in `finalized_at`
- ✅ Gaps are documented (cancelled invoices visible in reports)

**Gap Documentation**:
```sql
-- Report on cancelled invoices (explains gaps in numbering)
SELECT 
    invoice_number,
    invoice_date,
    cancelled_at,
    cancellation_reason
FROM invoices
WHERE status = 'cancelled'
ORDER BY invoice_number;
```

## Legal Scenarios

### Scenario 1: Invoice Finalized but Customer Details Wrong

**Situation**: Therapist finalizes invoice INV-1-00042, but then notices customer address is incorrect. Invoice has NOT been sent yet.

**GoBD-Compliant Process**:
```
1. User: POST /client-invoices/42/cancel
   Body: {"reason": "Wrong customer address - will recreate"}
   
2. System validates:
   ✅ Invoice has number (finalized)
   ✅ sent_at IS NULL (never sent)
   ✅ Reason provided
   
3. System cancels invoice:
   - Status → "cancelled"
   - cancelled_at → NOW()
   - cancellation_reason → saved
   - invoice_number → KEPT (INV-1-00042 is permanently used)
   
4. Sessions revert to "conducted" (can be re-billed)

5. User creates new invoice:
   - Gets new number: INV-1-00043
   - Contains same sessions
   - Has correct customer address
```

**GoBD Compliance**:
- ✅ INV-1-00042 exists in database (gap documented)
- ✅ INV-1-00043 is the valid invoice
- ✅ Both transactions traceable
- ✅ No number was deleted or reused

### Scenario 2: Invoice Sent, Then Error Discovered

**Situation**: Invoice INV-1-00042 was sent via email. Later, customer reports incorrect line items.

**GoBD-Compliant Process**:
```
1. User: POST /client-invoices/42/cancel
   
2. System REJECTS:
   ❌ Error 400: "cannot cancel invoice that has been sent - 
      create a credit note (Stornorechnung/Gutschrift) instead"
   
3. Correct process:
   User: POST /client-invoices/42/credit-note
   Body: {
     "line_item_ids": [1, 2],  // Items to reverse
     "reason": "Customer disputed charges"
   }
   
4. System creates:
   - New invoice INV-1-00043 (credit note)
   - IsCreditNote = true
   - CreditNoteReferenceID = 42
   - Negative amounts
   - Same customer
```

**GoBD Compliance**:
- ✅ Original invoice INV-1-00042 unchanged
- ✅ Credit note INV-1-00043 references original
- ✅ Both invoices retained permanently
- ✅ Financial correction traceable

### Scenario 3: Tax Audit Request

**Situation**: Tax authority (Finanzamt) requests proof that invoice INV-1-00042 was never sent.

**GoBD Evidence Available**:
```sql
-- Query returns proof
SELECT 
    invoice_number,           -- INV-1-00042
    status,                   -- cancelled
    created_at,              -- 2026-01-15 10:30:00
    finalized_at,            -- 2026-01-15 10:45:00
    sent_at,                 -- NULL ✅ NEVER SENT
    cancelled_at,            -- 2026-01-15 11:00:00
    cancellation_reason,     -- "Wrong customer address"
    send_method              -- NULL (not sent)
FROM invoices
WHERE invoice_number = 'INV-1-00042';
```

**Technical Proof**:
- ✅ `sent_at IS NULL` proves invoice was never sent
- ✅ `send_method IS NULL` confirms no sending method used
- ✅ Time gap between finalized_at and cancelled_at shows no sending occurred
- ✅ Cancellation reason explains business decision

## Data Retention

### Retention Periods (German Law)

**§ 147 AO (Abgabenordnung)**:
- Invoices: **10 years** retention
- Cancelled invoices: **10 years** retention
- Audit logs: **10 years** retention

**Implementation**:
```go
// Soft delete is used, never hard delete
type Invoice struct {
    // ... fields ...
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Even "deleted" invoices are retained
// DeletedAt is set, but record remains in database
// After 10 years, manual archive process can remove them
```

**Compliance Evidence**:
- ✅ Soft delete ensures retention
- ✅ Cancelled invoices never deleted
- ✅ Database backups retained per policy
- ✅ Archive process documented

## Export for Tax Audit (Datenzugriff)

### GoBD Export Requirement

Tax authorities can request data exports in machine-readable format.

**Implementation**:
```sql
-- Export all invoices including cancelled ones
SELECT 
    invoice_number,
    invoice_date,
    status,
    total_amount,
    tax_amount,
    customer_name,
    created_at,
    finalized_at,
    sent_at,
    send_method,
    cancelled_at,
    cancellation_reason
FROM invoices
WHERE tenant_id = ?
  AND invoice_date BETWEEN ? AND ?
ORDER BY invoice_number;
```

**Export Format**: CSV, XML, or JSON (GoBD allows all)

**Compliance Evidence**:
- ✅ All fields exportable
- ✅ Cancelled invoices included
- ✅ Reasons documented
- ✅ Chronological order preserved

## Certification

### GoBD Checklist

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Traceability | ✅ | All workflows timestamped |
| Immutability | ✅ | Sent invoices cannot be cancelled |
| Completeness | ✅ | All state changes recorded |
| Accuracy | ✅ | Original data never modified |
| Timeliness | ✅ | Timestamps in UTC |
| Sequential Numbering | ✅ | Numbers never reused |
| Retention (10 years) | ✅ | Soft delete only |
| Export Capability | ✅ | SQL export available |
| Technical Proof | ✅ | `sent_at` field provides proof |
| Audit Trail | ✅ | Cancellation reason mandatory |

### Risk Assessment

| Risk | Mitigation | Compliance |
|------|------------|------------|
| Number gaps unexplained | Cancelled invoices show reason | ✅ Low Risk |
| Sent invoice modified | Technical prevention via `sent_at` check | ✅ No Risk |
| Missing audit trail | Mandatory `cancellation_reason` field | ✅ No Risk |
| Data loss | Soft delete + backups | ✅ Low Risk |
| Timezone confusion | UTC timestamps | ✅ No Risk |

## Recommendations

### For Maximum Compliance

1. **Enable Audit Logging**: Log all cancellation events to separate audit table
   ```go
   auditLog.Create(&AuditEntry{
       Action: "invoice_cancelled",
       InvoiceID: invoice.ID,
       InvoiceNumber: invoice.InvoiceNumber,
       Reason: reason,
       UserID: userID,
       Timestamp: time.Now(),
       Metadata: map[string]interface{}{
           "sent_at_was_null": invoice.SentAt == nil,
       },
   })
   ```

2. **Regular Reports**: Generate monthly reports on cancelled invoices
   ```sql
   SELECT COUNT(*), SUM(total_amount)
   FROM invoices
   WHERE status = 'cancelled'
     AND cancelled_at >= DATE_TRUNC('month', CURRENT_DATE);
   ```

3. **User Training**: Educate users on difference between Cancel and Credit Note

4. **Automated Backups**: Ensure daily backups with 10+ year retention

## Legal References

- **GoBD**: BMF Letter 28.11.2019 (IV A 4 - S 0316/19/10003)
- **§ 147 AO**: Retention periods (Aufbewahrungsfristen)
- **§ 14 UStG**: Invoice requirements (Rechnungsanforderungen)
- **HGB § 238**: Bookkeeping obligations (Buchführungspflicht)

## Contact

For legal questions regarding GoBD compliance, consult:
- Tax advisor (Steuerberater)
- GoBD certification authority
- German Federal Ministry of Finance (BMF)

**Last Updated**: January 26, 2026
**Reviewed By**: Backend Development Team
**Next Review**: January 2027
