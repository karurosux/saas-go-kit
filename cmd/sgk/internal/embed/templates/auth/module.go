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

// Service providers for dependency injection

// ProvideRedisClient provides Redis client for the auth module
func ProvideRedisClient(i *do.Injector) (*redis.Client, error) {
	config := do.MustInvoke[*core.Config](i)
	
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}
	
	client := redis.NewClient(opts)
	return client, nil
}

// ProvideAccountRepository provides the account repository
func ProvideAccountRepository(i *do.Injector) (authinterface.AccountRepository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	
	// Run migrations
	if err := authgorm.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run auth migrations: %w", err)
	}
	
	return authgorm.NewAccountRepository(db), nil
}

// ProvideTokenRepository provides the token repository
func ProvideTokenRepository(i *do.Injector) (authinterface.TokenRepository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	return authgorm.NewTokenRepository(db), nil
}

// ProvideSessionStore provides the session store
func ProvideSessionStore(i *do.Injector) (authinterface.SessionStore, error) {
	redisClient := do.MustInvoke[*redis.Client](i)
	return authredis.NewSessionStore(redisClient, "session"), nil
}

// ProvidePasswordHasher provides the password hasher
func ProvidePasswordHasher(i *do.Injector) (authinterface.PasswordHasher, error) {
	return authservice.NewBcryptPasswordHasher(12), nil
}

// ProvideTokenGenerator provides the token generator
func ProvideTokenGenerator(i *do.Injector) (authinterface.TokenGenerator, error) {
	return authservice.NewTokenGenerator(), nil
}

// ProvideEmailSender provides the email sender
func ProvideEmailSender(i *do.Injector) (authinterface.EmailSender, error) {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	
	// For now, always use mock sender
	// TODO: Implement real SMTP sender when needed
	return authservice.NewMockEmailSender(baseURL), nil
}

// ProvideSMSSender provides the SMS sender
func ProvideSMSSender(i *do.Injector) (authinterface.SMSSender, error) {
	return authservice.NewMockSMSSender(), nil
}

// ProvideAuthConfig provides the auth configuration
func ProvideAuthConfig(i *do.Injector) (authinterface.AuthConfig, error) {
	return authservice.NewDefaultAuthConfig(), nil
}

// ProvideAuthService provides the main auth service
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

// ProvideAuthMiddleware provides the auth middleware
func ProvideAuthMiddleware(i *do.Injector) (*authmiddleware.AuthMiddleware, error) {
	authService := do.MustInvoke[authinterface.AuthService](i)
	return authmiddleware.NewAuthMiddleware(authService), nil
}

// ProvideAuthController provides the auth controller
func ProvideAuthController(i *do.Injector) (*authcontroller.AuthController, error) {
	authService := do.MustInvoke[authinterface.AuthService](i)
	return authcontroller.NewAuthController(authService), nil
}

// RegisterModule registers the auth module with the container
func RegisterModule(container *core.Container) error {
	// Register all auth services
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
	
	// Register routes
	e := do.MustInvoke[*echo.Echo](container)
	authController := do.MustInvoke[*authcontroller.AuthController](container)
	authMiddleware := do.MustInvoke[*authmiddleware.AuthMiddleware](container)
	
	authController.RegisterRoutes(e, "/auth", authMiddleware)
	
	return nil
}

