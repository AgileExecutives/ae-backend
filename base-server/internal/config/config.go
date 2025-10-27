package config

import (
	"os"
	"strconv"

	"github.com/ae-saas-basic/ae-saas-basic/internal/database"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Database  database.Config
	JWT       JWTConfig
	Email     EmailConfig
	PDF       PDFConfig
	RateLimit RateLimitConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
	Mode string // gin mode: debug, release, test
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	ExpiryHour int
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost                 string
	SMTPPort                 int
	SMTPUser                 string
	SMTPPassword             string
	FromEmail                string
	FromName                 string
	MockEmail                bool
	RequireEmailVerification bool
}

// PDFConfig holds PDF generation configuration
type PDFConfig struct {
	TemplateDir  string
	OutputDir    string
	PageSize     string
	Orientation  string
	MarginTop    string
	MarginRight  string
	MarginBottom string
	MarginLeft   string
	Quality      int
	EnableJS     bool
	LoadTimeout  int
	MaxFileSize  int64 // Maximum PDF file size in bytes
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled bool
}

// Load loads configuration from environment variables with defaults
func Load() Config {
	return Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: database.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "ae_saas_basic"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			ExpiryHour: getEnvAsInt("JWT_EXPIRY_HOUR", 24),
		},
		Email: EmailConfig{
			SMTPHost:                 getEnv("SMTP_HOST", "localhost"),
			SMTPPort:                 getEnvAsInt("SMTP_PORT", 587),
			SMTPUser:                 getEnv("SMTP_USER", ""),
			SMTPPassword:             getEnv("SMTP_PASSWORD", ""),
			FromEmail:                getEnv("FROM_EMAIL", "noreply@ae-saas-basic.com"),
			FromName:                 getEnv("FROM_NAME", "AE SaaS Basic"),
			MockEmail:                getEnvAsBool("MOCK_EMAIL", false),
			RequireEmailVerification: getEnvAsBool("FEATURE_EMAIL_VERIFICATION", true),
		},
		PDF: PDFConfig{
			TemplateDir:  getEnv("PDF_TEMPLATE_DIR", "./statics/templates/pdf"),
			OutputDir:    getEnv("PDF_OUTPUT_DIR", "./output/pdf"),
			PageSize:     getEnv("PDF_PAGE_SIZE", "A4"),
			Orientation:  getEnv("PDF_ORIENTATION", "Portrait"),
			MarginTop:    getEnv("PDF_MARGIN_TOP", "1cm"),
			MarginRight:  getEnv("PDF_MARGIN_RIGHT", "1cm"),
			MarginBottom: getEnv("PDF_MARGIN_BOTTOM", "1cm"),
			MarginLeft:   getEnv("PDF_MARGIN_LEFT", "1cm"),
			Quality:      getEnvAsInt("PDF_QUALITY", 80),
			EnableJS:     getEnvAsBool("PDF_ENABLE_JS", true),
			LoadTimeout:  getEnvAsInt("PDF_LOAD_TIMEOUT", 30),
			MaxFileSize:  getEnvAsInt64("PDF_MAX_FILE_SIZE", 50*1024*1024), // 50MB default
		},
		RateLimit: RateLimitConfig{
			Enabled: getEnvAsBool("RATE_LIMIT_ENABLED", true),
		},
	}
}

// getEnv gets environment variable with fallback to default value
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt gets environment variable as integer with fallback to default value
func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// getEnvAsBool gets environment variable as boolean with fallback to default value
func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// getEnvAsInt64 gets environment variable as int64 with fallback to default value
func getEnvAsInt64(key string, defaultVal int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultVal
}

// getEnvAsFloat64 gets environment variable as float64 with fallback to default value
func getEnvAsFloat64(key string, defaultVal float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultVal
}
