package container

import (
	"fmt"
	
	core "github.com/karurosux/saas-go-kit/core-go"
	"github.com/labstack/echo/v4"
)

// Module provides dependency injection functionality as a saas-go-kit module
type Module struct {
	*core.BaseModule
	container Container
}

// NewModule creates a new container module
func NewModule(container Container) *Module {
	module := &Module{
		BaseModule: core.NewBaseModule("container"),
		container:  container,
	}
	
	// Register diagnostic routes
	module.registerRoutes()
	
	return module
}

// registerRoutes adds diagnostic endpoints
func (m *Module) registerRoutes() {
	routes := []core.Route{
		{
			Method:      "GET",
			Path:        "/container/services",
			Handler:     m.listServices,
			Name:        "container.services",
			Description: "List all registered services",
		},
		{
			Method:      "GET",
			Path:        "/container/health",
			Handler:     m.healthCheck,
			Name:        "container.health",
			Description: "Check health of all services",
		},
	}
	
	m.AddRoutes(routes)
}

// listServices returns all registered services
func (m *Module) listServices(c echo.Context) error {
	// For now, we'll just return a simple response
	return c.JSON(200, map[string]interface{}{
		"message": "Service listing endpoint",
		"note":    "Detailed implementation pending",
	})
}

// healthCheck checks health of all services
func (m *Module) healthCheck(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"status": "healthy",
		"note":   "Container service is operational",
	})
}

// GetContainer returns the underlying container
func (m *Module) GetContainer() Container {
	return m.container
}

// Helper function to get type name
func getTypeName(v interface{}) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%T", v)
}