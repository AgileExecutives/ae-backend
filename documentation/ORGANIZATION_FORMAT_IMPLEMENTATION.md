# Organization Format Settings - Implementation Summary

## Overview
Added internationalization (i18n) support for date, time, and amount formatting in the Organization model. Invoices now use organization-specific formatting settings when generating PDFs.

## Changes Made

### 1. Database Schema
**File:** N/A (requires migration)

Added four new columns to the `organizations` table:
- `locale` (VARCHAR(10), default: 'de-DE')
- `date_format` (VARCHAR(50), default: '02.01.2006')  
- `time_format` (VARCHAR(50), default: '15:04')
- `amount_format` (VARCHAR(50), default: 'de')

### 2. Organization Model
**File:** [base-server/internal/models/organization.go](base-server/internal/models/organization.go)

- Added format fields to `Organization` struct
- Added format fields to `CreateOrganizationRequest`
- Added format fields to `UpdateOrganizationRequest`  
- Added format fields to `OrganizationResponse`
- Updated `ToResponse()` method to include format fields

### 3. Formatting Utilities
**File:** [base-server/pkg/formatting/formatter.go](base-server/pkg/formatting/formatter.go) (NEW)

Created new formatting package with:
- `Formatter` struct to hold formatting settings
- `NewFormatter()` - Creates formatter with organization settings
- `FormatDate()` - Formats dates according to configured format
- `FormatTime()` - Formats times (24h/12h support)
- `FormatDateTime()` - Formats date and time together
- `FormatAmount()` - Formats monetary amounts (German: 1.234,56 / English: 1,234.56)
- `GetSupportedDateFormats()` - Returns available date format examples
- `GetSupportedTimeFormats()` - Returns available time format examples  
- `GetSupportedAmountFormats()` - Returns available amount format examples
- `GetSupportedLocales()` - Returns list of supported locales

### 4. Invoice Handler
**File:** [unburdy_server/modules/client_management/handlers/invoice_handler.go](unburdy_server/modules/client_management/handlers/invoice_handler.go)

- Added import for `github.com/ae-base-server/pkg/formatting`
- Resolved TODO at line 228
- Creates formatter from `invoice.Organization` settings
- Formats invoice dates and amounts before passing to PDF template:
  - `invoice_date` - formatted using organization's date format
  - `net_total`, `tax_amount`, `gross_total` - formatted using organization's amount format
- Added `formatSessions()` helper function to format session dates
- Template data now includes:
  - Full `organization` object (instead of just ID)
  - `formatter` settings for template access
  - Formatted session data with both formatted strings and raw values

### 5. Organization Handler
**File:** [base-server/internal/organizations/handlers/organization_handler.go](base-server/internal/organizations/handlers/organization_handler.go)

- Added import for `github.com/ae-base-server/pkg/formatting`
- Added `GetSupportedFormats()` handler to return available format options

### 6. Organization Routes
**File:** [base-server/internal/organizations/routes/routes.go](base-server/internal/organizations/routes/routes.go)

- Added `GET /organizations/supported-formats` endpoint

## API Endpoints

### Get Supported Formats
```
GET /organizations/supported-formats
```

Response:
```json
{
  "success": true,
  "message": "Supported formats retrieved successfully",
  "data": {
    "date_formats": {
      "02.01.2006": "31.12.2024",
      "01/02/2006": "12/31/2024",
      "02/01/2006": "31/12/2024",
      "2006-01-02": "2024-12-31",
      "Jan 02, 2006": "Dec 31, 2024",
      "02 Jan 2006": "31 Dec 2024",
      "Monday, 02.01": "Tuesday, 31.12"
    },
    "time_formats": {
      "15:04": "14:30",
      "15:04:05": "14:30:00",
      "3:04 PM": "2:30 PM",
      "03:04 PM": "02:30 PM",
      "3:04:05 PM": "2:30:00 PM",
      "03:04:05 PM": "02:30:00 PM"
    },
    "amount_formats": {
      "de": "1.234,56",
      "de-DE": "1.234,56",
      "de-AT": "1.234,56",
      "de-CH": "1'234,56",
      "en": "1,234.56",
      "en-US": "1,234.56",
      "en-GB": "1,234.56"
    },
    "locales": [
      "de-DE",
      "de-AT",
      "de-CH",
      "en-US",
      "en-GB",
      "fr-FR",
      "es-ES",
      "it-IT"
    ]
  }
}
```

### Create/Update Organization with Format Settings
```
POST /organizations
PUT /organizations/:id
```

Request body can now include:
```json
{
  "name": "My Organization",
  "locale": "en-US",
  "date_format": "01/02/2006",
  "time_format": "3:04 PM",
  "amount_format": "en",
  ...
}
```

## Migration Required

Run this SQL migration on your database:

```sql
ALTER TABLE organizations 
ADD COLUMN locale VARCHAR(10) DEFAULT 'de-DE',
ADD COLUMN date_format VARCHAR(50) DEFAULT '02.01.2006',
ADD COLUMN time_format VARCHAR(50) DEFAULT '15:04',
ADD COLUMN amount_format VARCHAR(50) DEFAULT 'de';
```

## Template Updates

Invoice templates now receive formatted data:

```handlebars
<!-- Old (hardcoded format) -->
{{ invoice.invoice_date }}  <!-- Raw: 2024-12-31T00:00:00Z -->

<!-- New (formatted) -->
{{ invoice.invoice_date }}  <!-- Formatted: 31.12.2024 or 12/31/2024 -->
{{ invoice.net_total }}     <!-- Formatted: 1.234,56 or 1,234.56 -->

<!-- Sessions also formatted -->
{{#each sessions}}
  {{ this.original_date }}  <!-- Formatted: 15.12.2024 -->
  {{ this.original_date_raw }}  <!-- Raw for sorting: 2024-12-15T00:00:00Z -->
{{/each}}

<!-- Organization settings available -->
{{ organization.name }}
{{ organization.date_format }}
{{ organization.time_format }}
```

## Testing

### Test Format Settings
1. Get supported formats:
   ```bash
   curl http://localhost:8080/api/organizations/supported-formats \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

2. Update organization with US formatting:
   ```bash
   curl -X PUT http://localhost:8080/api/organizations/1 \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "locale": "en-US",
       "date_format": "01/02/2006",
       "time_format": "3:04 PM",
       "amount_format": "en"
     }'
   ```

3. Generate invoice and verify PDF uses US formatting

### Expected Results

**German Format (default):**
- Date: 31.12.2024
- Time: 14:30
- Amount: 1.234,56 €

**US Format:**
- Date: 12/31/2024
- Time: 2:30 PM
- Amount: $1,234.56

## Benefits

1. **Internationalization**: Support for multiple locales and formats
2. **Per-Organization Settings**: Each organization can have different formatting
3. **Consistency**: All dates/times/amounts formatted uniformly in invoices
4. **User Experience**: Invoices appear in familiar local format
5. **Extensibility**: Easy to add more locales and formats
6. **Template Flexibility**: Templates receive both formatted and raw values

## Build Status

✅ base-server builds successfully  
✅ unburdy_server builds successfully  
✅ No compilation errors

## Documentation

See [ORGANIZATION_FORMAT_MIGRATION.md](ORGANIZATION_FORMAT_MIGRATION.md) for detailed migration guide.
