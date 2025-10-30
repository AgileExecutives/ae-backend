package services

import (
	"github.com/ae-base-server/pkg/core"
	"gorm.io/gorm"
)

// AuthService provides authentication related services
type AuthService struct {
	db     *gorm.DB
	logger core.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(db *gorm.DB, logger core.Logger) *AuthService {
	return &AuthService{
		db:     db,
		logger: logger,
	}
}

// AuthServiceProvider implements core.ServiceProvider for AuthService
type AuthServiceProvider struct {
	service *AuthService
}

// NewAuthServiceProvider creates a new auth service provider
func NewAuthServiceProvider(service *AuthService) core.ServiceProvider {
	return &AuthServiceProvider{
		service: service,
	}
}

func (p *AuthServiceProvider) ServiceName() string {
	return "auth"
}

func (p *AuthServiceProvider) ServiceInterface() interface{} {
	return (*AuthService)(nil)
}

func (p *AuthServiceProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
	return p.service, nil
}
