package core

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module represents a pluggable module in the system
type Module interface {
	Name() string
	Version() string
	Dependencies() []string

	// Lifecycle methods
	Initialize(ctx ModuleContext) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Component providers
	Entities() []Entity
	Routes() []RouteProvider
	EventHandlers() []EventHandler
	Middleware() []MiddlewareProvider
	Services() []ServiceProvider
	SwaggerPaths() []string
}

// ModuleContext provides dependencies to modules
type ModuleContext struct {
	DB             *gorm.DB
	Router         *gin.Engine
	EventBus       EventBus
	Config         interface{}
	Logger         Logger
	Services       ServiceRegistry
	Auth           AuthService
	TokenService   TokenService
	ModuleRegistry ModuleRegistry
}

// Entity represents a database entity with migrations
type Entity interface {
	TableName() string
	GetModel() interface{}
	GetMigrations() []Migration
}

// Migration represents a database migration
type Migration interface {
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
	Version() string
}

// RouteProvider handles route registration
type RouteProvider interface {
	RegisterRoutes(router *gin.RouterGroup, ctx ModuleContext)
	GetPrefix() string
	GetMiddleware() []gin.HandlerFunc
	GetSwaggerTags() []string
}

// EventHandler manages event processing
type EventHandler interface {
	EventType() string
	Handle(event interface{}) error
	Priority() int
}

// ServiceProvider exposes module services
type ServiceProvider interface {
	ServiceName() string
	ServiceInterface() interface{}
	Factory(ctx ModuleContext) (interface{}, error)
}

// MiddlewareProvider provides middleware functions
type MiddlewareProvider interface {
	Name() string
	Handler() gin.HandlerFunc
	Priority() int
	ApplyTo() []string // Route patterns to apply to, empty means global
}

// EventBus defines the event system interface
type EventBus interface {
	Publish(eventType string, event interface{}) error
	Subscribe(eventType string, handler func(interface{}) error, priority int) error
	Unsubscribe(eventType string, handler func(interface{}) error) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Logger defines the logging interface
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	With(key string, value interface{}) Logger
}

// AuthService defines authentication service interface
type AuthService interface {
	ValidateToken(token string) (interface{}, error)
	GenerateToken(user interface{}) (string, error)
	GetCurrentUser(c *gin.Context) (interface{}, error)
	RequireAuth() gin.HandlerFunc
	RequireRole(roles ...string) gin.HandlerFunc
}

// TokenService defines generic token generation and validation service
// Modules can use this to generate tokens with custom payloads
type TokenService interface {
	// GenerateToken generates a JWT token with custom claims implementing jwt.Claims
	GenerateToken(claims interface{}) (string, error)
	// ValidateToken validates a JWT token and populates the provided claims structure
	ValidateToken(tokenString string, claims interface{}) error
	// ParseTokenID extracts the token ID without full validation
	ParseTokenID(tokenString string) (string, error)
	// GetTokenExpiration extracts expiration time without full validation
	GetTokenExpiration(tokenString string) (time.Time, error)
}

// ServiceRegistry manages service discovery
type ServiceRegistry interface {
	Register(name string, service interface{}) error
	Get(name string) (interface{}, bool)
	GetTyped(name string, target interface{}) error
	List() []string
}

// ModuleMetadata provides information about a module
type ModuleMetadata struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Entities     []string `json:"entities"`
	Routes       []string `json:"routes"`
	Services     []string `json:"services"`
	Status       string   `json:"status"` // "initialized", "started", "stopped", "error"
}

// ModuleRegistry manages all modules
type ModuleRegistry interface {
	Register(module Module) error
	Get(name string) (Module, bool)
	GetAll() []Module
	GetMetadata() []ModuleMetadata
	InitializeAll(ctx ModuleContext) error
	StartAll(ctx context.Context) error
	StopAll(ctx context.Context) error
	GetInitializationOrder() ([]string, error)
}

// SeedData represents data to be seeded into the database
type SeedData interface {
	GetTableName() string
	GetData() []interface{}
	GetDependencies() []string // Tables that must be seeded first
}
