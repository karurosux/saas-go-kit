package sse

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/sse/constants"
	"{{.Project.GoModule}}/internal/sse/controller"
	"{{.Project.GoModule}}/internal/sse/interface"
	"{{.Project.GoModule}}/internal/sse/service"
	"github.com/labstack/echo/v4"
)

// RegisterModule registers the SSE module with the container
func RegisterModule(c core.Container) error {
	// Get dependencies from container
	e, ok := c.Get("echo").(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance not found in container")
	}
	
	// Create SSE configuration from environment
	config := createSSEConfig()
	
	// Create hub
	hub := sseservice.NewHub(config)
	
	// Create SSE service
	sseService := sseservice.NewSSEService(hub, config)
	
	// Create controller
	sseController := ssecontroller.NewSSEController(sseService, hub)
	
	// Register routes
	sseController.RegisterRoutes(e, "/sse")
	
	// Start the SSE service
	if err := sseService.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start SSE service: %w", err)
	}
	
	// Register components in container for other modules to use
	c.Set("sse.service", sseService)
	c.Set("sse.hub", hub)
	
	// Register shutdown handler
	c.OnShutdown(func() {
		sseService.Stop()
	})
	
	return nil
}

// createSSEConfig creates SSE configuration from environment variables
func createSSEConfig() sseinterface.Config {
	config := sseinterface.Config{
		BufferSize:        getIntEnv(sseconstants.EnvSSEBufferSize, sseconstants.DefaultBufferSize),
		ClientTimeout:     getDurationEnv(sseconstants.EnvSSEClientTimeout, sseconstants.DefaultClientTimeout),
		HeartbeatInterval: getDurationEnv(sseconstants.EnvSSEHeartbeatInterval, sseconstants.DefaultHeartbeatInterval),
		MaxClients:        getIntEnv(sseconstants.EnvSSEMaxClients, sseconstants.DefaultMaxClients),
		MaxClientsPerUser: getIntEnv(sseconstants.EnvSSEMaxClientsPerUser, sseconstants.DefaultMaxClientsPerUser),
		EnableHeartbeat:   getBoolEnv(sseconstants.EnvSSEEnableHeartbeat, true),
		EnableMetrics:     getBoolEnv(sseconstants.EnvSSEEnableMetrics, true),
	}
	
	return config
}

// Helper functions to parse environment variables

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil && intValue > 0 {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}