package health

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/karurosux/saas-go-kit/core-go"
)

// Module provides health check endpoints
type Module struct {
	*core.BaseModule
	service Service
}

// ModuleConfig holds module configuration
type ModuleConfig struct {
	Service       Service
	RoutePrefix   string
	DetailedCheck bool // If true, returns detailed check information
}

// NewModule creates a new health module
func NewModule(config ModuleConfig) *Module {
	if config.RoutePrefix == "" {
		config.RoutePrefix = "/health"
	}

	module := &Module{
		BaseModule: core.NewBaseModule("health"),
		service:    config.Service,
	}

	// Register routes
	module.registerRoutes(config)

	return module
}

// registerRoutes registers all health routes
func (m *Module) registerRoutes(config ModuleConfig) {
	routes := []core.Route{
		{
			Method:      "GET",
			Path:        config.RoutePrefix,
			Handler:     m.handleHealthCheck(config.DetailedCheck),
			Name:        "health.check",
			Description: "Health check endpoint",
		},
		{
			Method:      "GET",
			Path:        config.RoutePrefix + "/live",
			Handler:     m.handleLiveness,
			Name:        "health.liveness",
			Description: "Kubernetes liveness probe",
		},
		{
			Method:      "GET",
			Path:        config.RoutePrefix + "/ready",
			Handler:     m.handleReadiness,
			Name:        "health.readiness",
			Description: "Kubernetes readiness probe",
		},
	}

	// Add detailed endpoint if enabled
	if config.DetailedCheck {
		routes = append(routes, core.Route{
			Method:      "GET",
			Path:        config.RoutePrefix + "/detailed",
			Handler:     m.handleDetailedCheck,
			Name:        "health.detailed",
			Description: "Detailed health check with all components",
		})
	}

	m.AddRoutes(routes)
}

// handleHealthCheck handles the main health check endpoint
func (m *Module) handleHealthCheck(detailed bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		if detailed {
			return m.handleDetailedCheck(c)
		}

		// Simple health check
		if m.service.IsHealthy() {
			return c.JSON(http.StatusOK, map[string]string{
				"status": "ok",
			})
		}

		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
		})
	}
}

// handleLiveness handles Kubernetes liveness probe
func (m *Module) handleLiveness(c echo.Context) error {
	// Liveness just checks if the service is running
	// It doesn't check external dependencies
	return c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}

// handleReadiness handles Kubernetes readiness probe
func (m *Module) handleReadiness(c echo.Context) error {
	// Readiness checks if the service is ready to accept traffic
	// This includes checking critical dependencies
	report := m.service.GetReport()
	
	if report.Status == StatusOK {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ready",
			"checks": report.Healthy,
			"total":  report.TotalChecks,
		})
	}

	return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
		"status": "not ready",
		"checks": report.Healthy,
		"total":  report.TotalChecks,
	})
}

// handleDetailedCheck handles detailed health check
func (m *Module) handleDetailedCheck(c echo.Context) error {
	ctx := c.Request().Context()
	report := m.service.CheckAll(ctx)

	status := http.StatusOK
	if report.Status == StatusDegraded {
		status = http.StatusOK // Still return 200 for degraded
	} else if report.Status == StatusDown {
		status = http.StatusServiceUnavailable
	}

	return c.JSON(status, report)
}