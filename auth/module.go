package auth

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/saas-go-kit/core-go"
	"github.com/saas-go-kit/errors-go"
	"github.com/saas-go-kit/response-go"
)

// Module provides auth module for saas-go-kit
type Module struct {
	*core.BaseModule
	service  AuthService
	config   ModuleConfig
	handlers *Handlers
}

// ModuleConfig holds module configuration
type ModuleConfig struct {
	Service         AuthService
	RoutePrefix     string
	RequireVerified bool
	RateLimiter     echo.MiddlewareFunc
}

// NewModule creates a new auth module
func NewModule(config ModuleConfig) *Module {
	if config.RoutePrefix == "" {
		config.RoutePrefix = "/auth"
	}

	module := &Module{
		BaseModule: core.NewBaseModule("auth"),
		service:    config.Service,
		config:     config,
		handlers:   NewHandlers(config.Service),
	}

	// Register routes
	module.registerRoutes()

	return module
}

// registerRoutes registers all auth routes
func (m *Module) registerRoutes() {
	// Public routes
	publicRoutes := []core.Route{
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/register",
			Handler:     m.handlers.Register,
			Name:        "auth.register",
			Description: "Register a new account",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/login",
			Handler:     m.handlers.Login,
			Name:        "auth.login",
			Description: "Login to an account",
		},
		{
			Method:      "GET",
			Path:        m.config.RoutePrefix + "/verify-email",
			Handler:     m.handlers.VerifyEmail,
			Name:        "auth.verify-email",
			Description: "Verify email address",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/forgot-password",
			Handler:     m.handlers.ForgotPassword,
			Name:        "auth.forgot-password",
			Description: "Request password reset",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/reset-password",
			Handler:     m.handlers.ResetPassword,
			Name:        "auth.reset-password",
			Description: "Reset password with token",
		},
	}

	// Add rate limiting to public routes if configured
	if m.config.RateLimiter != nil {
		for i := range publicRoutes {
			publicRoutes[i].Middlewares = append(publicRoutes[i].Middlewares, m.config.RateLimiter)
		}
	}

	// Add public routes
	for _, route := range publicRoutes {
		m.AddRoute(route)
	}

	// Protected routes
	protectedRoutes := []core.Route{
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/refresh",
			Handler:     m.handlers.RefreshToken,
			Name:        "auth.refresh",
			Description: "Refresh JWT token",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/logout",
			Handler:     m.handlers.Logout,
			Name:        "auth.logout",
			Description: "Logout from account",
		},
		{
			Method:      "GET",
			Path:        m.config.RoutePrefix + "/me",
			Handler:     m.handlers.GetProfile,
			Name:        "auth.me",
			Description: "Get current account profile",
		},
		{
			Method:      "PUT",
			Path:        m.config.RoutePrefix + "/profile",
			Handler:     m.handlers.UpdateProfile,
			Name:        "auth.update-profile",
			Description: "Update account profile",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/change-password",
			Handler:     m.handlers.ChangePassword,
			Name:        "auth.change-password",
			Description: "Change account password",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/send-verification",
			Handler:     m.handlers.ResendVerification,
			Name:        "auth.resend-verification",
			Description: "Resend verification email",
		},
	}

	// Add auth middleware to protected routes
	authMiddleware := m.RequireAuth()
	for _, route := range protectedRoutes {
		route.Middlewares = append([]echo.MiddlewareFunc{authMiddleware}, route.Middlewares...)
		m.AddRoute(route)
	}

	// Add global middleware
	m.AddMiddleware(m.OptionalAuth())
}

// RequireAuth returns middleware that requires authentication
func (m *Module) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c)
			if token == "" {
				return response.Error(c, errors.ErrUnauthorized)
			}

			claims, err := m.service.ValidateToken(token)
			if err != nil {
				return response.Error(c, err)
			}

			// Get account
			account, err := m.service.GetAccount(c.Request().Context(), claims.AccountID)
			if err != nil {
				return response.Error(c, errors.ErrUnauthorized)
			}

			// Check if email verification is required
			if m.config.RequireVerified && !account.IsEmailVerified() {
				return response.Error(c, errors.Forbidden("Email verification required"))
			}

			// Store in context
			c.Set("account", account)
			c.Set("account_id", claims.AccountID.String())
			c.Set("claims", claims)

			return next(c)
		}
	}
}

// OptionalAuth returns middleware that optionally authenticates
func (m *Module) OptionalAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c)
			if token != "" {
				claims, err := m.service.ValidateToken(token)
				if err == nil {
					// Get account
					account, err := m.service.GetAccount(c.Request().Context(), claims.AccountID)
					if err == nil {
						c.Set("account", account)
						c.Set("account_id", claims.AccountID.String())
						c.Set("claims", claims)
					}
				}
			}
			return next(c)
		}
	}
}

// RequireVerified returns middleware that requires email verification
func (m *Module) RequireVerified() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			account := GetAccount(c)
			if account == nil {
				return response.Error(c, errors.ErrUnauthorized)
			}

			if !account.IsEmailVerified() {
				return response.Error(c, errors.Forbidden("Email verification required"))
			}

			return next(c)
		}
	}
}

// Helper functions

// extractToken extracts the JWT token from the request
func extractToken(c echo.Context) string {
	// Check Authorization header
	auth := c.Request().Header.Get("Authorization")
	if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}

	// Check cookie
	cookie, err := c.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Check query parameter (for email verification links)
	token := c.QueryParam("token")
	if token != "" {
		return token
	}

	return ""
}

// GetAccount retrieves the authenticated account from context
func GetAccount(c echo.Context) Account {
	if account, ok := c.Get("account").(Account); ok {
		return account
	}
	return nil
}

// GetAccountID retrieves the authenticated account ID from context
func GetAccountID(c echo.Context) string {
	if id, ok := c.Get("account_id").(string); ok {
		return id
	}
	return ""
}

// GetClaims retrieves the JWT claims from context
func GetClaims(c echo.Context) *Claims {
	if claims, ok := c.Get("claims").(*Claims); ok {
		return claims
	}
	return nil
}

// HTTPAuthMiddleware returns standard HTTP middleware for authentication
func (m *Module) HTTPAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			
			// Check Authorization header
			auth := r.Header.Get("Authorization")
			if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
				token = auth[7:]
			}

			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := m.service.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "account_id", claims.AccountID.String())
			ctx = context.WithValue(ctx, "claims", claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}