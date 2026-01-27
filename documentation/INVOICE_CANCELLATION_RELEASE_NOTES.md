# Release Notes: Invoice Cancellation (Storno) Feature

**Version**: 1.0  
**Release Date**: January 26, 2026  
**Module**: Client Management / Invoicing  
**Status**: ✅ Production Ready

## Overview

Implemented GoBD-compliant invoice cancellation functionality for invoices that have been finalized but never sent to customers.

## What's New

### ✨ Features

#### 1. Cancel Never-Sent Invoices
- Cancel finalized invoices that were never sent (`sent_at IS NULL`)
- Invoice number is permanently retained (GoBD compliance)
- Sessions and extra efforts revert to unbilled status
- Mandatory cancellation reason for audit trail

#### 2. Send Method Tracking
- Track how invoices are sent: `email`, `manual`, or `xrechnung`
- Prevents cancellation once invoice is sent
- Clear differentiation between cancel and credit note workflows

#### 3. Enhanced Workflow Tracking
- `sent_at`: Timestamp when invoice was sent
- `send_method`: How invoice was delivered
- `cancelled_at`: Timestamp of cancellation
- `cancellation_reason`: Required audit trail text

### 🔧 API Changes

#### New Endpoint
```
POST /api/v1/client-invoices/:id/cancel
```
**Request**:
```json
{
  "reason": "Fehlerhafte Positionen – Rechnung nicht versendet"
}
```

#### Updated Endpoint
```
POST /api/v1/client-invoices/:id/mark-sent
```
**Breaking Change**: Now requires request body
```json
{
  "send_method": "email"  // Required: email, manual, or xrechnung
}
```

### 📊 Database Changes

**New Fields in `invoices` table**:
- `sent_at` (TIMESTAMP NULL) - When invoice was sent
- `send_method` (VARCHAR(50)) - How it was sent
- `cancelled_at` (TIMESTAMP NULL) - When cancelled
- `cancellation_reason` (TEXT) - Why cancelled

**Indexes Added**:
- `idx_invoices_sent_at`
- `idx_invoices_cancelled_at`

**Migration**: `add_cancellation_fields_to_invoices.sql`

## Validation Rules

### Cancel Conditions
✅ **Allowed**:
- Invoice has invoice number (finalized)
- Invoice has `sent_at IS NULL` (never sent)
- Invoice status is not already 'cancelled'
- Valid cancellation reason provided

❌ **Blocked**:
- Invoice was already sent (`sent_at IS NOT NULL`) → Use credit note
- Invoice has no number → Use delete endpoint
- Invoice is already cancelled → Cannot cancel twice

## GoBD Compliance

### ✅ Compliant Features
- **Unveränderbarkeit**: Sent invoices cannot be cancelled
- **Nachvollziehbarkeit**: Full audit trail with timestamps and reasons
- **Vollständigkeit**: All workflow states tracked
- **Fortlaufende Nummerierung**: Invoice numbers never reused
- **Aufbewahrung**: Cancelled invoices retained permanently

### 📋 Audit Trail
Every cancellation records:
- Original invoice number (never deleted)
- Cancellation timestamp (UTC)
- Cancellation reason (mandatory)
- Technical proof invoice was never sent (`sent_at IS NULL`)

## Testing

### ✅ Test Coverage
- `TestCancelInvoice_Success` - Happy path ✅
- `TestCancelInvoice_AlreadySent` - Rejects sent invoices ✅
- `TestCancelInvoice_NoNumber` - Rejects drafts ✅
- `TestCancelInvoice_AlreadyCancelled` - Prevents double cancellation ✅

**Run Tests**:
```bash
cd modules/client_management
go test ./tests/... -run TestCancelInvoice -v
```

## Migration Guide

### For Backend Developers

#### 1. Database Migration
```bash
# Migration runs automatically on next startup
# Or run manually:
cd unburdy_server/modules/client_management/migrations
psql -d ae_dev -f add_cancellation_fields_to_invoices.sql
```

#### 2. Update MarkAsSent Calls
**Before**:
```go
invoice, err := service.MarkAsSent(invoiceID, tenantID)
```

**After**:
```go
invoice, err := service.MarkAsSent(invoiceID, tenantID, "email")
```

### For Frontend Developers

#### 1. Update API Calls

**Cancel Invoice** (NEW):
```typescript
async function cancelInvoice(id: number, reason: string) {
  const response = await fetch(`/api/v1/client-invoices/${id}/cancel`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ reason })
  });
  return response.json();
}
```

**Mark as Sent** (UPDATED):
```typescript
async function markInvoiceAsSent(id: number, sendMethod: 'email' | 'manual' | 'xrechnung') {
  const response = await fetch(`/api/v1/client-invoices/${id}/mark-sent`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ send_method: sendMethod })  // NEW: Required body
  });
  return response.json();
}
```

#### 2. Update Invoice Interface
```typescript
interface Invoice {
  // Existing fields...
  sent_at?: string;
  send_method?: 'email' | 'manual' | 'xrechnung';
  finalized_at?: string;
  cancelled_at?: string;
  cancellation_reason?: string;
}
```

#### 3. Conditional UI Display
```typescript
function showCancelButton(invoice: Invoice): boolean {
  return invoice.status === 'finalized' 
      && !invoice.sent_at 
      && !!invoice.invoice_number;
}

function showCreditNoteButton(invoice: Invoice): boolean {
  return !!invoice.sent_at 
      && ['sent', 'paid', 'overdue'].includes(invoice.status);
}
```

## Error Handling

### Common Errors

| Error Code | Message | Solution |
|------------|---------|----------|
| 400 | "cannot cancel invoice that has been sent" | Use credit note endpoint instead |
| 400 | "invoice has no number - use DeleteInvoice" | Use DELETE endpoint for drafts |
| 400 | "invoice is already cancelled" | Invoice was already cancelled |
| 404 | "invoice not found" | Check invoice ID and tenant access |

### User-Friendly Messages

```typescript
function handleCancelError(error: any) {
  const message = error.response?.data?.message || '';
  
  if (message.includes('been sent')) {
    return 'This invoice was already sent. Please create a credit note instead.';
  } else if (message.includes('no number')) {
    return 'This draft invoice can be deleted instead of cancelled.';
  } else if (message.includes('already cancelled')) {
    return 'This invoice is already cancelled.';
  }
  return 'Failed to cancel invoice. Please try again.';
}
```

## Documentation

### 📖 Available Docs
- **[Developer Guide](INVOICE_CANCELLATION.md)** - Technical implementation
- **[GoBD Compliance (EN)](INVOICE_CANCELLATION_GOBD_EN.md)** - Legal requirements
- **[GoBD Konformität (DE)](INVOICE_CANCELLATION_GOBD_DE.md)** - Rechtliche Anforderungen
- **[Swagger API](INVOICE_CANCELLATION_SWAGGER.md)** - API reference

### 🎓 Training Materials
1. Read GoBD documentation to understand legal requirements
2. Review test cases for implementation examples
3. Check Swagger UI for API testing
4. Use developer guide for integration

## Known Limitations

1. **No batch cancellation**: Must cancel invoices one at a time
2. **No undo**: Cancellation is permanent (cannot uncancel)
3. **No partial cancellation**: Must cancel entire invoice (use credit notes for partial)

## Future Enhancements

### Planned for Next Release
- [ ] Batch cancellation API
- [ ] Audit log integration (separate table)
- [ ] Email notification on cancellation
- [ ] Dashboard widget for cancelled invoices
- [ ] Export cancelled invoices report

### Under Consideration
- [ ] Configurable cancellation reason templates
- [ ] User permission controls (restrict who can cancel)
- [ ] Cancellation approval workflow
- [ ] Automatic session re-allocation

## Performance Impact

### Database
- ✅ Minimal: Two new indexes on timestamp fields
- ✅ No schema changes to existing columns
- ✅ Soft delete maintains referential integrity

### API
- ✅ Single transaction per cancellation
- ✅ No N+1 query issues
- ✅ Efficient session reversion (batch update)

## Security

### Authorization
- ✅ Multi-tenant isolation enforced
- ✅ User must own invoice to cancel
- ✅ Bearer token authentication required

### Audit
- ✅ All cancellations logged with timestamp
- ✅ User ID tracked
- ✅ Reason required (no anonymous cancellations)

## Breaking Changes

### ⚠️ Breaking Change: MarkAsSent Endpoint

**Impact**: All clients calling `POST /client-invoices/:id/mark-sent`

**Before** (No request body):
```bash
POST /client-invoices/123/mark-sent
```

**After** (Requires send_method):
```bash
POST /client-invoices/123/mark-sent
Content-Type: application/json

{"send_method": "email"}
```

**Migration Deadline**: Before next production deployment

## Rollback Plan

If issues arise:

1. **Database**: Run rollback migration (provided)
   ```sql
   ALTER TABLE invoices DROP COLUMN sent_at;
   ALTER TABLE invoices DROP COLUMN send_method;
   ALTER TABLE invoices DROP COLUMN cancelled_at;
   ALTER TABLE invoices DROP COLUMN cancellation_reason;
   DROP INDEX idx_invoices_sent_at;
   DROP INDEX idx_invoices_cancelled_at;
   ```

2. **Code**: Revert to previous commit
   ```bash
   git revert <commit-hash>
   ```

3. **API**: Old MarkAsSent signature available in rollback branch

## Support

### Getting Help
1. Check error messages (they're descriptive)
2. Review documentation above
3. Contact backend team
4. File issue in project tracker

### Reporting Issues
Include:
- Invoice ID
- Error message
- Expected vs actual behavior
- Steps to reproduce

## Changelog

### v1.0 (January 26, 2026)
- ✅ Initial implementation
- ✅ GoBD validation logic
- ✅ Database migration
- ✅ Unit tests (4/4 passing)
- ✅ Swagger documentation
- ✅ Developer documentation (EN + DE)
- ✅ GoBD compliance documentation (EN + DE)

## Contributors

- Backend Development Team
- Compliance Review Team
- QA Testing Team

## License

Proprietary - Internal Use Only

---

**Questions?** Contact the backend development team or refer to the documentation links above.
