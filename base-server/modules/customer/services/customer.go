package services

import (
	"github.com/ae-base-server/pkg/core"
	"gorm.io/gorm"
)

// CustomerService provides customer management services
type CustomerService struct {
	db     *gorm.DB
	logger core.Logger
}

// NewCustomerService creates a new customer service
func NewCustomerService(db *gorm.DB, logger core.Logger) *CustomerService {
	return &CustomerService{
		db:     db,
		logger: logger,
	}
}

// CustomerServiceProvider implements core.ServiceProvider for CustomerService
type CustomerServiceProvider struct {
	service *CustomerService
}

// NewCustomerServiceProvider creates a new customer service provider
func NewCustomerServiceProvider(service *CustomerService) core.ServiceProvider {
	return &CustomerServiceProvider{
		service: service,
	}
}

func (p *CustomerServiceProvider) ServiceName() string {
	return "customer"
}

func (p *CustomerServiceProvider) ServiceInterface() interface{} {
	return (*CustomerService)(nil)
}

func (p *CustomerServiceProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
	return p.service, nil
}
