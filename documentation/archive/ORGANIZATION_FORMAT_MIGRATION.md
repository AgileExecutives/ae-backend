# Organization Format Settings Migration

## Overview
This migration adds internationalization (i18n) support for date, time, and amount formatting in the Organization model.

## Schema Changes

### New Columns in `organizations` Table

```sql
ALTER TABLE organizations 
ADD COLUMN locale VARCHAR(10) DEFAULT 'de-DE',
ADD COLUMN date_format VARCHAR(50) DEFAULT '02.01.2006',
ADD COLUMN time_format VARCHAR(50) DEFAULT '15:04',
ADD COLUMN amount_format VARCHAR(50) DEFAULT 'de';
```

### Column Descriptions

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `locale` | VARCHAR(10) | 'de-DE' | Locale identifier (e.g., de-DE, en-US, en-GB) |
| `date_format` | VARCHAR(50) | '02.01.2006' | Go time format string for dates |
| `time_format` | VARCHAR(50) | '15:04' | Go time format string for times (24h default) |
| `amount_format` | VARCHAR(50) | 'de' | Amount formatting style (de=1.234,56 / en=1,234.56) |

## Format Examples

### Date Formats
- `02.01.2006` → 31.12.2024 (German)
- `01/02/2006` → 12/31/2024 (US)
- `02/01/2006` → 31/12/2024 (UK)
- `2006-01-02` → 2024-12-31 (ISO)
- `Jan 02, 2006` → Dec 31, 2024
- `02 Jan 2006` → 31 Dec 2024

### Time Formats
- `15:04` → 14:30 (24-hour)
- `15:04:05` → 14:30:00 (24-hour with seconds)
- `3:04 PM` → 2:30 PM (12-hour)
- `03:04 PM` → 02:30 PM (12-hour with leading zero)
- `3:04:05 PM` → 2:30:00 PM (12-hour with seconds)

### Amount Formats
- `de` → 1.234,56 (German: dot separator, comma decimal)
- `en` → 1,234.56 (English: comma separator, dot decimal)

### Supported Locales
- `de-DE` - German (Germany)
- `de-AT` - German (Austria)
- `de-CH` - German (Switzerland)
- `en-US` - English (United States)
- `en-GB` - English (United Kingdom)
- `fr-FR` - French
- `es-ES` - Spanish
- `it-IT` - Italian

## Code Changes

### 1. Organization Model
File: `/base-server/internal/models/organization.go`

Added fields:
```go
Locale       string `gorm:"size:10;default:'de-DE'" json:"locale"`
DateFormat   string `gorm:"size:50;default:'02.01.2006'" json:"date_format"`
TimeFormat   string `gorm:"size:50;default:'15:04'" json:"time_format"`
AmountFormat string `gorm:"size:50;default:'de'" json:"amount_format"`
```

### 2. Formatting Utilities
File: `/base-server/pkg/formatting/formatter.go`

New formatter package with:
- `NewFormatter()` - Creates formatter with organization settings
- `FormatDate()` - Formats dates
- `FormatTime()` - Formats times
- `FormatDateTime()` - Formats date and time
- `FormatAmount()` - Formats monetary amounts
- Helper functions for supported formats and locales

### 3. Invoice Handler Updates
File: `/unburdy_server/modules/client_management/handlers/invoice_handler.go`

- Imports `github.com/ae-base-server/pkg/formatting`
- Creates formatter from invoice.Organization settings
- Formats invoice dates and amounts before passing to template
- Formats session dates and amounts in template data
- Passes full organization data to templates

## API Changes

### Organization Endpoints

#### Create Organization
```json
POST /organizations
{
  "name": "My Organization",
  "locale": "de-DE",
  "date_format": "02.01.2006",
  "time_format": "15:04",
  "amount_format": "de"
}
```

#### Update Organization
```json
PUT /organizations/:id
{
  "locale": "en-US",
  "date_format": "01/02/2006",
  "time_format": "3:04 PM",
  "amount_format": "en"
}
```

#### Response Format
```json
{
  "id": 1,
  "name": "My Organization",
  "locale": "de-DE",
  "date_format": "02.01.2006",
  "time_format": "15:04",
  "amount_format": "de",
  ...
}
```

## Template Changes

Invoice templates now receive:

```json
{
  "invoice": {
    "invoice_number": "INV-2024-123",
    "invoice_date": "31.12.2024",      // Formatted
    "net_total": "1.234,56",           // Formatted
    "tax_amount": "234,56",            // Formatted
    "gross_total": "1.469,12",         // Formatted
    "tax_rate": 19.0
  },
  "sessions": [
    {
      "original_date": "15.12.2024",   // Formatted
      "original_date_raw": "2024-12-15T00:00:00Z", // Raw for sorting
      "start_time": "14:30",           // Formatted
      "start_time_raw": "2024-12-15T14:30:00Z",
      "unit_price": "150,00",          // Formatted
      "unit_price_raw": 150.00,        // Raw for calculations
      "total_amount": "300,00",        // Formatted
      "total_amount_raw": 300.00
    }
  ],
  "organization": {
    // Full organization object with all fields
  },
  "formatter": {
    "date_format": "02.01.2006",
    "time_format": "15:04",
    "amount_format": "de",
    "locale": "de-DE"
  }
}
```

## Migration Steps

### Step 1: Apply Database Migration
```bash
# Connect to your database
psql -U your_user -d your_database

# Run migration
ALTER TABLE organizations 
ADD COLUMN locale VARCHAR(10) DEFAULT 'de-DE',
ADD COLUMN date_format VARCHAR(50) DEFAULT '02.01.2006',
ADD COLUMN time_format VARCHAR(50) DEFAULT '15:04',
ADD COLUMN amount_format VARCHAR(50) DEFAULT 'de';
```

### Step 2: Update Code
All code changes are already in place:
- Organization model updated
- Formatting utilities created
- Invoice handler updated

### Step 3: Build and Deploy
```bash
cd base-server
make build

cd ../unburdy_server
make build
```

### Step 4: Verify
1. Check existing organizations have default values:
```sql
SELECT id, name, locale, date_format, time_format, amount_format 
FROM organizations;
```

2. Test organization update:
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

3. Generate invoice PDF and verify formatting

## Rollback

If needed, remove the columns:

```sql
ALTER TABLE organizations 
DROP COLUMN locale,
DROP COLUMN date_format,
DROP COLUMN time_format,
DROP COLUMN amount_format;
```

And revert code changes:
```bash
git revert <commit-hash>
```

## Benefits

1. **Internationalization**: Support for multiple locales and formats
2. **Flexibility**: Each organization can have its own formatting preferences
3. **Consistency**: All dates, times, and amounts formatted uniformly across invoices
4. **User Experience**: Invoices appear in familiar local format
5. **Extensibility**: Easy to add more locales and format options

## Testing

Test with different locales:

```bash
# German formatting
curl -X POST http://localhost:8080/api/client-invoices/generate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d @test-invoice-de.json

# US formatting  
curl -X POST http://localhost:8080/api/client-invoices/generate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d @test-invoice-us.json
```

Expected results:
- German: Dates as `31.12.2024`, amounts as `1.234,56`, time as `14:30`
- US: Dates as `12/31/2024`, amounts as `1,234.56`, time as `2:30 PM`
