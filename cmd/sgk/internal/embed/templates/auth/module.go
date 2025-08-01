package auth

import (
	"fmt"
	"os"
	
	"{{.Project.GoModule}}/internal/auth/controller"
	"{{.Project.GoModule}}/internal/auth/interface"
	"{{.Project.GoModule}}/internal/auth/middleware"
	"{{.Project.GoModule}}/internal/auth/repository/gorm"
	"{{.Project.GoModule}}/internal/auth/repository/redis"
	"{{.Project.GoModule}}/internal/auth/service"
	"{{.Project.GoModule}}/internal/core"
	"github.com/labstack/echo/v4"
	goredis "github.com/redis/go-redis/v9"
	gormdb "gorm.io/gorm"
)

// RegisterModule registers the auth module with the container
func RegisterModule(c core.Container) error {
	// Get dependencies from container
	e, ok := c.Get("echo").(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance not found in container")
	}
	
	db, ok := c.Get("db").(*gormdb.DB)
	if !ok {
		return fmt.Errorf("database instance not found in container")
	}
	
	// Get Redis client (optional)
	var sessionStore authinterface.SessionStore
	if redisClient, ok := c.Get("redis").(*goredis.Client); ok {
		sessionStore = redis.NewSessionStore(redisClient, "session")
	} else {
		// Fallback to in-memory or database session store
		return fmt.Errorf("redis instance not found in container - required for session storage")
	}
	
	// Run migrations
	if err := gorm.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to run auth migrations: %w", err)
	}
	
	// Create repositories
	accountRepo := gorm.NewAccountRepository(db)
	tokenRepo := gorm.NewTokenRepository(db)
	
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
		if cfg, ok := config.(*authservice.DefaultAuthConfig); ok {
			cfg.jwtSecret = jwtSecret
		}
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