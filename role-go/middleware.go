package role

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/karurosux/saas-go-kit/errors-go"
	"github.com/karurosux/saas-go-kit/response-go"
)

// UserIDExtractor extracts user ID from the request context
type UserIDExtractor func(c echo.Context) (uuid.UUID, error)

// MiddlewareConfig configuration for RBAC middleware
type MiddlewareConfig struct {
	UserIDExtractor UserIDExtractor
	Skipper         func(c echo.Context) bool
	ErrorHandler    func(c echo.Context, err error) error
}

// DefaultUserIDExtractor extracts user ID from "user_id" key in context
func DefaultUserIDExtractor(c echo.Context) (uuid.UUID, error) {
	userID := c.Get("user_id")
	if userID == nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	switch v := userID.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID format")
	}
}

// DefaultErrorHandler returns unauthorized error
func DefaultErrorHandler(c echo.Context, err error) error {
	return response.Error(c, errors.Unauthorized("Access denied: insufficient permissions"))
}

// RBAC middleware factory
type RBACMiddleware struct {
	service RoleService
	config  MiddlewareConfig
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(service RoleService, config ...MiddlewareConfig) *RBACMiddleware {
	cfg := MiddlewareConfig{
		UserIDExtractor: DefaultUserIDExtractor,
		ErrorHandler:    DefaultErrorHandler,
	}

	if len(config) > 0 {
		if config[0].UserIDExtractor != nil {
			cfg.UserIDExtractor = config[0].UserIDExtractor
		}
		if config[0].Skipper != nil {
			cfg.Skipper = config[0].Skipper
		}
		if config[0].ErrorHandler != nil {
			cfg.ErrorHandler = config[0].ErrorHandler
		}
	}

	return &RBACMiddleware{
		service: service,
		config:  cfg,
	}
}

// RequirePermission middleware that requires a specific permission
func (m *RBACMiddleware) RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			hasPermission, err := m.service.UserHasPermission(c.Request().Context(), userID, permission)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			if !hasPermission {
				return m.config.ErrorHandler(c, echo.NewHTTPError(http.StatusForbidden, "Missing required permission: "+permission))
			}

			return next(c)
		}
	}
}

// RequireAnyPermission middleware that requires at least one of the specified permissions
func (m *RBACMiddleware) RequireAnyPermission(permissions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			hasPermission, err := m.service.UserHasAnyPermission(c.Request().Context(), userID, permissions)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			if !hasPermission {
				return m.config.ErrorHandler(c, echo.NewHTTPError(http.StatusForbidden, "Missing required permissions: "+strings.Join(permissions, ", ")))
			}

			return next(c)
		}
	}
}

// RequireAllPermissions middleware that requires all specified permissions
func (m *RBACMiddleware) RequireAllPermissions(permissions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			hasPermission, err := m.service.UserHasAllPermissions(c.Request().Context(), userID, permissions)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			if !hasPermission {
				return m.config.ErrorHandler(c, echo.NewHTTPError(http.StatusForbidden, "Missing required permissions: "+strings.Join(permissions, ", ")))
			}

			return next(c)
		}
	}
}

// RequireRole middleware that requires a specific role
func (m *RBACMiddleware) RequireRole(roleName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			roles, err := m.service.GetUserRoles(c.Request().Context(), userID)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			for _, role := range roles {
				if role.GetName() == roleName {
					return next(c)
				}
			}

			return m.config.ErrorHandler(c, echo.NewHTTPError(http.StatusForbidden, "Missing required role: "+roleName))
		}
	}
}

// RequireAnyRole middleware that requires at least one of the specified roles
func (m *RBACMiddleware) RequireAnyRole(roleNames ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			roles, err := m.service.GetUserRoles(c.Request().Context(), userID)
			if err != nil {
				return m.config.ErrorHandler(c, err)
			}

			userRoleNames := make(map[string]bool)
			for _, role := range roles {
				userRoleNames[role.GetName()] = true
			}

			for _, roleName := range roleNames {
				if userRoleNames[roleName] {
					return next(c)
				}
			}

			return m.config.ErrorHandler(c, echo.NewHTTPError(http.StatusForbidden, "Missing required roles: "+strings.Join(roleNames, ", ")))
		}
	}
}

// CheckPermission checks if user has permission and adds result to context
func (m *RBACMiddleware) CheckPermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				c.Set("has_permission_"+permission, false)
				return next(c)
			}

			hasPermission, err := m.service.UserHasPermission(c.Request().Context(), userID, permission)
			if err != nil {
				c.Set("has_permission_"+permission, false)
				return next(c)
			}

			c.Set("has_permission_"+permission, hasPermission)
			return next(c)
		}
	}
}

// InjectUserPermissions middleware that injects all user permissions into context
func (m *RBACMiddleware) InjectUserPermissions() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				c.Set("user_permissions", []string{})
				return next(c)
			}

			permissions, err := m.service.GetUserPermissions(c.Request().Context(), userID)
			if err != nil {
				c.Set("user_permissions", []string{})
				return next(c)
			}

			c.Set("user_permissions", permissions)
			return next(c)
		}
	}
}

// Helper functions for use in handlers

// HasPermissionInContext checks if user has permission from context
func HasPermissionInContext(c echo.Context, permission string) bool {
	if hasPermission := c.Get("has_permission_" + permission); hasPermission != nil {
		if value, ok := hasPermission.(bool); ok {
			return value
		}
	}
	return false
}

// GetUserPermissionsFromContext gets user permissions from context
func GetUserPermissionsFromContext(c echo.Context) []string {
	if permissions := c.Get("user_permissions"); permissions != nil {
		if value, ok := permissions.([]string); ok {
			return value
		}
	}
	return []string{}
}

// GetUserIDFromContext gets user ID from context (helper for handlers)
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	return DefaultUserIDExtractor(c)
}