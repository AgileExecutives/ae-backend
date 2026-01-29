# Invoice Generation Workflow

## Current Architecture

The system has three separate modules:
- **client_management** - Invoice database records
- **templates** - HTML template rendering
- **documents** - PDF generation and storage

## Workflow Options

### Option 1: Integrated Single Endpoint (RECOMMENDED)

**Endpoint:** `POST /api/v1/client-invoices/generate`

**Request:**
```json
{
  "invoice_number": "INV-2025-12-0001",
  "client_id": 4,
  "cost_provider_id": 17,
  "invoice_date": "2025-12-27",
  "due_date": "2026-01-10",
  "line_items": [
    {
      "session_id": 70,
      "date": "2025-12-04T14:00:00Z",
      "description": "Therapy Session",
      "units": 1,
      "unit_price": 100,
      "subtotal": 100
    }
  ],
  "net_total": 100,
  "tax_rate": 19,
  "tax_amount": 19,
  "gross_total": 119,
  "template_id": 1,  // Optional - uses default if not provided
  "auto_generate_pdf": true  // Optional - default true
}
```

**Response:**
```json
{
  "success": true,
  "message": "Invoice created and PDF generated successfully",
  "data": {
    "invoice": {
      "id": 123,
      "invoice_number": "INV-2025-12-0001",
      "client_id": 4,
      "status": "generated",
      ...
    },
    "document": {
      "id": 456,
      "storage_key": "invoices/2025/12/INV-2025-12-0001.pdf",
      "filename": "INV-2025-12-0001.pdf",
      "document_type": "invoice",
      "size_bytes": 45678,
      "download_url": "/api/v1/documents/456/download"
    },
    "pdf_url": "/api/v1/documents/456/download"
  }
}
```

**What it does:**
1. ✅ Creates invoice record in database
2. ✅ Validates sessions aren't already invoiced
3. ✅ Gets default template (or specified template)
4. ✅ Renders HTML with invoice data
5. ✅ Generates PDF using chromedp
6. ✅ Saves PDF to MinIO storage
7. ✅ Creates document record
8. ✅ Links document to invoice (add document_id to invoices table)
9. ✅ Returns complete invoice + document info

**Frontend usage:**
```typescript
// Single API call
const response = await api.post('/client-invoices/generate', invoiceData);

// Access invoice
const invoice = response.data.invoice;

// Download PDF
const pdfUrl = response.data.pdf_url;
window.open(pdfUrl, '_blank');
```

---

### Option 2: Multi-Step Manual Process (Current)

**Step 1:** Create invoice
```
POST /api/v1/client-invoices
```

**Step 2:** Generate PDF separately
```
POST /api/v1/pdfs/from-template
{
  "template_id": 1,
  "data": { /* invoice data */ },
  "save_document": true
}
```

**Step 3:** Update invoice with document_id
```
PATCH /api/v1/client-invoices/{id}
{
  "document_id": 456
}
```

**Issues:**
- ❌ Requires 3 API calls
- ❌ Frontend must manage invoice data transformation
- ❌ Risk of invoice without PDF
- ❌ Complex error handling across calls

---

## Database Schema Changes Needed

Add to `invoices` table:
```sql
ALTER TABLE invoices ADD COLUMN document_id INTEGER REFERENCES documents(id);
CREATE INDEX idx_invoice_document ON invoices(document_id);
```

Or add to invoice entity:
```go
type Invoice struct {
    // ... existing fields ...
    DocumentID *uint      `gorm:"index:idx_invoice_document" json:"document_id,omitempty"`
    Document   *Document  `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
}
```

---

## Implementation Plan

### 1. Update Invoice Entity
Add document relationship to track generated PDFs.

### 2. Create Integrated Service
```go
// InvoiceService method
func (s *InvoiceService) GenerateInvoiceWithPDF(
    req entities.CreateInvoiceDirectRequest,
    templateService *templateServices.TemplateService,
    pdfService *documentServices.PDFService,
    tenantID, userID uint,
) (*entities.Invoice, *documents.Document, error)
```

### 3. Add New Handler
```go
func (h *InvoiceHandler) GenerateInvoiceWithPDF(c *gin.Context)
```

### 4. Register Route
```go
clientInvoices.POST("/generate", rp.invoiceHandler.GenerateInvoiceWithPDF)
```

---

## Frontend API Client

```typescript
class InvoiceAPI {
  // Option 1: Integrated endpoint (recommended)
  async generateInvoice(data: CreateInvoiceRequest): Promise<InvoiceWithDocument> {
    const response = await api.post('/client-invoices/generate', data);
    return response.data;
  }

  // Download PDF
  async downloadPDF(documentId: number): Promise<Blob> {
    const response = await api.get(`/documents/${documentId}/download`, {
      responseType: 'blob'
    });
    return response.data;
  }

  // Get invoice with PDF
  async getInvoice(id: number): Promise<InvoiceWithDocument> {
    const response = await api.get(`/client-invoices/${id}`);
    return response.data;
  }

  // List invoices with PDFs
  async listInvoices(filters?: InvoiceFilters): Promise<Invoice[]> {
    const response = await api.get('/client-invoices', { params: filters });
    return response.data;
  }
}
```

---

## Usage Example

```typescript
// Create invoice and generate PDF in one call
const result = await invoiceAPI.generateInvoice({
  invoice_number: generatedNumber,
  client_id: selectedClient.id,
  cost_provider_id: selectedClient.cost_provider_id,
  invoice_date: new Date().toISOString().split('T')[0],
  due_date: addDays(new Date(), 14).toISOString().split('T')[0],
  line_items: selectedSessions.map(session => ({
    session_id: session.id,
    date: session.date,
    description: 'Therapy Session',
    units: 1,
    unit_price: session.unit_price || 0,
    subtotal: session.unit_price || 0
  })),
  net_total: calculateNetTotal(selectedSessions),
  tax_rate: 19,
  tax_amount: calculateTax(selectedSessions, 19),
  gross_total: calculateGrossTotal(selectedSessions),
  auto_generate_pdf: true
});

// Invoice created and PDF generated
console.log('Invoice ID:', result.invoice.id);
console.log('PDF URL:', result.pdf_url);

// Download PDF immediately
window.open(result.pdf_url, '_blank');

// Or save for later
saveInvoiceToState(result.invoice);
```

---

## Next Steps

1. **Add document_id to Invoice entity**
2. **Create GenerateInvoiceWithPDF service method**
3. **Inject template & PDF services into InvoiceHandler**
4. **Add /generate endpoint**
5. **Update frontend to use single endpoint**

Would you like me to implement the integrated endpoint?
