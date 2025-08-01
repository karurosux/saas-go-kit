package auth

import (
	"fmt"
	"os"
	authredis "{{.Project.GoModule}}/internal/auth/repository/redis"
	"{{.Project.GoModule}}/internal/core"

	authcontroller "{{.Project.GoModule}}/internal/auth/controller"
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmiddleware "{{.Project.GoModule}}/internal/auth/middleware"
	authgorm "{{.Project.GoModule}}/internal/auth/repository/gorm"

	authservice "{{.Project.GoModule}}/internal/auth/service"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterModule registers the auth module with the container
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

	dbInt, err := c.Get("db")
	if err != nil {
		return fmt.Errorf("database instance not found in container: %w", err)
	}
	db, ok := dbInt.(*gorm.DB)
	if !ok {
		return fmt.Errorf("database instance has invalid type")
	}

	// Get Redis client (optional)
	var sessionStore authinterface.SessionStore
	redisInt, err := c.Get("redis")
	if err != nil {
		return fmt.Errorf("redis instance not found in container - required for session storage: %w", err)
	}
	redisClient, ok := redisInt.(*redis.Client)
	if !ok {
		return fmt.Errorf("redis instance has invalid type")
	}
	sessionStore = authredis.NewSessionStore(redisClient, "session")

	// Run migrations
	if err := authgorm.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to run auth migrations: %w", err)
	}

	// Create repositories
	accountRepo := authgorm.NewAccountRepository(db)
	tokenRepo := authgorm.NewTokenRepository(db)

	// Create service dependencies
	passwordHasher := authservice.NewBcryptPasswordHasher(12)
	tokenGenerator := authservice.NewTokenGenerator()

	// Get base URL from environment
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	emailSender := authservice.NewMockEmailSender(baseURL)
	smsSender := authservice.NewMockSMSSender()

	// Create auth config
	config := authservice.NewDefaultAuthConfig()

	// Override config from environment
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		// TODO: Add method to set JWT secret in config
		// if cfg, ok := config.(*authservice.DefaultAuthConfig); ok {
		//	cfg.SetJWTSecret(jwtSecret)
		// }
	}

	// Create auth service
	authService := authservice.NewAuthService(
		accountRepo,
		tokenRepo,
		sessionStore,
		passwordHasher,
		tokenGenerator,
		emailSender,
		smsSender,
		config,
	)

	// Create middleware
	authMiddleware := authmiddleware.NewAuthMiddleware(authService)

	// Create controller
	authController := authcontroller.NewAuthController(authService)

	// Register routes
	authController.RegisterRoutes(e, "/auth", authMiddleware)

	// Register components in container for other modules to use
	c.Set("auth.service", authService)
	c.Set("auth.middleware", authMiddleware)
	c.Set("auth.accountRepository", accountRepo)
	c.Set("auth.tokenRepository", tokenRepo)
	c.Set("auth.sessionStore", sessionStore)

	return nil
}

