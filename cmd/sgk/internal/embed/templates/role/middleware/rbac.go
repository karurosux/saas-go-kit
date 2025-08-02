package rolemiddleware

import (
	"fmt"
	"net/http"
	"strings"

	"{{.Project.GoModule}}/internal/core"
	roleconstants "{{.Project.GoModule}}/internal/role/constants"
	roleinterface "{{.Project.GoModule}}/internal/role/interface"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserIDExtractor func(c echo.Context) (uuid.UUID, error)

type MiddlewareConfig struct {
	UserIDExtractor UserIDExtractor
	Skipper         func(c echo.Context) bool
	ErrorHandler    func(c echo.Context, err error) error
}

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

func DefaultErrorHandler(c echo.Context, err error) error {
	return core.Unauthorized(c, fmt.Errorf("Access denied: insufficient permissions"))
}

type RBACMiddleware struct {
	service roleinterface.RoleService
	config  MiddlewareConfig
}

func NewRBACMiddleware(service roleinterface.RoleService, config ...MiddlewareConfig) *RBACMiddleware {
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
				return m.config.ErrorHandler(c, echo.NewHTTPError(http.StatusForbidden, "Missing all required permissions: "+strings.Join(permissions, ", ")))
			}

			return next(c)
		}
	}
}

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

func (m *RBACMiddleware) InjectUserPermissions() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				c.Set(roleconstants.ContextKeyUserPermissions, []string{})
				return next(c)
			}

			permissions, err := m.service.GetUserPermissions(c.Request().Context(), userID)
			if err != nil {
				c.Set(roleconstants.ContextKeyUserPermissions, []string{})
				return next(c)
			}

			c.Set(roleconstants.ContextKeyUserPermissions, permissions)
			return next(c)
		}
	}
}

func (m *RBACMiddleware) InjectUserRoles() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				c.Set(roleconstants.ContextKeyUserRoles, []roleinterface.Role{})
				return next(c)
			}

			roles, err := m.service.GetUserRoles(c.Request().Context(), userID)
			if err != nil {
				c.Set(roleconstants.ContextKeyUserRoles, []roleinterface.Role{})
				return next(c)
			}

			c.Set(roleconstants.ContextKeyUserRoles, roles)
			return next(c)
		}
	}
}

func (m *RBACMiddleware) CheckPermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if m.config.Skipper != nil && m.config.Skipper(c) {
				return next(c)
			}

			userID, err := m.config.UserIDExtractor(c)
			if err != nil {
				c.Set(roleconstants.ContextKeyHasPermissionPrefix+permission, false)
				return next(c)
			}

			hasPermission, err := m.service.UserHasPermission(c.Request().Context(), userID, permission)
			if err != nil {
				c.Set(roleconstants.ContextKeyHasPermissionPrefix+permission, false)
				return next(c)
			}

			c.Set(roleconstants.ContextKeyHasPermissionPrefix+permission, hasPermission)
			return next(c)
		}
	}
}