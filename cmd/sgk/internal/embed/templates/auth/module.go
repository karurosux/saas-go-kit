package auth

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"gorm.io/gorm"

	authcontroller "{{.Project.GoModule}}/internal/auth/controller"
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmiddleware "{{.Project.GoModule}}/internal/auth/middleware"
	authgorm "{{.Project.GoModule}}/internal/auth/repository/gorm"
	authredis "{{.Project.GoModule}}/internal/auth/repository/redis"
	authservice "{{.Project.GoModule}}/internal/auth/service"
	"{{.Project.GoModule}}/internal/core"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
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

	emailService, err := do.Invoke[emailinterface.EmailService](i)
	if err != nil {
		return authservice.NewMockEmailSender(baseURL), nil
	}

	return authservice.NewEmailSenderAdapter(emailService, baseURL), nil
}

func ProvideAuthConfig(i *do.Injector) (authinterface.AuthConfig, error) {
	return authservice.NewDefaultAuthConfig(), nil
}

func ProvideStrategyRegistry(i *do.Injector) (authinterface.StrategyRegistry, error) {
	registry := authservice.NewStrategyRegistry()
	
	// Register email/password strategy
	accountRepo := do.MustInvoke[authinterface.AccountRepository](i)
	passwordHasher := do.MustInvoke[authinterface.PasswordHasher](i)
	emailPasswordStrategy := authservice.NewEmailPasswordStrategy(accountRepo, passwordHasher)
	if err := registry.Register(emailPasswordStrategy); err != nil {
		return nil, err
	}
	
	// Register OAuth strategies if configured
	googleClientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	if googleClientID != "" && googleClientSecret != "" {
		googleRedirectURI := os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
		if googleRedirectURI == "" {
			googleRedirectURI = "http://localhost:8080/api/v1/auth/oauth/google/callback"
		}
		googleStrategy := authservice.NewGoogleOAuthStrategy(
			accountRepo,
			googleClientID,
			googleClientSecret,
			googleRedirectURI,
		)
		if err := registry.Register(googleStrategy); err != nil {
			return nil, err
		}
	}
	
	return registry, nil
}

func ProvideAuthService(i *do.Injector) (authinterface.AuthService, error) {
	accountRepo := do.MustInvoke[authinterface.AccountRepository](i)
	tokenRepo := do.MustInvoke[authinterface.TokenRepository](i)
	sessionStore := do.MustInvoke[authinterface.SessionStore](i)
	passwordHasher := do.MustInvoke[authinterface.PasswordHasher](i)
	tokenGenerator := do.MustInvoke[authinterface.TokenGenerator](i)
	emailSender := do.MustInvoke[authinterface.EmailSender](i)
	authConfig := do.MustInvoke[authinterface.AuthConfig](i)
	strategyRegistry := do.MustInvoke[authinterface.StrategyRegistry](i)

	return authservice.NewAuthService(
		accountRepo,
		tokenRepo,
		sessionStore,
		passwordHasher,
		tokenGenerator,
		emailSender,
		authConfig,
		strategyRegistry,
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
	do.Provide(container, ProvideAuthConfig)
	do.Provide(container, ProvideStrategyRegistry)
	do.Provide(container, ProvideAuthService)
	do.Provide(container, ProvideAuthMiddleware)
	do.Provide(container, ProvideAuthController)

	e := do.MustInvoke[*echo.Echo](container)
	authController := do.MustInvoke[*authcontroller.AuthController](container)
	authMiddleware := do.MustInvoke[*authmiddleware.AuthMiddleware](container)

	authController.RegisterRoutes(e, "/api/v1/auth", authMiddleware)

	return nil
}
