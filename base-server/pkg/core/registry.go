package core

import (
	"context"
	"fmt"
	"sync"
)

// moduleRegistry implements ModuleRegistry interface
type moduleRegistry struct {
	modules     map[string]Module
	initialized map[string]bool
	started     map[string]bool
	services    ServiceRegistry
	context     ModuleContext
	mutex       sync.RWMutex
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry() ModuleRegistry {
	return &moduleRegistry{
		modules:     make(map[string]Module),
		initialized: make(map[string]bool),
		started:     make(map[string]bool),
		services:    NewServiceRegistry(),
	}
}

// Register registers a new module
func (r *moduleRegistry) Register(module Module) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := module.Name()
	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("module %s already registered", name)
	}

	// Validate dependencies exist (will be checked again during initialization)
	for _, dep := range module.Dependencies() {
		if _, exists := r.modules[dep]; !exists {
			return fmt.Errorf("module %s depends on %s which is not registered", name, dep)
		}
	}

	r.modules[name] = module
	return nil
}

// Get retrieves a module by name
func (r *moduleRegistry) Get(name string) (Module, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	module, exists := r.modules[name]
	return module, exists
}

// GetAll returns all registered modules
func (r *moduleRegistry) GetAll() []Module {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	modules := make([]Module, 0, len(r.modules))
	for _, module := range r.modules {
		modules = append(modules, module)
	}
	return modules
}

// GetMetadata returns metadata for all modules
func (r *moduleRegistry) GetMetadata() []ModuleMetadata {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	metadata := make([]ModuleMetadata, 0, len(r.modules))
	for name, module := range r.modules {
		status := "registered"
		if r.started[name] {
			status = "started"
		} else if r.initialized[name] {
			status = "initialized"
		}

		entities := make([]string, len(module.Entities()))
		for i, entity := range module.Entities() {
			entities[i] = entity.TableName()
		}

		routes := make([]string, len(module.Routes()))
		for i, route := range module.Routes() {
			routes[i] = route.GetPrefix()
		}

		services := make([]string, len(module.Services()))
		for i, service := range module.Services() {
			services[i] = service.ServiceName()
		}

		metadata = append(metadata, ModuleMetadata{
			Name:         name,
			Version:      module.Version(),
			Dependencies: module.Dependencies(),
			Entities:     entities,
			Routes:       routes,
			Services:     services,
			Status:       status,
		})
	}
	return metadata
}

// GetInitializationOrder returns the order modules should be initialized
func (r *moduleRegistry) GetInitializationOrder() ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.topologicalSort()
}

// InitializeAll initializes all modules in dependency order
func (r *moduleRegistry) InitializeAll(ctx ModuleContext) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.context = ctx
	r.context.Services = r.services

	order, err := r.topologicalSort()
	if err != nil {
		return fmt.Errorf("failed to resolve module dependencies: %w", err)
	}

	for _, moduleName := range order {
		if err := r.initializeModule(moduleName); err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", moduleName, err)
		}
	}

	return nil
}

// StartAll starts all initialized modules
func (r *moduleRegistry) StartAll(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	order, err := r.topologicalSort()
	if err != nil {
		return fmt.Errorf("failed to resolve module dependencies: %w", err)
	}

	for _, name := range order {
		if !r.initialized[name] {
			continue
		}

		module := r.modules[name]
		if err := module.Start(ctx); err != nil {
			return fmt.Errorf("failed to start module %s: %w", name, err)
		}
		r.started[name] = true
	}

	return nil
}

// StopAll stops all started modules in reverse order
func (r *moduleRegistry) StopAll(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	order, err := r.topologicalSort()
	if err != nil {
		return fmt.Errorf("failed to resolve module dependencies: %w", err)
	}

	// Stop in reverse order
	for i := len(order) - 1; i >= 0; i-- {
		name := order[i]
		if !r.started[name] {
			continue
		}

		module := r.modules[name]
		if err := module.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop module %s: %w", name, err)
		}
		r.started[name] = false
	}

	return nil
}

// initializeModule initializes a single module and its dependencies
func (r *moduleRegistry) initializeModule(name string) error {
	if r.initialized[name] {
		return nil
	}

	module := r.modules[name]

	// Initialize dependencies first
	for _, depName := range module.Dependencies() {
		if !r.initialized[depName] {
			if err := r.initializeModule(depName); err != nil {
				return fmt.Errorf("failed to initialize dependency %s: %w", depName, err)
			}
		}
	}

	// Register module services first
	for _, serviceProvider := range module.Services() {
		service, err := serviceProvider.Factory(r.context)
		if err != nil {
			return fmt.Errorf("failed to create service %s: %w", serviceProvider.ServiceName(), err)
		}
		if err := r.services.Register(serviceProvider.ServiceName(), service); err != nil {
			return fmt.Errorf("failed to register service %s: %w", serviceProvider.ServiceName(), err)
		}
	}

	// Initialize the module
	if err := module.Initialize(r.context); err != nil {
		return fmt.Errorf("failed to initialize module %s: %w", name, err)
	}

	// Register routes with proper prefixing
	apiV1 := r.context.Router.Group("/api/v1")
	for _, routeProvider := range module.Routes() {
		routeGroup := apiV1.Group(routeProvider.GetPrefix())

		// Apply middleware
		for _, middleware := range routeProvider.GetMiddleware() {
			routeGroup.Use(middleware)
		}

		routeProvider.RegisterRoutes(routeGroup, r.context)
	}

	// Register event handlers
	for _, handler := range module.EventHandlers() {
		if err := r.context.EventBus.Subscribe(handler.EventType(), handler.Handle, handler.Priority()); err != nil {
			return fmt.Errorf("failed to register event handler for %s: %w", handler.EventType(), err)
		}
	}

	r.initialized[name] = true
	return nil
}

// topologicalSort performs topological sorting of modules based on dependencies
func (r *moduleRegistry) topologicalSort() ([]string, error) {
	// Kahn's algorithm for topological sorting
	inDegree := make(map[string]int)
	graph := make(map[string][]string)

	// Initialize
	for name := range r.modules {
		inDegree[name] = 0
		graph[name] = []string{}
	}

	// Build graph and calculate in-degrees
	for name, module := range r.modules {
		for _, dep := range module.Dependencies() {
			if _, exists := r.modules[dep]; !exists {
				return nil, fmt.Errorf("module %s depends on %s which is not registered", name, dep)
			}
			graph[dep] = append(graph[dep], name)
			inDegree[name]++
		}
	}

	// Find nodes with no incoming edges
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]string, 0)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Remove edges and add new nodes with no incoming edges
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for circular dependencies
	if len(result) != len(r.modules) {
		return nil, fmt.Errorf("circular dependency detected in modules")
	}

	return result, nil
}
