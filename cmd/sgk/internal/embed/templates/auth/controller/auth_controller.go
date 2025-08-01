package authcontroller

import (
	"errors"
	"fmt"
	"net/http"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmiddleware "{{.Project.GoModule}}/internal/auth/middleware"
	authmodel "{{.Project.GoModule}}/internal/auth/model"
	"{{.Project.GoModule}}/internal/core"
	"github.com/labstack/echo/v4"
)

// AuthController handles authentication requests
type AuthController struct {
	service authinterface.AuthService
}

// NewAuthController creates a new auth controller
func NewAuthController(service authinterface.AuthService) *AuthController {
	return &AuthController{
		service: service,
	}
}

// handleError converts service errors to appropriate HTTP responses
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
	// Default to internal server error for unknown errors
	return core.InternalServerError(c, err)
}

// RegisterRoutes registers all auth-related routes
func (ac *AuthController) RegisterRoutes(e *echo.Echo, basePath string, authMiddleware *authmiddleware.AuthMiddleware) {
	group := e.Group(basePath)
	
	// Public endpoints
	group.POST("/register", ac.Register)
	group.POST("/login", ac.Login)
	group.POST("/refresh", ac.RefreshToken)
	group.POST("/forgot-password", ac.ForgotPassword)
	group.POST("/reset-password", ac.ResetPassword)
	group.POST("/verify-email", ac.VerifyEmail)
	
	// Protected endpoints
	protected := group.Group("")
	protected.Use(authMiddleware.RequireAuth())
	
	protected.POST("/logout", ac.Logout)
	protected.GET("/me", ac.GetCurrentUser)
	protected.PUT("/me", ac.UpdateProfile)
	protected.POST("/change-password", ac.ChangePassword)
	protected.POST("/resend-verification", ac.ResendVerification)
	protected.POST("/verify-phone", ac.VerifyPhone)
}

// Register godoc
// @Summary Register a new account
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authmodel.RegisterRequest true "Registration details"
// @Success 201 {object} authmodel.Account
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/register [post]
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

// Login godoc
// @Summary Login to an account
// @Description Authenticate and receive session tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authmodel.LoginRequest true "Login credentials"
// @Success 200 {object} authmodel.Session
// @Failure 400 {object} core.ErrorResponse
// @Failure 401 {object} core.ErrorResponse
// @Failure 403 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/login [post]
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

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} authmodel.Session
// @Failure 400 {object} core.ErrorResponse
// @Failure 401 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/refresh [post]
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

// Logout godoc
// @Summary Logout from account
// @Description Invalidate current session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/logout [post]
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

// GetCurrentUser godoc
// @Summary Get current user
// @Description Get authenticated user's account details
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} authmodel.Account
// @Failure 401 {object} core.ErrorResponse
// @Router /auth/me [get]
func (ac *AuthController) GetCurrentUser(c echo.Context) error {
	account := authmiddleware.GetAccountFromContext(c)
	if account == nil {
		return core.Unauthorized(c, fmt.Errorf("Not authenticated"))
	}
	
	return core.Success(c, account)
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone *string `json:"phone,omitempty" validate:"omitempty,e164"`
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile
// @Tags auth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile updates"
// @Success 200 {object} authmodel.Account
// @Failure 400 {object} core.ErrorResponse
// @Failure 401 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/me [put]
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

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ChangePassword godoc
// @Summary Change password
// @Description Change authenticated user's password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 401 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/change-password [post]
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

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Send password reset email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Email address"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/forgot-password [post]
func (ac *AuthController) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return core.BadRequest(c, fmt.Errorf("Invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.BadRequest(c, err)
	}
	
	if err := ac.service.SendPasswordReset(c.Request().Context(), req.Email); err != nil {
		// Don't reveal if email exists
		// Log the actual error but return generic message
	}
	
	return core.Success(c, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password using token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/reset-password [post]
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

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyEmail godoc
// @Summary Verify email
// @Description Verify email address using token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Verification token"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/verify-email [post]
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

// ResendVerificationRequest represents a resend verification request
type ResendVerificationRequest struct {
	Type string `json:"type" validate:"required,oneof=email phone"`
}

// ResendVerification godoc
// @Summary Resend verification
// @Description Resend email or phone verification
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResendVerificationRequest true "Verification type"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 401 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/resend-verification [post]
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

// VerifyPhoneRequest represents a phone verification request
type VerifyPhoneRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// VerifyPhone godoc
// @Summary Verify phone
// @Description Verify phone number using code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyPhoneRequest true "Verification code"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 401 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /auth/verify-phone [post]
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