# Template System – Module‑Driven Requirements Document

## 1. Purpose

This document defines the functional and non‑functional requirements for a **module‑driven template system** that supports:

* Multiple business modules (e.g. Billing, Identity)
* Multiple template purposes per module (e.g. Invoice, Password Reset)
* Multiple output channels per purpose (Email, PDF/Document)
* Strongly defined template variables owned by modules
* Safe rendering, previewing, and versioning of templates

The goal is to provide a **generic, reusable template platform** while keeping **business semantics and data ownership inside the modules** that use the templates.

---

## 2. Scope

### In Scope

* Email templates (HTML output)
* Document templates (PDF is rendered from HTML output in other module)
* Module‑defined template contracts
* Template preview with sample data
* Variable validation
* Versioning and compatibility handling

### Out of Scope

* Email sending infrastructure
* PDF storage and archival policies
* Localization / i18n (future extension)
* Access control beyond basic authorization

---

## 3. Definitions

| Term              | Definition                                         |
| ----------------- | -------------------------------------------------- |
| Module            | A bounded business domain (e.g. Billing, Identity) |
| Template Contract | A module‑owned definition of a template purpose    |
| Template Instance | A channel‑specific implementation of a contract    |
| Channel           | Output target (EMAIL, DOCUMENT)                    |
| Rendering         | Converting a template + data into final output     |

---

## 4. Architectural Principles

1. **Modules own meaning**
   Modules define template purposes, variables, and constraints.

2. **Template system is domain‑agnostic**
   The template service does not know business concepts like “invoice”.

3. **Single contract, multiple channels**
   One template purpose may have multiple channel‑specific implementations.

4. **Preview equals production rendering**
   Previews use the exact same pipeline as real rendering.

5. **Fail fast and explicitly**
   Invalid data or incompatible templates must fail early and visibly.

---

## 5. Domain Model

### 5.1 Module Template Contract

A module template contract defines **what can be rendered**, not how.

**Ownership:** Module

```json
{
  "module": "billing",
  "template_key": "invoice",
  "description": "Customer invoice communication",
  "supported_channels": ["EMAIL", "DOCUMENT"],
  "variable_schema": { /* structured schema */ },
  "default_sample_data": { /* valid example data */ }
}
```

#### Rules

* (`module`, `template_key`) must be unique
* Variable schema applies to **all channels**
* Supported channels are authoritative

---

### 5.2 Template Instance

A template instance defines **presentation and layout** for one channel.

**Ownership:** Template Service

```json
{
  "id": 101,
  "module": "billing",
  "template_key": "invoice",
  "channel": "EMAIL",
  "subject": "Your invoice {{ invoice.number }}",
  "body": "<html>...</html>",
  "version": 1,
  "is_active": true
}
```

#### Channel Rules

* EMAIL:

  * `subject` is required
  * Output is HTML
* DOCUMENT:

  * `subject` is not allowed
  * Output is PDF (via HTML → PDF)

---

## 6. Functional Requirements

### FR‑1: Module Contract Registration

Modules MUST be able to register template contracts.

* Registration occurs at startup or deployment
* Contracts include:

  * Template key
  * Supported channels
  * Variable schema
  * Default sample data

**Failure Conditions**

* Duplicate (`module`, `template_key`)
* Invalid schema definition

---

### FR‑2: Template Creation Bound to Contracts

Templates MUST reference an existing module contract.

* Template creation requires:

  * module
  * template_key
  * channel
* System enforces:

  * Channel is supported by contract
  * Variables used exist in schema

---

### FR‑3: Variable Schema Enforcement

The system MUST validate template usage against the module’s variable schema.

Validation applies to:

* Template save
* Preview rendering
* Production rendering

Rules:

* Unknown variables are rejected
* Missing required variables cause rendering failure
* Nested objects and arrays must conform to schema

---

### FR‑4: Preview Rendering

Users MUST be able to preview templates with realistic data.

Preview data resolution order:

1. Explicit preview override
2. Template‑specific sample override (optional)
3. Module default sample data

Preview rendering MUST:

* Use the production rendering pipeline
* Produce identical output to real rendering
* Have no side effects

---

### FR‑5: Multi‑Channel Support per Template Purpose

A single template purpose MAY have multiple channel implementations.

Example:

* billing / invoice / EMAIL
* billing / invoice / DOCUMENT

All channel implementations:

* Share the same variable schema
* Use the same sample data

---

### FR‑6: Rendering API

The system MUST expose a unified rendering interface.

```json
render({
  "module": "billing",
  "template_key": "invoice",
  "channel": "EMAIL",
  "data": { ... }
})
```

Rendering Steps:

1. Contract lookup
2. Schema validation
3. Template compilation
4. Channel‑specific output adaptation

---

### FR‑7: Versioning

Templates MUST be versioned.

Rules:

* Activated templates are immutable
* Changes create a new version
* Rendering can target:

  * Explicit version
  * Default active version

Historical renders MUST remain reproducible.

---

### FR‑8: Compatibility Management

Schema changes MUST be classified:

* Backward‑compatible

  * Adding optional fields
* Breaking

  * Removing or renaming fields

On breaking changes:

* Affected templates are flagged
* Rendering is blocked until resolved

---

### FR‑9: Introspection API

The system MUST expose template metadata for UIs and tooling.

```http
GET /templates/{id}/schema
```

Returns:

* Variable schema
* Sample data
* Supported channels
* Version information

---

## 7. Non‑Functional Requirements

### NFR‑1: Performance

* Email rendering < 200 ms (P95)
* PDF rendering < 2 s (P95)

### NFR‑2: Reliability

* Rendering failures must be deterministic and logged
* Invalid templates must not reach production use

### NFR‑3: Security

* Templates must not execute arbitrary code
* Data is treated as untrusted input

### NFR‑4: Extensibility

* New modules require no template system changes
* New channels can be added with isolated adapters

---

## 8. Example: Invoice Email + PDF

* One contract: `billing.invoice`
* Two templates:

  * invoice / EMAIL
  * invoice / DOCUMENT
* Same data, different presentation

Email:

* Subject line
* Short explanation
* PDF attached (out of scope here)

PDF:

* Printable layout
* Page headers and footers

---

## 9. Open Questions / Future Extensions

* Localization strategy
* Channel‑specific optional variables
* Template inheritance
* Approval workflows
* Audit logging of renders

---

## 10. MVP Cut

This section defines the **Minimum Viable Product (MVP)** for the template system. The MVP focuses on enabling **safe, module-driven rendering of email and PDF templates** with previews, while deliberately excluding advanced lifecycle and governance features.

---

### 10.1 MVP Goals

The MVP MUST:

* Support module-defined template contracts
* Support invoice-style use cases with email + PDF
* Provide reliable preview rendering
* Enforce variable correctness at render time

The MVP MUST NOT:

* Include localization
* Include approval workflows
* Include complex compatibility migration tooling

---

### 10.2 MVP Scope (Included)

The MVP explicitly includes the following capabilities:

#### Template Channels & Output

* Email templates

  * HTML output
  * Plain text output
* Document templates

  * PDF output (HTML → PDF)

#### Module & Contract Model

* Module-defined template contracts
* Variable definitions stored as JSONB arrays
* Sample data for preview stored as JSONB objects

#### Rendering & Preview

* Template preview using sample data
* Template preview output as HTML or plain text
* Template rendering using custom runtime data
* Rendering output as HTML or plain text
* Variable validation at render and preview time

#### Multi-Tenancy & Ownership

* Tenant isolation (hard separation of data)
* Organization-specific templates
* System-level default templates as fallback
* Default template per (module, template_key, channel, organization)

#### Storage & Assets

* Template content stored in MinIO
* Template assets (images, logos) stored in MinIO with structured paths
* **Public Asset Endpoint**: Template module provides public HTTP endpoint for asset delivery
* **Frontend HTML Preview**: Ability to serve rendered HTML (including embedded/referenced images) to frontend for display in a div element

---

### 10.3 MVP Functional Requirements

#### MVP-FR-1: Module Template Contract Registration

Modules MUST be able to register template contracts with:

* module
* template_key
* supported_channels
* variable_schema
* default_sample_data

Contracts MAY be updated by redeploying the module.

---

#### MVP-FR-2: Channel-Specific Template Instances

The system MUST allow creation of templates bound to:

* module
* template_key
* channel

Rules:

* EMAIL templates require subject + body
* DOCUMENT templates require body only

---

#### MVP-FR-3: Variable Schema Validation at Render Time

At render and preview time, the system MUST:

* Reject unknown variables
* Reject missing required variables
* Validate nested structures and arrays

Template save-time validation is OPTIONAL in MVP.

---

#### MVP-FR-4: Preview Rendering

The system MUST support preview rendering using:

1. Explicit preview data (if provided)
2. Module default sample data

Preview rendering MUST:

* Use the same pipeline as production rendering
* Produce HTML (EMAIL) or PDF (DOCUMENT)
* Support frontend display via HTML endpoint with properly resolved asset URLs

---

#### MVP-FR-4.1: Frontend HTML Preview Endpoint

The system MUST provide an endpoint to serve rendered HTML for frontend display:

```http
GET /templates/{id}/preview?format=html
GET /templates/render?module=billing&template_key=invoice&channel=EMAIL&format=html
```

Requirements:

* Returns rendered HTML with all variables resolved
* Images and assets must be accessible via absolute URLs or data URIs
* Response includes proper Content-Type header (text/html)
* CORS-compatible for frontend consumption
* Supports both EMAIL and DOCUMENT channel templates

Asset Resolution
  * Proxy images through template module public endpoint
  * Template assets stored in MinIO at: `templates/{tenant_id}/assets/{template_id}/{filename}`
  * Public endpoint: `GET /api/public/templates/assets/{tenant_id}/{template_id}/{filename}`
  * No authentication required (public access)
  * Backend retrieves from MinIO and proxies to client
  * Enables proper CORS, caching headers, and CDN integration
---

#### MVP-FR-5: Unified Render API

The system MUST expose a single render interface:

```json
render({
  "module": "billing",
  "template_key": "invoice",
  "channel": "EMAIL",
  "data": { ... }
})
```

---

#### MVP-FR-6: Public Asset Delivery

The system MUST provide a public endpoint for serving template assets:

```http
GET /api/public/templates/assets/{tenant_id}/{template_id}/{filename}
```

Requirements:

* No authentication required (publicly accessible)
* Returns image/file with appropriate Content-Type header
* Supports common image formats (PNG, JPG, SVG)
* Implements caching headers (Cache-Control, ETag)
* Tenant and template scoping prevents unauthorized access
* Returns 404 for non-existent or unauthorized assets

Asset Storage Structure in MinIO:

```
templates/
  {tenant_id}/
    assets/
      {template_id}/
        logo.png
        header-image.jpg
        signature.png
```

Usage in Templates:

```html
<img src="/api/public/templates/assets/1/invoice-email/logo.png" alt="Company Logo">
```

---

### 10.4 MVP Non-Functional Requirements

* Rendering errors must be explicit and logged
* PDF rendering may be synchronous
* Performance targets are best-effort

---

### 10.5 Explicitly Excluded from MVP

The following features are **out of scope for MVP**:

* Template version pinning
* Backward-compatibility analysis
* Breaking-change detection
* Localization / multi-language templates
* Channel-specific variable subsets
* Approval workflows
* Audit logging of renders
* Template inheritance

---

### 10.6 MVP Success Criteria

The MVP is considered successful when:

* A module can register an `invoice` template contract
* An invoice EMAIL template can be rendered and previewed
* An invoice PDF template can be rendered and previewed
* **Rendered HTML can be displayed in frontend div** with all images properly loaded
* Invalid data reliably causes render failure
* The system works without any invoice-specific logic

---

## 11. Summary

The MVP delivers a **lean but correct** template system:

* Strong contracts
* Safe rendering
* Email + PDF support

All advanced lifecycle and governance concerns are deferred, ensuring fast delivery while preserving a solid architectural foundation.
