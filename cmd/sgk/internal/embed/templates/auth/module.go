package auth

import (
	"fmt"
	"os"
	
	"github.com/samber/do"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	
	"{{.Project.GoModule}}/internal/core"
	authcontroller "{{.Project.GoModule}}/internal/auth/controller"
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmiddleware "{{.Project.GoModule}}/internal/auth/middleware"
	authgorm "{{.Project.GoModule}}/internal/auth/repository/gorm"
	authredis "{{.Project.GoModule}}/internal/auth/repository/redis"
	authservice "{{.Project.GoModule}}/internal/auth/service"
)


func ProvideRedisClient(i *do.Injector) (*redis.Client, error) {
	config := do.MustInvoke[*core.Config](i)
	
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}
	
	client := redis.NewClient(opts)
	return client, nil
}

func ProvideAccountRepository(i *do.Injector) (authinterface.AccountRepository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	
	if err := authgorm.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run auth migrations: %w", err)
	}
	
	return authgorm.NewAccountRepository(db), nil
}

func ProvideTokenRepository(i *do.Injector) (authinterface.TokenRepository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	return authgorm.NewTokenRepository(db), nil
}

func ProvideSessionStore(i *do.Injector) (authinterface.SessionStore, error) {
	redisClient := do.MustInvoke[*redis.Client](i)
	return authredis.NewSessionStore(redisClient, "session"), nil
}

func ProvidePasswordHasher(i *do.Injector) (authinterface.PasswordHasher, error) {
	return authservice.NewBcryptPasswordHasher(12), nil
}

func ProvideTokenGenerator(i *do.Injector) (authinterface.TokenGenerator, error) {
	return authservice.NewTokenGenerator(), nil
}

func ProvideEmailSender(i *do.Injector) (authinterface.EmailSender, error) {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	
	return authservice.NewMockEmailSender(baseURL), nil
}

func ProvideSMSSender(i *do.Injector) (authinterface.SMSSender, error) {
	return authservice.NewMockSMSSender(), nil
}

func ProvideAuthConfig(i *do.Injector) (authinterface.AuthConfig, error) {
	return authservice.NewDefaultAuthConfig(), nil
}

func ProvideAuthService(i *do.Injector) (authinterface.AuthService, error) {
	accountRepo := do.MustInvoke[authinterface.AccountRepository](i)
	tokenRepo := do.MustInvoke[authinterface.TokenRepository](i)
	sessionStore := do.MustInvoke[authinterface.SessionStore](i)
	passwordHasher := do.MustInvoke[authinterface.PasswordHasher](i)
	tokenGenerator := do.MustInvoke[authinterface.TokenGenerator](i)
	emailSender := do.MustInvoke[authinterface.EmailSender](i)
	smsSender := do.MustInvoke[authinterface.SMSSender](i)
	authConfig := do.MustInvoke[authinterface.AuthConfig](i)
	
	return authservice.NewAuthService(
		accountRepo,
		tokenRepo,
		sessionStore,
		passwordHasher,
		tokenGenerator,
		emailSender,
		smsSender,
		authConfig,
	), nil
}

func ProvideAuthMiddleware(i *do.Injector) (*authmiddleware.AuthMiddleware, error) {
	authService := do.MustInvoke[authinterface.AuthService](i)
	return authmiddleware.NewAuthMiddleware(authService), nil
}

func ProvideAuthController(i *do.Injector) (*authcontroller.AuthController, error) {
	authService := do.MustInvoke[authinterface.AuthService](i)
	return authcontroller.NewAuthController(authService), nil
}

func RegisterModule(container *core.Container) error {
	do.Provide(container, ProvideRedisClient)
	do.Provide(container, ProvideAccountRepository)
	do.Provide(container, ProvideTokenRepository)
	do.Provide(container, ProvideSessionStore)
	do.Provide(container, ProvidePasswordHasher)
	do.Provide(container, ProvideTokenGenerator)
	do.Provide(container, ProvideEmailSender)
	do.Provide(container, ProvideSMSSender)
	do.Provide(container, ProvideAuthConfig)
	do.Provide(container, ProvideAuthService)
	do.Provide(container, ProvideAuthMiddleware)
	do.Provide(container, ProvideAuthController)
	
	e := do.MustInvoke[*echo.Echo](container)
	authController := do.MustInvoke[*authcontroller.AuthController](container)
	authMiddleware := do.MustInvoke[*authmiddleware.AuthMiddleware](container)
	
	authController.RegisterRoutes(e, "/api/v1/auth", authMiddleware)
	
	return nil
}

