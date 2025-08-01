package health

import (
	"context"
	"fmt"
	"os"
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	healthcheckers "{{.Project.GoModule}}/internal/health/checkers"
	healthconstants "{{.Project.GoModule}}/internal/health/constants"
	healthcontroller "{{.Project.GoModule}}/internal/health/controller"
	healthservice "{{.Project.GoModule}}/internal/health/service"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterModule registers the health module with the container
func RegisterModule(c *core.Container) error {
	// Get dependencies from container
	eInt, err := c.Get("echo")
	if err != nil {
		return fmt.Errorf("echo instance not found in container: %w", err)
	}
	e, ok := eInt.(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance has invalid type")
	}
	
	// Get version from environment or default
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0"
	}
	
	// Create health service
	healthService := healthservice.NewHealthService(version)
	
	// Register database checker if available
	if dbInt, err := c.Get("db"); err == nil {
		if db, ok := dbInt.(*gorm.DB); ok {
			dbChecker := healthcheckers.NewDatabaseChecker(db, true) // Critical
			healthService.RegisterChecker(dbChecker)
		}
	}
	
	// Register Redis checker if available
	if redisInt, err := c.Get("redis"); err == nil {
		if redisClient, ok := redisInt.(*redis.Client); ok {
			redisChecker := healthcheckers.NewRedisChecker(redisClient, false) // Non-critical
			healthService.RegisterChecker(redisChecker)
		}
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
	
	// Create controller
	healthController := healthcontroller.NewHealthController(healthService)
	
	// Register routes
	healthController.RegisterRoutes(e, "/health")
	
	// Start periodic health checks
	checkInterval := healthconstants.DefaultPeriodicInterval
	if intervalStr := os.Getenv("HEALTH_CHECK_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			checkInterval = interval
		}
	}
	
	ctx := context.Background()
	healthService.StartPeriodicChecks(ctx, checkInterval)
	
	// Register health service in container for other modules to use
	c.Set("health.service", healthService)
	
	// TODO: Register shutdown handler to stop periodic checks
	// c.OnShutdown(func() {
	//	healthService.StopPeriodicChecks()
	// })
	
	return nil
}