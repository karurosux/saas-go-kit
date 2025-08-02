package health

import (
	"context"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"gorm.io/gorm"

	"{{.Project.GoModule}}/internal/core"
	healthcheckers "{{.Project.GoModule}}/internal/health/checkers"
	healthconstants "{{.Project.GoModule}}/internal/health/constants"
	healthcontroller "{{.Project.GoModule}}/internal/health/controller"
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
	healthservice "{{.Project.GoModule}}/internal/health/service"
)

// Service providers for dependency injection

// ProvideHealthService provides the health service
func ProvideHealthService(i *do.Injector) (healthinterface.HealthService, error) {
	// Get version from environment or default
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0"
	}

	healthService := healthservice.NewHealthService(version)

	// Register database checker
	db := do.MustInvoke[*gorm.DB](i)
	dbChecker := healthcheckers.NewDatabaseChecker(db, true) // Critical
	healthService.RegisterChecker(dbChecker)

	// Register Redis checker if Redis is available
	if redisClient, err := do.Invoke[*redis.Client](i); err == nil {
		redisChecker := healthcheckers.NewRedisChecker(redisClient, false) // Non-critical
		healthService.RegisterChecker(redisChecker)
	}

	// Register disk space checker
	diskPath := os.Getenv("HEALTH_CHECK_DISK_PATH")
	if diskPath == "" {
		diskPath = "/"
	}
	diskThreshold := 90.0 // Can be configured via env
	diskChecker := healthcheckers.NewDiskSpaceChecker(diskPath, diskThreshold, false)
	healthService.RegisterChecker(diskChecker)

	// Register memory checker
	memoryThreshold := 80.0 // Can be configured via env
	memoryChecker := healthcheckers.NewMemoryChecker(memoryThreshold, false)
	healthService.RegisterChecker(memoryChecker)

	// Register external service checkers if configured
	if externalHealthURL := os.Getenv("EXTERNAL_SERVICE_HEALTH_URL"); externalHealthURL != "" {
		httpChecker := healthcheckers.NewHTTPChecker("external_service", externalHealthURL, false)
		httpChecker.SetTimeout(5 * time.Second)
		healthService.RegisterChecker(httpChecker)
	}

	return healthService, nil
}

// ProvideHealthController provides the health controller
func ProvideHealthController(i *do.Injector) (*healthcontroller.HealthController, error) {
	healthService := do.MustInvoke[healthinterface.HealthService](i)
	return healthcontroller.NewHealthController(healthService), nil
}

// RegisterModule registers the health module with the container
func RegisterModule(container *core.Container) error {
	// Register health services
	do.Provide(container, ProvideHealthService)
	do.Provide(container, ProvideHealthController)

	// Register routes
	e := do.MustInvoke[*echo.Echo](container)
	healthController := do.MustInvoke[*healthcontroller.HealthController](container)
	healthController.RegisterRoutes(e, "/api/v1/health")

	// Start periodic health checks
	healthService := do.MustInvoke[healthinterface.HealthService](container)
	checkInterval := healthconstants.DefaultPeriodicInterval
	if intervalStr := os.Getenv("HEALTH_CHECK_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			checkInterval = interval
		}
	}

	ctx := context.Background()
	healthService.StartPeriodicChecks(ctx, checkInterval)

	return nil
}
