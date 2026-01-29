# Invoice Schema Migration Guide

## Overview

The invoice system has been refactored to support multiple clients per invoice through a new junction table `client_invoices`. This allows for more flexible invoice structures and better data organization.

## Schema Changes

### 1. Invoice Table (`invoices`)

**Removed Fields:**
- `client_id` - Moved to `client_invoices` table
- `cost_provider_id` - Moved to `client_invoices` table
- Relationship: `Client` - Now accessed through `ClientInvoices`
- Relationship: `CostProvider` - Now accessed through `ClientInvoices`

**Added Fields:**
- Relationship: `ClientInvoices []ClientInvoice` - New many-to-many relationship

**Retained Fields:**
- `id`, `tenant_id`, `user_id`, `organization_id`
- `invoice_date`, `invoice_number`
- `number_units`, `sum_amount`, `tax_amount`, `total_amount`
- `payed_date`, `status`, `num_reminders`, `latest_reminder`
- `document_id`
- Timestamps: `created_at`, `updated_at`, `deleted_at`

### 2. InvoiceItem Table (`invoice_items`)

**Removed Fields:**
- `session_id` - Moved to `client_invoices` table
- Relationship: `Session` - Now accessed through `client_invoices`

**Retained Fields:**
- `id`, `invoice_id`
- Timestamps: `created_at`, `updated_at`, `deleted_at`

### 3. New ClientInvoice Table (`client_invoices`)

**Purpose:** Junction table linking invoices to clients and their sessions

**Fields:**
- `id` (primary key)
- `invoice_id` - References `invoices.id`
- `client_id` - References `clients.id`
- `cost_provider_id` - References `cost_providers.id`
- `session_id` - References `sessions.id` (unique index)
- `invoice_item_id` - References `invoice_items.id`
- Timestamps: `created_at`, `updated_at`, `deleted_at`

**Relationships:**
- `Invoice *Invoice` - Belongs to invoice
- `Client *Client` - Belongs to client
- `CostProvider *CostProvider` - Belongs to cost provider
- `Session *Session` - Belongs to session
- `InvoiceItem *InvoiceItem` - Belongs to invoice item

**Indexes:**
- `idx_client_invoice_invoice` on `invoice_id`
- `idx_client_invoice_client` on `client_id`
- `idx_client_invoice_cost_provider` on `cost_provider_id`
- `idx_client_invoice_session` (unique) on `session_id`
- `idx_client_invoice_item` on `invoice_item_id`

## API Response Changes

### Old InvoiceResponse Structure

```json
{
  "id": 1,
  "client_id": 5,
  "client": { "id": 5, "first_name": "John", ... },
  "cost_provider_id": 10,
  "cost_provider": { "id": 10, "organization": "Provider Corp", ... },
  "invoice_items": [
    {
      "id": 1,
      "session_id": 100,
      "session": { "id": 100, ... }
    }
  ],
  ...
}
```

### New InvoiceResponse Structure

```json
{
  "id": 1,
  "clients": [
    {
      "client_id": 5,
      "client": { "id": 5, "first_name": "John", ... },
      "cost_provider_id": 10,
      "cost_provider": { "id": 10, "organization": "Provider Corp", ... },
      "sessions": [
        { "id": 100, "original_date": "2025-01-15", ... },
        { "id": 101, "original_date": "2025-01-22", ... }
      ]
    }
  ],
  ...
}
```

## Migration SQL (PostgreSQL)

```sql
-- Step 1: Create new client_invoices table
CREATE TABLE client_invoices (
    id SERIAL PRIMARY KEY,
    invoice_id INTEGER NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    client_id INTEGER NOT NULL REFERENCES clients(id),
    cost_provider_id INTEGER NOT NULL REFERENCES cost_providers(id),
    session_id INTEGER NOT NULL REFERENCES sessions(id),
    invoice_item_id INTEGER NOT NULL REFERENCES invoice_items(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_client_invoice_invoice ON client_invoices(invoice_id);
CREATE INDEX idx_client_invoice_client ON client_invoices(client_id);
CREATE INDEX idx_client_invoice_cost_provider ON client_invoices(cost_provider_id);
CREATE UNIQUE INDEX idx_client_invoice_session ON client_invoices(session_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_invoice_item ON client_invoices(invoice_item_id);

-- Step 2: Migrate existing data
-- For each invoice, create client_invoices entries linking invoice -> client -> session -> invoice_item
INSERT INTO client_invoices (invoice_id, client_id, cost_provider_id, session_id, invoice_item_id, created_at, updated_at)
SELECT 
    i.id AS invoice_id,
    i.client_id,
    i.cost_provider_id,
    ii.session_id,
    ii.id AS invoice_item_id,
    ii.created_at,
    ii.updated_at
FROM invoices i
INNER JOIN invoice_items ii ON ii.invoice_id = i.id
WHERE i.deleted_at IS NULL AND ii.deleted_at IS NULL;

-- Step 3: Remove old columns from invoices table
ALTER TABLE invoices DROP COLUMN client_id;
ALTER TABLE invoices DROP COLUMN cost_provider_id;

-- Step 4: Remove session_id from invoice_items table
ALTER TABLE invoice_items DROP COLUMN session_id;
```

## Code Changes Summary

### Entity Changes

1. **Invoice** entity:
   - Removed `ClientID`, `Client`, `CostProviderID`, `CostProvider` fields
   - Added `ClientInvoices []ClientInvoice` relationship

2. **InvoiceItem** entity:
   - Removed `SessionID`, `Session` fields

3. **ClientInvoice** entity (new):
   - Links `invoice_id`, `client_id`, `cost_provider_id`, `session_id`, `invoice_item_id`

### Service Changes

Updated methods in `InvoiceService`:
- `CreateInvoice()` - Now creates `ClientInvoice` entries instead of setting direct foreign keys
- `CreateInvoiceDirect()` - Creates invoice items and client_invoices for each session
- `GetInvoiceByID()` - Preloads `ClientInvoices` with nested relationships
- `GetInvoices()` - Preloads `ClientInvoices` with nested relationships
- `UpdateInvoice()` - Updates `ClientInvoices` when sessions change
- `GetClientsWithUnbilledSessions()` - Queries `client_invoices` instead of `invoice_items`

### Response Changes

- `InvoiceResponse` now has `Clients []ClientInvoiceResponse` instead of single `Client` and `CostProvider`
- `ClientInvoiceResponse` groups sessions by client with their cost provider
- `InvoiceItemResponse` simplified (removed `SessionID` and `Session`)

## Testing

After migration, verify:

1. **Create Invoice**: POST `/api/v1/client-invoices/generate`
   - Check that `client_invoices` entries are created
   - Verify response has `clients` array with sessions grouped by client

2. **Get Invoices**: GET `/api/v1/client-invoices`
   - Verify `clients` array is populated correctly
   - Check that sessions are associated with correct client

3. **Get Single Invoice**: GET `/api/v1/client-invoices/:id`
   - Verify all client data loads correctly
   - Check nested sessions

4. **Update Invoice**: PUT `/api/v1/client-invoices/:id`
   - Verify session updates work correctly
   - Check that client_invoices are updated

5. **Unbilled Sessions**: GET `/api/v1/client-invoices/unbilled-sessions`
   - Verify query uses `client_invoices` table
   - Check that sessions already in invoices are excluded

## Rollback Procedure

If you need to rollback:

```sql
-- 1. Add back columns to invoices
ALTER TABLE invoices ADD COLUMN client_id INTEGER;
ALTER TABLE invoices ADD COLUMN cost_provider_id INTEGER;

-- 2. Add back session_id to invoice_items
ALTER TABLE invoice_items ADD COLUMN session_id INTEGER;

-- 3. Restore data from client_invoices
UPDATE invoices i
SET client_id = ci.client_id,
    cost_provider_id = ci.cost_provider_id
FROM (
    SELECT DISTINCT ON (invoice_id) invoice_id, client_id, cost_provider_id
    FROM client_invoices
    ORDER BY invoice_id, id
) ci
WHERE i.id = ci.invoice_id;

UPDATE invoice_items ii
SET session_id = ci.session_id
FROM client_invoices ci
WHERE ii.id = ci.invoice_item_id;

-- 4. Drop client_invoices table
DROP TABLE client_invoices;

-- 5. Restore code from git
git checkout <previous-commit>
```

## Benefits of New Structure

1. **Flexibility**: Invoices can now have multiple clients (future enhancement)
2. **Better Data Organization**: All invoice-client-session relationships in one place
3. **Clearer Relationships**: Explicit junction table makes relationships obvious
4. **Cost Provider Tracking**: Each client-session pair can have its own cost provider
5. **Easier Queries**: JOIN through client_invoices is more intuitive than multiple foreign keys

## Notes

- The current implementation still creates one invoice per client, but the schema now supports multiple clients per invoice
- Session uniqueness is still enforced - each session can only belong to one invoice
- All existing queries now go through the `client_invoices` table
- Document metadata no longer includes `client_id` since invoices can have multiple clients
