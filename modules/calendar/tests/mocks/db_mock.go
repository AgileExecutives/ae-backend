package mocks

import (
	"errors"
	"reflect"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of GORM DB for testing
type MockDB struct {
	mock.Mock
}

// Create mocks the GORM Create method
func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)

	// Simulate setting ID on created entity
	if args.Bool(1) { // second arg indicates success
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			idField := v.FieldByName("ID")
			if idField.IsValid() && idField.CanSet() && idField.Kind() == reflect.Uint {
				idField.SetUint(1) // Set a mock ID
			}
		}
	}

	return &gorm.DB{Error: args.Error(0)}
}

// First mocks the GORM First method
func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(append([]interface{}{dest}, conds...)...)

	if args.Bool(1) { // second arg indicates if record found
		// Simulate populating the dest with mock data
		m.populateMockData(dest, args.Get(2)) // third arg contains mock data
	}

	return &gorm.DB{Error: args.Error(0)}
}

// Find mocks the GORM Find method
func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(append([]interface{}{dest}, conds...)...)

	if args.Bool(1) { // second arg indicates success
		// Simulate populating the slice with mock data
		m.populateSliceMockData(dest, args.Get(2)) // third arg contains mock data slice
	}

	return &gorm.DB{Error: args.Error(0)}
}

// Save mocks the GORM Save method
func (m *MockDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	return &gorm.DB{Error: args.Error(0)}
}

// Delete mocks the GORM Delete method
func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(append([]interface{}{value}, conds...)...)
	return &gorm.DB{Error: args.Error(0)}
}

// Where mocks the GORM Where method
func (m *MockDB) Where(query interface{}, args ...interface{}) *MockDB {
	m.Called(append([]interface{}{query}, args...)...)
	return m
}

// Preload mocks the GORM Preload method
func (m *MockDB) Preload(query string, args ...interface{}) *MockDB {
	m.Called(append([]interface{}{query}, args...)...)
	return m
}

// Model mocks the GORM Model method
func (m *MockDB) Model(value interface{}) *MockDB {
	m.Called(value)
	return m
}

// Count mocks the GORM Count method
func (m *MockDB) Count(count *int64) *gorm.DB {
	args := m.Called(count)
	if args.Bool(1) { // second arg indicates success
		*count = args.Get(2).(int64) // third arg contains count value
	}
	return &gorm.DB{Error: args.Error(0)}
}

// Offset mocks the GORM Offset method
func (m *MockDB) Offset(offset int) *MockDB {
	m.Called(offset)
	return m
}

// Limit mocks the GORM Limit method
func (m *MockDB) Limit(limit int) *MockDB {
	m.Called(limit)
	return m
}

// Helper methods for populating mock data

func (m *MockDB) populateMockData(dest interface{}, mockData interface{}) {
	if mockData == nil {
		return
	}

	destValue := reflect.ValueOf(dest)
	if destValue.Kind() == reflect.Ptr {
		destValue = destValue.Elem()
	}

	mockValue := reflect.ValueOf(mockData)
	if mockValue.Kind() == reflect.Ptr {
		mockValue = mockValue.Elem()
	}

	if destValue.Type() == mockValue.Type() {
		destValue.Set(mockValue)
	}
}

func (m *MockDB) populateSliceMockData(dest interface{}, mockData interface{}) {
	if mockData == nil {
		return
	}

	destValue := reflect.ValueOf(dest)
	if destValue.Kind() == reflect.Ptr {
		destValue = destValue.Elem()
	}

	if destValue.Kind() != reflect.Slice {
		return
	}

	mockValue := reflect.ValueOf(mockData)
	if mockValue.Kind() == reflect.Slice {
		destValue.Set(mockValue)
	}
}

// Common error constants for testing
var (
	ErrRecordNotFound = gorm.ErrRecordNotFound
	ErrDatabase       = errors.New("database error")
	ErrConstraint     = errors.New("constraint violation")
)
