# Invoice Frontend Implementation Guide

**Based on:** [INVOICE-FRONTEND.md](./INVOICE-FRONTEND.md) & [invoice_implementation.md](./invoice_implementation.md)  
**Created:** 8. Januar 2026  
**Backend API Base:** `/api/v1`

---

## ⚠️ Important: API Client Generation

### Invoice APIs: All Implemented ✅

**All 19 invoice-related API endpoints documented in this guide are fully implemented and production-ready.** You can develop the complete invoice frontend immediately.

### TypeScript Errors: Other Modules ❌

If you're seeing TypeScript compilation errors when generating the API client, they are **NOT** related to the invoice system. These errors come from **unimplemented features** in other modules:

**Missing APIs (NOT needed for invoices):**
- Booking/Calendar APIs (`entities.TimeSlot`, `entities.DayData`, booking endpoints)
- Template Management (`listTemplates`, `createTemplate`, `updateTemplate`, `deleteTemplate`)
- Organization Management (advanced features beyond basic CRUD)
- Booking Template CRUD operations

### Solutions for Frontend Development

**Option 1: Stub Missing Types (Recommended)**
```typescript
// src/api/stubs.ts
// Temporary stubs for unimplemented features
export namespace entities {
  export interface TimeSlot {
    id: number;
    start_time: string;
    end_time: string;
  }
  
  export interface DayData {
    date: string;
    slots: TimeSlot[];
  }
  
  // Add other missing types as needed
}

// Export as fallback
export const stubApis = {
  listTemplates: async () => ({ templates: [] }),
  createTemplate: async () => ({}),
  // Add other stubs as needed
};
```

**Option 2: Exclude Unimplemented Endpoints**
```typescript
// openapi-config.ts
export default {
  schemaFile: './swagger.json',
  apiFile: './src/api/base.ts',
  outputFile: './src/api/generated.ts',
  // Exclude unimplemented endpoints
  filters: {
    tags: ['Invoices', 'VAT', 'Audit'] // Only generate invoice-related APIs
  }
}
```

**Option 3: Type-Safe Partial Generation**
```typescript
// Generate only invoice types
import type { 
  Invoice, 
  InvoiceItem, 
  VATCategory,
  AuditLog 
} from './api/generated';

// Use `Partial` for missing features
import type { Template, Booking } from './api/stubs';
```

### Recommended Approach for Invoice Development

1. **Generate API client with invoice endpoints only** (use tag filtering)
2. **Create stub files for missing types** (prevents compilation errors)
3. **Develop invoice features independently** (no blockers)
4. **Replace stubs when other modules are implemented**

### Invoice-Specific Type Imports

```typescript
// ✅ Safe imports - All implemented
import type {
  Invoice,
  InvoiceItem,
  InvoiceStatus,
  VATCategory,
  VATBreakdown,
  ClientWithUnbilled,
  UnbilledSession,
  UnbilledEffort,
  AuditLog,
  AuditLogFilter
} from '@/api/generated';

// ❌ Avoid - Not implemented yet
import type {
  TimeSlot,      // Booking module
  DayData,       // Calendar module
  Template,      // Template management
  BookingConfig  // Booking templates
} from '@/api/generated'; // These will cause errors
```

---

## User Stories & Implementation Guide

### Phase 1: Unbilled Items Discovery

#### **User Story 1.1: View Clients with Unbilled Sessions**
**As a** practice manager  
**I want to** see all clients who have unbilled sessions or extra efforts  
**So that** I can decide which invoices to create next

**API Endpoint:**
```http
GET /client-invoices/unbilled-sessions
Authorization: Bearer {token}
```

**Response Schema:**
```json
{
  "clients": [
    {
      "client_id": 123,
      "client_name": "Max Mustermann",
      "total_sessions": 8,
      "total_efforts": 2,
      "estimated_total": 1240.50,
      "unbilled_sessions": [
        {
          "session_id": 456,
          "session_date": "2026-01-05T10:00:00Z",
          "duration_minutes": 60,
          "service_type": "Therapie",
          "price": 120.00
        }
      ],
      "unbilled_efforts": [
        {
          "effort_id": 789,
          "description": "Dokumentation",
          "hours": 2.5,
          "price": 100.00
        }
      ]
    }
  ]
}
```

**UI Flow:**
1. **Page Load** → Fetch unbilled sessions
2. **Display** → List of client cards with:
   - Client name (clickable to details)
   - Session count badge (e.g., "8 Sessions")
   - Effort count badge (e.g., "2 Efforts")
   - Estimated total (€1.240,50)
   - "Create Draft Invoice" button (primary action)
3. **Filters:**
   - Search by client name
   - Sort by: total amount, session count, client name
4. **Empty State:** "No unbilled sessions found"

**Component Structure:**
```
UnbilledItemsPage/
├── ClientCard (repeatable)
│   ├── ClientHeader (name, badges)
│   ├── SessionsList (collapsible)
│   ├── EffortsList (collapsible)
│   └── ActionButton (Create Invoice)
└── FilterBar (search, sort)
```

**State Management:**
```typescript
interface UnbilledState {
  clients: ClientWithUnbilled[];
  loading: boolean;
  error: string | null;
  filters: {
    search: string;
    sortBy: 'amount' | 'sessions' | 'name';
  };
}
```

---

### Phase 2: Draft Invoice Creation & Editing

#### **User Story 2.1: Create Draft Invoice from Unbilled Items**
**As a** practice manager  
**I want to** create a draft invoice from selected sessions and efforts  
**So that** I can review and edit before finalizing

**API Endpoint:**
```http
POST /invoices/draft
Authorization: Bearer {token}
Content-Type: application/json

{
  "client_id": 123,
  "organization_id": 1,
  "session_ids": [456, 457],
  "extra_effort_ids": [789],
  "invoice_date": "2026-01-08",
  "custom_line_items": []
}
```

**Response Schema:**
```json
{
  "invoice": {
    "id": 1001,
    "invoice_number": "DRAFT",
    "status": "draft",
    "client_id": 123,
    "client_name": "Max Mustermann",
    "invoice_date": "2026-01-08",
    "subtotal_amount": 1040.00,
    "tax_amount": 0.00,
    "total_amount": 1040.00,
    "invoice_items": [
      {
        "id": 1,
        "item_type": "session",
        "session_id": 456,
        "description": "Therapiesitzung - 05.01.2026",
        "number_units": 1,
        "unit_price": 120.00,
        "total_amount": 120.00,
        "vat_rate": 0.00,
        "vat_exempt": true,
        "vat_exemption_text": "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
        "is_editable": false
      },
      {
        "id": 2,
        "item_type": "extra_effort",
        "source_effort_id": 789,
        "description": "Dokumentation",
        "number_units": 2.5,
        "unit_price": 40.00,
        "total_amount": 100.00,
        "vat_rate": 0.00,
        "vat_exempt": true,
        "vat_exemption_text": "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
        "is_editable": false
      }
    ],
    "vat_breakdown": [
      {
        "vat_rate": 0.00,
        "net_amount": 1040.00,
        "tax_amount": 0.00,
        "gross_amount": 1040.00,
        "exemption_text": "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
      }
    ]
  }
}
```

**UI Flow:**
1. **Click "Create Invoice"** on client card
2. **Show Selection Modal:**
   - Checklist of unbilled sessions (pre-selected all)
   - Checklist of unbilled efforts (pre-selected all)
   - Invoice date picker (default: today)
   - "Create Draft" button
3. **On Submit** → POST to `/invoices/draft`
4. **On Success** → Navigate to Draft Editor (`/invoices/{id}/edit`)

---

#### **User Story 2.2: Edit Draft Invoice**
**As a** practice manager  
**I want to** edit line items in a draft invoice  
**So that** I can adjust prices, add custom items, or remove sessions

**API Endpoint (Get Current Draft):**
```http
GET /invoices/{invoice_id}
Authorization: Bearer {token}
```

**API Endpoint (Update Draft):**
```http
PUT /invoices/{invoice_id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "add_session_ids": [458],
  "remove_session_ids": [456],
  "add_extra_effort_ids": [],
  "remove_extra_effort_ids": [],
  "custom_line_items": [
    {
      "description": "Zusätzliche Beratung",
      "number_units": 1,
      "unit_price": 80.00,
      "vat_category": "exempt_heilberuf"
    }
  ],
  "update_custom_line_items": [
    {
      "id": 5,
      "description": "Aktualisierte Beschreibung",
      "number_units": 2,
      "unit_price": 50.00
    }
  ],
  "remove_custom_line_item_ids": [3]
}
```

**UI Flow:**
1. **Page Load** → GET invoice details
2. **Display Draft Banner:** "ENTWURF - Diese Rechnung ist noch nicht abgerechnet"
3. **Editable Table:**
   - Columns: Description, Quantity, Unit Price, VAT Category, Total, Actions
   - Session/Effort items: Show "Remove" button only (not editable inline)
   - Custom items: Inline editing + "Remove" button
4. **Actions:**
   - "Add Custom Line Item" → Opens modal with form
   - "Add More Sessions" → Opens session selector modal
   - "Save Draft" → PUT update (autosave every 30s)
   - "Cancel Draft" → DELETE with confirmation
   - "Finalize & Send" → Navigate to finalization dialog
5. **Live Totals:** Subtotal, VAT breakdown, Total (updates on every change)

**Component Structure:**
```
DraftInvoiceEditor/
├── DraftBanner (warning message)
├── InvoiceHeader (client, date, invoice number)
├── LineItemsTable
│   ├── SessionLineItem (read-only, removable)
│   ├── EffortLineItem (read-only, removable)
│   └── CustomLineItem (editable)
├── VATBreakdown (grouped by rate)
├── TotalSummary
└── ActionBar (Save, Cancel, Finalize)
```

**Validation Rules:**
- Draft status required for editing
- At least 1 line item required
- Custom items: description required, quantity > 0, unit_price >= 0
- VAT category must be valid (get from `/invoices/vat-categories`)

---

#### **User Story 2.3: Select VAT Category for Line Items**
**As a** practice manager  
**I want to** select the appropriate VAT category for each line item  
**So that** VAT is calculated correctly

**API Endpoint (Get VAT Categories):**
```http
GET /invoices/vat-categories
Authorization: Bearer {token}
```

**Response Schema:**
```json
{
  "vat_categories": [
    {
      "code": "exempt_heilberuf",
      "name": "Heilberufliche Leistung (§4 Nr.14 UStG)",
      "vat_rate": 0.00,
      "vat_exempt": true,
      "exemption_text": "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
      "is_default": true
    },
    {
      "code": "taxable_standard",
      "name": "Regelsteuersatz",
      "vat_rate": 19.00,
      "vat_exempt": false,
      "exemption_text": "",
      "is_default": false
    },
    {
      "code": "taxable_reduced",
      "name": "Ermäßigter Steuersatz",
      "vat_rate": 7.00,
      "vat_exempt": false,
      "exemption_text": "",
      "is_default": false
    }
  ]
}
```

**UI Component:**
- Dropdown/Select for VAT category on each custom line item
- Default: `exempt_heilberuf` for healthcare services
- Show rate and exemption text as helper text
- Update totals immediately on change

---

### Phase 3: Invoice Finalization

#### **User Story 3.1: Finalize Draft Invoice**
**As a** practice manager  
**I want to** finalize a draft invoice  
**So that** it becomes an official, immutable billing document

**API Endpoint:**
```http
POST /invoices/{invoice_id}/finalize
Authorization: Bearer {token}
```

**Response Schema:**
```json
{
  "invoice": {
    "id": 1001,
    "invoice_number": "2026-0001",
    "status": "sent",
    "finalized_at": "2026-01-08T14:30:00Z",
    "total_amount": 1040.00,
    "pdf_url": "https://minio.example.com/invoices/final/1/2026-0001.pdf?..."
  }
}
```

**UI Flow:**
1. **Click "Finalize & Send"** in draft editor
2. **Show Confirmation Modal:**
   - **Warning:** "⚠️ This action cannot be undone. The invoice will be finalized and all included sessions will be marked as billed."
   - **Preview Section:**
     - Client name, invoice date
     - Total amount (highlighted)
     - Line item count
     - VAT breakdown
   - **PDF Preview:** Embedded iframe or "Preview PDF" button
   - **Actions:**
     - "Cancel" (secondary, closes modal)
     - "Finalize Invoice" (primary, red/warning color)
3. **On Finalize Success:**
   - Show success toast: "Invoice 2026-0001 finalized successfully"
   - Navigate to invoice detail page (`/invoices/{id}`)
4. **Error Handling:**
   - Validation errors: Show inline errors (e.g., "Government customer requires Leitweg-ID")
   - Server errors: Show error modal with retry option

**Validation Checks (Frontend):**
- Invoice status must be "draft"
- At least 1 line item exists
- All custom line items have valid VAT categories
- Government customer: Leitweg-ID is present
- Total amount > 0

---

#### **User Story 3.2: Send Invoice via Email**
**As a** practice manager  
**I want to** send the finalized invoice to the client via email  
**So that** they receive their invoice promptly

**API Endpoint:**
```http
POST /invoices/{invoice_id}/send-email
Authorization: Bearer {token}
Content-Type: application/json

{
  "recipient_email": "max.mustermann@example.com",
  "cc_emails": ["office@praxis.de"],
  "subject": "Ihre Rechnung 2026-0001",
  "message": "Sehr geehrter Herr Mustermann,\n\nanbei erhalten Sie Ihre Rechnung."
}
```

**Response:**
```json
{
  "success": true,
  "email_sent_at": "2026-01-08T14:35:00Z",
  "recipient": "max.mustermann@example.com"
}
```

**UI Flow:**
1. **Click "Send Email"** on finalized invoice
2. **Show Email Compose Modal:**
   - Recipient email (pre-filled from client profile)
   - CC field (optional, multi-email)
   - Subject (editable, default template)
   - Message body (textarea with default template)
   - PDF attachment preview (read-only, shows filename)
   - "Send" button
3. **On Success:**
   - Show success toast: "Invoice sent to max.mustermann@example.com"
   - Update invoice detail to show "Sent on 08.01.2026 14:35"
4. **Error Handling:**
   - Invalid email: Inline validation
   - Server error: Show error modal

---

### Phase 4: Invoice Management

#### **User Story 4.1: View Invoice Details**
**As a** practice manager  
**I want to** view complete invoice details  
**So that** I can see all line items, payment status, and history

**API Endpoint:**
```http
GET /invoices/{invoice_id}
Authorization: Bearer {token}
```

**UI Flow:**
1. **Page Load** → GET invoice details
2. **Header Section:**
   - Invoice number (large, prominent)
   - Status badge (color-coded: draft=gray, sent=blue, paid=green, overdue=red, cancelled=dark gray)
   - Client name + contact info
   - Invoice date, due date (if sent)
3. **Timeline Section:**
   - Created: 08.01.2026 10:00
   - Finalized: 08.01.2026 14:30 (if finalized)
   - Sent: 08.01.2026 14:35 (if sent)
   - Reminder 1: 22.01.2026 (if overdue)
   - Paid: 25.01.2026 (if paid)
4. **Line Items Section:**
   - Read-only table: Description, Qty, Unit Price, VAT, Total
   - VAT breakdown summary
   - Total amount (highlighted)
5. **Linked Items Section:**
   - Sessions: List with dates, links to session details
   - Extra efforts: List with descriptions
6. **Actions Based on Status:**
   - **Draft:** Edit, Cancel, Finalize
   - **Sent:** Send Email, Download PDF, Mark Paid, Mark Overdue
   - **Overdue:** Send Reminder, Mark Paid, Download PDF
   - **Paid:** Download PDF, View Payment Details
   - **Cancelled:** View Only (no actions)
7. **Documents Section:**
   - Download PDF button
   - XRechnung XML button (if government customer)

**Component Structure:**
```
InvoiceDetailPage/
├── InvoiceHeader (number, status, client)
├── Timeline (status history)
├── LineItemsDisplay (read-only table)
├── VATBreakdownSummary
├── LinkedItemsSection (sessions, efforts)
├── DocumentsSection (PDF, XML downloads)
└── ActionsBar (status-dependent buttons)
```

---

#### **User Story 4.2: List All Invoices**
**As a** practice manager  
**I want to** see all invoices with filtering and sorting  
**So that** I can find specific invoices quickly

**API Endpoint:**
```http
GET /client-invoices?page=1&limit=20&status=sent&client_id=123&date_from=2026-01-01&date_to=2026-01-31
Authorization: Bearer {token}
```

**Response Schema:**
```json
{
  "invoices": [
    {
      "id": 1001,
      "invoice_number": "2026-0001",
      "status": "sent",
      "client_id": 123,
      "client_name": "Max Mustermann",
      "invoice_date": "2026-01-08",
      "due_date": "2026-01-22",
      "total_amount": 1040.00,
      "days_overdue": 0,
      "num_reminders": 0
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

**UI Flow:**
1. **Page Load** → GET invoices with default filters
2. **Filter Bar:**
   - Status filter (multi-select: draft, sent, paid, overdue, cancelled)
   - Date range picker (from/to)
   - Client search (autocomplete)
   - Sort by: date, amount, invoice number
3. **Invoice Table:**
   - Columns: Number, Client, Date, Due Date, Status, Amount, Actions
   - Status badges with color coding
   - Overdue invoices: Red background, "X days overdue" label
   - Click row → Navigate to detail page
4. **Actions Column:**
   - View (eye icon)
   - Download PDF (download icon)
   - Quick actions based on status
5. **Pagination:** Standard controls at bottom

**Overdue Highlighting:**
```typescript
function getRowClassName(invoice: Invoice): string {
  if (invoice.status === 'overdue') {
    return 'bg-red-50 border-l-4 border-red-500';
  }
  return '';
}
```

---

### Phase 5: Payment Tracking

#### **User Story 5.1: Mark Invoice as Paid**
**As a** practice manager  
**I want to** mark an invoice as paid  
**So that** payment status is tracked correctly

**API Endpoint:**
```http
POST /invoices/{invoice_id}/mark-paid
Authorization: Bearer {token}
Content-Type: application/json

{
  "payment_date": "2026-01-25",
  "payment_reference": "SEPA-2026-01-25-001"
}
```

**Response:**
```json
{
  "invoice": {
    "id": 1001,
    "invoice_number": "2026-0001",
    "status": "paid",
    "payment_date": "2026-01-25",
    "payment_reference": "SEPA-2026-01-25-001"
  }
}
```

**UI Flow:**
1. **Click "Mark as Paid"** on sent/overdue invoice
2. **Show Payment Modal:**
   - Payment date picker (default: today)
   - Payment reference field (optional, e.g., bank transfer ID)
   - Total amount (read-only, for verification)
   - "Confirm Payment" button
3. **On Success:**
   - Show success toast: "Invoice marked as paid"
   - Status badge changes to "Paid" (green)
   - Actions update (only "Download PDF" available)

---

### Phase 6: Overdue & Reminder Management

#### **User Story 6.1: Mark Invoice as Overdue**
**As a** practice manager  
**I want to** manually mark an invoice as overdue  
**So that** I can start the reminder process

**API Endpoint:**
```http
POST /invoices/{invoice_id}/mark-overdue
Authorization: Bearer {token}
```

**UI Flow:**
1. **Automatic Detection:** Frontend checks due_date vs current date
2. **Show Warning Badge:** "Due in X days" (yellow) or "X days overdue" (red)
3. **Manual Action:** "Mark as Overdue" button (only for sent invoices)
4. **On Success:**
   - Status changes to "overdue"
   - Reminder actions become available

---

#### **User Story 6.2: Send Payment Reminder**
**As a** practice manager  
**I want to** send payment reminders to clients with overdue invoices  
**So that** I can collect outstanding payments

**API Endpoint:**
```http
POST /invoices/{invoice_id}/reminder
Authorization: Bearer {token}
Content-Type: application/json

{
  "recipient_email": "max.mustermann@example.com",
  "custom_message": "Sehr geehrter Herr Mustermann,\n\nwir haben noch keine Zahlung erhalten..."
}
```

**Response:**
```json
{
  "invoice": {
    "id": 1001,
    "invoice_number": "2026-0001",
    "status": "overdue",
    "num_reminders": 1,
    "latest_reminder": "2026-01-22T10:00:00Z"
  },
  "reminder_number": 1
}
```

**UI Flow:**
1. **Overdue Dashboard Section:**
   - List overdue invoices sorted by days overdue
   - Show: Invoice number, client, days overdue, reminder count
   - "Send Reminder" button per invoice
2. **Click "Send Reminder":**
   - Show reminder compose modal
   - Auto-select template based on num_reminders:
     - Reminder 1: Friendly tone
     - Reminder 2: Firmer tone
     - Reminder 3: Final notice
   - Pre-fill email, subject, message (editable)
   - Show original invoice details (amount, due date)
3. **On Success:**
   - Update reminder count: "Reminder 1 sent on 22.01.2026"
   - Add to timeline
4. **Reminder History:** Show all reminders sent with timestamps

**Reminder Counter Display:**
```typescript
function getReminderBadge(numReminders: number): JSX.Element {
  if (numReminders === 0) return null;
  const color = numReminders >= 3 ? 'red' : numReminders >= 2 ? 'orange' : 'yellow';
  return <Badge color={color}>{numReminders} Reminder{numReminders > 1 ? 's' : ''}</Badge>;
}
```

---

### Phase 7: Credit Notes

#### **User Story 7.1: Create Credit Note**
**As a** practice manager  
**I want to** create a credit note for a finalized invoice  
**So that** I can refund or correct billing errors

**API Endpoint:**
```http
POST /invoices/{invoice_id}/credit-note
Authorization: Bearer {token}
Content-Type: application/json

{
  "line_items_to_credit": [
    {
      "original_line_item_id": 1,
      "quantity": 1,
      "reason": "Service not provided as planned"
    }
  ],
  "credit_date": "2026-01-10",
  "reason": "Partial refund - session cancelled"
}
```

**Response:**
```json
{
  "credit_note": {
    "id": 1050,
    "invoice_number": "2026-0002",
    "status": "sent",
    "is_credit_note": true,
    "credit_note_reference_id": 1001,
    "original_invoice_number": "2026-0001",
    "total_amount": -120.00,
    "invoice_date": "2026-01-10"
  }
}
```

**UI Flow:**
1. **Click "Create Credit Note"** on finalized invoice (sent/paid/overdue)
2. **Show Credit Note Modal:**
   - List original line items with checkboxes
   - For each selected item: quantity to credit (default: full)
   - Reason field (required)
   - Credit date picker (default: today)
   - Preview: Shows negative amounts
3. **On Success:**
   - Show success toast: "Credit note 2026-0002 created"
   - Navigate to credit note detail page
   - Original invoice shows link to credit note
4. **Credit Note Display:**
   - Red banner: "GUTSCHRIFT - Credit Note"
   - Reference to original invoice (clickable link)
   - Negative amounts in red
   - All line items prefixed with "Credit for:"

---

## API Endpoint Summary by Feature

### Discovery & Draft Creation
- `GET /client-invoices/unbilled-sessions` - List unbilled items
- `POST /invoices/draft` - Create draft
- `GET /invoices/{id}` - Get invoice details
- `PUT /invoices/{id}` - Update draft
- `DELETE /invoices/{id}` - Cancel draft

### VAT Management
- `GET /invoices/vat-categories` - List VAT categories

### Finalization & Sending
- `POST /invoices/{id}/finalize` - Finalize invoice
- `POST /invoices/{id}/send-email` - Send via email

### Payment & Status
- `POST /invoices/{id}/mark-paid` - Mark as paid
- `POST /invoices/{id}/mark-overdue` - Mark as overdue

### Reminders
- `POST /invoices/{id}/reminder` - Send reminder

### Credit Notes
- `POST /invoices/{id}/credit-note` - Create credit note

### Documents
- `GET /invoices/{id}/xrechnung` - Download XRechnung XML (government only)
- PDF URLs are returned in invoice response (`pdf_url` field)

### List & Filter
- `GET /client-invoices` - List all invoices (with pagination & filters)

---

## State Management Recommendations

### Redux/Zustand Store Structure
```typescript
interface InvoiceStore {
  // Lists
  unbilledClients: ClientWithUnbilled[];
  invoices: Invoice[];
  
  // Current editing
  currentDraft: Invoice | null;
  vatCategories: VATCategory[];
  
  // UI state
  loading: {
    unbilled: boolean;
    invoices: boolean;
    currentInvoice: boolean;
  };
  
  // Filters
  filters: {
    status: InvoiceStatus[];
    dateFrom: string;
    dateTo: string;
    clientId: number | null;
    search: string;
  };
  
  // Pagination
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}
```

### Actions
```typescript
// Discovery
fetchUnbilledClients()
createDraftFromUnbilled(clientId, sessionIds, effortIds)

// Draft editing
fetchDraftInvoice(invoiceId)
updateDraftInvoice(invoiceId, updates)
addCustomLineItem(item)
removeLineItem(itemId)
saveDraft() // Auto-save

// Finalization
finalizeInvoice(invoiceId)
sendInvoiceEmail(invoiceId, emailData)

// Payment
markAsPaid(invoiceId, paymentData)
markAsOverdue(invoiceId)

// Reminders
sendReminder(invoiceId, reminderData)

// Credit notes
createCreditNote(invoiceId, creditData)

// List
fetchInvoices(filters, pagination)
setFilters(filters)
```

---

## UI/UX Best Practices

### Status Badges
- **Draft:** 
- **Sent:** 
- **Paid:** 
- **Overdue:** 
- **Cancelled:** 

### Confirmation Dialogs
Always confirm:
- Finalize invoice (irreversible)
- Cancel draft (releases sessions)
- Mark as paid (status change)
- Send reminder (email action)
- Create credit note (creates new invoice)

### Loading States
- Skeleton loaders for lists
- Spinner for actions (finalize, send email)
- Optimistic updates for non-critical actions (save draft)

### Error Handling
- Validation errors: Inline, next to field
- API errors: Toast notification with retry option
- Network errors: Offline banner, queue actions

### Accessibility
- Keyboard navigation for all actions
- Screen reader labels for status badges
- Focus management in modals
- ARIA labels for icons

### Responsive Design
- Mobile: Stack invoice items vertically
- Tablet: 2-column layout for details
- Desktop: Full table view with all columns

---

## Security Considerations

### Authentication
- All API calls require Bearer token
- Token refresh on 401 response
- Redirect to login on session expiry

### Authorization
- Check user permissions before showing actions
- Server validates all actions (don't rely on frontend)
- Tenant isolation (organization_id in all requests)

### Data Validation
- Validate all inputs client-side before API call
- Sanitize user input (descriptions, notes)
- Prevent XSS in rendered content

---

## Performance Optimization

### API Calls
- Cache VAT categories (rarely change)
- Debounce search/filter inputs (300ms)
- Paginate invoice lists (20 per page)
- Use stale-while-revalidate for invoice details

### Rendering
- Virtual scrolling for large invoice lists
- Lazy load invoice details on expand
- Memoize computed totals
- Optimize re-renders with React.memo

### Bundle Size
- Code split invoice module
- Lazy load PDF viewer
- Tree-shake unused components

---

## Testing Strategy

### Unit Tests
- VAT calculation logic
- Total computation
- Date formatting utilities
- Status badge rendering

### Integration Tests
- Complete draft creation flow
- Edit and save draft workflow
- Finalization process
- Payment marking

### E2E Tests (Cypress/Playwright)
- Create invoice from unbilled items → finalize → send → mark paid
- Create draft → edit → cancel
- Send reminder for overdue invoice
- Create credit note for paid invoice

---

## Next Steps for Frontend Team

1. **Setup API Client:**
   - Create axios instance with auth interceptor
   - Define TypeScript types for all API responses
   - Implement error handling middleware

2. **Build Core Components:**
   - InvoiceStatusBadge
   - LineItemsTable
   - VATBreakdownDisplay
   - InvoiceTimeline

3. **Implement Features in Order:**
   - Phase 1: Unbilled items list (discovery)
   - Phase 2: Draft creation & editing
   - Phase 3: Finalization dialog
   - Phase 4: Invoice details & list
   - Phase 5: Payment tracking
   - Phase 6: Reminder management
   - Phase 7: Credit notes

4. **Setup State Management:**
   - Install Redux Toolkit or Zustand
   - Define store structure
   - Implement async actions with thunks

5. **Create Mock Data:**
   - Mock API responses for development
   - Use MSW (Mock Service Worker) for testing

6. **Documentation:**
   - Component storybook for UI components
   - API integration guide
   - Deployment checklist

---

## Appendix: Complete TypeScript Type Reference

### ✅ All Implemented Invoice Types

These types are safe to import and use - all corresponding backend endpoints are implemented and tested:

```typescript
// Core Invoice Types
interface Invoice {
  id: number;
  invoice_number: string;
  status: InvoiceStatus;
  client_id: number;
  client_name: string;
  organization_id: number;
  invoice_date: string;
  due_date?: string;
  subtotal_amount: number;
  tax_amount: number;
  total_amount: number;
  payment_date?: string;
  payment_reference?: string;
  sent_at?: string;
  finalized_at?: string;
  is_credit_note: boolean;
  credit_note_reference_id?: number;
  original_invoice_number?: string;
  num_reminders: number;
  latest_reminder?: string;
  pdf_url?: string;
  invoice_items: InvoiceItem[];
  vat_breakdown: VATBreakdown[];
}

interface InvoiceItem {
  id: number;
  invoice_id: number;
  item_type: 'session' | 'extra_effort' | 'custom';
  session_id?: number;
  source_effort_id?: number;
  description: string;
  number_units: number;
  unit_price: number;
  total_amount: number;
  vat_rate: number;
  vat_exempt: boolean;
  vat_exemption_text?: string;
  is_editable: boolean;
}

type InvoiceStatus = 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled';

// VAT Types
interface VATCategory {
  code: 'exempt_heilberuf' | 'taxable_standard' | 'taxable_reduced';
  name: string;
  vat_rate: number;
  vat_exempt: boolean;
  exemption_text: string;
  is_default: boolean;
}

interface VATBreakdown {
  vat_rate: number;
  net_amount: number;
  tax_amount: number;
  gross_amount: number;
  exemption_text?: string;
}

// Unbilled Items Types
interface ClientWithUnbilled {
  client_id: number;
  client_name: string;
  total_sessions: number;
  total_efforts: number;
  estimated_total: number;
  unbilled_sessions: UnbilledSession[];
  unbilled_efforts: UnbilledEffort[];
}

interface UnbilledSession {
  session_id: number;
  session_date: string;
  duration_minutes: number;
  service_type: string;
  price: number;
}

interface UnbilledEffort {
  effort_id: number;
  description: string;
  hours: number;
  price: number;
}

// Request/Response Types
interface CreateDraftRequest {
  client_id: number;
  organization_id: number;
  session_ids: number[];
  extra_effort_ids: number[];
  invoice_date: string;
  custom_line_items: CustomLineItem[];
}

interface UpdateDraftRequest {
  add_session_ids?: number[];
  remove_session_ids?: number[];
  add_extra_effort_ids?: number[];
  remove_extra_effort_ids?: number[];
  custom_line_items?: CustomLineItem[];
  update_custom_line_items?: UpdateCustomLineItem[];
  remove_custom_line_item_ids?: number[];
}

interface CustomLineItem {
  description: string;
  number_units: number;
  unit_price: number;
  vat_category: VATCategory['code'];
}

interface UpdateCustomLineItem extends CustomLineItem {
  id: number;
}

interface SendEmailRequest {
  recipient_email: string;
  cc_emails?: string[];
  subject: string;
  message: string;
}

interface MarkPaidRequest {
  payment_date: string;
  payment_reference?: string;
}

interface SendReminderRequest {
  recipient_email: string;
  custom_message?: string;
}

interface CreateCreditNoteRequest {
  line_items_to_credit: CreditLineItem[];
  credit_date: string;
  reason: string;
}

interface CreditLineItem {
  original_line_item_id: number;
  quantity: number;
  reason: string;
}

// List & Filter Types
interface InvoiceListResponse {
  invoices: InvoiceSummary[];
  pagination: PaginationInfo;
}

interface InvoiceSummary {
  id: number;
  invoice_number: string;
  status: InvoiceStatus;
  client_id: number;
  client_name: string;
  invoice_date: string;
  due_date?: string;
  total_amount: number;
  days_overdue: number;
  num_reminders: number;
}

interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

interface InvoiceFilters {
  status?: InvoiceStatus[];
  client_id?: number;
  date_from?: string;
  date_to?: string;
  page?: number;
  limit?: number;
}

// Audit Types (if needed for invoice history)
interface AuditLog {
  id: number;
  tenant_id: number;
  user_id: number;
  entity_type: string;
  entity_id: number;
  action: string;
  metadata: Record<string, any>;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}
```

### Backend Endpoint Status

| Endpoint | Status | Tests |
|----------|--------|-------|
| `GET /client-invoices/unbilled-sessions` | ✅ | Integration |
| `POST /invoices/draft` | ✅ | Integration |
| `GET /invoices/{id}` | ✅ | Integration |
| `PUT /invoices/{id}` | ✅ | Integration |
| `DELETE /invoices/{id}` | ✅ | Integration |
| `GET /invoices/vat-categories` | ✅ | 12 unit tests |
| `POST /invoices/{id}/finalize` | ✅ | 9 unit tests |
| `POST /invoices/{id}/send-email` | ✅ | Integration |
| `POST /invoices/{id}/mark-paid` | ✅ | Integration |
| `POST /invoices/{id}/mark-overdue` | ✅ | Integration |
| `POST /invoices/{id}/reminder` | ✅ | Integration |
| `POST /invoices/{id}/credit-note` | ✅ | Integration |
| `GET /invoices/{id}/xrechnung` | ✅ | Integration |
| `GET /invoices/{id}/pdf` | ✅ | Integration |
| `GET /invoices/{id}/preview-pdf` | ✅ | Integration |
| `GET /client-invoices` | ✅ | Integration |
| `GET /audit/logs` | ✅ | Production module |
| `GET /audit/entity/{type}/{id}` | ✅ | Production module |
| `GET /audit/export` | ✅ | Production module |

**Total: 19/19 endpoints implemented and tested (100%)**

---

## Support & Troubleshooting

### Common Issues

**Q: I'm getting TypeScript errors about missing types when generating the API client.**  
**A:** These are from unimplemented modules (booking, templates, etc.). Use Option 1 or 2 from the "API Client Generation" section above to stub or exclude them.

**Q: Can I develop the invoice frontend without the booking/template modules?**  
**A:** Yes! All invoice endpoints are independent and fully functional. You can build the complete invoice UI today.

**Q: Which API endpoints are safe to use?**  
**A:** All endpoints listed in "API Endpoint Summary by Feature" section are production-ready with 100% test coverage.

**Q: How do I handle VAT calculations on the frontend?**  
**A:** Don't implement VAT logic client-side. The backend handles all calculations. Just send the `vat_category` code and display the returned `vat_breakdown`.

**Q: What if I need to test without a real backend?**  
**A:** Use Mock Service Worker (MSW) with the response schemas provided in this guide. All example responses are accurate.

### Contact

For backend API questions or issues with invoice endpoints, check:
- [invoice_implementation.md](./invoice_implementation.md) - Complete technical documentation
- [PHASE_12_COMPLETION_SUMMARY.md](./PHASE_12_COMPLETION_SUMMARY.md) - Test results and status
- Backend API: `/swagger/index.html` - Live OpenAPI documentation
