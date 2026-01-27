# Invoice Cancellation - Quick Reference Card

> **One-page reference for the GoBD-compliant invoice cancellation feature**

## 🎯 Use Cases

| Scenario | Action | API Endpoint |
|----------|--------|--------------|
| Invoice finalized, never sent, needs correction | **Cancel** | `POST /client-invoices/:id/cancel` |
| Invoice sent, customer disputes charges | **Credit Note** | `POST /client-invoices/:id/credit-note` |
| Draft invoice without number | **Delete** | `DELETE /client-invoices/:id` |

## 📋 Cancel Prerequisites

```
✅ Invoice has invoice_number (finalized)
✅ Invoice sent_at IS NULL (never sent)
✅ Cancellation reason provided
❌ Invoice status NOT 'cancelled'
```

## 🔌 API Quick Start

### Cancel Invoice
```bash
curl -X POST https://api.example.com/api/v1/client-invoices/123/cancel \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Wrong customer address - will recreate"}'
```

**Success (200)**:
```json
{
  "success": true,
  "message": "Invoice cancelled successfully. Sessions and extra efforts reverted to unbilled status."
}
```

### Mark Invoice as Sent
```bash
curl -X POST https://api.example.com/api/v1/client-invoices/123/mark-sent \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"send_method": "email"}'
```

## 🚨 Common Errors

| Error | Meaning | Solution |
|-------|---------|----------|
| `"cannot cancel invoice that has been sent"` | Invoice was sent | Use credit note |
| `"invoice has no number"` | Draft without finalization | Use DELETE |
| `"invoice is already cancelled"` | Already cancelled | No action needed |

## 💾 Database Fields

```sql
-- New fields in invoices table
sent_at             TIMESTAMP    -- When sent to customer
send_method         VARCHAR(50)  -- email, manual, xrechnung
cancelled_at        TIMESTAMP    -- When cancelled
cancellation_reason TEXT         -- Why cancelled
```

## 🔍 Frontend Integration

### Show Cancel Button
```typescript
const canCancel = invoice.status === 'finalized' 
               && !invoice.sent_at 
               && invoice.invoice_number;
```

### Cancel Function
```typescript
async function cancelInvoice(id: number, reason: string) {
  const res = await fetch(`/api/v1/client-invoices/${id}/cancel`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ reason })
  });
  
  if (!res.ok) {
    const error = await res.json();
    if (error.message?.includes('been sent')) {
      alert('Use credit note for sent invoices');
    }
    throw error;
  }
  
  return res.json();
}
```

## 📊 What Happens on Cancel

```
1. Validates invoice is cancellable
2. Starts database transaction
3. Reverts sessions to "conducted"
4. Reverts extra efforts to "unbilled"
5. Deletes ClientInvoice links (hard delete)
6. Sets invoice:
   - status = 'cancelled'
   - cancelled_at = NOW()
   - cancellation_reason = <provided reason>
   - invoice_number = UNCHANGED ✅
7. Commits transaction
```

## ⚖️ GoBD Compliance

### ✅ What We Do
- Keep invoice number forever (no reuse)
- Record cancellation timestamp
- Require cancellation reason
- Prevent cancellation of sent invoices
- Maintain complete audit trail

### ❌ What We Don't Do
- Delete invoice records
- Modify original amounts
- Reuse invoice numbers
- Allow anonymous cancellations

## 🧪 Testing

### Test Scenarios
```bash
# Run all tests
cd modules/client_management
go test ./tests/... -run TestCancelInvoice -v

# Expected: 4/4 passing
# - Success case
# - Already sent (rejected)
# - No number (rejected)
# - Already cancelled (rejected)
```

### Manual Test
```bash
# 1. Create & finalize invoice
POST /client-invoices/from-sessions
POST /client-invoices/:id/finalize

# 2. Verify never sent
GET /client-invoices/:id
# Check: sent_at === null

# 3. Cancel
POST /client-invoices/:id/cancel
{"reason": "Test cancellation"}

# 4. Verify cancelled
GET /client-invoices/:id
# Check: status === 'cancelled', cancelled_at !== null

# 5. Try cancel again (should fail)
POST /client-invoices/:id/cancel
# Expect: 400 "already cancelled"
```

## 📚 Documentation

| Document | Purpose |
|----------|---------|
| [INVOICE_CANCELLATION.md](INVOICE_CANCELLATION.md) | Developer guide |
| [INVOICE_CANCELLATION_GOBD_EN.md](INVOICE_CANCELLATION_GOBD_EN.md) | Legal compliance (English) |
| [INVOICE_CANCELLATION_GOBD_DE.md](INVOICE_CANCELLATION_GOBD_DE.md) | Rechtliche Konformität (Deutsch) |
| [INVOICE_CANCELLATION_SWAGGER.md](INVOICE_CANCELLATION_SWAGGER.md) | API reference |
| [INVOICE_CANCELLATION_RELEASE_NOTES.md](INVOICE_CANCELLATION_RELEASE_NOTES.md) | Release notes |

## 🔧 Troubleshooting

### "Failed to revert session statuses"
**Cause**: Database transaction failed  
**Fix**: Check database connection and retry

### "Invoice not found"
**Cause**: Wrong ID or tenant isolation  
**Fix**: Verify invoice ID and user token

### Frontend shows cancel but API rejects
**Cause**: Race condition - invoice was sent between load and click  
**Fix**: Refresh invoice data before cancellation

## 📞 Support

- **Documentation**: See links above
- **API Issues**: Backend development team
- **Legal Questions**: Compliance team / Steuerberater
- **Bug Reports**: Project issue tracker

---

**Version**: 1.0 | **Updated**: January 26, 2026 | **Status**: ✅ Production Ready
