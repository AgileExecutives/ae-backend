# Extra Efforts & Invoice Management - Design Document

## Executive Summary

This document describes the design for tracking unbilled therapeutic efforts (preparation, consultations, meetings) and a flexible invoicing system with configurable billing modes, draft editing, and PDF preview.

## 1. Requirements Analysis

### 1.1 User Stories

**US-1**: As a therapist, I need to record extra efforts (copying materials, teacher meetings, parent consultations) related to client therapy, so I can track all billable activities.

**US-2**: As a practice owner, I need to configure how extra efforts are invoiced (ignore, bundle into double units, separate line items, or preparation time), so I can comply with different insurance/billing requirements.

**US-3**: As a billing manager, I need to edit draft invoices before finalizing, so I can make corrections or adjustments.

**US-4**: As a user, I need to preview invoice PDFs before finalizing, so I can verify formatting and content.

**US-5**: As a therapist, I need different types of extra efforts (preparation, consultation, meeting, documentation), so I can accurately categorize my work.

### 1.2 Extra Effort Types

| Type | Description | Typical Duration | Examples |
|------|-------------|------------------|----------|
| `preparation` | Session preparation | 5-30 min | Copying materials, preparing exercises |
| `consultation` | Professional consultation | 10-60 min | Teacher meetings, colleague discussions |
| `parent_meeting` | Parent/guardian meetings | 15-60 min | Progress discussions, planning meetings |
| `documentation` | Administrative work | 5-30 min | Reports, documentation, correspondence |
| `other` | Other billable efforts | Variable | Miscellaneous activities |

### 1.3 Billing Modes

#### Mode A: Ignore (`ignore`)
- Extra efforts are tracked but not billed
- Use case: Non-billable activities, internal tracking only

#### Mode B: Bundle into Double Units (`bundle_double_units`)
- When session + extra efforts >= unit threshold, invoice as 2 units
- Parameters:
  - `unit_duration_min` (45 or 60 minutes)
  - `threshold_percentage` (e.g., 90% = 40.5 min for 45 min unit)
- Example: 45 min session + 20 min preparation = 65 min → 2 units

#### Mode C: Separate Line Items (`separate_items`)
- Extra efforts billed as individual invoice items
- Parameters:
  - `round_to_min` (round to nearest 5, 15, or 30 minutes)
  - `minimum_duration_min` (minimum billable duration)
- Example: 20 min consultation → separate line item

#### Mode D: Preparation Time Allowance (`preparation_allowance`)
- Fixed preparation time per session unit
- Parameters:
  - `minutes_per_unit` (e.g., 15 min per 45 min session)
  - `billing_mode`: `automatic` (auto-add) or `track_actual` (bill actual up to limit)
- Example: 45 min session → automatically add 15 min preparation = 1 unit + 0.33 units

## 2. Database Schema Changes

### 2.1 New Table: `extra_efforts`

```sql
CREATE TABLE extra_efforts (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id),
    client_id INTEGER NOT NULL REFERENCES clients(id),
    session_id INTEGER REFERENCES sessions(id),  -- NULL if not session-related
    effort_type VARCHAR(50) NOT NULL,  -- preparation, consultation, parent_meeting, documentation, other
    effort_date DATE NOT NULL,
    duration_min INTEGER NOT NULL,
    description TEXT,
    billable BOOLEAN DEFAULT true,
    billing_status VARCHAR(20) DEFAULT 'unbilled',  -- unbilled, billed, excluded
    invoice_item_id INTEGER REFERENCES invoice_items(id),  -- Link when billed
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    INDEX idx_extra_efforts_tenant (tenant_id),
    INDEX idx_extra_efforts_client (client_id),
    INDEX idx_extra_efforts_session (session_id),
    INDEX idx_extra_efforts_billing_status (billing_status),
    INDEX idx_extra_efforts_date (effort_date)
);
```

### 2.2 Organization Table Updates

```sql
ALTER TABLE organizations 
ADD COLUMN extra_efforts_billing_mode VARCHAR(50) DEFAULT 'ignore',
ADD COLUMN extra_efforts_config JSONB DEFAULT '{}',
ADD COLUMN line_item_single_unit_text VARCHAR(100) DEFAULT 'Einzelstunde',  -- "Single hour"
ADD COLUMN line_item_double_unit_text VARCHAR(100) DEFAULT 'Doppelstunde';  -- "Double hour"

-- Example config for each mode:
-- Mode A (ignore):
-- {}

-- Mode B (bundle_double_units):
-- {
--   "unit_duration_min": 45,
--   "threshold_percentage": 90
-- }
-- Uses line_item_single_unit_text and line_item_double_unit_text for descriptions

-- Mode C (separate_items):
-- {
--   "round_to_min": 15,
--   "minimum_duration_min": 10,
--   "unit_price": 150.00  -- optional override
-- }

-- Mode D (preparation_allowance):
-- {
--   "minutes_per_unit": 15,
--   "billing_mode": "automatic",  -- or "track_actual"
--   "max_percentage": 33.33  -- 15/45 = 33.33%
-- }
```

### 2.3 Sessions Table Updates

```sql
ALTER TABLE sessions
ADD COLUMN internal_note TEXT;
```

**Purpose**: Add internal notes to sessions for therapist reference (not shown on invoices).

### 2.4 Invoice Table Updates

**Note**: The `invoices` table already has a `status` field. We'll use it for the state machine:
- `draft` - Invoice is being edited
- `finalized` - Invoice is locked and PDF generated
- `sent` - Invoice has been sent to client
- `paid` - Invoice has been paid
- `overdue` - Invoice is past due
- `cancelled` - Invoice was cancelled

No schema changes needed for invoices table.

### 2.5 Invoice Items Table Updates

```sql
ALTER TABLE invoice_items
ADD COLUMN item_type VARCHAR(50) DEFAULT 'session',  -- session, extra_effort, preparation, adjustment
ADD COLUMN source_effort_id INTEGER REFERENCES extra_efforts(id),
ADD COLUMN description TEXT,  -- Editable text like "Therapiestunde" or "Therapie Doppelstunde"
ADD COLUMN unit_duration_min INTEGER,  -- Store duration for extra efforts
ADD COLUMN is_editable BOOLEAN DEFAULT true;  -- Allow editing line item descriptions
```

**Line Item Descriptions**:
- Single unit sessions: "Therapiestunde" (Therapy hour)
- Double unit sessions: "Therapie Doppelstunde" (Therapy double hour)
- Extra efforts: Custom descriptions based on effort type
- Users can edit these descriptions in draft invoices

### 2.6 Client Invoices Junction Table Updates

```sql
ALTER TABLE client_invoices
ADD COLUMN extra_effort_id INTEGER REFERENCES extra_efforts(id);  -- Link to extra effort if applicable
```

## 3. Backend API Design

### 3.1 Extra Efforts Endpoints

#### Create Extra Effort
```http
POST /api/extra-efforts
Authorization: Bearer {token}
Content-Type: application/json

{
  "client_id": 123,
  "session_id": 456,  // optional
  "effort_type": "preparation",
  "effort_date": "2025-12-30",
  "duration_min": 20,
  "description": "Copied therapy materials for next session",
  "billable": true
}

Response 201:
{
  "success": true,
  "message": "Extra effort recorded successfully",
  "data": {
    "id": 789,
    "client_id": 123,
    "session_id": 456,
    "effort_type": "preparation",
    "effort_date": "2025-12-30",
    "duration_min": 20,
    "description": "Copied therapy materials",
    "billable": true,
    "billing_status": "unbilled",
    "created_at": "2025-12-30T10:00:00Z"
  }
}
```

#### Get Extra Efforts for Client
```http
GET /api/extra-efforts?client_id=123&billing_status=unbilled
Authorization: Bearer {token}

Response 200:
{
  "success": true,
  "data": [
    {
      "id": 789,
      "client_id": 123,
      "session_id": 456,
      "session": {
        "id": 456,
        "original_date": "2025-12-29",
        "type": "individual"
      },
      "effort_type": "preparation",
      "effort_date": "2025-12-30",
      "duration_min": 20,
      "description": "Copied therapy materials",
      "billing_status": "unbilled"
    }
  ],
  "total": 1
}
```

#### Update Extra Effort
```http
PUT /api/extra-efforts/:id
Authorization: Bearer {token}
Content-Type: application/json

{
  "duration_min": 25,
  "description": "Updated description"
}

Response 200:
{
  "success": true,
  "message": "Extra effort updated successfully"
}
```

#### Delete Extra Effort
```http
DELETE /api/extra-efforts/:id
Authorization: Bearer {token}

Response 200:
{
  "success": true,
  "message": "Extra effort deleted successfully"
}
```

### 3.2 Organization Billing Configuration

#### Update Organization Billing Config
```http
PUT /api/organizations/:id/billing-config
Authorization: Bearer {token}
Content-Type: application/json

{
  "extra_efforts_billing_mode": "bundle_double_units",
  "extra_efforts_config": {
    "unit_duration_min": 45,
    "threshold_percentage": 90
  },
  "line_item_single_unit_text": "Therapiestunde",
  "line_item_double_unit_text": "Therapie Doppelstunde"
}

Response 200:
{
  "success": true,
  "message": "Billing configuration updated successfully",
  "data": {
    "extra_efforts_billing_mode": "bundle_double_units",
    "extra_efforts_config": {
      "unit_duration_min": 45,
      "threshold_percentage": 90
    },
    "line_item_single_unit_text": "Therapiestunde",
    "line_item_double_unit_text": "Therapie Doppelstunde"
  }
}
```

### 3.3 Invoice Management Endpoints

#### Get Unbilled Sessions with Extra Efforts
```http
GET /api/client-invoices/unbilled-sessions
Authorization: Bearer {token}

Response 200:
{
  "success": true,
  "data": [
    {
      "client_id": 123,
      "client": { /* client details */ },
      "sessions": [
        {
          "id": 456,
          "original_date": "2025-12-29",
          "number_units": 1,
          "duration_min": 45,
          "billing_status": "unbilled",
          "internal_note": "Client was very engaged today"
        }
      ],
      "extra_efforts": [
        {
          "id": 789,
          "session_id": 456,
          "effort_type": "preparation",
          "effort_date": "2025-12-29",
          "duration_min": 20,
          "description": "Copied materials",
          "billing_status": "unbilled"
        },
        {
          "id": 790,
          "session_id": null,  // Not linked to specific session
          "effort_type": "consultation",
          "effort_date": "2025-12-28",
          "duration_min": 30,
          "description": "Teacher consultation",
          "billing_status": "unbilled"
        }
      ]
    }
  ]
}
```

**Note**: The endpoint returns unbilled sessions and unbilled extra efforts separately. The billing mode logic is **not** applied here - this is just raw data. The actual unit calculation and line item creation happens during invoice generation.

#### Generate Invoice (Draft)
```http
POST /api/client-invoices/generate
Authorization: Bearer {token}
Content-Type: application/json

{
  "client_ids": [123],  // Or select from unbilled-sessions response
  "parameters": {
    "invoice_number": "INV-2025-001",
    "invoice_date": "2025-12-30",
    "tax_rate": 19.0,
    "generate_pdf": false,  // Don't generate PDF yet, stay in draft
    "session_from_date": "2025-12-01",
    "session_to_date": "2025-12-31"
  }
}

Response 201:
{
  "success": true,
  "message": "Invoice created successfully (draft)",
  "data": {
    "id": 999,
    "invoice_number": "INV-2025-001",
    "invoice_date": "2025-12-30",
    "status": "draft",
    "clients": [
      {
        "client_id": 123,
        "client": { /* client details */ },
        "invoice_items": [
          {
            "id": 1001,
            "item_type": "session",
            "description": "Doppelstunde",  // From organization.line_item_double_unit_text
            "number_units": 2,  // Calculated: 45 min session + 20 min extra effort >= threshold
            "unit_price": 150.00,
            "total_amount": 300.00,
            "source_session_id": 456,
            "bundled_effort_ids": [789]  // Extra efforts bundled into this item
          },
          {
            "id": 1002,
            "item_type": "extra_effort",
            "description": "Beratung - 30 min",  // Separate item (mode C or standalone)
            "number_units": 0.5,
            "unit_price": 150.00,
            "total_amount": 75.00,
            "source_effort_id": 790
          }
        ]
      }
    ],
    "sum_amount": 375.00,
    "tax_amount": 71.25,
    "total_amount": 446.25,
    "editable": true
  }
}
```

**Important**: The billing mode logic is applied **during invoice creation**:
1. Backend retrieves unbilled sessions and extra efforts for selected clients
2. Groups extra efforts by session (using `session_id`) or keeps standalone
3. Applies organization's `extra_efforts_billing_mode` to calculate units
4. Creates `invoice_items` with appropriate descriptions from `line_item_single_unit_text` / `line_item_double_unit_text`
5. Marks sessions and extra efforts as `billed` and links them via `invoice_item_id`

#### Edit Draft Invoice
```http
PUT /api/client-invoices/:id/items
Authorization: Bearer {token}
Content-Type: application/json

{
  "items": [
    {
      "id": 1001,  // existing item
      "number_units": 2,
      "unit_price": 150.00,
      "description": "Therapie Doppelstunde"  // Editable - German for "Therapy double hour"
    },
    {
      "id": 1002,
      "number_units": 1,
      "unit_price": 150.00,
      "description": "Therapiestunde"  // "Therapy hour"
    },
    {
      // new item (no id)
      "item_type": "adjustment",
      "description": "Rabatt für Stammklient",  // "Discount for regular client"
      "number_units": -0.5,
      "unit_price": 150.00
    }
  ]
}

Response 200:
{
  "success": true,
  "message": "Invoice items updated successfully",
  "data": {
    "id": 999,
    "items": [ /* updated items */ ],
    "sum_amount": 300.00,  // recalculated
    "tax_amount": 57.00,
    "total_amount": 357.00
  }
}
```

**Line Item Description Guidelines**:
- Single unit (1): Uses `organization.line_item_single_unit_text` (default: "Einzelstunde")
- Double unit (2): Uses `organization.line_item_double_unit_text` (default: "Doppelstunde")
- Mode B (bundle_double_units): Automatically uses configured text based on calculated units
- Custom descriptions allowed for special cases (editable in draft invoices)
- Extra efforts: Auto-generated but editable (e.g., "Vorbereitung - 20 min")
- Adjustments: Fully custom descriptions
- Organizations can customize the text (e.g., "Therapiestunde", "Therapie Doppelstunde", "Einzeltherapie", etc.)
- Session `internal_note` is for therapist reference only and **not shown on invoices**

#### Delete Draft Invoice Item
```http
DELETE /api/client-invoices/:invoice_id/items/:item_id
Authorization: Bearer {token}

Response 200:
{
  "success": true,
  "message": "Invoice item deleted successfully"
}
```

#### Preview Invoice PDF
```http
POST /api/client-invoices/:id/preview-pdf
Authorization: Bearer {token}

Response 200:
{
  "success": true,
  "message": "PDF preview generated successfully",
  "data": {
    "preview_url": "https://minio.../invoices/preview-999.pdf?expires=...",
    "expires_at": "2025-12-30T11:00:00Z"
  }
}
```

#### Finalize Invoice
```http
POST /api/client-invoices/:id/finalize
Authorization: Bearer {token}
Content-Type: application/json

{
  "generate_pdf": true,
  "template_id": 5  // optional
}

Response 200:
{
  "success": true,
  "message": "Invoice finalized successfully",
  "data": {
    "id": 999,
    "status": "finalized",
    "finalized_at": "2025-12-30T10:30:00Z",
    "document_id": 555,
    "pdf_url": "https://minio.../invoices/INV-2025-001.pdf?expires=..."
  }
}
```

#### Revert to Draft
```http
POST /api/client-invoices/:id/revert-to-draft
Authorization: Bearer {token}

Response 200:
{
  "success": true,
  "message": "Invoice reverted to draft",
  "data": {
    "id": 999,
    "status": "draft",
    "finalized_at": null
  }
}
```

## 4. Business Logic Implementation

### 4.1 Extra Efforts Billing Calculator

**When to Use**: Called during `GenerateInvoice` / `CreateInvoice` to calculate units and create line items.

```go
// Service: extra_efforts_service.go or invoice_service.go

type BillingCalculator struct {
    mode   string
    config map[string]interface{}
    org    *Organization  // Reference to organization for line item text
}

// CalculateUnitsForSession calculates billing units including extra efforts
// This is called during invoice creation, NOT during unbilled sessions query
func (c *BillingCalculator) CalculateUnitsForSession(
    session Session,
    sessionExtraEfforts []ExtraEffort,  // Extra efforts linked to this session
) (*BillingResult, error) {
    sessionDurationMin := session.DurationMin
    
    switch c.mode {
    case "ignore":
        return &BillingResult{
            Units:        sessionDurationMin / 45,  // or config unit
            Items:        []BillingItem{{Type: "session", Units: ...}},
            ExtraEfforts: []ExtraEffort{},  // tracked but not billed
        }, nil
        
    case "bundle_double_units":
        totalMin := sessionDurationMin + sumEffortDuration(extraEfforts)
        unitMin := c.config["unit_duration_min"].(int)
        threshold := c.config["threshold_percentage"].(float64)
        
        units := 1
        description := c.org.LineItemSingleUnitText  // e.g., "Einzelstunde"
        if totalMin >= int(float64(unitMin*2) * threshold / 100) {
            units = 2
            description = c.org.LineItemDoubleUnitText  // e.g., "Doppelstunde"
        }
        
        return &BillingResult{
            Units: units,
            Items: []BillingItem{{
                Type:        "session",
                Units:       units,
                Description: description,  // Uses configured text from organization
            }},
            BundledEfforts: extraEfforts,
        }, nil
        
    case "separate_items":
        items := []BillingItem{{Type: "session", Units: sessionDurationMin/45}}
        
        for _, effort := range extraEfforts {
            roundedMin := roundDuration(effort.DurationMin, c.config["round_to_min"].(int))
            if roundedMin >= c.config["minimum_duration_min"].(int) {
                items = append(items, BillingItem{
                    Type:        "extra_effort",
                    EffortID:    effort.ID,
                    EffortType:  effort.Type,
                    Units:       float64(roundedMin) / 45.0,
                    Description: fmt.Sprintf("%s - %dmin", effort.Type, roundedMin),
                })
            }
        }
        
        return &BillingResult{Items: items}, nil
        
    case "preparation_allowance":
        sessionUnits := sessionDurationMin / 45
        prepMin := c.config["minutes_per_unit"].(int)
        billingMode := c.config["billing_mode"].(string)
        
        if billingMode == "automatic" {
            // Auto-add fixed preparation time
            prepUnits := float64(sessionUnits*prepMin) / 45.0
            return &BillingResult{
                Items: []BillingItem{
                    {Type: "session", Units: sessionUnits},
                    {Type: "preparation", Units: prepUnits, Description: "Preparation allowance"},
                },
            }, nil
        } else {
            // Track actual up to limit
            actualPrepMin := sumEffortDurationByType(extraEfforts, "preparation")
            maxPrepMin := sessionUnits * prepMin
            billedPrepMin := min(actualPrepMin, maxPrepMin)
            
            // ... implementation
        }
    }
}
```

### 4.2 Invoice State Machine

**Note**: Uses existing `invoices.status` field.

```go
// Valid state transitions for invoice.status field
var invoiceStateMachine = map[string][]string{
    "draft":     {"finalized", "cancelled"},
    "finalized": {"draft", "sent", "cancelled"}, // Can revert to draft if needed
    "sent":      {"paid", "cancelled"},
    "paid":      {}, // Terminal state
    "cancelled": {}, // Terminal state
}

func (s *InvoiceService) CanTransition(from, to string) bool {
    allowedStates, exists := invoiceStateMachine[from]
    if !exists {
        return false
    }
    return contains(allowedStates, to)
}

func (s *InvoiceService) FinalizeInvoice(invoiceID uint, generatePDF bool) error {
    invoice, err := s.GetInvoiceByID(invoiceID)
    if err != nil {
        return err
    }
    
    if invoice.Status != "draft" {
        return errors.New("can only finalize draft invoices")
    }
    
    // Lock all related sessions and extra efforts
    if err := s.lockBilledItems(invoice); err != nil {
        return err
    }
    
    // Generate PDF if requested
    if generatePDF {
        if err := s.generateInvoicePDF(invoice); err != nil {
            return err
        }
    }
    
    // Update status
    invoice.Status = "finalized"
    invoice.FinalizedAt = time.Now()
    invoice.FinalizedBy = getCurrentUserID()
    
    return s.db.Save(invoice).Error
}
```

## 5. Data Models (Go Entities)

### 5.1 ExtraEffort Entity

```go
package entities

import (
    "time"
    baseAPI "github.com/ae-base-server/api"
    "gorm.io/gorm"
)

type ExtraEffort struct {
    ID            uint                  `gorm:"primaryKey" json:"id"`
    TenantID      uint                  `gorm:"not null;index:idx_extra_efforts_tenant" json:"tenant_id"`
    ClientID      uint                  `gorm:"not null;index:idx_extra_efforts_client" json:"client_id"`
    Client        *Client               `gorm:"foreignKey:ClientID" json:"client,omitempty"`
    SessionID     *uint                 `gorm:"index:idx_extra_efforts_session" json:"session_id"`
    Session       *Session              `gorm:"foreignKey:SessionID" json:"session,omitempty"`
    EffortType    string                `gorm:"size:50;not null" json:"effort_type"` // preparation, consultation, parent_meeting, documentation, other
    EffortDate    time.Time             `gorm:"not null;index:idx_extra_efforts_date" json:"effort_date"`
    DurationMin   int                   `gorm:"not null" json:"duration_min"`
    Description   string                `gorm:"type:text" json:"description"`
    Billable      bool                  `gorm:"default:true" json:"billable"`
    BillingStatus string                `gorm:"size:20;default:'unbilled';index:idx_extra_efforts_billing_status" json:"billing_status"` // unbilled, billed, excluded
    InvoiceItemID *uint                 `json:"invoice_item_id,omitempty"`
    InvoiceItem   *InvoiceItem          `gorm:"foreignKey:InvoiceItemID" json:"invoice_item,omitempty"`
    CreatedBy     uint                  `json:"created_by"`
    CreatedAt     time.Time             `json:"created_at"`
    UpdatedAt     time.Time             `json:"updated_at"`
    DeletedAt     gorm.DeletedAt        `gorm:"index" json:"-"`
}

func (ExtraEffort) TableName() string {
    return "extra_efforts"
}

type ExtraEffortResponse struct {
    ID            uint             `json:"id"`
    ClientID      uint             `json:"client_id"`
    SessionID     *uint            `json:"session_id"`
    Session       *SessionResponse `json:"session,omitempty"`
    EffortType    string           `json:"effort_type"`
    EffortDate    time.Time        `json:"effort_date"`
    DurationMin   int              `json:"duration_min"`
    Description   string           `json:"description"`
    Billable      bool             `json:"billable"`
    BillingStatus string           `json:"billing_status"`
    CreatedAt     time.Time        `json:"created_at"`
}

type CreateExtraEffortRequest struct {
    ClientID    uint      `json:"client_id" binding:"required"`
    SessionID   *uint     `json:"session_id"`
    EffortType  string    `json:"effort_type" binding:"required,oneof=preparation consultation parent_meeting documentation other"`
    EffortDate  time.Time `json:"effort_date" binding:"required"`
    DurationMin int       `json:"duration_min" binding:"required,min=1,max=480"`
    Description string    `json:"description"`
    Billable    *bool     `json:"billable"`
}

type UpdateExtraEffortRequest struct {
    EffortType  *string    `json:"effort_type,omitempty" binding:"omitempty,oneof=preparation consultation parent_meeting documentation other"`
    EffortDate  *time.Time `json:"effort_date,omitempty"`
    DurationMin *int       `json:"duration_min,omitempty" binding:"omitempty,min=1,max=480"`
    Description *string    `json:"description,omitempty"`
    Billable    *bool      `json:"billable,omitempty"`
}
```

### 5.2 Updated Invoice Entity

```go
type Invoice struct {
    // ... existing fields ...
    // Note: Status field already exists in current schema
    // Values: draft, finalized, sent, paid, cancelled
    Status       string         `gorm:"size:20;default:'draft';index:idx_invoices_status" json:"status"`
    // These may need to be added if not present:
    FinalizedAt  *time.Time     `json:"finalized_at,omitempty"`
    FinalizedBy  *uint          `json:"finalized_by,omitempty"`
    FinalizedByUser *User       `gorm:"foreignKey:FinalizedBy" json:"finalized_by_user,omitempty"`
}

type InvoiceItem struct {
    // ... existing fields ...
    ItemType        string  `gorm:"size:50;default:'session'" json:"item_type"` // session, extra_effort, preparation, adjustment
    SourceEffortID  *uint   `json:"source_effort_id,omitempty"`
    SourceEffort    *ExtraEffort `gorm:"foreignKey:SourceEffortID" json:"source_effort,omitempty"`
    Description     string  `gorm:"type:text" json:"description"`
    UnitDurationMin *int    `json:"unit_duration_min,omitempty"`
    IsEditable      bool    `gorm:"default:true" json:"is_editable"`
}

type ClientInvoice struct {
    // ... existing fields ...
    ExtraEffortID *uint        `json:"extra_effort_id,omitempty"`
    ExtraEffort   *ExtraEffort `gorm:"foreignKey:ExtraEffortID" json:"extra_effort,omitempty"`
}
```

### 5.3 Organization Billing Config

```go
type OrganizationBillingConfig struct {
    ExtraEffortsBillingMode string                 `json:"extra_efforts_billing_mode"` // ignore, bundle_double_units, separate_items, preparation_allowance
    ExtraEffortsConfig      map[string]interface{} `json:"extra_efforts_config"`
}

// Config validation
func (c *OrganizationBillingConfig) Validate() error {
    switch c.ExtraEffortsBillingMode {
    case "ignore":
        return nil
    case "bundle_double_units":
        if _, ok := c.ExtraEffortsConfig["unit_duration_min"]; !ok {
            return errors.New("unit_duration_min required for bundle_double_units mode")
        }
        if _, ok := c.ExtraEffortsConfig["threshold_percentage"]; !ok {
            return errors.New("threshold_percentage required for bundle_double_units mode")
        }
    case "separate_items":
        if _, ok := c.ExtraEffortsConfig["round_to_min"]; !ok {
            return errors.New("round_to_min required for separate_items mode")
        }
    case "preparation_allowance":
        if _, ok := c.ExtraEffortsConfig["minutes_per_unit"]; !ok {
            return errors.New("minutes_per_unit required for preparation_allowance mode")
        }
    default:
        return errors.New("invalid billing mode")
    }
    return nil
}
```

## 6. Frontend Integration Guide

### 6.1 Recording Extra Efforts

```typescript
// components/ExtraEffortForm.tsx
interface ExtraEffortFormProps {
  clientId: number;
  sessionId?: number;
  onSuccess: () => void;
}

const ExtraEffortForm: React.FC<ExtraEffortFormProps> = ({
  clientId,
  sessionId,
  onSuccess
}) => {
  const [formData, setFormData] = useState({
    effort_type: 'preparation',
    effort_date: new Date().toISOString().split('T')[0],
    duration_min: 15,
    description: '',
    billable: true
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const response = await fetch('/api/extra-efforts', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        client_id: clientId,
        session_id: sessionId,
        ...formData
      })
    });
    
    if (response.ok) {
      onSuccess();
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <select 
        value={formData.effort_type}
        onChange={(e) => setFormData({...formData, effort_type: e.target.value})}
      >
        <option value="preparation">Preparation</option>
        <option value="consultation">Consultation</option>
        <option value="parent_meeting">Parent Meeting</option>
        <option value="documentation">Documentation</option>
        <option value="other">Other</option>
      </select>
      
      <input
        type="date"
        value={formData.effort_date}
        onChange={(e) => setFormData({...formData, effort_date: e.target.value})}
      />
      
      <input
        type="number"
        placeholder="Duration (minutes)"
        value={formData.duration_min}
        onChange={(e) => setFormData({...formData, duration_min: parseInt(e.target.value)})}
        min="1"
        max="480"
      />
      
      <textarea
        placeholder="Description"
        value={formData.description}
        onChange={(e) => setFormData({...formData, description: e.target.value})}
      />
      
      <label>
        <input
          type="checkbox"
          checked={formData.billable}
          onChange={(e) => setFormData({...formData, billable: e.target.checked})}
        />
        Billable
      </label>
      
      <button type="submit">Save Extra Effort</button>
    </form>
  );
};
```

### 6.2 Configuring Billing Mode

```typescript
// components/BillingConfigForm.tsx
interface BillingConfigFormProps {
  organizationId: number;
  currentConfig: {
    extra_efforts_billing_mode: string;
    extra_efforts_config: any;
  };
}

const BillingConfigForm: React.FC<BillingConfigFormProps> = ({
  organizationId,
  currentConfig
}) => {
  const [mode, setMode] = useState(currentConfig.extra_efforts_billing_mode);
  const [config, setConfig] = useState(currentConfig.extra_efforts_config);
  const [singleUnitText, setSingleUnitText] = useState(currentConfig.line_item_single_unit_text || 'Einzelstunde');
  const [doubleUnitText, setDoubleUnitText] = useState(currentConfig.line_item_double_unit_text || 'Doppelstunde');

  const handleModeChange = (newMode: string) => {
    setMode(newMode);
    
    // Set default config for each mode
    switch (newMode) {
      case 'ignore':
        setConfig({});
        break;
      case 'bundle_double_units':
        setConfig({
          unit_duration_min: 45,
          threshold_percentage: 90
        });
        break;
      case 'separate_items':
        setConfig({
          round_to_min: 15,
          minimum_duration_min: 10
        });
        break;
      case 'preparation_allowance':
        setConfig({
          minutes_per_unit: 15,
          billing_mode: 'automatic'
        });
        break;
    }
  };

  const handleSave = async () => {
    const response = await fetch(`/api/organizations/${organizationId}/billing-config`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        extra_efforts_billing_mode: mode,
        extra_efforts_config: config,
        line_item_single_unit_text: singleUnitText,
        line_item_double_unit_text: doubleUnitText
      })
    });
    
    if (response.ok) {
      toast.success('Billing configuration updated');
    }
  };

  return (
    <div className="billing-config-form">
      <h3>Extra Efforts Billing Configuration</h3>
      
      <div className="line-item-text-config">
        <h4>Line Item Descriptions</h4>
        <label>
          Single Unit Text:
          <input
            type="text"
            value={singleUnitText}
            onChange={(e) => setSingleUnitText(e.target.value)}
            placeholder="e.g., Einzelstunde, Therapiestunde"
          />
        </label>
        <label>
          Double Unit Text:
          <input
            type="text"
            value={doubleUnitText}
            onChange={(e) => setDoubleUnitText(e.target.value)}
            placeholder="e.g., Doppelstunde, Therapie Doppelstunde"
          />
        </label>
        <p className="hint">
          These texts will be used in invoice line items when billing mode B calculates 1 or 2 units.
        </p>
      </div>
      
      <div className="mode-selector">
        <label>Billing Mode:</label>
        <select value={mode} onChange={(e) => handleModeChange(e.target.value)}>
          <option value="ignore">Ignore (Track Only)</option>
          <option value="bundle_double_units">Bundle into Double Units</option>
          <option value="separate_items">Separate Line Items</option>
          <option value="preparation_allowance">Preparation Allowance</option>
        </select>
      </div>

      {mode === 'bundle_double_units' && (
        <div className="mode-config">
          <label>
            Unit Duration (min):
            <input
              type="number"
              value={config.unit_duration_min}
              onChange={(e) => setConfig({...config, unit_duration_min: parseInt(e.target.value)})}
              options={[45, 60]}
            />
          </label>
          <label>
            Threshold (%):
            <input
              type="number"
              value={config.threshold_percentage}
              onChange={(e) => setConfig({...config, threshold_percentage: parseInt(e.target.value)})}
              min="50"
              max="100"
            />
          </label>
          <p className="hint">
            When session + extra efforts ≥ {config.threshold_percentage}% of 2 units 
            ({config.unit_duration_min * 2 * config.threshold_percentage / 100} min), 
            bill as 2 units.
          </p>
        </div>
      )}

      {mode === 'separate_items' && (
        <div className="mode-config">
          <label>
            Round To (min):
            <select
              value={config.round_to_min}
              onChange={(e) => setConfig({...config, round_to_min: parseInt(e.target.value)})}
            >
              <option value="5">5 minutes</option>
              <option value="15">15 minutes</option>
              <option value="30">30 minutes</option>
            </select>
          </label>
          <label>
            Minimum Duration (min):
            <input
              type="number"
              value={config.minimum_duration_min}
              onChange={(e) => setConfig({...config, minimum_duration_min: parseInt(e.target.value)})}
              min="1"
            />
          </label>
        </div>
      )}

      {mode === 'preparation_allowance' && (
        <div className="mode-config">
          <label>
            Minutes per Unit:
            <input
              type="number"
              value={config.minutes_per_unit}
              onChange={(e) => setConfig({...config, minutes_per_unit: parseInt(e.target.value)})}
              min="1"
              max="60"
            />
          </label>
          <label>
            Billing Mode:
            <select
              value={config.billing_mode}
              onChange={(e) => setConfig({...config, billing_mode: e.target.value})}
            >
              <option value="automatic">Automatic (Always Add)</option>
              <option value="track_actual">Track Actual (Up to Limit)</option>
            </select>
          </label>
          <p className="hint">
            {config.billing_mode === 'automatic' 
              ? `Automatically add ${config.minutes_per_unit} min preparation per unit.`
              : `Bill actual preparation time, up to ${config.minutes_per_unit} min per unit.`
            }
          </p>
        </div>
      )}

      <button onClick={handleSave}>Save Configuration</button>
    </div>
  );
};
```

### 6.3 Invoice Draft Editing

```typescript
// components/InvoiceEditor.tsx
interface InvoiceEditorProps {
  invoiceId: number;
}

const InvoiceEditor: React.FC<InvoiceEditorProps> = ({ invoiceId }) => {
  const [invoice, setInvoice] = useState<Invoice | null>(null);
  const [items, setItems] = useState<InvoiceItem[]>([]);

  useEffect(() => {
    fetchInvoice();
  }, [invoiceId]);

  const fetchInvoice = async () => {
    const response = await fetch(`/api/client-invoices/${invoiceId}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    const data = await response.json();
    setInvoice(data.data);
    
    // Flatten items from all clients
    const allItems = data.data.clients.flatMap(c => 
      c.sessions.map(s => s.invoice_item).concat(
        c.extra_efforts?.map(e => e.invoice_item) || []
      )
    );
    setItems(allItems);
  };

  const updateItem = (itemId: number, updates: Partial<InvoiceItem>) => {
    setItems(items.map(item => 
      item.id === itemId ? { ...item, ...updates } : item
    ));
  };

  const addItem = () => {
    setItems([...items, {
      item_type: 'adjustment',
      description: '',
      number_units: 0,
      unit_price: invoice?.unit_price || 0,
      is_editable: true
    }]);
  };

  const removeItem = (itemId: number) => {
    setItems(items.filter(item => item.id !== itemId));
  };

  const saveChanges = async () => {
    const response = await fetch(`/api/client-invoices/${invoiceId}/items`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ items })
    });
    
    if (response.ok) {
      toast.success('Invoice updated');
      fetchInvoice(); // Refresh to get recalculated totals
    }
  };

  const previewPDF = async () => {
    const response = await fetch(`/api/client-invoices/${invoiceId}/preview-pdf`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    const data = await response.json();
    window.open(data.data.preview_url, '_blank');
  };

  const finalizeInvoice = async () => {
    if (!confirm('Finalize this invoice? You can still revert it later.')) {
      return;
    }
    
    const response = await fetch(`/api/client-invoices/${invoiceId}/finalize`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ generate_pdf: true })
    });
    
    if (response.ok) {
      toast.success('Invoice finalized');
      fetchInvoice();
    }
  };

  if (!invoice) return <div>Loading...</div>;

  return (
    <div className="invoice-editor">
      <div className="invoice-header">
        <h2>Invoice {invoice.invoice_number}</h2>
        <span className={`status-badge status-${invoice.status}`}>
          {invoice.status}
        </span>
      </div>

      <div className="invoice-items">
        <table>
          <thead>
            <tr>
              <th>Type</th>
              <th>Description</th>
              <th>Units</th>
              <th>Unit Price</th>
              <th>Total</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {items.map((item, index) => (
              <tr key={item.id || `new-${index}`}>
                <td>{item.item_type}</td>
                <td>
                  <input
                    type="text"
                    value={item.description}
                    onChange={(e) => updateItem(item.id, { description: e.target.value })}
                    disabled={!item.is_editable && invoice.status !== 'draft'}
                  />
                </td>
                <td>
                  <input
                    type="number"
                    step="0.25"
                    value={item.number_units}
                    onChange={(e) => updateItem(item.id, { number_units: parseFloat(e.target.value) })}
                    disabled={!item.is_editable && invoice.status !== 'draft'}
                  />
                </td>
                <td>
                  <input
                    type="number"
                    step="0.01"
                    value={item.unit_price}
                    onChange={(e) => updateItem(item.id, { unit_price: parseFloat(e.target.value) })}
                    disabled={invoice.status !== 'draft'}
                  />
                </td>
                <td>{(item.number_units * item.unit_price).toFixed(2)} €</td>
                <td>
                  {item.is_editable && invoice.status === 'draft' && (
                    <button onClick={() => removeItem(item.id)}>Delete</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {invoice.status === 'draft' && (
          <button onClick={addItem}>Add Line Item</button>
        )}
      </div>

      <div className="invoice-totals">
        <div>Net Total: {invoice.sum_amount.toFixed(2)} €</div>
        <div>Tax ({invoice.tax_rate}%): {invoice.tax_amount.toFixed(2)} €</div>
        <div className="total">Total: {invoice.total_amount.toFixed(2)} €</div>
      </div>

      <div className="invoice-actions">
        {invoice.status === 'draft' && (
          <>
            <button onClick={saveChanges}>Save Changes</button>
            <button onClick={previewPDF}>Preview PDF</button>
            <button onClick={finalizeInvoice} className="primary">Finalize Invoice</button>
          </>
        )}
        
        {invoice.status === 'finalized' && (
          <button onClick={() => {/* revert to draft */}}>Revert to Draft</button>
        )}
      </div>
    </div>
  );
};
```

### 6.4 Unbilled Sessions View with Extra Efforts

```typescript
// components/UnbilledSessionsView.tsx
const UnbilledSessionsView: React.FC = () => {
  const [unbilledData, setUnbilledData] = useState([]);

  useEffect(() => {
    fetchUnbilledSessions();
  }, []);

  const fetchUnbilledSessions = async () => {
    const response = await fetch(
      `/api/client-invoices/unbilled-sessions`,
      { headers: { 'Authorization': `Bearer ${token}` } }
    );
    const data = await response.json();
    setUnbilledData(data.data);
  };

  return (
    <div className="unbilled-sessions">
      <div className="controls">
        <label>
          <input
            type="checkbox"
            checked={showExtraEfforts}
            onChange={(e) => setShowExtraEfforts(e.target.checked)}
          />
          Include Extra Efforts
        </label>
      </div>

      {unbilledData.map((clientData) => (
        <div key={clientData.client_id} className="client-section">
          <h3>{clientData.client.first_name} {clientData.client.last_name}</h3>
          
          <h4>Unbilled Sessions</h4>
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Type</th>
                <th>Duration</th>
                <th>Units</th>
              </tr>
            </thead>
            <tbody>
              {clientData.sessions.map((session) => (
                <tr key={session.id}>
                  <td>{new Date(session.original_date).toLocaleDateString()}</td>
                  <td>{session.type}</td>
                  <td>{session.duration_min} min</td>
                  <td>{session.number_units}</td>
                </tr>
              ))}
            </tbody>
          </table>
          
          {clientData.extra_efforts?.length > 0 && (
            <>
              <h4>Unbilled Extra Efforts</h4>
              <table>
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Type</th>
                    <th>Duration</th>
                    <th>Description</th>
                    <th>Linked Session</th>
                  </tr>
                </thead>
                <tbody>
                  {clientData.extra_efforts.map((effort) => (
                    <tr key={effort.id}>
                      <td>{new Date(effort.effort_date).toLocaleDateString()}</td>
                      <td>{effort.effort_type}</td>
                      <td>{effort.duration_min} min</td>
                      <td>{effort.description}</td>
                      <td>
                        {effort.session_id ? (
                          <span>Session #{effort.session_id}</span>
                        ) : (
                          <span className="hint">Standalone</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}
          
          <div className="client-info">
            <p className="hint">
              When generating invoice, billing mode will calculate final units based on sessions + extra efforts.
            </p>
          </div>
        </div>
      ))}
    </div>
  );
};
```

## 7. Database Setup (No Migration)

**Approach**: Fresh database creation with updated schema in seed scripts.

### 7.1 Update Seed Scripts

Update the following files to include new schema:

**File: `unburdy_server/seed/schema.sql`** (or equivalent)

```sql
-- Add to existing schema file

-- Extra efforts table
CREATE TABLE extra_efforts (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    session_id INTEGER REFERENCES sessions(id) ON DELETE SET NULL,
    effort_type VARCHAR(50) NOT NULL CHECK (effort_type IN ('preparation', 'consultation', 'parent_meeting', 'documentation', 'other')),
    effort_date DATE NOT NULL,
    duration_min INTEGER NOT NULL CHECK (duration_min > 0 AND duration_min <= 480),
    description TEXT,
    billable BOOLEAN DEFAULT true,
    billing_status VARCHAR(20) DEFAULT 'unbilled' CHECK (billing_status IN ('unbilled', 'billed', 'excluded')),
    invoice_item_id INTEGER REFERENCES invoice_items(id) ON DELETE SET NULL,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_extra_efforts_tenant ON extra_efforts(tenant_id);
CREATE INDEX idx_extra_efforts_client ON extra_efforts(client_id);
CREATE INDEX idx_extra_efforts_session ON extra_efforts(session_id);
CREATE INDEX idx_extra_efforts_billing_status ON extra_efforts(billing_status);
CREATE INDEX idx_extra_efforts_date ON extra_efforts(effort_date);
CREATE INDEX idx_extra_efforts_deleted_at ON extra_efforts(deleted_at);

-- Update sessions table (add column)
ALTER TABLE sessions
ADD COLUMN internal_note TEXT;

-- Update organizations table (add columns)
ALTER TABLE organizations 
ADD COLUMN extra_efforts_billing_mode VARCHAR(50) DEFAULT 'ignore',
ADD COLUMN extra_efforts_config JSONB DEFAULT '{}',
ADD COLUMN line_item_single_unit_text VARCHAR(100) DEFAULT 'Einzelstunde',
ADD COLUMN line_item_double_unit_text VARCHAR(100) DEFAULT 'Doppelstunde';

-- Update invoice_items table (add columns)
ALTER TABLE invoice_items
ADD COLUMN item_type VARCHAR(50) DEFAULT 'session' CHECK (item_type IN ('session', 'extra_effort', 'preparation', 'adjustment')),
ADD COLUMN source_effort_id INTEGER REFERENCES extra_efforts(id) ON DELETE SET NULL,
ADD COLUMN description TEXT,  -- "Therapiestunde", "Therapie Doppelstunde", etc.
ADD COLUMN unit_duration_min INTEGER,
ADD COLUMN is_editable BOOLEAN DEFAULT true;

-- Update client_invoices junction table (add column)
ALTER TABLE client_invoices
ADD COLUMN extra_effort_id INTEGER REFERENCES extra_efforts(id) ON DELETE SET NULL;
```

### 7.2 Sample Data for Testing

**File: `unburdy_server/seed/seed-data.json`** (or equivalent)

Add sample extra efforts (note: `session_id` links effort to a session for billing purposes):

```json
{
  "extra_efforts": [
    {
      "tenant_id": 1,
      "client_id": 1,
      "session_id": 1,  // Linked to session - will be considered during billing calculation
      "effort_type": "preparation",
      "effort_date": "2025-12-15",
      "duration_min": 20,
      "description": "Therapy materials copied and prepared",
      "billable": true,
      "billing_status": "unbilled"
    },
    {
      "tenant_id": 1,
      "client_id": 1,
      "session_id": null,
      "effort_type": "consultation",
      "effort_date": "2025-12-16",
      "duration_min": 30,
      "description": "Teacher consultation regarding client progress",
      "billable": true,
      "billing_status": "unbilled"
    },
    {
      "tenant_id": 1,
      "client_id": 2,
      "session_id": null,
      "effort_type": "parent_meeting",
      "effort_date": "2025-12-17",
      "duration_min": 45,
      "description": "Parent meeting - discussed therapy goals",
      "billable": true,
      "billing_status": "unbilled"
    }
  ],
  "organizations": [
    {
      "extra_efforts_billing_mode": "bundle_double_units",
      "extra_efforts_config": {
        "unit_duration_min": 45,
        "threshold_percentage": 90
      },
      "line_item_single_unit_text": "Therapiestunde",
      "line_item_double_unit_text": "Therapie Doppelstunde"
    }
  ]
}
```

### 7.3 Database Recreation Steps

```bash
# 1. Drop existing database
psql -U postgres -c "DROP DATABASE IF EXISTS unburdy_db;"

# 2. Create fresh database
psql -U postgres -c "CREATE DATABASE unburdy_db;"

# 3. Run updated schema
psql -U postgres -d unburdy_db -f seed/schema.sql

# 4. Run seed data
go run seed/main.go

# 5. Verify tables
psql -U postgres -d unburdy_db -c "\dt"
psql -U postgres -d unburdy_db -c "SELECT * FROM extra_efforts;"
```

## 8. Testing Strategy

### 8.1 Unit Tests

```go
// Test billing calculator for each mode
func TestBillingCalculator_BundleDoubleUnits(t *testing.T) {
    calc := NewBillingCalculator("bundle_double_units", map[string]interface{}{
        "unit_duration_min":    45,
        "threshold_percentage": 90,
    })
    
    // Test case 1: Session + efforts < threshold (should be 1 unit)
    result, err := calc.CalculateUnitsForSession(45, []ExtraEffort{
        {DurationMin: 10},
    })
    assert.NoError(t, err)
    assert.Equal(t, 1, result.Units)
    
    // Test case 2: Session + efforts >= threshold (should be 2 units)
    result, err = calc.CalculateUnitsForSession(45, []ExtraEffort{
        {DurationMin: 20},
        {DurationMin: 15},
    })
    assert.NoError(t, err)
    assert.Equal(t, 2, result.Units)
    assert.Equal(t, 80, result.TotalDurationMin) // 45 + 35
}

func TestInvoiceStateMachine(t *testing.T) {
    service := NewInvoiceService(db)
    
    // Test valid transition: draft -> finalized
    assert.True(t, service.CanTransition("draft", "finalized"))
    
    // Test invalid transition: paid -> draft
    assert.False(t, service.CanTransition("paid", "draft"))
    
    // Test finalize operation
    invoice := createTestInvoice(t, "draft")
    err := service.FinalizeInvoice(invoice.ID, false)
    assert.NoError(t, err)
    
    reloaded, _ := service.GetInvoiceByID(invoice.ID)
    assert.Equal(t, "finalized", reloaded.Status)
    assert.NotNil(t, reloaded.FinalizedAt)
}
```

### 8.2 Integration Tests

```bash
# Test extra efforts creation
./tests/test-extra-efforts.sh

# Test invoice generation with extra efforts
./tests/test-invoice-with-extra-efforts.sh

# Test draft invoice editing
./tests/test-invoice-editing.sh

# Test PDF preview
./tests/test-pdf-preview.sh
```

## 9. Deployment Checklist

- [ ] Update seed scripts with new schema
- [ ] Drop and recreate database from updated seeds
- [ ] Update environment variables (if any)
- [ ] Deploy base-server with new endpoints
- [ ] Deploy unburdy_server with updated invoice generation
- [ ] Update Swagger documentation
- [ ] Test extra efforts recording
- [ ] Test each billing mode configuration
- [ ] Test draft invoice editing with German descriptions ("Therapiestunde", "Therapie Doppelstunde")
- [ ] Test PDF preview generation
- [ ] Verify invoice status state machine (draft → finalized → sent → paid)
- [ ] Train users on new features
- [ ] Update user documentation

## 10. Future Enhancements

### 10.1 Phase 2 Features
- Bulk extra effort import (CSV)
- Extra effort templates (common activities)
- Time tracking integration
- Mobile app for recording efforts on-the-go
- Analytics: time spent vs billed ratio

### 10.2 Phase 3 Features
- Automated reminders for unbilled efforts
- Client-specific billing rules
- Multi-currency support for international clients
- Integration with accounting software (DATEV, etc.)

## 11. Glossary

- **Extra Effort**: Therapeutic work beyond scheduled sessions (preparation, consultations, meetings)
- **Billing Mode**: Method for calculating invoice units from sessions and extra efforts
- **Draft Invoice**: Editable invoice before finalization
- **Finalized Invoice**: Locked invoice, ready for sending
- **Unit**: Billable time increment (typically 45 or 60 minutes)
- **Bundle**: Combining session and extra effort time into higher unit count

---

**Document Version**: 1.0  
**Last Updated**: 2025-12-30  
**Author**: Technical Architecture Team  
**Status**: Draft for Review
