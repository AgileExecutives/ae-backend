# Template Module Refactoring Plan

## Executive Summary

Refactor the existing template module in `base-server/modules/templates` to fulfill the requirements defined in `New_Template_Module_Req.md`, implementing a **module-driven template system** with contracts, channels, and robust rendering.

---

## Current State Analysis

### What We Have

**Entities:**
- `Template` - Single monolithic entity
- Fields: `TemplateType`, `Name`, `Variables`, `SampleData`
- Storage in MinIO via `StorageKey`

**Services:**
- `TemplateService` - CRUD operations
- Basic rendering with Go `html/template`
- No contract/channel separation
- No variable validation

**API:**
- POST `/templates` - Create
- GET `/templates/{id}` - Get
- PUT `/templates/{id}` - Update
- DELETE `/templates/{id}` - Delete

### What's Missing (Per Requirements)

1. **Template Contracts** - Module-owned definitions
2. **Channel Separation** - EMAIL vs DOCUMENT
3. **Variable Schema Validation** - Runtime validation
4. **Unified Render API** - `render(module, template_key, channel, data)`
5. **Public Asset Endpoint** - `/api/public/templates/assets/{tenant}/{template}/{file}`
6. **Preview API** - HTML output for frontend display
7. **Contract Registration** - Module startup registration

---

## Gap Analysis

| Requirement | Current Status | Gap |
|-------------|----------------|-----|
| Module Contracts (FR-1) | ❌ Missing | Need `TemplateContract` entity |
| Channel Support (FR-5) | ❌ Missing | Need `Channel` enum and template-contract binding |
| Variable Validation (FR-3) | ❌ Missing | Need schema validation engine |
| Unified Render API (FR-6) | ❌ Missing | Need render service with lookup |
| Public Assets (MVP-FR-6) | ❌ Missing | Need public endpoint + MinIO proxy |
| HTML Preview (MVP-FR-4.1) | ❌ Missing | Need preview endpoint |
| Multi-tenant (MVP) | ✅ Partial | Have tenant_id, need org-level defaults |

---

## Refactoring Strategy

### Phase 1: Data Model Migration
Add new entities alongside existing ones (non-breaking).

### Phase 2: Contract Registration
Implement contract management without touching templates.

### Phase 3: Rendering Engine
Build new render service with validation.

### Phase 4: Asset Management
Implement public asset endpoint.

### Phase 5: API Migration
Expose new endpoints, deprecate old ones.

### Phase 6: Cleanup
Remove old code after migration.

---

## Detailed Implementation Plan

### **Phase 1: Data Model & Migrations** (Day 1-2)

#### Step 1.1: Create TemplateContract Entity

**File:** `base-server/modules/templates/entities/contract.go`

```go
type TemplateContract struct {
    ID                uint
    Module            string // "billing", "identity"
    TemplateKey       string // "invoice", "password_reset"
    Description       string
    SupportedChannels []string // ["EMAIL", "DOCUMENT"]
    VariableSchema    datatypes.JSON
    DefaultSampleData datatypes.JSON
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

// Unique index on (module, template_key)
```

**Migration:**
```sql
CREATE TABLE template_contracts (
    id SERIAL PRIMARY KEY,
    module VARCHAR(100) NOT NULL,
    template_key VARCHAR(100) NOT NULL,
    description TEXT,
    supported_channels JSONB NOT NULL,
    variable_schema JSONB,
    default_sample_data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(module, template_key)
);
```

**Test:**
- [ ] Create contract with unique module+key
- [ ] Reject duplicate module+key
- [ ] Validate JSONB fields parse correctly

---

#### Step 1.2: Add Channel to Template Entity

**File:** `base-server/modules/templates/entities/template.go`

```go
type Channel string

const (
    ChannelEmail    Channel = "EMAIL"
    ChannelDocument Channel = "DOCUMENT"
)

type Template struct {
    // ... existing fields ...
    
    // NEW FIELDS
    Module       string  // Reference to contract
    TemplateKey  string  // Reference to contract
    Channel      Channel // EMAIL or DOCUMENT
    Subject      *string // Required for EMAIL, null for DOCUMENT
    
    // DEPRECATED (keep for backward compat)
    TemplateType string `gorm:"column:template_type"` // Keep during migration
}
```

**Migration:**
```sql
ALTER TABLE templates 
ADD COLUMN module VARCHAR(100),
ADD COLUMN template_key VARCHAR(100),
ADD COLUMN channel VARCHAR(20),
ADD COLUMN subject TEXT;

CREATE INDEX idx_template_contract ON templates(module, template_key, channel);
```

**Test:**
- [ ] Insert template with channel
- [ ] Query by (module, template_key, channel)
- [ ] Email template requires subject
- [ ] Document template subject is null

---

### **Phase 2: Contract Management** (Day 3)

#### Step 2.1: Contract Service

**File:** `base-server/modules/templates/services/contract_service.go`

```go
type ContractService struct {
    db *gorm.DB
}

func (s *ContractService) RegisterContract(req *RegisterContractRequest) error
func (s *ContractService) GetContract(module, templateKey string) (*entities.TemplateContract, error)
func (s *ContractService) ListContracts(module string) ([]entities.TemplateContract, error)
func (s *ContractService) ValidateChannel(module, templateKey, channel string) error
```

**Test:**
- [ ] Register new contract
- [ ] Update existing contract
- [ ] Reject invalid channel for contract
- [ ] List all contracts for module

---

#### Step 2.2: Contract Registration API

**File:** `base-server/modules/templates/handlers/contract_handler.go`

```go
// POST /templates/contracts
func (h *ContractHandler) RegisterContract(c *gin.Context)

// GET /templates/contracts
func (h *ContractHandler) ListContracts(c *gin.Context)

// GET /templates/contracts/{module}/{template_key}
func (h *ContractHandler) GetContract(c *gin.Context)
```

**Test:**
- [ ] POST contract returns 201
- [ ] GET contract returns full schema
- [ ] LIST filters by module

---

### **Phase 3: Variable Validation** (Day 4)

#### Step 3.1: Schema Validator

**File:** `base-server/modules/templates/services/validator.go`

```go
type SchemaValidator struct{}

func (v *SchemaValidator) Validate(schema datatypes.JSON, data map[string]interface{}) error {
    // Validate data against schema
    // Check required fields
    // Check types
    // Check nested objects
}
```

**Test:**
- [ ] Valid data passes
- [ ] Missing required field fails
- [ ] Wrong type fails
- [ ] Unknown field fails (strict mode)
- [ ] Nested object validation

---

### **Phase 4: Rendering Engine** (Day 5-6)

#### Step 4.1: Render Service

**File:** `base-server/modules/templates/services/render_service.go`

```go
type RenderService struct {
    db              *gorm.DB
    storage         storage.DocumentStorage
    contractService *ContractService
    validator       *SchemaValidator
}

func (s *RenderService) Render(ctx context.Context, req *RenderRequest) (*RenderResponse, error) {
    // 1. Lookup contract
    contract, err := s.contractService.GetContract(req.Module, req.TemplateKey)
    
    // 2. Validate channel
    if !contract.SupportsChannel(req.Channel) {
        return nil, ErrUnsupportedChannel
    }
    
    // 3. Validate data against schema
    if err := s.validator.Validate(contract.VariableSchema, req.Data); err != nil {
        return nil, err
    }
    
    // 4. Find template
    template, err := s.getActiveTemplate(req.TenantID, req.Module, req.TemplateKey, req.Channel)
    
    // 5. Load content from MinIO
    content, err := s.storage.Retrieve(ctx, "templates", template.StorageKey)
    
    // 6. Render with Go html/template
    tmpl, err := template.New("render").Parse(string(content))
    output, err := tmpl.Execute(req.Data)
    
    return &RenderResponse{
        Output: output,
        Format: req.Channel,
    }, nil
}

type RenderRequest struct {
    Module      string
    TemplateKey string
    Channel     Channel
    TenantID    uint
    OrgID       *uint
    Data        map[string]interface{}
}
```

**Test:**
- [ ] Render with valid data succeeds
- [ ] Invalid data rejected
- [ ] Unknown module/key returns 404
- [ ] Unsupported channel rejected
- [ ] Variables interpolated correctly

---

#### Step 4.2: Preview Service

**File:** `base-server/modules/templates/services/preview_service.go`

```go
func (s *RenderService) Preview(ctx context.Context, req *PreviewRequest) (*RenderResponse, error) {
    contract, _ := s.contractService.GetContract(req.Module, req.TemplateKey)
    
    // Use explicit preview data OR default sample data
    data := req.PreviewData
    if data == nil {
        data = contract.DefaultSampleData
    }
    
    return s.Render(ctx, &RenderRequest{
        Module:      req.Module,
        TemplateKey: req.TemplateKey,
        Channel:     req.Channel,
        Data:        data,
    })
}
```

**Test:**
- [ ] Preview with sample data
- [ ] Preview with override data
- [ ] HTML output for EMAIL channel
- [ ] HTML output for DOCUMENT channel

---

### **Phase 5: Asset Management** (Day 7)

#### Step 5.1: Asset Storage Structure

**MinIO Structure:**
```
templates/
  {tenant_id}/
    assets/
      {template_id}/
        logo.png
        header.jpg
```

---

#### Step 5.2: Public Asset Endpoint

**File:** `base-server/modules/templates/handlers/asset_handler.go`

```go
// GET /api/public/templates/assets/{tenant_id}/{template_id}/{filename}
func (h *AssetHandler) GetAsset(c *gin.Context) {
    tenantID := c.Param("tenant_id")
    templateID := c.Param("template_id")
    filename := c.Param("filename")
    
    // Construct MinIO key
    key := fmt.Sprintf("templates/%s/assets/%s/%s", tenantID, templateID, filename)
    
    // Retrieve from MinIO
    data, err := h.storage.Retrieve(ctx, "templates", key)
    if err != nil {
        c.JSON(404, gin.H{"error": "asset not found"})
        return
    }
    
    // Determine content type
    contentType := mime.TypeByExtension(filepath.Ext(filename))
    
    // Set caching headers
    c.Header("Content-Type", contentType)
    c.Header("Cache-Control", "public, max-age=86400")
    
    c.Data(200, contentType, data)
}
```

**Routes:**
```go
// Public routes (no auth required)
public := router.Group("/api/public/templates")
{
    public.GET("/assets/:tenant_id/:template_id/*filename", handler.GetAsset)
}
```

**Test:**
- [ ] GET asset returns image
- [ ] Correct content-type
- [ ] 404 for missing asset
- [ ] Cache headers present
- [ ] No auth required

---

#### Step 5.3: Asset Upload API

**File:** `base-server/modules/templates/handlers/template_handler.go`

```go
// POST /templates/{id}/assets
func (h *TemplateHandler) UploadAsset(c *gin.Context) {
    templateID := c.Param("id")
    file, _ := c.FormFile("file")
    
    // Validate file type
    if !isAllowedAssetType(file.Filename) {
        c.JSON(400, gin.H{"error": "invalid file type"})
        return
    }
    
    // Read file
    data, _ := file.Open()
    content, _ := io.ReadAll(data)
    
    // Store in MinIO
    key := fmt.Sprintf("templates/%d/assets/%s/%s", 
        tenantID, templateID, file.Filename)
    
    _, err := h.storage.Store(ctx, storage.StoreRequest{
        Bucket: "templates",
        Key: key,
        Data: content,
        ContentType: mime.TypeByExtension(filepath.Ext(file.Filename)),
    })
    
    // Return public URL
    publicURL := fmt.Sprintf("/api/public/templates/assets/%d/%s/%s",
        tenantID, templateID, file.Filename)
    
    c.JSON(200, gin.H{"url": publicURL})
}
```

**Test:**
- [ ] Upload PNG succeeds
- [ ] Upload JPG succeeds
- [ ] Upload SVG succeeds
- [ ] Reject .exe file
- [ ] Returns public URL
- [ ] Asset retrievable via public endpoint

---

### **Phase 6: API Integration** (Day 8-9)

#### Step 6.1: Unified Render Endpoint

**File:** `base-server/modules/templates/handlers/render_handler.go`

```go
// POST /templates/render
func (h *RenderHandler) Render(c *gin.Context) {
    var req services.RenderRequest
    c.BindJSON(&req)
    
    req.TenantID, _ = baseAPI.GetTenantID(c)
    
    result, err := h.renderService.Render(c.Request.Context(), &req)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "output": result.Output,
        "channel": result.Format,
    })
}
```

**Request:**
```json
{
  "module": "billing",
  "template_key": "invoice",
  "channel": "EMAIL",
  "data": {
    "invoice_number": "INV-001",
    "customer_name": "John Doe"
  }
}
```

**Response:**
```json
{
  "output": "<html>...</html>",
  "channel": "EMAIL"
}
```

**Test:**
- [ ] Render EMAIL template
- [ ] Render DOCUMENT template
- [ ] Returns HTML output
- [ ] Variables replaced
- [ ] Invalid data rejected

---

#### Step 6.2: Preview Endpoint

**File:** `base-server/modules/templates/handlers/render_handler.go`

```go
// GET /templates/preview?module=billing&template_key=invoice&channel=EMAIL&format=html
func (h *RenderHandler) Preview(c *gin.Context) {
    module := c.Query("module")
    templateKey := c.Query("template_key")
    channel := c.Query("channel")
    format := c.DefaultQuery("format", "html")
    
    result, err := h.renderService.Preview(c.Request.Context(), &services.PreviewRequest{
        Module:      module,
        TemplateKey: templateKey,
        Channel:     services.Channel(channel),
        TenantID:    getTenantID(c),
    })
    
    if format == "html" {
        c.Header("Content-Type", "text/html")
        c.String(200, result.Output)
    } else {
        c.JSON(200, gin.H{"output": result.Output})
    }
}

// GET /templates/{id}/preview?format=html
func (h *RenderHandler) PreviewByID(c *gin.Context) {
    templateID := c.Param("id")
    
    // Load template to get module/key/channel
    template, _ := h.templateService.GetTemplate(ctx, tenantID, templateID)
    
    result, err := h.renderService.Preview(ctx, &services.PreviewRequest{
        Module:      template.Module,
        TemplateKey: template.TemplateKey,
        Channel:     template.Channel,
        TenantID:    tenantID,
    })
    
    c.Header("Content-Type", "text/html")
    c.String(200, result.Output)
}
```

**Test:**
- [ ] Preview returns HTML
- [ ] Can be displayed in iframe
- [ ] Images load from public asset endpoint
- [ ] Sample data used by default
- [ ] Override data works

---

### **Phase 7: Testing & Validation** (Day 10)

#### Integration Tests

**File:** `base-server/modules/templates/tests/integration_test.go`

```go
func TestEndToEndInvoiceWorkflow(t *testing.T) {
    // 1. Register contract
    contract := RegisterContract("billing", "invoice", []string{"EMAIL", "DOCUMENT"})
    
    // 2. Create EMAIL template
    emailTemplate := CreateTemplate(contract, "EMAIL", emailHTML)
    
    // 3. Upload logo asset
    UploadAsset(emailTemplate.ID, "logo.png", logoData)
    
    // 4. Preview template
    preview := PreviewTemplate(emailTemplate.ID)
    assert.Contains(t, preview, "logo.png")
    
    // 5. Render with real data
    output := RenderTemplate("billing", "invoice", "EMAIL", invoiceData)
    assert.Contains(t, output, "INV-001")
    
    // 6. Access asset via public URL
    asset := GetPublicAsset(tenantID, emailTemplate.ID, "logo.png")
    assert.NotNil(t, asset)
}
```

**Test Cases:**
- [ ] Full invoice EMAIL workflow
- [ ] Full invoice DOCUMENT workflow
- [ ] Multi-tenant isolation
- [ ] Organization-level defaults
- [ ] Invalid data rejection
- [ ] Missing template handling
- [ ] Asset access control

---

## Migration Strategy

### Backward Compatibility

**Option 1: Dual API (Recommended)**
- Keep existing `/templates` endpoints
- Add new `/templates/contracts` and `/templates/render`
- Gradually migrate clients
- Deprecate old API in 6 months

**Option 2: Data Migration**
- Script to convert old templates to contracts
- Map `TemplateType` → `Module + TemplateKey`
- Set default channel based on type

---

## Testing Checklist

### Unit Tests
- [ ] Contract CRUD operations
- [ ] Template CRUD with contracts
- [ ] Schema validation logic
- [ ] Render pipeline
- [ ] Asset storage/retrieval

### Integration Tests
- [ ] Contract registration
- [ ] Template creation bound to contract
- [ ] Render with validation
- [ ] Preview with sample data
- [ ] Public asset access

### E2E Tests
- [ ] Invoice EMAIL template workflow
- [ ] Invoice DOCUMENT template workflow
- [ ] Frontend HTML preview display
- [ ] Multi-tenant isolation
- [ ] Error handling

---

## Success Criteria

MVP is complete when:

1. ✅ Billing module can register `invoice` contract
2. ✅ EMAIL template can be created for invoice
3. ✅ DOCUMENT template can be created for invoice
4. ✅ Render API validates data against schema
5. ✅ Preview endpoint serves HTML with images
6. ✅ Public asset endpoint works without auth
7. ✅ Invalid data reliably rejected
8. ✅ No invoice-specific logic in template module

---

## Timeline

| Phase | Days | Deliverable |
|-------|------|-------------|
| Phase 1 | 2 | Data model + migrations |
| Phase 2 | 1 | Contract management |
| Phase 3 | 1 | Variable validation |
| Phase 4 | 2 | Rendering engine |
| Phase 5 | 1 | Asset management |
| Phase 6 | 2 | API integration |
| Phase 7 | 1 | Testing |
| **Total** | **10 days** | **MVP Complete** |

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing templates | High | Dual API, data migration |
| Schema validation too strict | Medium | Configurable strict/loose mode |
| Performance of render pipeline | Medium | Caching, template compilation |
| Asset CDN requirement | Low | Public endpoint supports CDN proxy |

---

## Next Steps

1. **Review this plan** with team
2. **Create Phase 1 branch** `feature/template-contracts`
3. **Write migrations** for TemplateContract
4. **Implement Step 1.1** (TemplateContract entity)
5. **Test contract CRUD** before proceeding

---

## Questions for Discussion

1. Should we support JSON Schema for variable schemas?
2. Do we need template approval workflow in MVP?
3. Should preview be cached?
4. CDN strategy for public assets?
5. Rollback strategy if migration fails?
