package rolecontroller

import (
	"fmt"
	"net/http"
	"{{.Project.GoModule}}/internal/core"
	"strconv"
	"time"

	roleinterface "{{.Project.GoModule}}/internal/role/interface"
	rolemiddleware "{{.Project.GoModule}}/internal/role/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RoleController struct {
	service roleinterface.RoleService
}

func NewRoleController(service roleinterface.RoleService) *RoleController {
	return &RoleController{
		service: service,
	}
}

func (rc *RoleController) RegisterRoutes(e *echo.Echo, basePath string, rbacMiddleware *rolemiddleware.RBACMiddleware) {
	group := e.Group(basePath)

	roles := group.Group("/roles")
	roles.Use(rbacMiddleware.RequirePermission("roles:read"))

	roles.GET("", rc.ListRoles)
	roles.GET("/:id", rc.GetRole)

	roles.POST("", rc.CreateRole, rbacMiddleware.RequirePermission("roles:create"))
	roles.PUT("/:id", rc.UpdateRole, rbacMiddleware.RequirePermission("roles:update"))
	roles.DELETE("/:id", rc.DeleteRole, rbacMiddleware.RequirePermission("roles:delete"))

	userRoles := group.Group("/users/:userId/roles")
	userRoles.Use(rbacMiddleware.RequirePermission("users:roles:read"))

	userRoles.GET("", rc.GetUserRoles)
	userRoles.POST("/:roleId", rc.AssignRole, rbacMiddleware.RequirePermission("users:roles:assign"))
	userRoles.DELETE("/:roleId", rc.UnassignRole, rbacMiddleware.RequirePermission("users:roles:unassign"))

	permissions := group.Group("/permissions")
	permissions.Use(rbacMiddleware.InjectUserPermissions())

	permissions.GET("/my", rc.GetMyPermissions)
	permissions.POST("/check", rc.CheckPermissions)
}

func (rc *RoleController) ListRoles(c echo.Context) error {
	filters := roleinterface.RoleFilters{
		Name: c.QueryParam("name"),
	}

	if isSystem := c.QueryParam("is_system"); isSystem != "" {
		if b, err := strconv.ParseBool(isSystem); err == nil {
			filters.IsSystem = &b
		}
	}

	if limit := c.QueryParam("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters.Limit = l
		}
	}

	if offset := c.QueryParam("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filters.Offset = o
		}
	}

	roles, err := rc.service.GetRoles(c.Request().Context(), filters)
	if err != nil {
		return core.InternalServerError(c, fmt.Errorf("failed to fetch roles"))
	}

	return core.Success(c, roles)
}

func (rc *RoleController) GetRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid role ID"))
	}

	role, err := rc.service.GetRole(c.Request().Context(), id)
	if err != nil {
		return core.NotFound(c, fmt.Errorf("role not found"))
	}

	return core.Success(c, role)
}

type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=3,max=50"`
	Description string   `json:"description" validate:"max=200"`
	Permissions []string `json:"permissions" validate:"required,min=1"`
}

func (rc *RoleController) CreateRole(c echo.Context) error {
	var req CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid request body"))
	}

	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}

	role, err := rc.service.CreateRole(
		c.Request().Context(),
		req.Name,
		req.Description,
		req.Permissions,
		false, // Not a system role
	)
	if err != nil {
		return core.InternalServerError(c, fmt.Errorf("failed to create role"))
	}

	return core.Created(c, role)
}

type UpdateRoleRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=3,max=50"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=200"`
	Permissions []string `json:"permissions,omitempty" validate:"omitempty,min=1"`
}

func (rc *RoleController) UpdateRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid role ID"))
	}

	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid request body"))
	}

	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}

	updates := roleinterface.RoleUpdates{
		Name:        req.Name,
		Description: req.Description,
	}
	if len(req.Permissions) > 0 {
		updates.Permissions = &req.Permissions
	}

	role, err := rc.service.UpdateRole(c.Request().Context(), id, updates)
	if err != nil {
		return core.InternalServerError(c, fmt.Errorf("failed to update role"))
	}

	return core.Success(c, role)
}

func (rc *RoleController) DeleteRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid role ID"))
	}

	if err := rc.service.DeleteRole(c.Request().Context(), id); err != nil {
		return core.InternalServerError(c, fmt.Errorf("failed to delete role"))
	}

	return c.NoContent(http.StatusNoContent)
}

func (rc *RoleController) GetUserRoles(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid user ID"))
	}

	roles, err := rc.service.GetUserRoles(c.Request().Context(), userID)
	if err != nil {
		return core.InternalServerError(c, fmt.Errorf("failed to fetch user roles"))
	}

	return core.Success(c, roles)
}

type AssignRoleRequest struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func (rc *RoleController) AssignRole(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid user ID"))
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid role ID"))
	}

	assignedBy, err := rolemiddleware.GetUserIDFromContext(c)
	if err != nil {
		return core.Unauthorized(c, fmt.Errorf("could not determine assigner"))
	}

	var req AssignRoleRequest
	c.Bind(&req) // Optional body

	err = rc.service.AssignRoleToUser(c.Request().Context(), userID, roleID, assignedBy, req.ExpiresAt)
	if err != nil {
		return core.InternalServerError(c, fmt.Errorf("failed to assign role"))
	}

	return core.Success(c, map[string]string{
		"message": "Role assigned successfully",
	})
}

func (rc *RoleController) UnassignRole(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid user ID"))
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid role ID"))
	}

	err = rc.service.UnassignRoleFromUser(c.Request().Context(), userID, roleID)
	if err != nil {
		return core.InternalServerError(c, fmt.Errorf("failed to unassign role"))
	}

	return core.Success(c, map[string]string{
		"message": "Role unassigned successfully",
	})
}

func (rc *RoleController) GetMyPermissions(c echo.Context) error {
	permissions := rolemiddleware.GetUserPermissionsFromContext(c)
	return core.Success(c, permissions)
}

type CheckPermissionsRequest struct {
	Permissions []string `json:"permissions" validate:"required,min=1"`
	RequireAll  bool     `json:"require_all"`
}

type CheckPermissionsResponse struct {
	HasPermissions bool            `json:"has_permissions"`
	Details        map[string]bool `json:"details"`
}

func (rc *RoleController) CheckPermissions(c echo.Context) error {
	var req CheckPermissionsRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("invalid request body"))
	}

	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}

	details := make(map[string]bool)
	hasAll := true
	hasAny := false

	for _, permission := range req.Permissions {
		hasPermission := rolemiddleware.UserHasPermission(c, permission)
		details[permission] = hasPermission

		if hasPermission {
			hasAny = true
		} else {
			hasAll = false
		}
	}

	hasPermissions := hasAny
	if req.RequireAll {
		hasPermissions = hasAll
	}

	return core.Success(c, CheckPermissionsResponse{
		HasPermissions: hasPermissions,
		Details:        details,
	})
}

