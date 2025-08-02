package healthcontroller

import (
	"fmt"
	"net/http"
	
	"{{.Project.GoModule}}/internal/core"
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
	"github.com/labstack/echo/v4"
)

type HealthController struct {
	service healthinterface.HealthService
}

func NewHealthController(service healthinterface.HealthService) *HealthController {
	return &HealthController{
		service: service,
	}
}

func (hc *HealthController) RegisterRoutes(e *echo.Echo, basePath string) {
	group := e.Group(basePath)
	
	group.GET("", hc.GetHealth)
	group.GET("/live", hc.GetLiveness)
	group.GET("/ready", hc.GetReadiness)
	group.GET("/detailed", hc.GetDetailedHealth)
	group.GET("/check/:name", hc.GetSpecificCheck)
}

func (hc *HealthController) GetHealth(c echo.Context) error {
	if hc.service.IsHealthy() {
		return c.JSON(http.StatusOK, map[string]any{
			"status": "healthy",
		})
	}
	
	return c.JSON(http.StatusServiceUnavailable, map[string]any{
		"status": "unhealthy",
	})
}

func (hc *HealthController) GetLiveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status": "alive",
	})
}

func (hc *HealthController) GetReadiness(c echo.Context) error {
	report := hc.service.CheckAll(c.Request().Context())
	
	response := map[string]any{
		"status":         report.GetStatus(),
		"total_checks":   report.GetTotalChecks(),
		"healthy_checks": report.GetHealthyChecks(),
	}
	
	if report.GetStatus() == healthinterface.StatusOK {
		return c.JSON(http.StatusOK, response)
	}
	
	return c.JSON(http.StatusServiceUnavailable, response)
}

func (hc *HealthController) GetDetailedHealth(c echo.Context) error {
	report := hc.service.CheckAll(c.Request().Context())
	
	httpStatus := http.StatusOK
	if report.GetStatus() == healthinterface.StatusDown {
		httpStatus = http.StatusServiceUnavailable
	} else if report.GetStatus() == healthinterface.StatusDegraded {
		httpStatus = http.StatusOK
	}
	
	return c.JSON(httpStatus, report)
}

func (hc *HealthController) GetSpecificCheck(c echo.Context) error {
	name := c.Param("name")
	
	check, err := hc.service.Check(c.Request().Context(), name)
	if err != nil {
		return core.NotFound(c, fmt.Errorf("Health check not found"))
	}
	
	httpStatus := http.StatusOK
	if check.GetStatus() == healthinterface.StatusDown {
		httpStatus = http.StatusServiceUnavailable
	}
	
	return c.JSON(httpStatus, check)
}
