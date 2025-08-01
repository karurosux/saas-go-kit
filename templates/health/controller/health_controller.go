package healthcontroller

import (
	"net/http"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/health/interface"
	"github.com/labstack/echo/v4"
)

// HealthController handles health check requests
type HealthController struct {
	service healthinterface.HealthService
}

// NewHealthController creates a new health controller
func NewHealthController(service healthinterface.HealthService) *HealthController {
	return &HealthController{
		service: service,
	}
}

// RegisterRoutes registers all health-related routes
func (hc *HealthController) RegisterRoutes(e *echo.Echo, basePath string) {
	group := e.Group(basePath)
	
	// Health check endpoints
	group.GET("", hc.GetHealth)
	group.GET("/live", hc.GetLiveness)
	group.GET("/ready", hc.GetReadiness)
	group.GET("/detailed", hc.GetDetailedHealth)
	group.GET("/check/:name", hc.GetSpecificCheck)
}

// GetHealth godoc
// @Summary Get health status
// @Description Get basic health status
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Success 503 {object} map[string]interface{}
// @Router /health [get]
func (hc *HealthController) GetHealth(c echo.Context) error {
	if hc.service.IsHealthy() {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "healthy",
		})
	}
	
	return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
		"status": "unhealthy",
	})
}

// GetLiveness godoc
// @Summary Get liveness status
// @Description Check if the service is alive (for Kubernetes liveness probe)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health/live [get]
func (hc *HealthController) GetLiveness(c echo.Context) error {
	// Liveness check - always return OK unless the service is completely broken
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "alive",
	})
}

// GetReadiness godoc
// @Summary Get readiness status
// @Description Check if the service is ready to accept traffic (for Kubernetes readiness probe)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Success 503 {object} map[string]interface{}
// @Router /health/ready [get]
func (hc *HealthController) GetReadiness(c echo.Context) error {
	// Run all health checks to determine readiness
	report := hc.service.CheckAll(c.Request().Context())
	
	response := map[string]interface{}{
		"status":         report.GetStatus(),
		"total_checks":   report.GetTotalChecks(),
		"healthy_checks": report.GetHealthyChecks(),
	}
	
	if report.GetStatus() == healthinterface.StatusOK {
		return c.JSON(http.StatusOK, response)
	}
	
	return c.JSON(http.StatusServiceUnavailable, response)
}

// GetDetailedHealth godoc
// @Summary Get detailed health report
// @Description Get detailed health report with all checks
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} healthmodel.Report
// @Success 503 {object} healthmodel.Report
// @Router /health/detailed [get]
func (hc *HealthController) GetDetailedHealth(c echo.Context) error {
	report := hc.service.CheckAll(c.Request().Context())
	
	// Determine HTTP status based on health status
	httpStatus := http.StatusOK
	if report.GetStatus() == healthinterface.StatusDown {
		httpStatus = http.StatusServiceUnavailable
	} else if report.GetStatus() == healthinterface.StatusDegraded {
		// Still return 200 for degraded status, but the status field indicates degradation
		httpStatus = http.StatusOK
	}
	
	return c.JSON(httpStatus, report)
}

// GetSpecificCheck godoc
// @Summary Get specific health check
// @Description Get the result of a specific health check
// @Tags health
// @Accept json
// @Produce json
// @Param name path string true "Check name"
// @Success 200 {object} healthmodel.Check
// @Failure 404 {object} core.ErrorResponse
// @Failure 503 {object} healthmodel.Check
// @Router /health/check/{name} [get]
func (hc *HealthController) GetSpecificCheck(c echo.Context) error {
	name := c.Param("name")
	
	check, err := hc.service.Check(c.Request().Context(), name)
	if err != nil {
		return core.Error(c, core.NotFound("Health check not found"))
	}
	
	// Determine HTTP status based on check status
	httpStatus := http.StatusOK
	if check.GetStatus() == healthinterface.StatusDown {
		httpStatus = http.StatusServiceUnavailable
	}
	
	return c.JSON(httpStatus, check)
}