package models

// LoginRequest represents login request data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

// UserCreateRequest represents user registration request data
type UserCreateRequest struct {
	Email       string `json:"email" binding:"required,email" example:"user@example.com"`
	Password    string `json:"password" binding:"required,min=8" example:"password123"`
	FirstName   string `json:"first_name" binding:"required" example:"John"`
	LastName    string `json:"last_name" binding:"required" example:"Doe"`
	CompanyName string `json:"company_name" example:"Acme Corp"`
	TenantName  string `json:"tenant_name" binding:"required" example:"acme"`
}

// ForgotPasswordRequest represents forgot password request data
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// ResetPasswordRequest represents reset password request data
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required" example:"reset-token-123"`
	Password string `json:"password" binding:"required,min=8" example:"newpassword123"`
}

// ContactFormRequest represents contact form request data
type ContactFormRequest struct {
	Name    string `json:"name" binding:"required" example:"John Doe"`
	Email   string `json:"email" binding:"required,email" example:"john@example.com"`
	Subject string `json:"subject" binding:"required" example:"Inquiry about services"`
	Message string `json:"message" binding:"required" example:"I would like to know more about your services."`
}

// LoginResponse represents login response data
type LoginResponse struct {
	Token        string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"refresh-token-123"`
	User         UserResponse `json:"user"`
}

// UserResponse represents user response data
type UserResponse struct {
	ID          int            `json:"id" example:"1"`
	Email       string         `json:"email" example:"user@example.com"`
	FirstName   string         `json:"first_name" example:"John"`
	LastName    string         `json:"last_name" example:"Doe"`
	CompanyName string         `json:"company_name" example:"Acme Corp"`
	IsActive    bool           `json:"is_active" example:"true"`
	Tenant      TenantResponse `json:"tenant"`
}

// TenantResponse represents tenant response data
type TenantResponse struct {
	ID          int    `json:"id" example:"1"`
	Name        string `json:"name" example:"acme"`
	DisplayName string `json:"display_name" example:"Acme Corp"`
	IsActive    bool   `json:"is_active" example:"true"`
}

// ContactFormResponse represents contact form response data
type ContactFormResponse struct {
	ID      int    `json:"id" example:"1"`
	Message string `json:"message" example:"Thank you for your message! We will get back to you soon."`
}

// Newsletter represents newsletter subscription data
type Newsletter struct {
	ID        int    `json:"id" example:"1"`
	Email     string `json:"email" example:"user@example.com"`
	IsActive  bool   `json:"is_active" example:"true"`
	CreatedAt string `json:"created_at" example:"2023-01-01T00:00:00Z"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"10"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"10"`
}

// Customer-related models
type CustomerRequest struct {
	Name        string `json:"name" binding:"required" example:"John Doe"`
	Email       string `json:"email" binding:"required,email" example:"john@example.com"`
	Phone       string `json:"phone" example:"+1-555-123-4567"`
	CompanyName string `json:"company_name" example:"Acme Corp"`
}

type CustomerResponse struct {
	ID          int    `json:"id" example:"1"`
	Name        string `json:"name" example:"John Doe"`
	Email       string `json:"email" example:"john@example.com"`
	Phone       string `json:"phone" example:"+1-555-123-4567"`
	CompanyName string `json:"company_name" example:"Acme Corp"`
	CreatedAt   string `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// Plan-related models
type Plan struct {
	ID          int     `json:"id" example:"1"`
	Name        string  `json:"name" example:"Basic Plan"`
	Description string  `json:"description" example:"Basic subscription plan"`
	Price       float64 `json:"price" example:"29.99"`
	Currency    string  `json:"currency" example:"USD"`
	IsActive    bool    `json:"is_active" example:"true"`
	CreatedAt   string  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   string  `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

type PlanRequest struct {
	Name        string  `json:"name" binding:"required" example:"Basic Plan"`
	Description string  `json:"description" example:"Basic subscription plan"`
	Price       float64 `json:"price" binding:"required" example:"29.99"`
	Currency    string  `json:"currency" binding:"required" example:"USD"`
}

// Email-related models
type EmailSendRequest struct {
	To        []string               `json:"to" binding:"required" example:"user@example.com,user2@example.com"`
	Subject   string                 `json:"subject" binding:"required" example:"Welcome to our service"`
	Content   string                 `json:"content" binding:"required" example:"Thank you for signing up!"`
	Template  string                 `json:"template" example:"welcome"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type EmailResponse struct {
	ID           int                    `json:"id" example:"1"`
	To           []string               `json:"to" example:"user@example.com"`
	Subject      string                 `json:"subject" example:"Welcome to our service"`
	Content      string                 `json:"content" example:"Thank you for signing up!"`
	Status       string                 `json:"status" example:"pending"`
	Template     string                 `json:"template,omitempty" example:"welcome"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	SentAt       *string                `json:"sent_at,omitempty" example:"2023-01-01T00:00:00Z"`
	DeliveredAt  *string                `json:"delivered_at,omitempty" example:"2023-01-01T00:00:00Z"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	CreatedAt    string                 `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt    string                 `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}
