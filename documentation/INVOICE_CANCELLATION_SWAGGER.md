# Invoice Cancellation API - Swagger Summary

## Overview

This document summarizes the Swagger/OpenAPI annotations for the invoice cancellation feature.

## Updated Endpoints

### 1. Cancel Invoice (NEW)

**Path**: `/api/v1/client-invoices/{id}/cancel`
**Method**: `POST`
**Tags**: `client-invoices`
**Operation ID**: `cancelInvoice`

**Swagger Annotations**:
```go
// @Summary Cancel an invoice
// @Description Cancel an invoice that has not been sent (sent_at IS NULL). For sent invoices, use credit note instead.
// @Tags client-invoices
// @ID cancelInvoice
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param request body entities.CancelInvoiceRequest true "Cancellation reason"
// @Success 200 {object} models.SuccessMessageResponse
// @Failure 400 {object} models.ErrorResponse "Invoice already sent or invalid request"
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/cancel [post]
```

**Request Schema**:
```json
{
  "reason": "string (required, min 1 char)"
}
```

**Example**:
```json
{
  "reason": "Fehlerhafte Positionen – Rechnung nicht versendet"
}
```

**Responses**:

**200 OK**:
```json
{
  "success": true,
  "message": "Invoice cancelled successfully. Sessions and extra efforts reverted to unbilled status."
}
```

**400 Bad Request** (Already Sent):
```json
{
  "success": false,
  "error": "Cannot cancel invoice",
  "message": "cannot cancel invoice that has been sent - create a credit note (Stornorechnung/Gutschrift) instead"
}
```

**400 Bad Request** (No Number):
```json
{
  "success": false,
  "error": "Cannot cancel invoice",
  "message": "invoice has no number - use DeleteInvoice for draft invoices without numbers"
}
```

**400 Bad Request** (Already Cancelled):
```json
{
  "success": false,
  "error": "Cannot cancel invoice",
  "message": "invoice is already cancelled"
}
```

**404 Not Found**:
```json
{
  "success": false,
  "error": "Invoice not found",
  "message": "invoice not found"
}
```

### 2. Mark Invoice as Sent (UPDATED)

**Path**: `/api/v1/client-invoices/{id}/mark-sent`
**Method**: `POST`
**Tags**: `client-invoices`
**Operation ID**: `markInvoiceAsSent`

**Updated Swagger Annotations**:
```go
// @Summary Mark invoice as sent
// @Description Mark a finalized invoice as sent (changes status from finalized to sent). Requires send_method to be specified.
// @Tags client-invoices
// @ID markInvoiceAsSent
// @Accept json
// @Produce json
// @Param id path int true "Invoice ID"
// @Param request body entities.MarkInvoiceAsSentRequest true "Send method (email, manual, xrechnung)"
// @Success 200 {object} entities.InvoiceAPIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /client-invoices/{id}/mark-sent [post]
```

**Request Schema** (NEW):
```json
{
  "send_method": "string (required, enum: email, manual, xrechnung)"
}
```

**Example**:
```json
{
  "send_method": "email"
}
```

**Changes from Previous Version**:
- ❌ **Before**: No request body required
- ✅ **After**: Requires `send_method` in request body
- ✅ **Validation**: `binding:"required,oneof=email manual xrechnung"`

## New Request DTOs

### CancelInvoiceRequest

**Location**: `entities/invoice.go`

```go
type CancelInvoiceRequest struct {
    Reason string `json:"reason" binding:"required" example:"Fehlerhafte Positionen – Rechnung nicht versendet"`
}
```

**Validation**:
- `reason` is required
- No minimum length enforced (but recommended 10+ chars for audit)

### MarkInvoiceAsSentRequest

**Location**: `entities/invoice.go`

```go
type MarkInvoiceAsSentRequest struct {
    SendMethod string `json:"send_method" binding:"required,oneof=email manual xrechnung" example:"email"`
}
```

**Validation**:
- `send_method` is required
- Must be one of: `email`, `manual`, `xrechnung`

## Updated Response DTOs

### InvoiceResponse (Extended)

**Location**: `entities/invoice.go`

**New Fields**:
```go
type InvoiceResponse struct {
    // ... existing fields ...
    
    SentAt             *time.Time `json:"sent_at,omitempty"`
    SendMethod         string     `json:"send_method,omitempty"`
    FinalizedAt        *time.Time `json:"finalized_at,omitempty"`
    CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
    CancellationReason string     `json:"cancellation_reason,omitempty"`
}
```

**Example**:
```json
{
  "id": 123,
  "invoice_number": "INV-1-00042",
  "status": "cancelled",
  "total_amount": 150.00,
  "sent_at": null,
  "send_method": null,
  "finalized_at": "2026-01-26T10:30:00Z",
  "cancelled_at": "2026-01-26T11:15:00Z",
  "cancellation_reason": "Fehlerhafte Positionen – Rechnung nicht versendet",
  "...": "..."
}
```

## Swagger Generation

### Generate Swagger Docs

```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Navigate to project root
cd /Users/alex/src/ae/backend/unburdy_server

# Generate swagger docs
swag init

# Docs will be created at:
# docs/docs.go
# docs/swagger.json
# docs/swagger.yaml
```

### Access Swagger UI

**Local Development**:
```
http://localhost:8080/swagger/index.html
```

**Production**:
```
https://api.yourdomain.com/swagger/index.html
```

## Breaking Changes

### ⚠️ Mark Invoice as Sent

**Impact**: API clients must update to include `send_method` in request body

**Migration**:

**Before**:
```javascript
// Old API call (no body)
await fetch(`/api/v1/client-invoices/${id}/mark-sent`, {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` }
});
```

**After**:
```javascript
// New API call (with send_method)
await fetch(`/api/v1/client-invoices/${id}/mark-sent`, {
  method: 'POST',
  headers: { 
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({ send_method: 'email' })
});
```

### Timeline

- **Implemented**: January 26, 2026
- **Recommended Migration Window**: Before next production deployment
- **Backward Compatibility**: ❌ None (required field)

## Testing in Swagger UI

### Test Cancel Endpoint

1. Open Swagger UI
2. Navigate to **client-invoices** section
3. Find `POST /client-invoices/{id}/cancel`
4. Click "Try it out"
5. Enter:
   - `id`: Invoice ID (e.g., 123)
   - `request`:
     ```json
     {
       "reason": "Test cancellation for development"
     }
     ```
6. Click "Execute"
7. Check response

### Expected Results

**Successful Cancellation** (200):
```json
{
  "success": true,
  "message": "Invoice cancelled successfully. Sessions and extra efforts reverted to unbilled status."
}
```

**Invoice Already Sent** (400):
```json
{
  "success": false,
  "error": "Cannot cancel invoice",
  "message": "cannot cancel invoice that has been sent - create a credit note (Stornorechnung/Gutschrift) instead"
}
```

## OpenAPI 3.0 Schema Snippet

```yaml
paths:
  /client-invoices/{id}/cancel:
    post:
      tags:
        - client-invoices
      summary: Cancel an invoice
      description: Cancel an invoice that has not been sent (sent_at IS NULL). For sent invoices, use credit note instead.
      operationId: cancelInvoice
      security:
        - BearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
          description: Invoice ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CancelInvoiceRequest'
            example:
              reason: "Fehlerhafte Positionen – Rechnung nicht versendet"
      responses:
        '200':
          description: Invoice cancelled successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SuccessMessageResponse'
        '400':
          description: Invoice already sent or invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '401':
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '404':
          description: Invoice not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

components:
  schemas:
    CancelInvoiceRequest:
      type: object
      required:
        - reason
      properties:
        reason:
          type: string
          description: Reason for cancellation (required for audit trail)
          example: "Fehlerhafte Positionen – Rechnung nicht versendet"
    
    MarkInvoiceAsSentRequest:
      type: object
      required:
        - send_method
      properties:
        send_method:
          type: string
          enum: [email, manual, xrechnung]
          description: Method used to send the invoice
          example: "email"
```

## Documentation Links

For detailed information, see:
- [Developer Documentation](INVOICE_CANCELLATION.md)
- [GoBD Compliance (English)](INVOICE_CANCELLATION_GOBD_EN.md)
- [GoBD Konformität (Deutsch)](INVOICE_CANCELLATION_GOBD_DE.md)

## Status

- ✅ Swagger annotations complete
- ✅ Request/Response DTOs defined
- ✅ Validation rules documented
- ✅ Examples provided
- ✅ Breaking changes documented
- ⏳ Frontend client update required
- ⏳ Swagger docs regeneration needed

**Last Updated**: January 26, 2026
