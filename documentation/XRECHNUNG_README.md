# XRechnung Implementation Guide

## Overview

XRechnung is the German standard for electronic invoicing to government entities. This implementation provides full support for generating XRechnung-compliant XML documents in UBL 2.1 format.

## Features

✅ **XRechnung 3.0.1 Compliance** - Implements the latest XRechnung standard  
✅ **UBL 2.1 XML Format** - Universal Business Language version 2.1  
✅ **PEPPOL BIS Billing 3.0** - European e-procurement standard  
✅ **Leitweg-ID Routing** - German government routing identifier  
✅ **VAT Exemption Support** - Healthcare exemption under §4 Nr.14 UStG  
✅ **Multiple Tax Rates** - Support for standard (19%), reduced (7%), and exempt rates  
✅ **Credit Note Support** - Generate XML for credit notes (type code 384)  

## Database Schema

### Government Customer Fields (CostProvider)

```sql
ALTER TABLE cost_providers 
  ADD COLUMN IF NOT EXISTS is_government_customer BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS leitweg_id VARCHAR(100),
  ADD COLUMN IF NOT EXISTS authority_name VARCHAR(255),
  ADD COLUMN IF NOT EXISTS reference_number VARCHAR(100);
```

**Field Descriptions:**
- `is_government_customer`: Flag to identify government entities
- `leitweg_id`: German government routing ID (required for XRechnung)
- `authority_name`: Official name of the government authority
- `reference_number`: Optional reference like cost center or department code

## API Endpoint

### Export XRechnung XML

```http
GET /invoices/{id}/xrechnung
```

**Authentication:** Bearer Token required

**Preconditions:**
1. Invoice status must be `sent`, `paid`, or `overdue` (not `draft` or `cancelled`)
2. Customer must be a government customer (`is_government_customer = true`)
3. Customer must have a valid Leitweg-ID

**Response:**
- **Success (200):** XML file download with filename `xrechnung_{invoice_number}.xml`
- **Error (400):** Invalid invoice status or missing government customer data
- **Error (404):** Invoice not found
- **Error (500):** XML generation failed

**Example Request:**
```bash
curl -X GET "https://api.example.com/invoices/123/xrechnung" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o xrechnung_INV-2025-001.xml
```

## XRechnung XML Structure

### Key Elements

1. **Invoice Identification**
   - CustomizationID: `urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0`
   - ProfileID: `urn:fdc:peppol.eu:2017:poacc:billing:01:1.0`
   - Invoice Number: Unique invoice identifier
   - Issue Date: ISO 8601 format (YYYY-MM-DD)
   - Due Date: Calculated from payment terms

2. **Supplier Information (AccountingSupplierParty)**
   - Tax ID (VAT number)
   - Legal name
   - Postal address
   - Contact information (email, phone)

3. **Customer Information (AccountingCustomerParty)**
   - Leitweg-ID as endpoint (scheme ID: 0204)
   - Authority name
   - Postal address
   - Country code: DE (Germany)

4. **Payment Information**
   - Payment means code: 58 (SEPA credit transfer)
   - IBAN
   - BIC
   - Account holder name

5. **Tax Information**
   - Tax subtotals by rate
   - VAT exemption reasons (e.g., §4 Nr.14 UStG)
   - Total tax amount

6. **Line Items**
   - Description and name
   - Quantity with UN/CECE unit codes
     - HUR: Hours (for therapy sessions)
     - C62: Units/pieces (for extra efforts)
   - Unit price and total amount
   - Tax category per line

## Code Components

### XRechnungService

**Location:** `internal/services/xrechnung_service.go`

**Key Methods:**
- `GenerateXRechnungXML()` - Main entry point for XML generation
- `buildSupplierParty()` - Maps organization data to UBL supplier structure
- `buildCustomerParty()` - Maps cost provider data to UBL customer structure
- `buildPaymentMeans()` - Generates SEPA payment information
- `buildTaxTotal()` - Calculates and groups VAT by rate
- `buildMonetaryTotal()` - Invoice totals (subtotal, tax, total)
- `buildInvoiceLines()` - Invoice line items with tax categories

### Invoice Handler

**Location:** `modules/client_management/handlers/invoice_handler.go`

**New Method:** `ExportXRechnung()`

**Validation Flow:**
1. Parse invoice ID from URL parameter
2. Authenticate user and get tenant context
3. Fetch invoice with relations (ClientInvoices, InvoiceItems)
4. Validate invoice status (sent/paid/overdue)
5. Fetch cost provider from first ClientInvoice relationship
6. Validate government customer requirements
7. Fetch organization details
8. Generate XRechnung XML
9. Return XML file for download

### Routes

**Location:** `modules/client_management/routes/routes.go`

```go
invoices.GET("/:id/xrechnung", rp.invoiceHandler.ExportXRechnung)
```

## Usage Examples

### 1. Mark Customer as Government Entity

```go
costProvider := &models.CostProvider{
    Organization:          "Bundesministerium für Gesundheit",
    IsGovernmentCustomer:  true,
    LeitwegID:            "99123-ABC123-45",
    AuthorityName:        "Bundesministerium für Gesundheit",
    ReferenceNumber:      "KST-2025-001",
    StreetAddress:        "Wilhelmstraße 49",
    Zip:                  "10117",
    City:                 "Berlin",
}
```

### 2. Finalize Invoice for Government Customer

```bash
POST /invoices/123/finalize
```

The finalization endpoint validates that government customers have a Leitweg-ID before allowing finalization.

### 3. Export XRechnung XML

```bash
GET /invoices/123/xrechnung
```

Downloads: `xrechnung_INV-2025-001.xml`

### 4. Sample XML Output Structure

```xml
<?xml version="1.0" encoding="UTF-8"?>
<ubl:Invoice xmlns:ubl="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
             xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
             xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2">
  <cbc:CustomizationID>urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0</cbc:CustomizationID>
  <cbc:ProfileID>urn:fdc:peppol.eu:2017:poacc:billing:01:1.0</cbc:ProfileID>
  <cbc:ID>INV-2025-001</cbc:ID>
  <cbc:IssueDate>2025-01-08</cbc:IssueDate>
  <cbc:DueDate>2025-01-22</cbc:DueDate>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cbc:BuyerReference>99123-ABC123-45</cbc:BuyerReference>
  
  <cac:AccountingSupplierParty>
    <!-- Supplier/Organization details -->
  </cac:AccountingSupplierParty>
  
  <cac:AccountingCustomerParty>
    <cac:Party>
      <cbc:EndpointID schemeID="0204">99123-ABC123-45</cbc:EndpointID>
      <!-- Government authority details -->
    </cac:Party>
  </cac:AccountingCustomerParty>
  
  <cac:PaymentMeans>
    <cbc:PaymentMeansCode>58</cbc:PaymentMeansCode>
    <!-- SEPA payment details -->
  </cac:PaymentMeans>
  
  <cac:TaxTotal>
    <!-- VAT breakdown by rate -->
  </cac:TaxTotal>
  
  <cac:LegalMonetaryTotal>
    <!-- Invoice totals -->
  </cac:LegalMonetaryTotal>
  
  <cac:InvoiceLine>
    <!-- Line items -->
  </cac:InvoiceLine>
</ubl:Invoice>
```

## Validation Rules

### Pre-Finalization
- Government customers must have `leitweg_id` populated
- Validation implemented in `FinalizeInvoice()` service method

### Pre-Export
1. **Invoice Status Check:**
   - Must be: `sent`, `paid`, or `overdue`
   - Cannot be: `draft` or `cancelled`

2. **Government Customer Check:**
   - `is_government_customer` must be `true`
   - `leitweg_id` must not be empty

3. **Organization Data:**
   - Must have valid organization with address
   - Bank details recommended for payment means

## Error Handling

| Error | HTTP Status | Description |
|-------|-------------|-------------|
| Invalid invoice ID | 400 | ID parameter is not a valid integer |
| Invoice not found | 404 | Invoice doesn't exist for this tenant |
| Wrong status | 400 | Invoice is still draft or cancelled |
| Not government customer | 400 | Customer is not marked as government entity |
| Missing Leitweg-ID | 400 | Government customer has no Leitweg-ID |
| Organization not found | 500 | Invoice references invalid organization |
| XML generation failed | 500 | Internal error during XML creation |

## Testing

### Manual Test Checklist

1. ✅ Create cost provider as government customer with Leitweg-ID
2. ✅ Create draft invoice for government customer
3. ✅ Finalize invoice (validates Leitweg-ID presence)
4. ✅ Export XRechnung XML via API
5. ✅ Verify XML structure and content
6. ✅ Test with VAT-exempt items (§4 Nr.14 UStG)
7. ✅ Test with multiple tax rates
8. ✅ Test credit note export (type code 384)
9. ✅ Verify error handling for non-government customers
10. ✅ Verify error handling for draft invoices

### XML Validation

The generated XML should be validated against:
- XRechnung 3.0.1 specification
- UBL 2.1 schema
- PEPPOL BIS Billing 3.0 rules

Use official validators:
- KoSIT XRechnung Validator
- PEPPOL Validator

## Compliance Notes

### XRechnung 3.0.1
- Mandatory for German public procurement since November 27, 2020
- Based on European standard EN 16931
- Implements CEN semantic data model

### Leitweg-ID Format
Format: `{Grobadressierung}-{Feinadressierung}-{Prüfziffer}`
- Example: `99123-ABC123-45`
- Used for electronic invoice routing in Germany
- Scheme ID: 0204 (as per XRechnung specification)

### VAT Exemption §4 Nr.14 UStG
German tax exemption for healthcare services:
- Category ID: E (Exempt)
- Percent: omitted
- Tax exemption reason: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"

## Future Enhancements

- [ ] XML schema validation before returning
- [ ] Support for CII (Cross Industry Invoice) format as alternative to UBL
- [ ] Attachment support (e.g., signed PDFs)
- [ ] Digital signature (XAdES or PAdES)
- [ ] Automatic upload to government portals
- [ ] Multi-language support for descriptions
- [ ] Advanced line item grouping
- [ ] Dunning procedures integration

## References

- [XRechnung Standard](https://www.xrechnung.de/)
- [PEPPOL BIS Billing 3.0](https://docs.peppol.eu/poacc/billing/3.0/)
- [UBL 2.1 Specification](http://docs.oasis-open.org/ubl/UBL-2.1.html)
- [EN 16931 European Standard](https://ec.europa.eu/digital-building-blocks/sites/display/DIGITAL/Invoice+formats)
- [Leitweg-ID Information](https://www.e-rechnung-bund.de/)

---

**Last Updated:** January 8, 2026  
**Version:** 1.0  
**Status:** Production Ready
