package teammiddleware

import (
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/team/constants"
	"{{.Project.GoModule}}/internal/team/interface"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TeamMiddleware provides team-related middleware
type TeamMiddleware struct {
	teamService teaminterface.TeamService
}

// NewTeamMiddleware creates a new team middleware
func NewTeamMiddleware(teamService teaminterface.TeamService) *TeamMiddleware {
	return &TeamMiddleware{
		teamService: teamService,
	}
}

// RequireTeamMember middleware that requires user to be a team member
func (m *TeamMiddleware) RequireTeamMember() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := GetUserIDFromContext(c)
			if err != nil {
				return core.Error(c, core.Unauthorized("user not authenticated"))
			}
			
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return core.Error(c, core.BadRequest("account ID not provided"))
			}
			
			// Check if user is a team member
			member, err := m.teamService.GetMemberRole(c.Request().Context(), accountID, userID)
			if err != nil {
				return core.Error(c, core.Forbidden("not a team member"))
			}
			
			// Store member role in context
			c.Set(teamconstants.ContextKeyMemberRole, member)
			
			return next(c)
		}
	}
}

// RequireRole middleware that requires a specific role or higher
func (m *TeamMiddleware) RequireRole(minRole teaminterface.MemberRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := GetUserIDFromContext(c)
			if err != nil {
				return core.Error(c, core.Unauthorized("user not authenticated"))
			}
			
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return core.Error(c, core.BadRequest("account ID not provided"))
			}
			
			// Get user's role
			role, err := m.teamService.GetMemberRole(c.Request().Context(), accountID, userID)
			if err != nil {
				return core.Error(c, core.Forbidden("not a team member"))
			}
			
			// Check role hierarchy
			if !hasMinimumRole(role, minRole) {
				return core.Error(c, core.Forbidden(teamconstants.ErrInsufficientPermission))
			}
			
			// Store member role in context
			c.Set(teamconstants.ContextKeyMemberRole, role)
			
			return next(c)
		}
	}
}

// RequireOwner middleware that requires owner role
func (m *TeamMiddleware) RequireOwner() echo.MiddlewareFunc {
	return m.RequireRole(teaminterface.RoleOwner)
}

// RequireAdmin middleware that requires admin role or higher
func (m *TeamMiddleware) RequireAdmin() echo.MiddlewareFunc {
	return m.RequireRole(teaminterface.RoleAdmin)
}

// RequirePermission middleware that requires a specific permission
func (m *TeamMiddleware) RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := GetUserIDFromContext(c)
			if err != nil {
				return core.Error(c, core.Unauthorized("user not authenticated"))
			}
			
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return core.Error(c, core.BadRequest("account ID not provided"))
			}
			
			// Check permission
			hasPermission, err := m.teamService.CheckPermission(
				c.Request().Context(),
				accountID,
				userID,
				permission,
			)
			if err != nil || !hasPermission {
				return core.Error(c, core.Forbidden(teamconstants.ErrInsufficientPermission))
			}
			
			return next(c)
		}
	}
}

// InjectTeamMember middleware that injects team member info into context
func (m *TeamMiddleware) InjectTeamMember() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := GetUserIDFromContext(c)
			if err != nil {
				// Not authenticated, continue without injecting
				return next(c)
			}
			
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				// No account context, continue without injecting
				return next(c)
			}
			
			// Try to get member role
			role, _ := m.teamService.GetMemberRole(c.Request().Context(), accountID, userID)
			if role != "" {
				c.Set(teamconstants.ContextKeyMemberRole, role)
			}
			
			return next(c)
		}
	}
}

// hasMinimumRole checks if a role meets the minimum requirement
func hasMinimumRole(userRole, minRole teaminterface.MemberRole) bool {
	roleHierarchy := map[teaminterface.MemberRole]int{
		teaminterface.RoleOwner:  4,
		teaminterface.RoleAdmin:  3,
		teaminterface.RoleMember: 2,
		teaminterface.RoleViewer: 1,
	}
	
	userLevel, ok1 := roleHierarchy[userRole]
	minLevel, ok2 := roleHierarchy[minRole]
	
	if !ok1 || !ok2 {
		return false
	}
	
	return userLevel >= minLevel
}