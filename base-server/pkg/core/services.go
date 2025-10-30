package core

import (
	"fmt"
	"reflect"
	"sync"
)

// serviceRegistry implements ServiceRegistry interface
type serviceRegistry struct {
	services map[string]interface{}
	mutex    sync.RWMutex
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() ServiceRegistry {
	return &serviceRegistry{
		services: make(map[string]interface{}),
	}
}

// Register registers a service by name
func (r *serviceRegistry) Register(name string, service interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	r.services[name] = service
	return nil
}

// Get retrieves a service by name
func (r *serviceRegistry) Get(name string) (interface{}, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	service, exists := r.services[name]
	return service, exists
}

// GetTyped retrieves a service and attempts to cast it to the target type
func (r *serviceRegistry) GetTyped(name string, target interface{}) error {
	service, exists := r.Get(name)
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	// Use reflection to set the target
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	serviceValue := reflect.ValueOf(service)
	targetType := targetValue.Elem().Type()

	if !serviceValue.Type().AssignableTo(targetType) {
		return fmt.Errorf("service %s of type %s is not assignable to %s",
			name, serviceValue.Type(), targetType)
	}

	targetValue.Elem().Set(serviceValue)
	return nil
}

// List returns all registered service names
func (r *serviceRegistry) List() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.services))
	for name := range r.services {
		names = append(names, name)
	}
	return names
}
