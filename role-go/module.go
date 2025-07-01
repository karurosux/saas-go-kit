package role

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/karurosux/saas-go-kit/core-go"
	"github.com/karurosux/saas-go-kit/errors-go"
	"github.com/karurosux/saas-go-kit/response-go"
	"github.com/karurosux/saas-go-kit/validator-go"
)

type Module struct {
	*core.BaseModule
	service    RoleService
	rbac       *RBACMiddleware
	routePrefix string
}

type ModuleConfig struct {
	RoutePrefix string
	RBAC        *RBACMiddleware
}

// NewModule creates a new role module
func NewModule(service RoleService, config ...ModuleConfig) core.Module {
	cfg := ModuleConfig{RoutePrefix: "/api/roles"}
	if len(config) > 0 {
		if config[0].RoutePrefix != "" {
			cfg.RoutePrefix = config[0].RoutePrefix
		}
		if config[0].RBAC != nil {
			cfg.RBAC = config[0].RBAC
		}
	}

	module := &Module{
		BaseModule:  core.NewBaseModule("role"),
		service:     service,
		rbac:        cfg.RBAC,
		routePrefix: cfg.RoutePrefix,
	}

	module.setupRoutes()
	return module
}

func (m *Module) setupRoutes() {
	routes := []core.Route{
		// Role management
		{
			Method:  http.MethodPost,
			Path:    m.routePrefix,
			Handler: m.createRole,
			Name:    "create_role",
			Middlewares: m.getMiddleware("roles:create"),
		},
		{
			Method:  http.MethodGet,
			Path:    m.routePrefix,
			Handler: m.getRoles,
			Name:    "list_roles",
			Middlewares: m.getMiddleware("roles:read"),
		},
		{
			Method:  http.MethodGet,
			Path:    m.routePrefix + "/:id",
			Handler: m.getRole,
			Name:    "get_role",
			Middlewares: m.getMiddleware("roles:read"),
		},
		{
			Method:  http.MethodPut,
			Path:    m.routePrefix + "/:id",
			Handler: m.updateRole,
			Name:    "update_role",
			Middlewares: m.getMiddleware("roles:update"),
		},
		{
			Method:  http.MethodDelete,
			Path:    m.routePrefix + "/:id",
			Handler: m.deleteRole,
			Name:    "delete_role",
			Middlewares: m.getMiddleware("roles:delete"),
		},
		
		// User role assignments
		{
			Method:  http.MethodPost,
			Path:    m.routePrefix + "/:id/users",
			Handler: m.assignRoleToUser,
			Name:    "assign_role_to_user",
			Middlewares: m.getMiddleware("roles:assign"),
		},
		{
			Method:  http.MethodDelete,
			Path:    m.routePrefix + "/:id/users/:userId",
			Handler: m.unassignRoleFromUser,
			Name:    "unassign_role_from_user",
			Middlewares: m.getMiddleware("roles:assign"),
		},
		{
			Method:  http.MethodGet,
			Path:    m.routePrefix + "/:id/users",
			Handler: m.getUsersWithRole,
			Name:    "get_users_with_role",
			Middlewares: m.getMiddleware("roles:read"),
		},
		
		// User permissions - these endpoints are more permissive
		{
			Method:  http.MethodGet,
			Path:    m.routePrefix + "/users/:userId/roles",
			Handler: m.getUserRoles,
			Name:    "get_user_roles",
			Middlewares: m.getMiddleware("users:read", "roles:read"),
		},
		{
			Method:  http.MethodGet,
			Path:    m.routePrefix + "/users/:userId/permissions",
			Handler: m.getUserPermissions,
			Name:    "get_user_permissions",
			Middlewares: m.getMiddleware("users:read", "roles:read"),
		},
		{
			Method:  http.MethodPost,
			Path:    m.routePrefix + "/users/:userId/check",
			Handler: m.checkUserPermission,
			Name:    "check_user_permission",
			Middlewares: m.getMiddleware("users:read", "roles:read"),
		},
		
		// System roles - admin only
		{
			Method:  http.MethodGet,
			Path:    m.routePrefix + "/system",
			Handler: m.getSystemRoles,
			Name:    "get_system_roles",
			Middlewares: m.getMiddleware("admin:*"),
		},
	}

	m.AddRoutes(routes)
}

// getMiddleware returns middleware for the given permissions
func (m *Module) getMiddleware(permissions ...string) []echo.MiddlewareFunc {
	if m.rbac == nil {
		return []echo.MiddlewareFunc{}
	}

	if len(permissions) == 1 {
		return []echo.MiddlewareFunc{m.rbac.RequirePermission(permissions[0])}
	}

	return []echo.MiddlewareFunc{m.rbac.RequireAnyPermission(permissions...)}
}

// DTOs
type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Description string   `json:"description" validate:"max=500"`
	Permissions []string `json:"permissions" validate:"required,min=1"`
	IsSystem    bool     `json:"is_system"`
}

type UpdateRoleRequest struct {
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string   `json:"description,omitempty" validate:"omitempty,max=500"`
	Permissions *[]string `json:"permissions,omitempty" validate:"omitempty,min=1"`
}

type AssignRoleRequest struct {
	UserID    string `json:"user_id" validate:"required,uuid"`
	ExpiresAt *int64 `json:"expires_at,omitempty"`
}

type CheckPermissionRequest struct {
	Permissions []string `json:"permissions" validate:"required,min=1"`
	RequireAll  bool     `json:"require_all"`
}

// Role management handlers
func (m *Module) createRole(c echo.Context) error {
	var req CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, err)
	}

	// Validate permissions format
	utils := NewPermissionUtils()
	for _, permission := range req.Permissions {
		if !utils.IsValidPermission(permission) {
			return response.Error(c, errors.BadRequest("Invalid permission format: "+permission))
		}
	}

	role, err := m.service.CreateRole(c.Request().Context(), req.Name, req.Description, req.Permissions, req.IsSystem)
	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return response.Success(c, role)
}

func (m *Module) getRoles(c echo.Context) error {
	filters := RoleFilters{}
	
	if name := c.QueryParam("name"); name != "" {
		filters.Name = name
	}
	
	if isSystemStr := c.QueryParam("is_system"); isSystemStr != "" {
		if isSystem, err := strconv.ParseBool(isSystemStr); err == nil {
			filters.IsSystem = &isSystem
		}
	}
	
	if permission := c.QueryParam("has_permission"); permission != "" {
		filters.HasPermission = permission
	}
	
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}
	
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	roles, err := m.service.GetRoles(c.Request().Context(), filters)
	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return response.Success(c, roles)
}

func (m *Module) getRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid role ID"))
	}

	role, err := m.service.GetRole(c.Request().Context(), id)
	if err != nil {
		return response.Error(c, errors.NotFound("Role not found"))
	}

	return response.Success(c, role)
}

func (m *Module) updateRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid role ID"))
	}

	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, err)
	}

	// Validate permissions format if provided
	if req.Permissions != nil {
		utils := NewPermissionUtils()
		for _, permission := range *req.Permissions {
			if !utils.IsValidPermission(permission) {
				return response.Error(c, errors.BadRequest("Invalid permission format: "+permission))
			}
		}
	}

	updates := RoleUpdates{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	}

	role, err := m.service.UpdateRole(c.Request().Context(), id, updates)
	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return response.Success(c, role)
}

func (m *Module) deleteRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid role ID"))
	}

	if err := m.service.DeleteRole(c.Request().Context(), id); err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return c.NoContent(http.StatusNoContent)
}

// User role assignment handlers
func (m *Module) assignRoleToUser(c echo.Context) error {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid role ID"))
	}

	var req AssignRoleRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, err)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid user ID"))
	}

	// TODO: Get assigned_by from authenticated user context
	assignedBy := uuid.New() // Placeholder

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		timestamp := time.Unix(*req.ExpiresAt, 0)
		expiresAt = &timestamp
	}

	err = m.service.AssignRoleToUser(c.Request().Context(), userID, roleID, assignedBy, expiresAt)
	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return c.NoContent(http.StatusNoContent)
}

func (m *Module) unassignRoleFromUser(c echo.Context) error {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid role ID"))
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid user ID"))
	}

	err = m.service.UnassignRoleFromUser(c.Request().Context(), userID, roleID)
	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return c.NoContent(http.StatusNoContent)
}

func (m *Module) getUsersWithRole(c echo.Context) error {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid role ID"))
	}

	userRoles, err := m.service.GetUsersWithRole(c.Request().Context(), roleID)
	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return response.Success(c, userRoles)
}

// User permission handlers
func (m *Module) getUserRoles(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid user ID"))
	}

	roles, err := m.service.GetUserRoles(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return response.Success(c, roles)
}

func (m *Module) getUserPermissions(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid user ID"))
	}

	permissions, err := m.service.GetUserPermissions(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return response.Success(c, map[string]interface{}{
		"user_id":     userID,
		"permissions": permissions,
	})
}

func (m *Module) checkUserPermission(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return response.Error(c, errors.BadRequest("Invalid user ID"))
	}

	var req CheckPermissionRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, err)
	}

	var hasPermission bool
	if req.RequireAll {
		hasPermission, err = m.service.UserHasAllPermissions(c.Request().Context(), userID, req.Permissions)
	} else {
		hasPermission, err = m.service.UserHasAnyPermission(c.Request().Context(), userID, req.Permissions)
	}

	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return response.Success(c, map[string]interface{}{
		"user_id":        userID,
		"permissions":    req.Permissions,
		"require_all":    req.RequireAll,
		"has_permission": hasPermission,
	})
}

// System role handlers
func (m *Module) getSystemRoles(c echo.Context) error {
	roles, err := m.service.GetSystemRoles(c.Request().Context())
	if err != nil {
		return response.Error(c, errors.Internal(err.Error()))
	}

	return response.Success(c, roles)
}