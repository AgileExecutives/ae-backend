# Invoice Workflow – Complete Requirements Document

## 1. Goals & Principles

* Clear separation between **drafting** and **final billing**
* No irreversible backend state changes before user explicitly finalizes
* Full editability of invoice line items during draft phase
* Transparent state machine for invoices, sessions, and extra efforts
* Support for reminders with configurable organization-level settings
* Auditability: all state transitions are explicit and traceable

---

## 2. Domain States

### 2.1 Invoice Statuses

```text
Draft      – invoice exists but is not finalized or sent (not accounting-relevant)
Sent       – invoice finalized and sent to client (accounting-relevant)
Paid       – payment received and confirmed
Overdue    – invoice sent, payment overdue (system-derived or manual)
Cancelled  – invoice voided BEFORE finalization; linked items released
```

### 2.2 Session States

```text
scheduled
canceled
re-scheduled
conducted
invoice-draft   (reserved by a draft invoice)
billed          (finalized & invoiced)
```

### 2.3 Extra Effort States

```text
delivered
invoice-draft   (reserved by a draft invoice)
billed          (finalized & invoiced)
```

---

## 3. Backend Requirements

### 3.1 Core Concepts

#### Draft Invoice

* Draft invoices **reserve** sessions and extra efforts
* Reserved items are marked as `invoice-draft`
* Draft invoices can be edited or cancelled without billing
* No accounting-relevant actions occur in draft state

#### Finalization

* Finalization is explicit
* On finalization:

  * Invoice status → `sent`
  * Sessions & extra efforts → `billed`
  * PDF is frozen (immutable)
* From an accounting perspective (EU/Germany):

  * Invoice number becomes legally binding
  * Invoice date and amount are fixed
  * Invoice must never be deleted or altered

#### Cancellation vs Credit Note (EU/German Accounting)

**Cancellation**

* Only allowed while invoice status = `draft`
* Represents: "This invoice should never have existed"
* Legally safe because no invoice number was issued
* Sessions & extra efforts revert to pre-invoice state

**Credit Note (Gutschrift / Stornorechnung)**

* Required once an invoice has been finalized (`sent`)
* Original invoice remains valid and immutable
* A credit note offsets the original invoice partially or fully
* Mandatory for German/EU tax compliance (UStG §14)

---

## 4. Backend API Endpoints

### 4.1 Discovery & Preparation

#### `GET /client-invoices/unbilled-sessions`

* Returns clients with sessions in `conducted` and extra efforts in `delivered`
* Grouped by client
* Purpose: entry point for invoice creation

---

### 4.2 Draft Invoice Lifecycle

#### `POST /invoices/draft`

* Creates a new draft invoice
* Input: client_id, selected sessions, selected extra efforts, optional custom line items
* Behavior: sets items to `invoice-draft` and generates draft PDF

#### `GET /invoices/{invoice_id}`

* Returns invoice details: line items, totals, status

#### `PUT /invoices/{invoice_id}`

* Updates draft invoice (only if status = draft)
* Supports editing line items, adding/removing sessions and extra efforts
* Recalculates totals and regenerates draft PDF

#### `DELETE /invoices/{invoice_id}`

* Cancels draft invoice
* Invoice status → `cancelled`
* Reverts sessions and extra efforts to original state

### 4.3 Finalization & Sending

#### `POST /invoices/{invoice_id}/finalize`

* Preconditions: invoice status = draft
* Behavior: status → sent, sessions & extra efforts → billed, final PDF generated, timestamp stored

#### `POST /invoices/{invoice_id}/send-email`

* Sends finalized invoice PDF using HTML template
* Stores `email_sent_at`

### 4.4 Payment Handling

#### `POST /invoices/{invoice_id}/mark-paid`

* Marks invoice as paid
* Stores payment date & optional reference

### 4.5 Credit Notes

#### `POST /invoices/{invoice_id}/credit-note`

* Preconditions: invoice status ∈ {sent, paid, overdue}
* Input: line items to credit, optional reason
* Behavior: creates new invoice record with negative line items, reference to original invoice, immutable PDF
* VAT and totals adjusted accordingly

### 4.6 Overdue & Reminders

#### Organization Settings

```text
payment_due_days
first_reminder_days
second_reminder_days
```

#### `POST /invoices/{invoice_id}/mark-overdue`

* Marks invoice as overdue (manual or scheduled)

#### `POST /invoices/{invoice_id}/reminder`

* Generates reminder PDF
* Selects template based on number of reminders sent: reminder_1, reminder_2, reminder_3
* Stores `reminder_sent_at`
* Invoice remains overdue

### 4.7 Templates

* Invoice
* Reminder 1
* Reminder 2
* Reminder 3

---

## 5. Frontend Requirements

### 5.1 Entry Flow

* List clients with unbilled sessions and extra efforts
* Show totals per client
* CTA: Create Draft Invoice

### 5.2 Draft Invoice Editor

* Editable line items
* Inline price/quantity editing
* Add custom line items
* Remove sessions/extra efforts
* Live total calculation
* Banner: Draft – not yet billed
* Actions: Save Draft, Cancel Draft, Finalize & Send

### 5.3 Finalization UX

* Confirmation dialog: This will bill all included sessions and extra efforts
* Preview PDF, send via email or print

### 5.4 Invoice Detail View

* Status timeline
* Linked sessions/extra efforts (read-only)
* Download PDFs
* Actions based on status (Draft/Edit, Sent/Mark Paid, Overdue/Send Reminder, Paid/View only, Cancelled/View only)

### 5.5 Overdue & Reminder Flow

* Overdue invoices highlighted
* Countdown labels
* Next Reminder button using reminder template logic

### 5.6 Error Prevention & Clarity

* Sessions in invoice-draft are locked elsewhere
* Clear messaging if items already reserved
* No silent state changes

---

## 6. Non-Functional Requirements

* All state transitions logged
* PDFs immutable after finalization
* Draft PDFs watermarked
* Idempotent finalize & send endpoints

---

## 7. Open Design Decisions

* Manual overdue marking
* Partial payments out of scope
* Draft-only invoice cancellations
* Corrections after finalization via credit notes only

---

## 8. Invoice Numbering Rules (EU/Germany)

* Unique sequential number per org per year
* Only assigned on finalization
* Credit notes share sequence but reference original invoice

---

## 9. VAT Handling – NOW (Germany, Heilberuf → Government)

* Domestic DE only
* Customer = government agency
* VAT-exempt services flagged per line item
* Exemption text required on invoice: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
* Backend stores `vat_exempt`, `vat_rate`, exemption text
* VAT immutable after finalization

### 9.4 Service Catalog

* Each service has `vat_category`: `exempt_heilberuf` or `taxable_standard`
* Default VAT text prefills draft line items
* User may override only in draft

---

## 10. VAT Handling – FUTURE (EU, OSS, Reverse Charge)

* EU B2B reverse charge
* EU B2C OSS
* Non-EU clients
* Deferred to future implementation

---

## 11. GoBD Audit & Export Requirements

* Complete, immutable, traceable, machine-readable
* Audit logs: creation, edits, finalization, sending, payment, reminders, credit notes
* Logs include timestamp, user, action, entity
* Export via `GET /audit/export` (CSV mandatory, DATEV optional)
* Include VAT category/text
* Retain for 10 years, no early deletion

---

## 12. Government Customer & Invoice Validation

* Required fields: authority name, address, reference/cost center
* Validation before finalization:

  * All line items have VAT category
  * Exemption text present if exempt
  * Invoice number not assigned
  * Customer fields complete
* Finalization blocked if validation fails
