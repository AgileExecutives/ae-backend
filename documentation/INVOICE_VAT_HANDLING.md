# Invoice VAT Handling (Phase 9)

## Overview

The VAT handling system provides comprehensive support for German VAT (Umsatzsteuer) requirements, including healthcare exemptions (§4 Nr.14 UStG), standard rates, and reduced rates. The system automatically calculates VAT, validates configuration, and provides detailed breakdowns for compliance.

## VAT Categories

The system supports three main VAT categories:

### 1. Exempt (Heilberuf) - `exempt_heilberuf`
- **Rate**: 0%
- **Legal Basis**: §4 Nr.14 UStG
- **Exemption Text**: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
- **Use Case**: Healthcare services provided by licensed healthcare professionals (Heilberufe)
- **Description**: "Heilberufliche Leistungen (§4 Nr.14 UStG)"

### 2. Standard Rate - `taxable_standard`
- **Rate**: 19%
- **Legal Basis**: Standard German VAT rate
- **Use Case**: General taxable services and products
- **Description**: "Standard VAT rate"

### 3. Reduced Rate - `taxable_reduced`
- **Rate**: 7%
- **Legal Basis**: Reduced German VAT rate
- **Use Case**: Specific services and products qualifying for reduced rate
- **Description**: "Reduced VAT rate"

## Architecture

### Core Components

1. **VATService** (`modules/client_management/services/vat_service.go`)
   - Central service for VAT management
   - Category definitions and application
   - VAT calculation and validation
   - Breakdown generation

2. **VAT Category Endpoint** (`GET /invoices/vat-categories`)
   - Returns available VAT categories
   - Includes rates and exemption information

3. **Enhanced Invoice Responses**
   - VAT breakdown included in invoice responses
   - Grouped by VAT rate
   - Shows net, tax, and gross amounts per rate

## Usage

### Getting Available VAT Categories

```bash
GET /client-invoices/vat-categories
```

**Response:**
```json
{
  "success": true,
  "message": "VAT categories retrieved successfully",
  "data": {
    "categories": [
      {
        "category": "exempt_heilberuf",
        "rate": 0,
        "is_exempt": true,
        "exemption_text": "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
        "description": "Heilberufliche Leistungen (§4 Nr.14 UStG)"
      },
      {
        "category": "taxable_standard",
        "rate": 19,
        "is_exempt": false,
        "exemption_text": "",
        "description": "Standard VAT rate"
      },
      {
        "category": "taxable_reduced",
        "rate": 7,
        "is_exempt": false,
        "exemption_text": "",
        "description": "Reduced VAT rate"
      }
    ]
  }
}
```

### Creating Draft Invoice with VAT Categories

When creating custom line items, specify the VAT category:

```json
{
  "organization_id": 1,
  "invoice_date": "2024-01-15",
  "custom_line_items": [
    {
      "description": "Consultation",
      "number_units": 1,
      "unit_price": 150.00,
      "vat_category": "exempt_heilberuf"
    },
    {
      "description": "Training Materials",
      "number_units": 2,
      "unit_price": 50.00,
      "vat_category": "taxable_standard"
    },
    {
      "description": "Book",
      "number_units": 1,
      "unit_price": 20.00,
      "vat_category": "taxable_reduced"
    }
  ]
}
```

### Updating Draft Invoice

The same VAT category system applies when updating:

```json
{
  "custom_line_items": [
    {
      "description": "Updated item",
      "number_units": 1,
      "unit_price": 100.00,
      "vat_category": "taxable_standard"
    }
  ]
}
```

### Invoice Response with VAT Breakdown

When retrieving an invoice (`GET /invoices/:id`), the response includes a detailed VAT breakdown:

```json
{
  "success": true,
  "message": "Invoice retrieved successfully",
  "data": {
    "id": 123,
    "invoice_number": "2024-001",
    "sum_amount": 220.00,
    "tax_amount": 19.00,
    "total_amount": 239.00,
    "vat_breakdown": {
      "subtotal": 220.00,
      "total_tax": 19.00,
      "grand_total": 239.00,
      "items": [
        {
          "vat_rate": 0,
          "net_amount": 150.00,
          "tax_amount": 0,
          "gross_amount": 150.00,
          "exemption_text": "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
        },
        {
          "vat_rate": 19,
          "net_amount": 100.00,
          "tax_amount": 19.00,
          "gross_amount": 119.00,
          "exemption_text": ""
        },
        {
          "vat_rate": 7,
          "net_amount": 20.00,
          "tax_amount": 1.40,
          "gross_amount": 21.40,
          "exemption_text": ""
        }
      ]
    },
    "invoice_items": [...]
  }
}
```

## VAT Service Methods

### `GetVATCategories() []VATCategoryInfo`

Returns all available VAT categories with their configuration.

### `ApplyVATCategory(item *entities.InvoiceItem, category VATCategory) error`

Applies a VAT category to an invoice line item:
- Sets VAT rate
- Sets exempt flag
- Adds exemption text if applicable

**Parameters:**
- `item`: Invoice item to modify
- `category`: VAT category to apply

**Returns:** Error if category is invalid

### `GetDefaultVATCategory(itemType string) VATCategory`

Automatically selects appropriate VAT category based on item type:
- Session items → `exempt_heilberuf`
- Extra efforts → `exempt_heilberuf`
- Custom items → `taxable_standard`

### `CalculateInvoiceVAT(items []entities.InvoiceItem) InvoiceVATSummary`

Calculates comprehensive VAT breakdown for an invoice:
- Groups items by VAT rate
- Calculates net, tax, and gross amounts per rate
- Returns summary with totals

**Returns:**
```go
type InvoiceVATSummary struct {
    SubtotalAmount float64
    TaxAmount      float64
    TotalAmount    float64
    VATBreakdown   []VATBreakdownItem
}
```

### `ValidateVATConfiguration(items []entities.InvoiceItem) error`

Validates that all invoice items have proper VAT configuration:
- Non-exempt items must have VAT rate > 0
- Exempt items must have exemption text

**Use case:** Called before finalizing invoice

## Workflow Integration

### Create Draft Invoice
1. User specifies VAT category for custom line items
2. System applies category using `ApplyVATCategory()`
3. Falls back to manual VAT if category invalid
4. Default VAT rate applied if none specified
5. VAT calculated using `CalculateInvoiceVAT()`

### Update Draft Invoice
1. Same category application logic as creation
2. Recalculates totals with VAT service
3. Preserves existing items' VAT configuration

### Finalize Invoice
1. Validates VAT configuration with `ValidateVATConfiguration()`
2. Ensures all items have proper rates/exemptions
3. Prevents finalization if validation fails

### Get Invoice
1. Retrieves invoice with items
2. Calculates VAT breakdown using `CalculateInvoiceVAT()`
3. Includes breakdown in response

## Default Behavior

### Session Items
- Automatically use `exempt_heilberuf` category
- Set via `GetDefaultVATCategory("session")`
- Exemption text automatically added

### Extra Effort Items
- Automatically use `exempt_heilberuf` category
- Set via `GetDefaultVATCategory("extra_effort")`
- Exemption text automatically added

### Custom Line Items
- **With category specified**: Uses specified category
- **Without category**: Falls back to manual VAT settings
- **No manual settings**: Uses organization default rate

## Error Handling

### Invalid Category
If an invalid VAT category is specified, the system:
1. Logs a warning
2. Falls back to manual VAT settings from request
3. Continues processing (graceful degradation)

### Missing VAT Configuration (Finalization)
If VAT is not properly configured on finalization:
- Returns error: "VAT validation failed"
- Lists specific items with issues
- Prevents finalization

### Example Error:
```json
{
  "success": false,
  "message": "VAT validation failed: line item 'Consultation' is missing VAT rate"
}
```

## Validation Rules

1. **Non-exempt items** must have `vat_rate > 0`
2. **Exempt items** must have `vat_exemption_text != ""`
3. All items must have either:
   - Valid VAT category applied, OR
   - Manual VAT configuration (rate/exempt/exemption text)

## Database Schema

### InvoiceItem Fields
```sql
vat_rate           DECIMAL(5,2)  -- VAT rate in percentage (e.g., 19.00)
vat_exempt         BOOLEAN       -- Whether VAT exempt
vat_exemption_text TEXT          -- Legal exemption text (required if exempt)
```

## API Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/client-invoices/vat-categories` | Get available VAT categories |
| POST | `/client-invoices/draft` | Create draft with VAT categories |
| PUT | `/client-invoices/:id` | Update draft with VAT categories |
| GET | `/client-invoices/:id` | Get invoice with VAT breakdown |
| POST | `/client-invoices/:id/finalize` | Finalize with VAT validation |

## Testing

### Test VAT Categories Endpoint
```bash
curl -X GET "http://localhost:9091/client-invoices/vat-categories" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Test Draft with VAT Category
```bash
curl -X POST "http://localhost:9091/client-invoices/draft" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": 1,
    "invoice_date": "2024-01-15",
    "custom_line_items": [
      {
        "description": "Healthcare consultation",
        "number_units": 1,
        "unit_price": 150.00,
        "vat_category": "exempt_heilberuf"
      }
    ]
  }'
```

### Test Invoice Retrieval
```bash
curl -X GET "http://localhost:9091/client-invoices/1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Expected response includes `vat_breakdown` with grouped rates.

## Compliance Notes

### GoBD Compliance
- VAT breakdown provides audit trail
- All rates properly documented
- Exemption legal basis included

### German VAT Law (UStG)
- §4 Nr.14 UStG exemption for healthcare services
- Proper documentation of exemption reasons
- Correct application of standard and reduced rates

## Future Enhancements

1. **Additional Categories**
   - Import/Export VAT handling
   - Reverse charge mechanism
   - Small business exemption (Kleinunternehmerregelung)

2. **Rate Changes**
   - Historical rate tracking
   - Automatic rate updates based on date

3. **Advanced Validation**
   - Customer-specific VAT rules
   - Cross-border VAT validation
   - B2B vs B2C rate differences

## Related Documentation

- [XRechnung Implementation](XRECHNUNG_README.md) - E-invoicing for government customers
- [MinIO Integration](INVOICE_MINIO_INTEGRATION.md) - PDF storage
- [Schema Migration](../SCHEMA_MIGRATION.md) - Database schema details
- [Module Development Guide](MODULE_DEVELOPMENT_GUIDE.md) - Development practices

## Summary

Phase 9 implements a comprehensive VAT handling system with:
- ✅ Three VAT categories (exempt, standard 19%, reduced 7%)
- ✅ Automatic category application
- ✅ VAT calculation and breakdown
- ✅ Validation before finalization
- ✅ Detailed response with grouped rates
- ✅ Full German compliance support
- ✅ Graceful error handling
- ✅ GoBD audit trail support
