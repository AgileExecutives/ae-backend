package handlers

// GenerateInvoiceNumberRequest represents invoice number generation request
type GenerateInvoiceNumberRequest struct {
	OrganizationID uint   `json:"organization_id,omitempty" example:"10"` // Optional - will use authenticated user's organization if not provided
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
