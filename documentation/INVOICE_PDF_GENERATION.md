# Integrated Invoice + PDF Generation

## Overview

The integrated invoice generation endpoint creates an invoice record and optionally generates a PDF document in a single atomic operation.

## Endpoint

```
POST /api/v1/client-invoices/generate
```

## Features

✅ **Single-call invoice creation** - Create invoice with one API call  
✅ **Optional PDF generation** - Control via `auto_generate_pdf` flag  
✅ **Automatic template lookup** - Uses default invoice template if not specified  
✅ **Document storage** - PDF saved to MinIO storage automatically  
✅ **Invoice-document linking** - `document_id` set on invoice record  
✅ **Service availability check** - Graceful degradation if PDF services unavailable  

## Request Body

The request body has a clean two-part structure:
1. `unbilledClient` - Complete client object from `/client-invoices/unbilled-sessions`
2. `parameters` - Invoice generation settings

```json
{
  "unbilledClient": {
    "id": 1,
    "first_name": "John",
    "last_name": "Doe",
    "cost_provider_id": 1,
    "street_address": "123 Main St",
    "zip": "12345",
    "city": "New York",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "therapy_title": "Cognitive Behavioral Therapy",
    "provider_approval_code": "PROV123",
    "unit_price": 100.00,
    "cost_provider": {
      "id": 1,
      "name": "Insurance Co"
    },
    "sessions": [
      {
        "id": 1,
        "original_date": "2024-01-15T00:00:00Z",
        "original_start_time": "2024-01-15T10:00:00Z",
        "duration_min": 60,
        "type": "therapy",
        "number_units": 1,
        "status": "conducted"
      },
      {
        "id": 2,
        "original_date": "2024-01-22T00:00:00Z",
        "original_start_time": "2024-01-22T10:00:00Z",
        "duration_min": 60,
        "type": "therapy",
        "number_units": 1,
        "status": "conducted"
      }
    ]
  },
  "parameters": {
    "invoice_number": "INV-2024-001",
    "invoice_date": "2024-01-31",
    "tax_rate": 19.0,
    "generate_pdf": true,
    "session_from_date": "2024-01-01",
    "session_to_date": "2024-01-31"
  }
}
```

## Request Fields

The request has two main parts:

### 1. unbilledClient Object
The complete client object from `/client-invoices/unbilled-sessions` endpoint, including:

- `id` - Client ID (required)
- `first_name`, `last_name` - Client name
- `date_of_birth`, `gender`, `primary_language`
- `contact_first_name`, `contact_last_name`, `contact_email`, `contact_phone`
- `street_address`, `zip`, `city`, `email`, `phone`
- `therapy_title`, `provider_approval_code`, `provider_approval_date`
- `unit_price` - Used for calculating invoice totals
- `cost_provider_id`, `cost_provider` - Cost provider reference
- `sessions` - Array of SessionResponse objects
- And all other client fields available in the template context

### 2. parameters Object
Invoice generation settings:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `invoice_number` | string | No | Auto-generated | Invoice number |
| `invoice_date` | string | No | Current date | Invoice date (YYYY-MM-DD) |
| `tax_rate` | float | No | 19.0 | Tax rate percentage |
| `generate_pdf` | bool | No | **true** | Generate PDF? |
| `template_id` | uint | No | Auto-select | Specific template ID |
| `session_from_date` | string | No | - | Filter sessions from date (YYYY-MM-DD) |
| `session_to_date` | string | No | - | Filter sessions to date (YYYY-MM-DD) |

**Note:** `generate_pdf` defaults to **true**. Set explicitly to `false` to skip PDF generation.

## Response

### Without PDF (`generate_pdf: false`)

```json
{
  "success": true,
  "message": "Invoice created successfully",
  "data": {
    "id": 123,
    "invoice_number": "INV-2024-001",
    "sum_amount": 200.00,
    "tax_amount": 38.00,
    "total_amount": 238.00,
    "status": "generated"
  }
}
```

### With PDF (`generate_pdf: true`, default)

```json
{
  "success": true,
  "message": "Invoice created and PDF generated successfully",
  "data": {
    "invoice": {
      "id": 123,
      "invoice_number": "INV-2024-001",
      "document_id": 456,
      "sum_amount": 200.00,
      "tax_amount": 38.00,
      "total_amount": 238.00,
      "status": "generated"
    },
    "document": {
      "id": 456,
      "storage_key": "tenants/1/documents/invoice/INV-2024-001.pdf",
      "filename": "INV-2024-001.pdf",
      "size_bytes": 45678,
      "download_url": "/api/v1/documents/456/download"
    },
    "pdf_url": "/api/v1/documents/456/download"
  }
}
```

## Error Responses

### Service Unavailable (503)

When PDF generation is requested but services are not configured:

```json
{
  "success": false,
  "error": "Service unavailable",
  "message": "PDF generation service is not configured"
}
```

### Template Not Found (500)

When no invoice template is available:

```json
{
  "success": false,
  "error": "Failed to find invoice template",
  "message": "No invoice template available"
}
```

## Implementation Details

### Service Dependencies

The handler requires two services for PDF generation:
- **Template Service** - For rendering invoice HTML from template
- **PDF Service** - For converting HTML to PDF and storing

These services are injected via the service registry in the CoreModule initialization.

### Workflow

1. **Parse Request** - Extract `unbilledClient` and `parameters` from request body
2. **Filter Sessions** - Apply `session_from_date` and `session_to_date` from parameters if specified
3. **Create Invoice** - Invoice record created in database with filtered sessions from `unbilledClient.sessions`
4. **Check Generate PDF** - If `parameters.generate_pdf: false`, return invoice and exit (default is true)
5. **Fetch Template** - Lookup template by `parameters.template_id` or use default invoice template
6. **Prepare Template Data** - Combine invoice data with all fields from `unbilledClient` object
7. **Render HTML** - Template service renders HTML with invoice + client data
8. **Generate PDF** - PDF service converts HTML to PDF
9. **Store Document** - PDF stored in MinIO with invoice reference
10. **Update Invoice** - Invoice record updated with `document_id` and `template_id`
11. **Return Response** - Invoice and document details returned to client
5. **Verify Services** - Check template/PDF services available
6. **Lookup Template** - Use specified `template_id` or find default active invoice template
7. **Prepare Data** - Build template data from complete client object and filtered sessions
8. **Render HTML** - Use template service to render invoice HTML
9. **Generate PDF** - Use PDF service to convert HTML to PDF
10. **Store Document** - Save PDF to storage and create document record
11. **Link Invoice** - Update invoice with `document_id`
12. **Return Response** - Return invoice + document details

### Database Changes

Added to `Invoice` entity:
- `document_id` (uint, nullable) - Links to generated PDF document
- `template_id` (uint, nullable) - Template used for generation
- `auto_generate_pdf` (bool, transient) - Controls PDF generation

## Testing

Run the test script:

```bash
cd /Users/alex/src/ae/backend/unburdy_server
./tests/test-invoice-with-pdf.sh
```

Or test manually with curl:

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8082/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}' \
  | jq -r '.data.token')

# Create invoice with PDF
curl -X POST http://localhost:8082/api/v1/client-invoices/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": 1,
    "client_name": "Test Client",
    "cost_provider_id": 1,
    "cost_provider_name": "Test Provider",
    "invoice_date": "2024-01-15",
    "tax_rate": 19.0,
    "auto_generate_pdf": true,
    "line_items": [{
      "description": "Session 1",
      "quantity": 1,
      "unit_price": 100.00,
      "total": 100.00
    }]
  }' | jq .
```

## Migration Notes

### Legacy Module Support

The legacy `NewModule` function (non-bootstrap) passes `nil` for template and PDF services, meaning PDF generation will not work in that context. The handler gracefully returns a 503 error in this case.

### Service Registry

Services are registered in the documents module:
- Service name: `"pdf_service"`
- Type: `*documentServices.PDFService`

The client management module retrieves it during initialization:

```go
if pdfSvcRaw, ok := ctx.Services.Get("pdf_service"); ok {
    if pdfSvc, ok := pdfSvcRaw.(*documentServices.PDFService); ok {
        pdfService = pdfSvc
    }
}
```

## Related Files

- Handler: [unburdy_server/modules/client_management/handlers/invoice_handler.go](../modules/client_management/handlers/invoice_handler.go#L76-L234)
- Routes: [unburdy_server/modules/client_management/routes/routes.go](../modules/client_management/routes/routes.go#L90)
- Entity: [unburdy_server/modules/client_management/entities/invoice.go](../modules/client_management/entities/invoice.go)
- Module: [unburdy_server/modules/client_management/module.go](../modules/client_management/module.go#L186-L197)
- Test: [tests/test-invoice-with-pdf.sh](../tests/test-invoice-with-pdf.sh)
