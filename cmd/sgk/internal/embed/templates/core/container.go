package core

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/do"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Container = do.Injector

func NewContainer() *Container {
	return do.New()
}

func ProvideConfig(i *do.Injector) (*Config, error) {
	config := LoadConfig()
	if err := config.ValidateRequired(); err != nil {
		return nil, err
	}
	return config, nil
}

func ProvideEcho(i *do.Injector) (*echo.Echo, error) {
	config := do.MustInvoke[*Config](i)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.CORSAllowedOrigins,
		AllowMethods: config.CORSAllowedMethods,
		AllowHeaders: config.CORSAllowedHeaders,
	}))

	if config.RateLimitEnabled {
		e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
			rate.Limit(config.RateLimitRPM),
		)))
	}

	return e, nil
}

func ProvideDatabase(i *do.Injector) (*gorm.DB, error) {
	config := do.MustInvoke[*Config](i)

	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf(ErrMsgDatabaseConnection, err)
	}

	return db, nil
}

func RegisterCoreServices(container *Container) error {
	do.Provide(container, ProvideConfig)
	do.Provide(container, ProvideEcho)
	do.Provide(container, ProvideDatabase)

	return nil
}
