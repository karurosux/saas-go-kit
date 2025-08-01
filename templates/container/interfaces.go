package container

import (
	"context"
	"errors"
)

// ServiceKey is a unique identifier for a service
type ServiceKey string

// Service represents a service that can be registered in the container
type Service interface {
	// Name returns the service name for logging/debugging
	Name() string
	// Initialize is called after all services are registered
	Initialize(ctx context.Context, container Container) error
	// Shutdown is called when the application is shutting down
	Shutdown(ctx context.Context) error
}

// Container manages all application services
type Container interface {
	// Register adds a service to the container
	Register(key ServiceKey, service interface{}) error
	
	// Get retrieves a service by key
	Get(key ServiceKey) (interface{}, error)
	
	// MustGet retrieves a service by key or panics
	MustGet(key ServiceKey) interface{}
	
	// Has checks if a service is registered
	Has(key ServiceKey) bool
	
	// Initialize calls Initialize on all services that implement Service interface
	InitializeAll(ctx context.Context) error
	
	// Shutdown calls Shutdown on all services in reverse order
	ShutdownAll(ctx context.Context) error
}

// Errors
var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrServiceAlreadyExists = errors.New("service already exists")
	ErrInvalidServiceType   = errors.New("invalid service type")
)