package core

import (
	"fmt"
	
	"github.com/samber/do"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Container is an alias for samber/do injector
type Container = do.Injector

// NewContainer creates a new dependency injection container
func NewContainer() *Container {
	return do.New()
}

// Service constructors for common dependencies

// ProvideConfig provides the application configuration
func ProvideConfig(i *do.Injector) (*Config, error) {
	config := LoadConfig()
	if err := config.ValidateRequired(); err != nil {
		return nil, err
	}
	return config, nil
}

// ProvideEcho provides the Echo web server instance
func ProvideEcho(i *do.Injector) (*echo.Echo, error) {
	config := do.MustInvoke[*Config](i)
	
	e := echo.New()
	
	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// Configure CORS with settings from config
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.CORSAllowedOrigins,
		AllowMethods: config.CORSAllowedMethods,
		AllowHeaders: config.CORSAllowedHeaders,
	}))
	
	// Add rate limiting if enabled
	if config.RateLimitEnabled {
		e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(config.RateLimitRPM),
		)))
	}
	
	return e, nil
}

// ProvideDatabase provides the database connection
func ProvideDatabase(i *do.Injector) (*gorm.DB, error) {
	config := do.MustInvoke[*Config](i)
	
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf(ErrMsgDatabaseConnection, err)
	}
	
	return db, nil
}

// RegisterCoreServices registers all core services with the container
func RegisterCoreServices(container *Container) error {
	// Register core services
	do.Provide(container, ProvideConfig)
	do.Provide(container, ProvideEcho)  
	do.Provide(container, ProvideDatabase)
	
	return nil
}

