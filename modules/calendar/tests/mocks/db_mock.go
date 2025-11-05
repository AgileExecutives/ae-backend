package mocks

import (
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormDB is an interface that defines the GORM methods we use
// This allows us to mock the database in tests
type GormDB interface {
	Create(value interface{}) *gorm.DB
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Delete(value interface{}, conds ...interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(query string, args ...interface{}) *gorm.DB
	Model(value interface{}) *gorm.DB
	Count(count *int64) *gorm.DB
	Offset(offset int) *gorm.DB
	Limit(limit int) *gorm.DB
	Updates(values interface{}) *gorm.DB
	Clauses(conds ...clause.Expression) *gorm.DB
}

// MockDB is a mock implementation of GormDB for testing
type MockDB struct {
	mock.Mock
	err error
}

// NewMockDB creates a new MockDB instance
func NewMockDB() *MockDB {
	return &MockDB{}
}

// SetError sets the error to be returned by subsequent calls
func (m *MockDB) SetError(err error) {
	m.err = err
}

// Create mocks the GORM Create method
func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: args.Error(0)}
}

// First mocks the GORM First method
func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(append([]interface{}{dest}, conds...)...)
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: args.Error(0)}
}

// Find mocks the GORM Find method
func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(append([]interface{}{dest}, conds...)...)
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: args.Error(0)}
}

// Save mocks the GORM Save method
func (m *MockDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: args.Error(0)}
}

// Delete mocks the GORM Delete method
func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(append([]interface{}{value}, conds...)...)
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: args.Error(0)}
}

// Where mocks the GORM Where method
func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	m.Called(append([]interface{}{query}, args...)...)
	return &gorm.DB{Error: m.err}
}

// Preload mocks the GORM Preload method
func (m *MockDB) Preload(query string, args ...interface{}) *gorm.DB {
	m.Called(append([]interface{}{query}, args...)...)
	return &gorm.DB{Error: m.err}
}

// Model mocks the GORM Model method
func (m *MockDB) Model(value interface{}) *gorm.DB {
	m.Called(value)
	return &gorm.DB{Error: m.err}
}

// Count mocks the GORM Count method
func (m *MockDB) Count(count *int64) *gorm.DB {
	args := m.Called(count)
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: args.Error(0)}
}

// Offset mocks the GORM Offset method
func (m *MockDB) Offset(offset int) *gorm.DB {
	m.Called(offset)
	return &gorm.DB{Error: m.err}
}

// Limit mocks the GORM Limit method
func (m *MockDB) Limit(limit int) *gorm.DB {
	m.Called(limit)
	return &gorm.DB{Error: m.err}
}

// Updates mocks the GORM Updates method
func (m *MockDB) Updates(values interface{}) *gorm.DB {
	args := m.Called(values)
	if m.err != nil {
		return &gorm.DB{Error: m.err}
	}
	return &gorm.DB{Error: args.Error(0)}
}

// Clauses mocks the GORM Clauses method
func (m *MockDB) Clauses(conds ...clause.Expression) *gorm.DB {
	m.Called(conds)
	return &gorm.DB{Error: m.err}
}
