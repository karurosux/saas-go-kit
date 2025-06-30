package auth

import (
	"github.com/labstack/echo/v4"
	"github.com/karurosux/saas-go-kit/errors-go"
	"github.com/karurosux/saas-go-kit/response-go"
	"github.com/karurosux/saas-go-kit/validator-go"
)

// Handlers provides HTTP handlers for auth operations
type Handlers struct {
	service AuthService
}

// NewHandlers creates new auth handlers
func NewHandlers(service AuthService) *Handlers {
	return &Handlers{
		service: service,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email       string                 `json:"email" validate:"required,email"`
	Password    string                 `json:"password" validate:"required,min=8,max=128"`
	CompanyName string                 `json:"company_name,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token   string                 `json:"token"`
	Account map[string]interface{} `json:"account"`
}

// Register handles account registration
func (h *Handlers) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, err)
	}

	// Build metadata
	metadata := req.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	if req.CompanyName != "" {
		metadata["company_name"] = req.CompanyName
	}
	if req.Phone != "" {
		metadata["phone"] = req.Phone
	}

	// Register account
	account, err := h.service.Register(c.Request().Context(), req.Email, req.Password, metadata)
	if err != nil {
		return response.Error(c, err)
	}

	// Generate token for immediate login
	token, err := h.service.GenerateToken(account)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, LoginResponse{
		Token:   token,
		Account: accountToMap(account),
	})
}

// Login handles account login
func (h *Handlers) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, err)
	}

	// Login
	token, account, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, LoginResponse{
		Token:   token,
		Account: accountToMap(account),
	})
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(c echo.Context) error {
	oldToken := extractToken(c)
	if oldToken == "" {
		return response.Error(c, errors.ErrUnauthorized)
	}

	// Refresh token
	newToken, err := h.service.RefreshToken(c.Request().Context(), oldToken)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"token": newToken,
	})
}

// Logout handles account logout
func (h *Handlers) Logout(c echo.Context) error {
	account := GetAccount(c)
	if account == nil {
		return response.Error(c, errors.ErrUnauthorized)
	}

	// Logout
	if err := h.service.Logout(c.Request().Context(), account.GetID()); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Logged out successfully",
	})
}

// GetProfile returns the current account profile
func (h *Handlers) GetProfile(c echo.Context) error {
	account := GetAccount(c)
	if account == nil {
		return response.Error(c, errors.ErrUnauthorized)
	}

	return response.Success(c, accountToMap(account))
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	CompanyName string `json:"company_name,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

// UpdateProfile updates the account profile
func (h *Handlers) UpdateProfile(c echo.Context) error {
	account := GetAccount(c)
	if account == nil {
		return response.Error(c, errors.ErrUnauthorized)
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	// Build updates
	updates := make(map[string]interface{})
	if req.CompanyName != "" {
		updates["company_name"] = req.CompanyName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	// Update account
	updated, err := h.service.UpdateAccount(c.Request().Context(), account.GetID(), updates)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, accountToMap(updated))
}

// VerifyEmail verifies an email address
func (h *Handlers) VerifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return response.Error(c, errors.BadRequest("Token is required"))
	}

	// Verify email
	if err := h.service.VerifyEmail(c.Request().Context(), token); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Email verified successfully",
	})
}

// ResendVerification resends verification email
func (h *Handlers) ResendVerification(c echo.Context) error {
	account := GetAccount(c)
	if account == nil {
		return response.Error(c, errors.ErrUnauthorized)
	}

	// Send verification email
	if err := h.service.SendVerificationEmail(c.Request().Context(), account.GetID()); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Verification email sent",
	})
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPassword handles password reset request
func (h *Handlers) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, err)
	}

	// Request password reset
	if err := h.service.RequestPasswordReset(c.Request().Context(), req.Email); err != nil {
		// Don't reveal if email exists
		return response.Success(c, map[string]string{
			"message": "If an account exists with this email, a password reset link has been sent",
		})
	}

	return response.Success(c, map[string]string{
		"message": "If an account exists with this email, a password reset link has been sent",
	})
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// ResetPassword handles password reset
func (h *Handlers) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, err)
	}

	// Reset password
	if err := h.service.ResetPassword(c.Request().Context(), req.Token, req.NewPassword); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Password reset successfully",
	})
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// ChangePassword handles password change
func (h *Handlers) ChangePassword(c echo.Context) error {
	account := GetAccount(c)
	if account == nil {
		return response.Error(c, errors.ErrUnauthorized)
	}

	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, errors.BadRequest("Invalid request format"))
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, err)
	}

	// Change password
	if err := h.service.ChangePassword(c.Request().Context(), account.GetID(), req.OldPassword, req.NewPassword); err != nil {
		return response.Error(c, err)
	}

	return response.Success(c, map[string]string{
		"message": "Password changed successfully",
	})
}

// Helper function to convert account to map
func accountToMap(account Account) map[string]interface{} {
	// Basic fields that all accounts should have
	result := map[string]interface{}{
		"id":              account.GetID().String(),
		"email":           account.GetEmail(),
		"is_active":       account.IsActive(),
		"email_verified":  account.IsEmailVerified(),
	}

	// Add metadata
	for k, v := range account.GetMetadata() {
		result[k] = v
	}

	// If it's a DefaultAccount, add more fields
	if defAccount, ok := account.(*DefaultAccount); ok {
		result["company_name"] = defAccount.CompanyName
		result["phone"] = defAccount.Phone
		result["created_at"] = defAccount.CreatedAt
		result["updated_at"] = defAccount.UpdatedAt
		
		if defAccount.EmailVerifiedAt != nil {
			result["email_verified_at"] = defAccount.EmailVerifiedAt
		}
	}

	return result
}