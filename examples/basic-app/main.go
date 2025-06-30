package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/karurosux/saas-go-kit/auth-go"
	"github.com/karurosux/saas-go-kit/core-go"
	"github.com/karurosux/saas-go-kit/errors-go"
	"github.com/karurosux/saas-go-kit/ratelimit-go"
	"github.com/karurosux/saas-go-kit/response-go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Setup database
	db, err := setupDatabase()
	if err != nil {
		log.Fatal("Failed to setup database:", err)
	}

	// Setup auth service
	authService := setupAuthService(db)

	// Create Echo instance with saas-go-kit
	e, err := setupApp(authService)
	if err != nil {
		log.Fatal("Failed to setup app:", err)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func setupDatabase() (*gorm.DB, error) {
	// Use SQLite for example
	db, err := gorm.Open(sqlite.Open("example.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&AccountModel{},
		&TokenModel{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

func setupAuthService(db *gorm.DB) auth.AuthService {
	// Create stores
	accountStore := NewGormAccountStore(db)
	tokenStore := NewGormTokenStore(db)
	
	// Create email provider
	emailProvider := NewConsoleEmailProvider()
	
	// Create config provider
	configProvider := &AppConfig{
		jwtSecret:      getEnvOrDefault("JWT_SECRET", "super-secret-key"),
		jwtExpiration:  24 * time.Hour,
		appURL:         getEnvOrDefault("APP_URL", "http://localhost:8080"),
		appName:        getEnvOrDefault("APP_NAME", "SaaS Go Kit Example"),
		isDev:          getEnvOrDefault("ENV", "development") == "development",
	}

	// Create auth service
	return auth.NewService(
		accountStore,
		tokenStore,
		emailProvider,
		configProvider,
		auth.WithEventListener(&LogEventListener{}),
	)
}

func setupApp(authService auth.AuthService) (*echo.Echo, error) {
	// Create rate limiter
	rateLimiter := ratelimit.New(10, time.Minute)

	// Create auth module
	authModule := auth.NewModule(auth.ModuleConfig{
		Service:         authService,
		RoutePrefix:     "/api/auth",
		RequireVerified: false,
		RateLimiter:     rateLimiter.EchoMiddleware(),
	})

	// Build app with saas-go-kit
	return core.NewBuilder().
		WithDebug(true).
		WithRoutePrefix("/api/v1").
		WithMiddleware(
			middleware.Logger(),
			middleware.Recover(),
			middleware.CORS(),
		).
		WithErrorHandler(customErrorHandler).
		WithModule(authModule).
		Build()
}

// Custom error handler
func customErrorHandler(err error, c echo.Context) {
	// Handle Echo HTTP errors
	if he, ok := err.(*echo.HTTPError); ok {
		response.Error(c, errors.New(
			fmt.Sprintf("HTTP_%d", he.Code),
			fmt.Sprintf("%v", he.Message),
			he.Code,
		))
		return
	}

	// Handle other errors
	response.Error(c, err)
}

// Helper function
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Example event listener
type LogEventListener struct{}

func (l *LogEventListener) OnRegister(ctx context.Context, account auth.Account) error {
	log.Printf("New account registered: %s", account.GetEmail())
	return nil
}

func (l *LogEventListener) OnLogin(ctx context.Context, account auth.Account) error {
	log.Printf("Account logged in: %s", account.GetEmail())
	return nil
}

func (l *LogEventListener) OnLogout(ctx context.Context, account auth.Account) error {
	log.Printf("Account logged out: %s", account.GetEmail())
	return nil
}

func (l *LogEventListener) OnPasswordReset(ctx context.Context, account auth.Account) error {
	log.Printf("Password reset for: %s", account.GetEmail())
	return nil
}

func (l *LogEventListener) OnEmailVerified(ctx context.Context, account auth.Account) error {
	log.Printf("Email verified for: %s", account.GetEmail())
	return nil
}

func (l *LogEventListener) OnEmailChanged(ctx context.Context, account auth.Account, oldEmail string) error {
	log.Printf("Email changed from %s to %s", oldEmail, account.GetEmail())
	return nil
}

func (l *LogEventListener) OnAccountDeactivated(ctx context.Context, account auth.Account) error {
	log.Printf("Account deactivated: %s", account.GetEmail())
	return nil
}