package models

import (
	"fmt"
	"time"
)

// PDFTemplate represents a PDF template configuration
type PDFTemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"unique;not null"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	FilePath    string    `json:"file_path" gorm:"not null"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Version     string    `json:"version" gorm:"default:'1.0'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Template metadata
	RequiredFields []PDFTemplateField `json:"required_fields" gorm:"foreignKey:TemplateID"`
	SampleData     string             `json:"sample_data,omitempty" gorm:"type:text"` // JSON string
}

// PDFTemplateField represents required fields for a template
type PDFTemplateField struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	TemplateID   uint   `json:"template_id"`
	FieldName    string `json:"field_name" gorm:"not null"`
	FieldType    string `json:"field_type" gorm:"not null"` // string, number, date, boolean, array, object
	IsRequired   bool   `json:"is_required" gorm:"default:false"`
	DefaultValue string `json:"default_value,omitempty"`
	Description  string `json:"description,omitempty"`
	Validation   string `json:"validation,omitempty"` // JSON validation rules
}

// PDFGenerationLog represents a log entry for PDF generation
type PDFGenerationLog struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	TemplateID     uint      `json:"template_id,omitempty"`
	TemplateName   string    `json:"template_name"`
	UserID         uint      `json:"user_id,omitempty"`
	Status         string    `json:"status"` // pending, success, error
	FilePath       string    `json:"file_path,omitempty"`
	FileSize       int64     `json:"file_size,omitempty"`
	GenerationTime int64     `json:"generation_time"` // milliseconds
	Error          string    `json:"error,omitempty"`
	InputData      string    `json:"input_data,omitempty" gorm:"type:text"` // JSON string
	Config         string    `json:"config,omitempty" gorm:"type:text"`     // JSON string
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// PDFTemplateCategory represents template categories
type PDFTemplateCategory struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"unique;not null"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	SortOrder   int    `json:"sort_order" gorm:"default:0"`
}

// Common template data structures for different document types

// InvoiceData represents data structure for invoice templates
type InvoiceData struct {
	// Header information
	InvoiceNumber string    `json:"invoice_number" validate:"required"`
	InvoiceDate   time.Time `json:"invoice_date" validate:"required"`
	DueDate       time.Time `json:"due_date"`

	// Company information
	Company CompanyInfo `json:"company" validate:"required"`

	// Customer information
	Customer CustomerInfo `json:"customer" validate:"required"`

	// Invoice items
	Items []InvoiceItem `json:"items" validate:"required,min=1"`

	// Totals
	Subtotal  float64 `json:"subtotal"`
	TaxRate   float64 `json:"tax_rate"`
	TaxAmount float64 `json:"tax_amount"`
	Discount  float64 `json:"discount"`
	Total     float64 `json:"total"`
	Currency  string  `json:"currency" validate:"required"`

	// Additional information
	Notes       string `json:"notes,omitempty"`
	Terms       string `json:"terms,omitempty"`
	PaymentInfo string `json:"payment_info,omitempty"`
}

// CompanyInfo represents company information
type CompanyInfo struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`
	TaxID   string `json:"tax_id"`
	Logo    string `json:"logo,omitempty"` // Logo URL or path
}

// CustomerInfo represents customer information
type CustomerInfo struct {
	ID      uint   `json:"id,omitempty"`
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

// InvoiceItem represents an invoice line item
type InvoiceItem struct {
	Description string  `json:"description" validate:"required"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gte=0"`
	Total       float64 `json:"total"`
	Category    string  `json:"category,omitempty"`
}

// ReportData represents data structure for report templates
type ReportData struct {
	// Header information
	Title      string    `json:"title" validate:"required"`
	SubTitle   string    `json:"sub_title,omitempty"`
	ReportDate time.Time `json:"report_date" validate:"required"`
	Period     string    `json:"period"`

	// Company information
	Company CompanyInfo `json:"company" validate:"required"`

	// Report sections
	Summary ReportSummary `json:"summary,omitempty"`
	Charts  []ChartData   `json:"charts,omitempty"`
	Tables  []TableData   `json:"tables,omitempty"`
	Metrics []MetricData  `json:"metrics,omitempty"`

	// Additional information
	Notes  string `json:"notes,omitempty"`
	Footer string `json:"footer,omitempty"`
}

// ReportSummary represents summary section of a report
type ReportSummary struct {
	Text       string      `json:"text"`
	KeyMetrics []KeyMetric `json:"key_metrics,omitempty"`
	Highlights []string    `json:"highlights,omitempty"`
}

// KeyMetric represents a key metric
type KeyMetric struct {
	Label      string `json:"label" validate:"required"`
	Value      string `json:"value" validate:"required"`
	Change     string `json:"change,omitempty"`
	ChangeType string `json:"change_type,omitempty"` // positive, negative, neutral
}

// ChartData represents chart data for reports
type ChartData struct {
	Title   string      `json:"title" validate:"required"`
	Type    string      `json:"type" validate:"required"` // bar, line, pie, etc.
	Data    interface{} `json:"data" validate:"required"`
	Options interface{} `json:"options,omitempty"`
}

// TableData represents table data for reports
type TableData struct {
	Title   string     `json:"title" validate:"required"`
	Headers []string   `json:"headers" validate:"required"`
	Rows    [][]string `json:"rows" validate:"required"`
	Footer  []string   `json:"footer,omitempty"`
}

// MetricData represents metric data
type MetricData struct {
	Name        string  `json:"name" validate:"required"`
	Value       float64 `json:"value" validate:"required"`
	Unit        string  `json:"unit,omitempty"`
	Format      string  `json:"format,omitempty"` // currency, percentage, number
	Description string  `json:"description,omitempty"`
}

// CertificateData represents data structure for certificate templates
type CertificateData struct {
	// Certificate information
	CertificateNumber string    `json:"certificate_number" validate:"required"`
	IssueDate         time.Time `json:"issue_date" validate:"required"`
	ExpiryDate        time.Time `json:"expiry_date,omitempty"`

	// Recipient information
	RecipientName  string `json:"recipient_name" validate:"required"`
	RecipientTitle string `json:"recipient_title,omitempty"`

	// Certificate details
	Title        string   `json:"title" validate:"required"`
	Description  string   `json:"description"`
	CourseInfo   string   `json:"course_info,omitempty"`
	Achievements []string `json:"achievements,omitempty"`

	// Issuer information
	IssuerName      string `json:"issuer_name" validate:"required"`
	IssuerTitle     string `json:"issuer_title,omitempty"`
	IssuerSignature string `json:"issuer_signature,omitempty"`
	IssuerLogo      string `json:"issuer_logo,omitempty"`

	// Additional information
	CredentialID    string `json:"credential_id,omitempty"`
	VerificationURL string `json:"verification_url,omitempty"`
}

// LetterData represents data structure for letter templates
type LetterData struct {
	// Header information
	Date      time.Time `json:"date" validate:"required"`
	Reference string    `json:"reference,omitempty"`

	// Sender information
	Sender ContactInfo `json:"sender" validate:"required"`

	// Recipient information
	Recipient ContactInfo `json:"recipient" validate:"required"`

	// Letter content
	Subject    string `json:"subject" validate:"required"`
	Salutation string `json:"salutation"`
	Body       string `json:"body" validate:"required"`
	Closing    string `json:"closing"`

	// Additional information
	Attachments []string `json:"attachments,omitempty"`
	CopyTo      []string `json:"copy_to,omitempty"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	Name    string `json:"name" validate:"required"`
	Title   string `json:"title,omitempty"`
	Company string `json:"company,omitempty"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
}

// PDF Template Validation Interface

// TemplateValidator defines interface for template validation
type TemplateValidator interface {
	Validate(data interface{}) error
	GetRequiredFields() []string
	GetSampleData() interface{}
}

// Common validation functions

// ValidateInvoiceData validates invoice template data
func ValidateInvoiceData(data InvoiceData) []string {
	var errors []string

	if data.InvoiceNumber == "" {
		errors = append(errors, "invoice_number is required")
	}

	if data.Company.Name == "" {
		errors = append(errors, "company name is required")
	}

	if data.Customer.Name == "" {
		errors = append(errors, "customer name is required")
	}

	if len(data.Items) == 0 {
		errors = append(errors, "at least one invoice item is required")
	}

	for i, item := range data.Items {
		if item.Description == "" {
			errors = append(errors, fmt.Sprintf("item %d: description is required", i+1))
		}
		if item.Quantity <= 0 {
			errors = append(errors, fmt.Sprintf("item %d: quantity must be greater than 0", i+1))
		}
		if item.UnitPrice < 0 {
			errors = append(errors, fmt.Sprintf("item %d: unit price cannot be negative", i+1))
		}
	}

	return errors
}

// ValidateReportData validates report template data
func ValidateReportData(data ReportData) []string {
	var errors []string

	if data.Title == "" {
		errors = append(errors, "title is required")
	}

	if data.Company.Name == "" {
		errors = append(errors, "company name is required")
	}

	return errors
}

// ValidateCertificateData validates certificate template data
func ValidateCertificateData(data CertificateData) []string {
	var errors []string

	if data.CertificateNumber == "" {
		errors = append(errors, "certificate_number is required")
	}

	if data.RecipientName == "" {
		errors = append(errors, "recipient_name is required")
	}

	if data.Title == "" {
		errors = append(errors, "title is required")
	}

	if data.IssuerName == "" {
		errors = append(errors, "issuer_name is required")
	}

	return errors
}
