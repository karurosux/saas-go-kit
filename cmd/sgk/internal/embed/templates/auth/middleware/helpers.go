package authmiddleware

import (
	"{{.Project.GoModule}}/internal/auth/constants"
	"{{.Project.GoModule}}/internal/auth/interface"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userID := c.Get(authconstants.ContextKeyUserID)
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

// GetAccountFromContext retrieves account from context
func GetAccountFromContext(c echo.Context) authinterface.Account {
	if account := c.Get(authconstants.ContextKeyAccount); account != nil {
		if acc, ok := account.(authinterface.Account); ok {
			return acc
		}
	}
	return nil
}

// GetSessionFromContext retrieves session from context
func GetSessionFromContext(c echo.Context) authinterface.Session {
	if session := c.Get(authconstants.ContextKeySession); session != nil {
		if sess, ok := session.(authinterface.Session); ok {
			return sess
		}
	}
	return nil
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated(c echo.Context) bool {
	if isAuth := c.Get(authconstants.ContextKeyIsAuthenticated); isAuth != nil {
		if value, ok := isAuth.(bool); ok {
			return value
		}
	}
	return false
}

// RequireEmailVerified checks if user's email is verified
func RequireEmailVerified(c echo.Context) bool {
	account := GetAccountFromContext(c)
	if account == nil {
		return false
	}
	return account.GetEmailVerified()
}

// RequirePhoneVerified checks if user's phone is verified
func RequirePhoneVerified(c echo.Context) bool {
	account := GetAccountFromContext(c)
	if account == nil {
		return false
	}
	return account.GetPhoneVerified()
}