package testutils

import (
	"context"

	"github.com/ae-base-server/pkg/core"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of core.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Info(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Fatal(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) With(key string, value interface{}) core.Logger {
	return m
}

// NewMockLogger creates a new mock logger that accepts any calls
func NewMockLogger() *MockLogger {
	logger := new(MockLogger)
	logger.On("Info", mock.Anything).Maybe()
	logger.On("Debug", mock.Anything).Maybe()
	logger.On("Warn", mock.Anything).Maybe()
	logger.On("Error", mock.Anything).Maybe()
	return logger
}

// MockEmailService is a mock implementation of email service
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	args := m.Called(ctx, to, subject, body)
	return args.Error(0)
}

func (m *MockEmailService) SendEmailWithTemplate(ctx context.Context, to, subject, templateName string, data interface{}) error {
	args := m.Called(ctx, to, subject, templateName, data)
	return args.Error(0)
}

// MockPDFGenerator is a mock implementation of PDF generator
type MockPDFGenerator struct {
	mock.Mock
}

func NewMockPDFGenerator() *MockPDFGenerator {
	return &MockPDFGenerator{}
}

func (m *MockPDFGenerator) GeneratePDF(ctx context.Context, htmlContent string) ([]byte, error) {
	args := m.Called(ctx, htmlContent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPDFGenerator) ConvertHtmlStringToPDF(ctx context.Context, htmlContent string) ([]byte, error) {
	args := m.Called(ctx, htmlContent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPDFGenerator) GenerateInvoicePDF(ctx context.Context, data map[string]interface{}) ([]byte, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPDFGenerator) RenderTemplate(ctx context.Context, template string, data map[string]interface{}) (string, error) {
	args := m.Called(ctx, template, data)
	return args.String(0), args.Error(1)
}

func (m *MockPDFGenerator) GenerateXRechnung(ctx context.Context, data map[string]interface{}) ([]byte, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// MockStorageService is a mock implementation of document storage
type MockStorageService struct {
	mock.Mock
}

func NewMockStorageService() *MockStorageService {
	return &MockStorageService{}
}

func (m *MockStorageService) Upload(ctx context.Context, fileName string, data []byte, tenantID uint) (string, error) {
	args := m.Called(ctx, fileName, data, tenantID)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) Download(ctx context.Context, path string, tenantID uint) ([]byte, error) {
	args := m.Called(ctx, path, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStorageService) Delete(ctx context.Context, path string, tenantID uint) error {
	args := m.Called(ctx, path, tenantID)
	return args.Error(0)
}

func (m *MockStorageService) Search(ctx context.Context, tenantID uint, docType string) ([]string, error) {
	args := m.Called(ctx, tenantID, docType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorageService) SearchByDateRange(ctx context.Context, tenantID uint, startDate, endDate interface{}) ([]string, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorageService) CreateBackup(ctx context.Context, tenantID uint) (string, error) {
	args := m.Called(ctx, tenantID)
	return args.String(0), args.Error(1)
}

// MockAuditService is a mock implementation of audit service
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogAction(ctx context.Context, userID uint, entityType, entityID, action string, details interface{}) error {
	args := m.Called(ctx, userID, entityType, entityID, action, details)
	return args.Error(0)
}

func (m *MockAuditService) GetAuditLogs(ctx context.Context, filters map[string]interface{}) ([]interface{}, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

// MockEventBus is a mock implementation of event bus
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, eventName string, data interface{}) error {
	args := m.Called(ctx, eventName, data)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(eventName string, handler func(interface{}) error) {
	m.Called(eventName, handler)
}

// MockCache is a mock implementation of cache service
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// MockInvoiceNumberService is a mock for invoice number generation
type MockInvoiceNumberService struct {
	mock.Mock
}

func (m *MockInvoiceNumberService) GenerateInvoiceNumber(ctx context.Context, tenantID, orgID uint, year int) (string, error) {
	args := m.Called(ctx, tenantID, orgID, year)
	return args.String(0), args.Error(1)
}

func (m *MockInvoiceNumberService) ReserveInvoiceNumber(ctx context.Context, tenantID, orgID uint) (string, error) {
	args := m.Called(ctx, tenantID, orgID)
	return args.String(0), args.Error(1)
}

func (m *MockInvoiceNumberService) ReleaseInvoiceNumber(ctx context.Context, number string) error {
	args := m.Called(ctx, number)
	return args.Error(0)
}
