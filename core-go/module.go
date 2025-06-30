package core

import (
	"github.com/labstack/echo/v4"
)

// Module represents a modular component that can be registered with the kit
type Module interface {
	// Name returns the module name
	Name() string
	
	// Routes returns the routes to be registered
	Routes() []Route
	
	// Middleware returns global middleware to be applied
	Middleware() []echo.MiddlewareFunc
	
	// Dependencies returns the names of required modules
	Dependencies() []string
	
	// Init initializes the module with dependencies
	Init(deps map[string]Module) error
}

// Route represents a single route definition
type Route struct {
	Method      string
	Path        string
	Handler     echo.HandlerFunc
	Middlewares []echo.MiddlewareFunc
	Name        string
	Description string
}

// BaseModule provides a basic implementation of Module
type BaseModule struct {
	name         string
	routes       []Route
	middleware   []echo.MiddlewareFunc
	dependencies []string
}

// NewBaseModule creates a new base module
func NewBaseModule(name string) *BaseModule {
	return &BaseModule{
		name:         name,
		routes:       []Route{},
		middleware:   []echo.MiddlewareFunc{},
		dependencies: []string{},
	}
}

// Name returns the module name
func (m *BaseModule) Name() string {
	return m.name
}

// Routes returns the routes
func (m *BaseModule) Routes() []Route {
	return m.routes
}

// Middleware returns the middleware
func (m *BaseModule) Middleware() []echo.MiddlewareFunc {
	return m.middleware
}

// Dependencies returns the dependencies
func (m *BaseModule) Dependencies() []string {
	return m.dependencies
}

// Init initializes the module
func (m *BaseModule) Init(deps map[string]Module) error {
	return nil
}

// AddRoute adds a route to the module
func (m *BaseModule) AddRoute(route Route) {
	m.routes = append(m.routes, route)
}

// AddRoutes adds multiple routes to the module
func (m *BaseModule) AddRoutes(routes []Route) {
	m.routes = append(m.routes, routes...)
}

// AddMiddleware adds middleware to the module
func (m *BaseModule) AddMiddleware(middleware echo.MiddlewareFunc) {
	m.middleware = append(m.middleware, middleware)
}

// AddDependency adds a dependency to the module
func (m *BaseModule) AddDependency(dep string) {
	m.dependencies = append(m.dependencies, dep)
}