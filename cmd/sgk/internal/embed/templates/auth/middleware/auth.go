package authmiddleware

import (
	"fmt"
	"strings"
	
	authconstants "{{.Project.GoModule}}/internal/auth/constants"
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	"{{.Project.GoModule}}/internal/core"
	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	authService authinterface.AuthService
	config      MiddlewareConfig
}

type MiddlewareConfig struct {
	Skipper      func(c echo.Context) bool
	TokenLookup  string // "header:Authorization" or "cookie:token"
	TokenPrefix  string // "Bearer " for header
	ErrorHandler func(c echo.Context, err error) error
}

func DefaultConfig() MiddlewareConfig {
	return MiddlewareConfig{
		TokenLookup:  "header:Authorization",
		TokenPrefix:  "Bearer ",
		ErrorHandler: DefaultErrorHandler,
	}
}

func DefaultErrorHandler(c echo.Context, err error) error {
	return core.Unauthorized(c, fmt.Errorf(authconstants.ErrUnauthorized))
}

func NewAuthMiddleware(authService authinterface.AuthService, config ...MiddlewareConfig) *AuthMiddleware {
	cfg := DefaultConfig()
	if len(config) > 0 {
		if config[0].Skipper != nil {
			cfg.Skipper = config[0].Skipper
		}
		if config[0].TokenLookup != "" {
			cfg.TokenLookup = config[0].TokenLookup
		}
		if config[0].TokenPrefix != "" {
			cfg.TokenPrefix = config[0].TokenPrefix
		}
		if config[0].ErrorHandler != nil {
			cfg.ErrorHandler = config[0].ErrorHandler
		}
	}
	
	return &AuthMiddleware{
		authService: authService,
		config:      cfg,
	}
}

func (m *AuthMiddleware) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}
			
			token, err := m.extractToken(c)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}
			
			account, err := m.authService.ValidateSession(c.Request().Context(), token)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}
			
			c.Set(authconstants.ContextKeyUserID, account.GetID())
			c.Set(authconstants.ContextKeyAccount, account)
			c.Set(authconstants.ContextKeyIsAuthenticated, true)
			
			return next(c)
		}
	}
}

func (m *AuthMiddleware) OptionalAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}
			
			token, err := m.extractToken(c)
			if err != nil || token == "" {
				c.Set(authconstants.ContextKeyIsAuthenticated, false)
				return next(c)
			}
			
			account, err := m.authService.ValidateSession(c.Request().Context(), token)
			if err != nil {
				c.Set(authconstants.ContextKeyIsAuthenticated, false)
				return next(c)
			}
			
			c.Set(authconstants.ContextKeyUserID, account.GetID())
			c.Set(authconstants.ContextKeyAccount, account)
			c.Set(authconstants.ContextKeyIsAuthenticated, true)
			
			return next(c)
		}
	}
}

func (m *AuthMiddleware) extractToken(c echo.Context) (string, error) {
	parts := strings.Split(m.config.TokenLookup, ":")
	if len(parts) != 2 {
		return "", echo.NewHTTPError(500, "invalid token lookup format")
	}
	
	switch parts[0] {
	case "header":
		header := c.Request().Header.Get(parts[1])
		if header == "" {
			return "", echo.NewHTTPError(401, "missing authorization header")
		}
		
		if m.config.TokenPrefix != "" {
			if !strings.HasPrefix(header, m.config.TokenPrefix) {
				return "", echo.NewHTTPError(401, "invalid authorization header format")
			}
			header = strings.TrimPrefix(header, m.config.TokenPrefix)
		}
		
		return strings.TrimSpace(header), nil
		
	case "cookie":
		cookie, err := c.Cookie(parts[1])
		if err != nil {
			return "", echo.NewHTTPError(401, "missing authentication cookie")
		}
		return cookie.Value, nil
		
	default:
		return "", echo.NewHTTPError(500, "unsupported token lookup method")
	}
}