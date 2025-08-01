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

// ListRoles godoc
// @Summary List all roles
// @Description Get a list of all roles with optional filtering
// @Tags roles
// @Accept json
// @Produce json
// @Param name query string false "Filter by role name"
// @Param is_system query bool false "Filter by system roles"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {array} roleinterface.Role
// @Failure 500 {object} core.ErrorResponse
// @Router /roles [get]
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

// GetRole godoc
// @Summary Get a role by ID
// @Description Get detailed information about a specific role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} roleinterface.Role
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /roles/{id} [get]
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

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with specified permissions
// @Tags roles
// @Accept json
// @Produce json
// @Param role body CreateRoleRequest true "Role details"
// @Success 201 {object} roleinterface.Role
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /roles [post]
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

// UpdateRole godoc
// @Summary Update a role
// @Description Update an existing role's details
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body UpdateRoleRequest true "Role updates"
// @Success 200 {object} roleinterface.Role
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /roles/{id} [put]
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

// DeleteRole godoc
// @Summary Delete a role
// @Description Delete an existing role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 204 {object} nil
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /roles/{id} [delete]
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

// GetUserRoles godoc
// @Summary Get user's roles
// @Description Get all roles assigned to a user
// @Tags user-roles
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {array} roleinterface.Role
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /users/{userId}/roles [get]
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

// AssignRole godoc
// @Summary Assign role to user
// @Description Assign a role to a user with optional expiration
// @Tags user-roles
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param roleId path string true "Role ID"
// @Param request body AssignRoleRequest false "Assignment options"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /users/{userId}/roles/{roleId} [post]
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

// UnassignRole godoc
// @Summary Remove role from user
// @Description Remove a role assignment from a user
// @Tags user-roles
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param roleId path string true "Role ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /users/{userId}/roles/{roleId} [delete]
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

// GetMyPermissions godoc
// @Summary Get current user's permissions
// @Description Get all permissions for the authenticated user
// @Tags permissions
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} core.ErrorResponse
// @Router /permissions/my [get]
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

// CheckPermissions godoc
// @Summary Check if user has permissions
// @Description Check if the current user has the specified permissions
// @Tags permissions
// @Accept json
// @Produce json
// @Param request body CheckPermissionsRequest true "Permissions to check"
// @Success 200 {object} CheckPermissionsResponse
// @Failure 400 {object} core.ErrorResponse
// @Router /permissions/check [post]
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

