# Invoice PDF Generation

## Overview

The InvoicePDFService provides PDF generation for invoices using chromedp (headless Chrome). It supports:

- **Draft invoices** with "ENTWURF" watermark
- **Final invoices** without watermark (immutable)
- **Credit notes** with special header and reference to original invoice

## Prerequisites

### Chromedp Dependency

PDF generation requires Chrome/Chromium to be installed:

**macOS:**
```bash
brew install chromium --no-quarantine
```

**Ubuntu/Debian:**
```bash
apt-get install chromium-browser
```

**Docker:**
```dockerfile
RUN apt-get update && apt-get install -y \
    chromium \
    chromium-driver
```

## Usage

### 1. Initialize the Service

```go
pdfService := services.NewInvoicePDFService(db)
```

### 2. Generate Draft PDF (with watermark)

```go
ctx := context.Background()
pdfData, err := pdfService.GenerateDraftPDF(ctx, invoice)
if err != nil {
    log.Fatal(err)
}

// Save to file (development/testing)
err = pdfService.SavePDFToFile(pdfData, "invoice-draft.pdf")
```

### 3. Generate Final PDF (immutable, no watermark)

```go
pdfData, err := pdfService.GenerateFinalPDF(ctx, invoice)
if err != nil {
    log.Fatal(err)
}
```

### 4. Generate Credit Note PDF

```go
pdfData, err := pdfService.GenerateCreditNotePDF(
    ctx, 
    creditNoteInvoice,
    "INV-2026-001", // Original invoice number
    "Customer dissatisfaction", // Reason
)
```

## Template Data Structure

The template receives the following data:

```go
type InvoicePDFData struct {
    Invoice               *entities.Invoice
    Organization          interface{} // Organization details
    Client                interface{} // Client details
    CostProvider          interface{} // Cost provider/insurance
    Sessions              []map[string]interface{} // Session list
    InvoiceItems          []entities.InvoiceItem
    IsDraft               bool // Shows "ENTWURF" watermark
    IsCreditNote          bool // Shows credit note header
    OriginalInvoiceNumber string // For credit notes
    CreditNoteReason      string // Credit note reason
    VATExempt             bool // Shows VAT exemption notice
    VATExemptionText      string // e.g., "§4 Nr. 14 UStG"
    PaymentDueDate        string // Formatted due date
}
```

## Template Location

Default: `statics/pdf_templates/invoice_units_template.html`

Override via environment variable:
```bash
export INVOICE_TEMPLATE_PATH=/path/to/custom/template.html
```

## Features

### Draft Watermark

Drafts display a large diagonal "ENTWURF" watermark:

```html
{{if .IsDraft}}
<div class="draft-watermark">ENTWURF</div>
{{end}}
```

### VAT Exemption Notice

When invoice items are VAT-exempt:

```html
{{if .VATExempt}}
<div class="vat-exemption-notice">
    ℹ️ {{.VATExemptionText}}
</div>
{{end}}
```

### Payment Information

Final invoices show payment details:

```html
{{if not .IsDraft}}
<div class="payment-info">
    <h4>Zahlungsinformationen</h4>
    <p><strong>Fällig bis:</strong> {{.PaymentDueDate}}</p>
    ...
</div>
{{end}}
```

### Credit Note Header

Credit notes display a highlighted header:

```html
{{if .IsCreditNote}}
<div class="credit-note-header">
    <strong>GUTSCHRIFT</strong> zur Originalrechnung Nr. {{.OriginalInvoiceNumber}}<br>
    Grund: {{.CreditNoteReason}}
</div>
{{end}}
```

## Error Handling

All methods return errors that should be handled:

```go
pdfData, err := pdfService.GenerateFinalPDF(ctx, invoice)
if err != nil {
    // Handle error
    // Common issues:
    // - Chrome/Chromium not installed
    // - Template file not found
    // - Missing required invoice data
    return err
}
```

## Performance

- PDF generation takes 1-3 seconds depending on complexity
- Use context with timeout to prevent hanging
- Chromedp context is created per request (stateless)

## Future Enhancements

The following are optional and not required for core functionality:

- Integration with MinIO for document storage
- Document database records with metadata
- Template versioning
- Multiple language support
- Custom template selection per organization

## Integration with Invoice Workflow

PDF generation is **optional** in the invoice workflow:

1. **Draft Creation**: Generate draft PDF with watermark
2. **Draft Editing**: Regenerate draft PDF if needed
3. **Finalization**: Generate final immutable PDF
4. **Credit Notes**: Generate credit note PDF with reference

All PDF generation is done synchronously but can be made async if needed.
