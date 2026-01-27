# Invoice Cancellation (Storno) - Developer Documentation

## Overview

The invoice cancellation feature implements GoBD-compliant cancellation of invoices that have been finalized but **never sent** to clients. This distinction is crucial for German accounting compliance.

**Status**: ✅ Implemented and tested (January 2026)

## Architecture

### Key Principle: Two Cancellation Scenarios

```
┌─────────────────────────────────────────────────────────────┐
│                   Invoice Lifecycle                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Draft → Finalized → Sent → Paid/Overdue                   │
│            │          │                                      │
│            │          └─→ Credit Note (Stornorechnung)      │
│            │             (for sent invoices)                │
│            │                                                 │
│            └─→ CANCEL (Storno)                             │
│               (only if never sent!)                         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Decision Logic

| Scenario | Invoice Status | sent_at | Action | API Endpoint |
|----------|---------------|---------|---------|-------------|
| A | Finalized | NULL | **Cancel** - Reverts sessions, keeps number | `POST /client-invoices/:id/cancel` |
| B | Sent/Paid/Overdue | NOT NULL | **Credit Note** - Creates reversal invoice | `POST /client-invoices/:id/credit-note` |
| C | Draft | NULL | **Delete** - Removes completely (no number assigned) | `DELETE /client-invoices/:id` |

## Database Schema

### New Fields (entities/invoice.go)

```go
type Invoice struct {
    // ... existing fields ...
    
    // Workflow tracking
    SentAt         *time.Time `json:"sent_at,omitempty"`
    SendMethod     string     `gorm:"size:50" json:"send_method,omitempty"` // email, manual, xrechnung
    FinalizedAt    *time.Time `json:"finalized_at,omitempty"`
    CancelledAt    *time.Time `json:"cancelled_at,omitempty"`
    CancellationReason string `gorm:"type:text" json:"cancellation_reason,omitempty"`
}
```

### Migration

```sql
-- Add tracking fields
ALTER TABLE invoices ADD COLUMN sent_at TIMESTAMP NULL;
ALTER TABLE invoices ADD COLUMN send_method VARCHAR(50) NULL;
ALTER TABLE invoices ADD COLUMN cancelled_at TIMESTAMP NULL;
ALTER TABLE invoices ADD COLUMN cancellation_reason TEXT NULL;

-- Indexes for performance
CREATE INDEX idx_invoices_sent_at ON invoices(sent_at);
CREATE INDEX idx_invoices_cancelled_at ON invoices(cancelled_at);
```

**Migration File**: `migrations/add_cancellation_fields_to_invoices.sql`

## API Endpoints

### 1. Cancel Invoice (Never Sent)

**Endpoint**: `POST /api/v1/client-invoices/:id/cancel`

**Authentication**: Bearer Token required

**Request Body**:
```json
{
  "reason": "Fehlerhafte Positionen – Rechnung wird neu erstellt"
}
```

**Validation**:
- `reason` is required (min 10 characters recommended for audit trail)

**Success Response** (200 OK):
```json
{
  "success": true,
  "message": "Invoice cancelled successfully. Sessions and extra efforts reverted to unbilled status."
}
```

**Error Responses**:

| Status | Error | Reason |
|--------|-------|--------|
| 400 | "cannot cancel invoice that has been sent" | Invoice was already sent (`sent_at IS NOT NULL`) |
| 400 | "invoice has no number - use DeleteInvoice" | Draft invoice without finalization |
| 400 | "invoice is already cancelled" | Invoice status is already 'cancelled' |
| 404 | "invoice not found" | Invoice doesn't exist or wrong tenant |

**Example (curl)**:
```bash
curl -X POST https://api.example.com/api/v1/client-invoices/123/cancel \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Incorrect line items - invoice will be recreated"}'
```

### 2. Mark Invoice as Sent

**Endpoint**: `POST /api/v1/client-invoices/:id/mark-sent`

**Request Body** (Updated):
```json
{
  "send_method": "email"  // REQUIRED: email, manual, or xrechnung
}
```

**Behavior**:
- Sets `sent_at` to current timestamp
- Sets `send_method` to provided value
- Changes status from `finalized` → `sent`
- **Prevents future cancellation** (only credit notes allowed)

## Service Layer

### CancelInvoice Method

**Location**: `services/invoice_service.go`

**Signature**:
```go
func (s *InvoiceService) CancelInvoice(id, tenantID, userID uint, reason string) error
```

**Validation Logic**:
```go
// 1. Invoice must exist and belong to tenant
// 2. Must have invoice number (finalized)
if invoice.InvoiceNumber == "" {
    return errors.New("invoice has no number - use DeleteInvoice for draft invoices without numbers")
}

// 3. Must NOT have been sent (GoBD requirement)
if invoice.SentAt != nil {
    return errors.New("cannot cancel invoice that has been sent - create a credit note (Stornorechnung/Gutschrift) instead")
}

// 4. Cannot cancel twice
if invoice.Status == entities.InvoiceStatusCancelled {
    return errors.New("invoice is already cancelled")
}
```

**Transaction Flow**:
```go
tx.Begin()
├── Fetch ClientInvoices
├── Collect Session IDs & Extra Effort IDs
├── Revert Sessions → "conducted"
├── Revert Extra Efforts → "unbilled"
├── Hard Delete ClientInvoices
├── Update Invoice:
│   ├── status = "cancelled"
│   ├── cancelled_at = NOW()
│   └── cancellation_reason = reason
└── tx.Commit()
```

**Critical**: Invoice number is **NEVER** deleted or modified. It remains permanently in the database for audit purposes.

## Frontend Integration

### Conditional UI Display

```typescript
interface Invoice {
  status: 'draft' | 'finalized' | 'sent' | 'paid' | 'overdue' | 'cancelled';
  invoice_number: string;
  sent_at?: string;
  cancelled_at?: string;
  cancellation_reason?: string;
}

// Show Cancel button only for finalized invoices that were never sent
function showCancelButton(invoice: Invoice): boolean {
  return invoice.status === 'finalized' 
      && invoice.sent_at === null 
      && invoice.invoice_number !== '';
}

// For sent invoices, show Credit Note button
function showCreditNoteButton(invoice: Invoice): boolean {
  return invoice.sent_at !== null 
      && ['sent', 'paid', 'overdue'].includes(invoice.status);
}
```

### Cancel Dialog Example

```typescript
async function cancelInvoice(invoiceId: number) {
  const reason = await showReasonDialog({
    title: 'Cancel Invoice',
    message: 'This invoice has not been sent yet. You can cancel it and the sessions will be available for billing again.',
    placeholder: 'Enter reason for cancellation (required for audit trail)',
    minLength: 10
  });
  
  if (!reason) return;
  
  try {
    await api.post(`/client-invoices/${invoiceId}/cancel`, { reason });
    toast.success('Invoice cancelled. Sessions are now billable again.');
    refreshInvoiceList();
  } catch (error) {
    if (error.response?.data?.message?.includes('been sent')) {
      toast.error('This invoice was already sent. Please create a credit note instead.');
    } else {
      toast.error(error.response?.data?.message || 'Failed to cancel invoice');
    }
  }
}
```

## Testing

### Unit Tests

**Location**: `tests/invoice_service_test.go`

**Test Coverage**:
- ✅ `TestCancelInvoice_Success` - Happy path
- ✅ `TestCancelInvoice_AlreadySent` - Rejects sent invoices
- ✅ `TestCancelInvoice_NoNumber` - Rejects drafts without numbers
- ✅ `TestCancelInvoice_AlreadyCancelled` - Prevents double cancellation

**Run tests**:
```bash
cd modules/client_management
go test ./tests/... -run TestCancelInvoice -v
```

### Manual Testing Workflow

1. **Setup**: Create invoice with sessions
   ```bash
   POST /client-invoices/from-sessions
   {
     "unbilledClient": {...},
     "generationParameters": {...}
   }
   ```

2. **Finalize**: Give it an invoice number
   ```bash
   POST /client-invoices/:id/finalize
   ```

3. **Verify**: Check invoice has number but no sent_at
   ```bash
   GET /client-invoices/:id
   # Response should have invoice_number but sent_at: null
   ```

4. **Cancel**: Attempt cancellation
   ```bash
   POST /client-invoices/:id/cancel
   {"reason": "Test cancellation"}
   ```

5. **Verify Sessions**: Check sessions reverted to "conducted"
   ```bash
   GET /clients/:clientId/sessions
   # Sessions should show status: "conducted"
   ```

6. **Try Cancel Again**: Should fail
   ```bash
   POST /client-invoices/:id/cancel
   # Should return 400 "already cancelled"
   ```

## Error Handling

### Common Pitfalls

| Issue | Symptom | Solution |
|-------|---------|----------|
| User tries to cancel sent invoice | Error 400 "been sent" | Guide to Credit Note workflow |
| Draft invoice without number | Error 400 "no number" | Use DELETE endpoint instead |
| Missing cancellation reason | Error 400 validation failed | Enforce required field in UI |
| Sessions not reverting | Data inconsistency | Check transaction rollback in logs |

### Logging

All cancellations should be logged:

```go
// Recommended: Add audit log
log.Info("Invoice cancelled",
    "invoice_id", invoiceID,
    "invoice_number", invoice.InvoiceNumber,
    "reason", reason,
    "user_id", userID,
    "sent_at_was_null", invoice.SentAt == nil,
)
```

## Performance Considerations

### Indexes

The migration creates indexes on:
- `sent_at` - For filtering sendable invoices
- `cancelled_at` - For reporting on cancelled invoices

### Query Optimization

Common query for "cancellable invoices":
```sql
SELECT id, invoice_number, status, sent_at
FROM invoices
WHERE tenant_id = ?
  AND status = 'finalized'
  AND sent_at IS NULL
  AND deleted_at IS NULL
ORDER BY invoice_date DESC;
```

## Security

### Authorization Checks

```go
// Multi-tenant isolation
if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", 
    id, tenantID, userID).First(&invoice).Error
```

All endpoints verify:
1. User is authenticated (JWT token)
2. Invoice belongs to user's tenant
3. Invoice belongs to requesting user (in current implementation)

### Audit Trail

Every cancellation creates an immutable record:
- `cancelled_at` timestamp (UTC)
- `cancellation_reason` text
- Original `invoice_number` retained
- Original creation date retained
- Status change logged

## Migration Path

### For Existing Invoices

Run migration to add columns:
```bash
cd unburdy_server/modules/client_management/migrations
# Migration runs automatically on next startup with AutoMigrate
```

Existing invoices will have:
- `sent_at = NULL` (can be cancelled if finalized)
- `send_method = NULL` (will be set on next send)
- `cancelled_at = NULL` (not cancelled)

### Backward Compatibility

The system remains backward compatible:
- Old invoices without `sent_at` can still be cancelled if finalized
- Old code checking only `status` still works
- New validation is additive, not breaking

## Related Documentation

- [GoBD Compliance Guide](INVOICE_CANCELLATION_GOBD_EN.md) - Legal requirements
- [GoBD Konformität (Deutsch)](INVOICE_CANCELLATION_GOBD_DE.md) - Rechtliche Anforderungen
- [Invoice Workflow](INVOICING.md) - Overall invoice lifecycle
- [API Documentation](SWAGGER_DOCUMENTATION.md) - Full API reference

## Status & Changelog

### Version 1.0 (January 2026)
- ✅ Initial implementation
- ✅ GoBD validation logic
- ✅ Unit tests (4/4 passing)
- ✅ Migration scripts
- ✅ Swagger documentation
- ⏳ Frontend integration (pending)
- ⏳ Audit log integration (pending)

## Support

For questions or issues:
1. Check error message - they contain specific guidance
2. Review test cases for examples
3. Check GoBD documentation for compliance questions
4. Contact backend team for implementation support
