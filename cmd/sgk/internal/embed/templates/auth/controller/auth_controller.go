package authcontroller

import (
	"errors"
	"fmt"
	"net/http"
	"time"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmiddleware "{{.Project.GoModule}}/internal/auth/middleware"
	authmodel "{{.Project.GoModule}}/internal/auth/model"
	"{{.Project.GoModule}}/internal/core"
	"github.com/labstack/echo/v4"
)

type AuthController struct {
	service authinterface.AuthService
}

func NewAuthController(service authinterface.AuthService) *AuthController {
	return &AuthController{
		service: service,
	}
}

func (ac *AuthController) handleError(c echo.Context, err error) error {
	var appErr *core.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case core.ErrCodeNotFound:
			return core.NotFound(c, err)
		case core.ErrCodeUnauthorized:
			return core.Unauthorized(c, err)
		case core.ErrCodeForbidden:
			return core.Forbidden(c, err)
		case core.ErrCodeBadRequest, core.ErrCodeValidation:
			return core.BadRequest(c, err)
		case core.ErrCodeConflict:
			return core.Error(c, http.StatusConflict, err)
		default:
			return core.InternalServerError(c, err)
		}
	}
	return core.InternalServerError(c, err)
}

func (ac *AuthController) RegisterRoutes(e *echo.Echo, basePath string, authMiddleware *authmiddleware.AuthMiddleware) {
	group := e.Group(basePath)
	
	group.POST("/register", ac.Register)
	group.POST("/login", ac.Login)
	group.POST("/refresh", ac.RefreshToken)
	group.POST("/forgot-password", ac.ForgotPassword)
	group.POST("/reset-password", ac.ResetPassword)
	group.POST("/verify-email", ac.VerifyEmail)
	
	// OAuth routes
	group.GET("/providers", ac.GetProviders)
	group.GET("/oauth/:provider", ac.OAuthLogin)
	group.GET("/oauth/:provider/callback", ac.OAuthCallback)
	
	protected := group.Group("")
	protected.Use(authMiddleware.RequireAuth())
	
	protected.POST("/logout", ac.Logout)
	protected.GET("/me", ac.GetCurrentUser)
	protected.PUT("/me", ac.UpdateProfile)
	protected.POST("/change-password", ac.ChangePassword)
	protected.POST("/resend-verification", ac.ResendVerification)
	protected.POST("/verify-phone", ac.VerifyPhone)
}

func (ac *AuthController) Register(c echo.Context) error {
	var req authmodel.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	account, err := ac.service.Register(c.Request().Context(), &req)
	if err != nil {
		return ac.handleError(c, err)
	}
	
	return core.Created(c, account)
}

func (ac *AuthController) Login(c echo.Context) error {
	var req authmodel.LoginRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	session, err := ac.service.Login(c.Request().Context(), &req)
	if err != nil {
		return ac.handleError(c, err)
	}
	
	return core.Success(c, session)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (ac *AuthController) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	session, err := ac.service.RefreshSession(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return ac.handleError(c, err)
	}
	
	return core.Success(c, session)
}

func (ac *AuthController) Logout(c echo.Context) error {
	userID, err := authmiddleware.GetUserIDFromContext(c)
	if err != nil {
		return core.Unauthorized(c, fmt.Errorf("Not authenticated"))
	}
	
	if err := ac.service.Logout(c.Request().Context(), userID); err != nil {
		return core.InternalServerError(c, fmt.Errorf("Failed to logout"))
	}
	
	return core.Success(c, map[string]string{
		"message": "Logged out successfully",
	})
}

func (ac *AuthController) GetCurrentUser(c echo.Context) error {
	account := authmiddleware.GetAccountFromContext(c)
	if account == nil {
		return core.Unauthorized(c, fmt.Errorf("Not authenticated"))
	}
	
	return core.Success(c, account)
}

type UpdateProfileRequest struct {
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone *string `json:"phone,omitempty" validate:"omitempty,e164"`
}

func (ac *AuthController) UpdateProfile(c echo.Context) error {
	userID, err := authmiddleware.GetUserIDFromContext(c)
	if err != nil {
		return core.Unauthorized(c, fmt.Errorf("Not authenticated"))
	}
	
	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	updates := authinterface.AccountUpdates{
		Email: req.Email,
		Phone: req.Phone,
	}
	
	account, err := ac.service.UpdateAccount(c.Request().Context(), userID, updates)
	if err != nil {
		return ac.handleError(c, err)
	}
	
	return core.Success(c, account)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

func (ac *AuthController) ChangePassword(c echo.Context) error {
	userID, err := authmiddleware.GetUserIDFromContext(c)
	if err != nil {
		return core.Unauthorized(c, fmt.Errorf("Not authenticated"))
	}
	
	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	if err := ac.service.ChangePassword(c.Request().Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		return ac.handleError(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Password changed successfully",
	})
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (ac *AuthController) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	if err := ac.service.SendPasswordReset(c.Request().Context(), req.Email); err != nil {
	}
	
	return core.Success(c, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

func (ac *AuthController) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	if err := ac.service.ResetPassword(c.Request().Context(), req.Token, req.NewPassword); err != nil {
		return ac.handleError(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Password reset successfully",
	})
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

func (ac *AuthController) VerifyEmail(c echo.Context) error {
	var req VerifyEmailRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	if err := ac.service.VerifyEmail(c.Request().Context(), req.Token); err != nil {
		return ac.handleError(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Email verified successfully",
	})
}

type ResendVerificationRequest struct {
	Type string `json:"type" validate:"required,oneof=email phone"`
}

func (ac *AuthController) ResendVerification(c echo.Context) error {
	userID, err := authmiddleware.GetUserIDFromContext(c)
	if err != nil {
		return core.Unauthorized(c, fmt.Errorf("Not authenticated"))
	}
	
	var req ResendVerificationRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	switch req.Type {
	case "email":
		if err := ac.service.SendEmailVerification(c.Request().Context(), userID); err != nil {
			return ac.handleError(c, err)
		}
	case "phone":
		if err := ac.service.SendPhoneVerification(c.Request().Context(), userID); err != nil {
			return ac.handleError(c, err)
		}
	}
	
	return core.Success(c, map[string]string{
		"message": "Verification sent successfully",
	})
}

type VerifyPhoneRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

func (ac *AuthController) VerifyPhone(c echo.Context) error {
	userID, err := authmiddleware.GetUserIDFromContext(c)
	if err != nil {
		return core.Unauthorized(c, fmt.Errorf("Not authenticated"))
	}
	
	var req VerifyPhoneRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	if err := ac.service.VerifyPhone(c.Request().Context(), userID, req.Code); err != nil {
		return ac.handleError(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Phone verified successfully",
	})
}

func (ac *AuthController) GetProviders(c echo.Context) error {
	providers := ac.service.GetAvailableProviders(c.Request().Context())
	
	return core.Success(c, map[string]interface{}{
		"providers": providers,
	})
}

func (ac *AuthController) OAuthLogin(c echo.Context) error {
	provider := c.Param("provider")
	state := c.QueryParam("state")
	
	if state == "" {
		state = fmt.Sprintf("%d", time.Now().Unix())
	}
	
	authURL, err := ac.service.GetOAuthURL(c.Request().Context(), provider, state)
	if err != nil {
		return ac.handleError(c, err)
	}
	
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (ac *AuthController) OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	errorParam := c.QueryParam("error")
	
	if errorParam != "" {
		errorDesc := c.QueryParam("error_description")
		return core.BadRequest(c, fmt.Errorf("OAuth error: %s - %s", errorParam, errorDesc))
	}
	
	if code == "" {
		return core.BadRequest(c, fmt.Errorf("Authorization code is required"))
	}
	
	session, err := ac.service.HandleOAuthCallback(c.Request().Context(), provider, code, state)
	if err != nil {
		return ac.handleError(c, err)
	}
	
	sessionData := map[string]interface{}{
		"access_token":  session.GetToken(),
		"refresh_token": session.GetRefreshToken(),
		"expires_at":    session.GetExpiresAt(),
		"token_type":    "Bearer",
	}
	
	return core.Success(c, sessionData)
}