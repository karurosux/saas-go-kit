package rolemiddleware

import (
	roleconstants "{{.Project.GoModule}}/internal/role/constants"
	roleinterface "{{.Project.GoModule}}/internal/role/interface"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func GetUserPermissionsFromContext(c echo.Context) []string {
	if permissions := c.Get(roleconstants.ContextKeyUserPermissions); permissions != nil {
		if value, ok := permissions.([]string); ok {
			return value
		}
	}
	return []string{}
}

func GetUserRolesFromContext(c echo.Context) []roleinterface.Role {
	if roles := c.Get(roleconstants.ContextKeyUserRoles); roles != nil {
		if value, ok := roles.([]roleinterface.Role); ok {
			return value
		}
	}
	return []roleinterface.Role{}
}

func HasPermissionInContext(c echo.Context, permission string) bool {
	key := roleconstants.ContextKeyHasPermissionPrefix + permission
	if hasPermission := c.Get(key); hasPermission != nil {
		if value, ok := hasPermission.(bool); ok {
			return value
		}
	}
	return false
}

func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	return DefaultUserIDExtractor(c)
}

func UserHasRole(c echo.Context, roleName string) bool {
	roles := GetUserRolesFromContext(c)
	for _, role := range roles {
		if role.GetName() == roleName {
			return true
		}
	}
	return false
}

func UserHasAnyRole(c echo.Context, roleNames ...string) bool {
	roles := GetUserRolesFromContext(c)
	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role.GetName()] = true
	}
	
	for _, roleName := range roleNames {
		if roleMap[roleName] {
			return true
		}
	}
	return false
}

func UserHasPermission(c echo.Context, permission string) bool {
	permissions := GetUserPermissionsFromContext(c)
	for _, p := range permissions {
		if p == permission || p == roleconstants.PermissionAll {
			return true
		}
		if len(p) > 2 && p[len(p)-2:] == ":*" {
			prefix := p[:len(p)-1]
			if len(permission) >= len(prefix) && permission[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

func UserHasAnyPermission(c echo.Context, permissions ...string) bool {
	for _, permission := range permissions {
		if UserHasPermission(c, permission) {
			return true
		}
	}
	return false
}

func UserHasAllPermissions(c echo.Context, permissions ...string) bool {
	for _, permission := range permissions {
		if !UserHasPermission(c, permission) {
			return false
		}
	}
	return true
}