package container

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// serviceContainer is the default implementation of Container
type serviceContainer struct {
	mu            sync.RWMutex
	services      map[ServiceKey]interface{}
	initialized   map[ServiceKey]bool
	initOrder     []ServiceKey
	shutdownOrder []ServiceKey
}

// New creates a new service container
func New() Container {
	return &serviceContainer{
		services:      make(map[ServiceKey]interface{}),
		initialized:   make(map[ServiceKey]bool),
		initOrder:     make([]ServiceKey, 0),
		shutdownOrder: make([]ServiceKey, 0),
	}
}

// Register adds a service to the container
func (c *serviceContainer) Register(key ServiceKey, service interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.services[key]; exists {
		return fmt.Errorf("%w: %s", ErrServiceAlreadyExists, key)
	}
	
	c.services[key] = service
	c.initOrder = append(c.initOrder, key)
	// Shutdown order is reverse of init order
	c.shutdownOrder = append([]ServiceKey{key}, c.shutdownOrder...)
	
	return nil
}

// Get retrieves a service by key
func (c *serviceContainer) Get(key ServiceKey) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if service, exists := c.services[key]; exists {
		return service, nil
	}
	
	return nil, fmt.Errorf("%w: %s", ErrServiceNotFound, key)
}

// MustGet retrieves a service by key or panics
func (c *serviceContainer) MustGet(key ServiceKey) interface{} {
	service, err := c.Get(key)
	if err != nil {
		panic(fmt.Sprintf("service %s not found: %v", key, err))
	}
	return service
}

// Has checks if a service is registered
func (c *serviceContainer) Has(key ServiceKey) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	_, exists := c.services[key]
	return exists
}

// InitializeAll calls Initialize on all services that implement Service interface
func (c *serviceContainer) InitializeAll(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for _, key := range c.initOrder {
		if c.initialized[key] {
			continue
		}
		
		if service, exists := c.services[key]; exists {
			if svc, ok := service.(Service); ok {
				if err := svc.Initialize(ctx, c); err != nil {
					return fmt.Errorf("failed to initialize service %s: %w", key, err)
				}
				c.initialized[key] = true
			}
		}
	}
	
	return nil
}

// ShutdownAll calls Shutdown on all services in reverse order
func (c *serviceContainer) ShutdownAll(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	var errs []error
	
	for _, key := range c.shutdownOrder {
		if service, exists := c.services[key]; exists {
			if svc, ok := service.(Service); ok {
				if err := svc.Shutdown(ctx); err != nil {
					errs = append(errs, fmt.Errorf("failed to shutdown service %s: %w", key, err))
				}
			}
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	
	return nil
}

// GetTyped is a generic helper for type-safe service retrieval
func GetTyped[T any](c Container, key ServiceKey) (T, error) {
	var zero T
	
	service, err := c.Get(key)
	if err != nil {
		return zero, err
	}
	
	typed, ok := service.(T)
	if !ok {
		return zero, fmt.Errorf("%w: expected %s, got %s", 
			ErrInvalidServiceType, 
			reflect.TypeOf(zero).String(),
			reflect.TypeOf(service).String())
	}
	
	return typed, nil
}

// MustGetTyped is a generic helper that panics if service is not found
func MustGetTyped[T any](c Container, key ServiceKey) T {
	typed, err := GetTyped[T](c, key)
	if err != nil {
		panic(fmt.Sprintf("failed to get typed service %s: %v", key, err))
	}
	return typed
}