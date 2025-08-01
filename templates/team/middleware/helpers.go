package teammiddleware

import (
	"{{.Project.GoModule}}/internal/team/constants"
	"{{.Project.GoModule}}/internal/team/interface"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// GetUserIDFromContext gets user ID from context
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	// Try to get from auth context first
	userID := c.Get("user_id")
	if userID == nil {
		return uuid.Nil, echo.NewHTTPError(401, "user not authenticated")
	}
	
	switch v := userID.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, echo.NewHTTPError(401, "invalid user ID format")
	}
}

// GetAccountIDFromContext gets account ID from context
func GetAccountIDFromContext(c echo.Context) (uuid.UUID, error) {
	// Try from context
	if accountID := c.Get(teamconstants.ContextKeyAccountID); accountID != nil {
		switch v := accountID.(type) {
		case string:
			return uuid.Parse(v)
		case uuid.UUID:
			return v, nil
		}
	}
	
	// Try from path parameter
	if accountIDStr := c.Param("accountId"); accountIDStr != "" {
		return uuid.Parse(accountIDStr)
	}
	
	// Try from query parameter
	if accountIDStr := c.QueryParam("account_id"); accountIDStr != "" {
		return uuid.Parse(accountIDStr)
	}
	
	return uuid.Nil, echo.NewHTTPError(400, "account ID not provided")
}

// GetMemberRoleFromContext gets member role from context
func GetMemberRoleFromContext(c echo.Context) teaminterface.MemberRole {
	if role := c.Get(teamconstants.ContextKeyMemberRole); role != nil {
		if r, ok := role.(teaminterface.MemberRole); ok {
			return r
		}
	}
	return ""
}

// GetTeamMemberFromContext gets team member from context
func GetTeamMemberFromContext(c echo.Context) teaminterface.TeamMember {
	if member := c.Get(teamconstants.ContextKeyTeamMember); member != nil {
		if m, ok := member.(teaminterface.TeamMember); ok {
			return m
		}
	}
	return nil
}

// IsTeamOwner checks if current user is team owner
func IsTeamOwner(c echo.Context) bool {
	role := GetMemberRoleFromContext(c)
	return role == teaminterface.RoleOwner
}

// IsTeamAdmin checks if current user is team admin or owner
func IsTeamAdmin(c echo.Context) bool {
	role := GetMemberRoleFromContext(c)
	return role == teaminterface.RoleOwner || role == teaminterface.RoleAdmin
}

// CanManageTeam checks if current user can manage team
func CanManageTeam(c echo.Context) bool {
	return IsTeamAdmin(c)
}

// SetAccountIDInContext sets account ID in context
func SetAccountIDInContext(c echo.Context, accountID uuid.UUID) {
	c.Set(teamconstants.ContextKeyAccountID, accountID)
}