package handlers

// Swagger request/response type definitions for documentation

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Error message"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation successful"`
}

// DocumentUploadRequest represents document upload parameters
type DocumentUploadRequest struct {
	File           interface{} `json:"file" swaggertype:"file" example:"binary"`           // File to upload
	DocumentType   string      `json:"document_type" binding:"required" example:"invoice"` // Document type
	Bucket         string      `json:"bucket" example:"documents"`                         // Storage bucket
	Path           string      `json:"path" example:"invoices/2025/invoice-001.pdf"`       // Storage path
	ReferenceType  string      `json:"reference_type" example:"invoice"`                   // Reference type
	ReferenceID    uint        `json:"reference_id" example:"123"`                         // Reference ID
	OrganizationID uint        `json:"organization_id" example:"10"`                       // Organization ID
}

// ListDocumentsResponse represents documents list response
type ListDocumentsResponse struct {
	Success    bool                  `json:"success" example:"true"`
	Data       []DocumentResponseDTO `json:"data"`
	Total      int                   `json:"total" example:"42"`
	Page       int                   `json:"page" example:"1"`
	PageSize   int                   `json:"page_size" example:"20"`
	TotalPages int                   `json:"total_pages" example:"3"`
}

// DocumentResponseDTO represents a single document response
type DocumentResponseDTO struct {
	ID             uint                   `json:"id" example:"1"`
	TenantID       uint                   `json:"tenant_id" example:"1"`
	OrganizationID *uint                  `json:"organization_id,omitempty" example:"10"`
	UserID         uint                   `json:"user_id" example:"5"`
	DocumentType   string                 `json:"document_type" example:"invoice"`
	ReferenceType  string                 `json:"reference_type,omitempty" example:"invoice"`
	ReferenceID    *uint                  `json:"reference_id,omitempty" example:"123"`
	FileName       string                 `json:"file_name" example:"invoice-001.pdf"`
	FileSizeBytes  int64                  `json:"file_size_bytes" example:"102400"`
	ContentType    string                 `json:"content_type" example:"application/pdf"`
	StorageKey     string                 `json:"storage_key" example:"tenants/1/documents/invoice/invoice-001.pdf"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      string                 `json:"created_at" example:"2025-12-26T10:00:00Z"`
}

// DownloadURLResponse represents download URL response
type DownloadURLResponse struct {
	Success     bool   `json:"success" example:"true"`
	Message     string `json:"message" example:"Download URL generated successfully"`
	DocumentID  uint   `json:"document_id" example:"1"`
	FileName    string `json:"file_name" example:"invoice-001.pdf"`
	DownloadURL string `json:"download_url" example:"https://minio.example.com/..."`
	ExpiresAt   string `json:"expires_at" example:"2025-12-26T11:00:00Z"`
}

// GenerateInvoiceNumberRequest represents invoice number generation request
type GenerateInvoiceNumberRequest struct {
	OrganizationID uint   `json:"organization_id" binding:"required" example:"10"`
	Prefix         string `json:"prefix" example:"INV"`
	YearFormat     string `json:"year_format" example:"YYYY"`              // "YYYY" or "YY"
	MonthFormat    string `json:"month_format" example:"MM"`               // "MM", "M", or ""
	Padding        int    `json:"padding" example:"4"`                     // e.g., 4 for "0001"
	Separator      string `json:"separator" example:"-"`                   // e.g., "-"
	ResetMonthly   *bool  `json:"reset_monthly,omitempty" example:"false"` // pointer to distinguish false from not set
	Year           *int   `json:"year,omitempty" example:"2025"`
	Month          *int   `json:"month,omitempty" example:"12"`
}

// InvoiceNumberResponse represents invoice number generation response
type InvoiceNumberResponse struct {
	Success       bool   `json:"success" example:"true"`
	InvoiceNumber string `json:"invoice_number" example:"INV-2025-0001"`
	Sequence      int    `json:"sequence" example:"1"`
	Year          int    `json:"year" example:"2025"`
	Month         int    `json:"month" example:"12"`
}

// CreateTemplateRequest represents template creation request
type CreateTemplateRequest struct {
	OrganizationID *uint                  `json:"organization_id,omitempty" example:"10"`
	TemplateType   string                 `json:"template_type" binding:"required" example:"invoice"`
	Name           string                 `json:"name" binding:"required" example:"Standard Invoice"`
	Description    string                 `json:"description" example:"Default invoice template"`
	Content        string                 `json:"content" binding:"required" example:"<!DOCTYPE html>..."`
	Variables      []string               `json:"variables" example:"invoice_number,customer_name"`
	SampleData     map[string]interface{} `json:"sample_data,omitempty"`
	IsActive       bool                   `json:"is_active" example:"true"`
	IsDefault      bool                   `json:"is_default" example:"false"`
}

// UpdateTemplateRequest represents template update request
type UpdateTemplateRequest struct {
	Name        *string                `json:"name,omitempty" example:"Updated Invoice Template"`
	Description *string                `json:"description,omitempty" example:"Updated description"`
	Content     *string                `json:"content,omitempty" example:"<!DOCTYPE html>..."`
	Variables   []string               `json:"variables,omitempty"`
	SampleData  map[string]interface{} `json:"sample_data,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty" example:"true"`
	IsDefault   *bool                  `json:"is_default,omitempty" example:"false"`
}

// TemplateResponse represents template response
type TemplateResponse struct {
	ID             uint                   `json:"id" example:"1"`
	TenantID       uint                   `json:"tenant_id" example:"1"`
	OrganizationID *uint                  `json:"organization_id,omitempty" example:"10"`
	TemplateType   string                 `json:"template_type" example:"invoice"`
	Name           string                 `json:"name" example:"Standard Invoice"`
	Description    string                 `json:"description" example:"Default invoice template"`
	Version        int                    `json:"version" example:"1"`
	IsActive       bool                   `json:"is_active" example:"true"`
	IsDefault      bool                   `json:"is_default" example:"false"`
	StorageKey     string                 `json:"storage_key" example:"tenants/1/templates/invoice/..."`
	Variables      []string               `json:"variables"`
	SampleData     map[string]interface{} `json:"sample_data,omitempty"`
	CreatedAt      string                 `json:"created_at" example:"2025-12-26T10:00:00Z"`
	UpdatedAt      string                 `json:"updated_at" example:"2025-12-26T10:00:00Z"`
}

// TemplateWithContentResponse represents template with HTML content
type TemplateWithContentResponse struct {
	Template TemplateResponse `json:"template"`
	Content  string           `json:"content"`
}

// ListTemplatesResponse represents templates list response
type ListTemplatesResponse struct {
	Success    bool               `json:"success" example:"true"`
	Data       []TemplateResponse `json:"data"`
	Total      int                `json:"total" example:"10"`
	Page       int                `json:"page" example:"1"`
	PageSize   int                `json:"page_size" example:"20"`
	TotalPages int                `json:"total_pages" example:"1"`
}

// RenderTemplateRequest represents template rendering request
type RenderTemplateRequest struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

// DuplicateTemplateRequest represents template duplication request
type DuplicateTemplateRequest struct {
	Name string `json:"name" binding:"required" example:"Copy of Standard Invoice"`
}

// GeneratePDFRequest represents PDF generation from HTML request
type GeneratePDFRequest struct {
	HTML           string                 `json:"html" binding:"required" example:"<!DOCTYPE html>..."`
	Filename       string                 `json:"filename" example:"document.pdf"`
	DocumentType   string                 `json:"document_type" example:"report"`
	SaveDocument   bool                   `json:"save_document" example:"true"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	OrganizationID *uint                  `json:"organization_id,omitempty" example:"10"`
}

// GeneratePDFFromTemplateRequest represents PDF generation from template request
type GeneratePDFFromTemplateRequest struct {
	TemplateID     uint                   `json:"template_id" binding:"required" example:"1"`
	OrganizationID *uint                  `json:"organization_id,omitempty" example:"10"`
	Data           map[string]interface{} `json:"data" binding:"required"`
	Filename       string                 `json:"filename" example:"invoice.pdf"`
	DocumentType   string                 `json:"document_type" example:"invoice"`
	SaveDocument   bool                   `json:"save_document" example:"true"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// InvoicePDFRequest represents invoice PDF generation request
type InvoicePDFRequest struct {
	InvoiceNumber   string                 `json:"invoice_number" binding:"required" example:"INV-2025-001"`
	InvoiceDate     string                 `json:"invoice_date" binding:"required" example:"2025-12-26"`
	CustomerName    string                 `json:"customer_name" binding:"required" example:"Acme Corp"`
	CustomerAddress string                 `json:"customer_address" example:"123 Main St"`
	Items           []InvoiceItemRequest   `json:"items" binding:"required"`
	Subtotal        float64                `json:"subtotal" example:"1000.00"`
	TaxRate         float64                `json:"tax_rate" example:"0.10"`
	TaxAmount       float64                `json:"tax_amount" example:"100.00"`
	Total           float64                `json:"total" example:"1100.00"`
	PaymentTerms    string                 `json:"payment_terms" example:"Net 30"`
	Notes           string                 `json:"notes" example:"Thank you for your business"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// InvoiceItemRequest represents an invoice line item
type InvoiceItemRequest struct {
	Description string  `json:"description" example:"Consulting Services"`
	Quantity    int     `json:"quantity" example:"10"`
	UnitPrice   float64 `json:"unit_price" example:"100.00"`
	Amount      float64 `json:"amount" example:"1000.00"`
}
