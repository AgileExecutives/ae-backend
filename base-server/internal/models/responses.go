package models

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Database  string                 `json:"database"`
	Settings  HealthSettingsResponse `json:"settings"`
}

// HealthSettingsResponse represents configuration settings in health response
type HealthSettingsResponse struct {
	MockEmail        bool `json:"mock_email"`
	RateLimitEnabled bool `json:"rate_limit_enabled"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse creates a success API response
func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponseFunc creates an error API response
func ErrorResponseFunc(message string, err string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Error:   err,
	}
}
