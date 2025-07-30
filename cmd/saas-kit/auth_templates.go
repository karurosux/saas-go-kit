package main

// authTemplates contains all template files for the auth module
var authTemplates = map[string]string{
	"module.go": authModuleTemplate,
	"handlers.go": authHandlersTemplate,
	"service.go": authServiceTemplate,
	"models.go": authModelsTemplate,
	"interfaces.go": authInterfacesTemplate,
	"repositories/gorm/account_repository.go": authAccountRepoTemplate,
	"repositories/gorm/token_repository.go": authTokenRepoTemplate,
	"repositories/gorm/migrations.go": authMigrationsTemplate,
}

var authModuleTemplate = `package {{.Module.Name}}

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"{{.Project.GoModule}}/internal/core"
)

// Module provides auth module for {{.Project.Name}}
type Module struct {
	*core.BaseModule
	service  AuthService
	config   ModuleConfig
	handlers *Handlers
}

// ModuleConfig holds module configuration
type ModuleConfig struct {
	Service            AuthService
	RoutePrefix        string
	RequireVerified    bool
	RateLimiter        echo.MiddlewareFunc
}

// NewModule creates a new auth module
func NewModule(config ModuleConfig) *Module {
	if config.RoutePrefix == "" {
		config.RoutePrefix = "{{.Module.RoutePrefix}}"
	}

	module := &Module{
		BaseModule: core.NewBaseModule("{{.Module.Name}}"),
		service:    config.Service,
		config:     config,
		handlers:   NewHandlers(config.Service),
	}

	// Register routes
	module.registerRoutes()

	return module
}

// registerRoutes registers all auth routes
func (m *Module) registerRoutes() {
	// Public routes
	publicRoutes := []core.Route{
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/register",
			Handler:     m.handlers.Register,
			Name:        "auth.register",
			Description: "Register a new account",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/login",
			Handler:     m.handlers.Login,
			Name:        "auth.login",
			Description: "Login to an account",
		},
		{
			Method:      "GET",
			Path:        m.config.RoutePrefix + "/verify-email",
			Handler:     m.handlers.VerifyEmail,
			Name:        "auth.verify-email",
			Description: "Verify email address",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/forgot-password",
			Handler:     m.handlers.ForgotPassword,
			Name:        "auth.forgot-password",
			Description: "Request password reset",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/reset-password",
			Handler:     m.handlers.ResetPassword,
			Name:        "auth.reset-password",
			Description: "Reset password with token",
		},
	}

	m.AddRoutes(publicRoutes)

	// Protected routes
	protectedRoutes := []core.Route{
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/refresh",
			Handler:     m.handlers.RefreshToken,
			Name:        "auth.refresh",
			Description: "Refresh JWT token",
		},
		{
			Method:      "POST",
			Path:        m.config.RoutePrefix + "/logout",
			Handler:     m.handlers.Logout,
			Name:        "auth.logout",
			Description: "Logout from account",
		},
		{
			Method:      "GET",
			Path:        m.config.RoutePrefix + "/me",
			Handler:     m.handlers.GetProfile,
			Name:        "auth.me",
			Description: "Get current account profile",
		},
	}

	// Add auth middleware to protected routes
	authMiddleware := m.RequireAuth()
	for i := range protectedRoutes {
		protectedRoutes[i].Middlewares = []echo.MiddlewareFunc{authMiddleware}
	}
	m.AddRoutes(protectedRoutes)
}

// RequireAuth returns middleware that requires authentication
func (m *Module) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c)
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing token")
			}

			claims, err := m.service.ValidateToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			// Get account
			account, err := m.service.GetAccount(c.Request().Context(), claims.AccountID)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Account not found")
			}

			// Check if email verification is required
			if m.config.RequireVerified && !account.IsEmailVerified() {
				return echo.NewHTTPError(http.StatusForbidden, "Email verification required")
			}

			// Store in context
			c.Set("account", account)
			c.Set("account_id", claims.AccountID.String())
			c.Set("claims", claims)

			return next(c)
		}
	}
}

// Helper functions

// extractToken extracts the JWT token from the request
func extractToken(c echo.Context) string {
	// Check Authorization header
	auth := c.Request().Header.Get("Authorization")
	if auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}

	// Check cookie
	cookie, err := c.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// GetAccount retrieves the authenticated account from context
func GetAccount(c echo.Context) Account {
	if account, ok := c.Get("account").(Account); ok {
		return account
	}
	return nil
}
`

var authHandlersTemplate = `package {{.Module.Name}}

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"{{.Project.GoModule}}/internal/core"
)

// Handlers handles HTTP requests for auth
type Handlers struct {
	service AuthService
}

// NewHandlers creates new auth handlers
func NewHandlers(service AuthService) *Handlers {
	return &Handlers{
		service: service,
	}
}

// Register handles account registration
func (h *Handlers) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	account, err := h.service.Register(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return core.SuccessResponse(c, map[string]interface{}{
		"account": account,
		"message": "Account created successfully. Please verify your email.",
	})
}

// Login handles account login
func (h *Handlers) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	result, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return core.ErrorResponse(c, http.StatusUnauthorized, err.Error())
	}

	return core.SuccessResponse(c, result)
}

// VerifyEmail handles email verification
func (h *Handlers) VerifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return core.ErrorResponse(c, http.StatusBadRequest, "Token is required")
	}

	err := h.service.VerifyEmail(c.Request().Context(), token)
	if err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return core.SuccessResponse(c, map[string]interface{}{
		"message": "Email verified successfully",
	})
}

// ForgotPassword handles password reset requests
func (h *Handlers) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	err := h.service.RequestPasswordReset(c.Request().Context(), req.Email)
	if err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return core.SuccessResponse(c, map[string]interface{}{
		"message": "Password reset email sent",
	})
}

// ResetPassword handles password reset
func (h *Handlers) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	err := h.service.ResetPassword(c.Request().Context(), req.Token, req.Password)
	if err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return core.SuccessResponse(c, map[string]interface{}{
		"message": "Password reset successfully",
	})
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return core.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	result, err := h.service.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return core.ErrorResponse(c, http.StatusUnauthorized, err.Error())
	}

	return core.SuccessResponse(c, result)
}

// Logout handles account logout
func (h *Handlers) Logout(c echo.Context) error {
	claims := GetClaims(c)
	if claims == nil {
		return core.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
	}

	err := h.service.Logout(c.Request().Context(), claims.AccountID)
	if err != nil {
		return core.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return core.SuccessResponse(c, map[string]interface{}{
		"message": "Logged out successfully",
	})
}

// GetProfile handles getting account profile
func (h *Handlers) GetProfile(c echo.Context) error {
	account := GetAccount(c)
	if account == nil {
		return core.ErrorResponse(c, http.StatusUnauthorized, "Account not found")
	}

	return core.SuccessResponse(c, account)
}

// GetClaims retrieves the JWT claims from context
func GetClaims(c echo.Context) *Claims {
	if claims, ok := c.Get("claims").(*Claims); ok {
		return claims
	}
	return nil
}
`

var authServiceTemplate = `package {{.Module.Name}}

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service implements AuthService
type Service struct {
	accountRepo AccountRepository
	tokenRepo   TokenRepository
	jwtSecret   string
}

// NewService creates a new auth service
func NewService(accountRepo AccountRepository, tokenRepo TokenRepository, jwtSecret string) *Service {
	return &Service{
		accountRepo: accountRepo,
		tokenRepo:   tokenRepo,
		jwtSecret:   jwtSecret,
	}
}

// Register creates a new account
func (s *Service) Register(ctx context.Context, email, password string) (Account, error) {
	// Check if account already exists
	existing, err := s.accountRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("account with email %s already exists", email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create account
	account := &AccountModel{
		ID:                uuid.New(),
		Email:            email,
		HashedPassword:   string(hashedPassword),
		IsEmailVerified:  false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %v", err)
	}

	// Generate verification token
	verificationToken, err := s.generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %v", err)
	}

	tokenModel := &TokenModel{
		ID:        uuid.New(),
		AccountID: account.ID,
		Token:     verificationToken,
		Type:      TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, tokenModel); err != nil {
		return nil, fmt.Errorf("failed to create verification token: %v", err)
	}

	// TODO: Send verification email

	return account, nil
}

// Login authenticates an account
func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	account, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(account.GetHashedPassword()), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateJWT(account.GetID())
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Generate refresh token
	refreshToken, err := s.generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	// Store refresh token
	tokenModel := &TokenModel{
		ID:        uuid.New(),
		AccountID: account.GetID(),
		Token:     refreshToken,
		Type:      TokenTypeRefresh,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, tokenModel); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %v", err)
	}

	return &LoginResult{
		Account:      account,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}, nil
}

// ValidateToken validates a JWT token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetAccount retrieves an account by ID
func (s *Service) GetAccount(ctx context.Context, accountID uuid.UUID) (Account, error) {
	return s.accountRepo.GetByID(ctx, accountID)
}

// VerifyEmail verifies an email with a token
func (s *Service) VerifyEmail(ctx context.Context, tokenString string) error {
	token, err := s.tokenRepo.GetByToken(ctx, tokenString, TokenTypeEmailVerification)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	if token.GetExpiresAt().Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	// Update account
	account, err := s.accountRepo.GetByID(ctx, token.GetAccountID())
	if err != nil {
		return fmt.Errorf("account not found")
	}

	account.SetEmailVerified(true)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update account: %v", err)
	}

	// Delete the token
	if err := s.tokenRepo.Delete(ctx, token.GetID()); err != nil {
		return fmt.Errorf("failed to delete token: %v", err)
	}

	return nil
}

// RequestPasswordReset creates a password reset token
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	account, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if account exists
		return nil
	}

	// Generate reset token
	resetToken, err := s.generateToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %v", err)
	}

	tokenModel := &TokenModel{
		ID:        uuid.New(),
		AccountID: account.GetID(),
		Token:     resetToken,
		Type:      TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, tokenModel); err != nil {
		return fmt.Errorf("failed to create reset token: %v", err)
	}

	// TODO: Send reset email

	return nil
}

// ResetPassword resets password with a token
func (s *Service) ResetPassword(ctx context.Context, tokenString, newPassword string) error {
	token, err := s.tokenRepo.GetByToken(ctx, tokenString, TokenTypePasswordReset)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	if token.GetExpiresAt().Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update account
	account, err := s.accountRepo.GetByID(ctx, token.GetAccountID())
	if err != nil {
		return fmt.Errorf("account not found")
	}

	account.SetHashedPassword(string(hashedPassword))
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	// Delete the token
	if err := s.tokenRepo.Delete(ctx, token.GetID()); err != nil {
		return fmt.Errorf("failed to delete token: %v", err)
	}

	return nil
}

// RefreshToken generates a new JWT token from a refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshTokenString string) (*LoginResult, error) {
	token, err := s.tokenRepo.GetByToken(ctx, refreshTokenString, TokenTypeRefresh)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if token.GetExpiresAt().Before(time.Now()) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	// Get account
	account, err := s.accountRepo.GetByID(ctx, token.GetAccountID())
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	// Generate new JWT token
	jwtToken, err := s.generateJWT(account.GetID())
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return &LoginResult{
		Account:      account,
		Token:        jwtToken,
		RefreshToken: refreshTokenString,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}, nil
}

// Logout invalidates refresh tokens
func (s *Service) Logout(ctx context.Context, accountID uuid.UUID) error {
	return s.tokenRepo.DeleteByAccountID(ctx, accountID, TokenTypeRefresh)
}

// Helper methods

func (s *Service) generateJWT(accountID uuid.UUID) (string, error) {
	claims := &Claims{
		AccountID: accountID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
`

var authModelsTemplate = `package {{.Module.Name}}

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Account interface
type Account interface {
	GetID() uuid.UUID
	GetEmail() string
	GetHashedPassword() string
	IsEmailVerified() bool
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetEmailVerified(verified bool)
	SetHashedPassword(password string)
}

// Token interface
type Token interface {
	GetID() uuid.UUID
	GetAccountID() uuid.UUID
	GetToken() string
	GetType() TokenType
	GetExpiresAt() time.Time
	GetCreatedAt() time.Time
}

// AccountModel implements Account interface
type AccountModel struct {
	ID                uuid.UUID ` + "`json:\"id\" gorm:\"primaryKey;type:uuid;default:gen_random_uuid()\"`" + `
	Email            string    ` + "`json:\"email\" gorm:\"unique;not null\"`" + `
	HashedPassword   string    ` + "`json:\"-\" gorm:\"not null\"`" + `
	IsEmailVerified  bool      ` + "`json:\"is_email_verified\" gorm:\"default:false\"`" + `
	CreatedAt        time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt        time.Time ` + "`json:\"updated_at\"`" + `
}

func (a *AccountModel) GetID() uuid.UUID               { return a.ID }
func (a *AccountModel) GetEmail() string               { return a.Email }
func (a *AccountModel) GetHashedPassword() string      { return a.HashedPassword }
func (a *AccountModel) IsEmailVerified() bool          { return a.IsEmailVerified }
func (a *AccountModel) GetCreatedAt() time.Time        { return a.CreatedAt }
func (a *AccountModel) GetUpdatedAt() time.Time        { return a.UpdatedAt }
func (a *AccountModel) SetEmailVerified(verified bool) { a.IsEmailVerified = verified }
func (a *AccountModel) SetHashedPassword(password string) { a.HashedPassword = password }

// TokenType represents the type of token
type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
	TokenTypeRefresh           TokenType = "refresh"
)

// TokenModel implements Token interface
type TokenModel struct {
	ID        uuid.UUID ` + "`json:\"id\" gorm:\"primaryKey;type:uuid;default:gen_random_uuid()\"`" + `
	AccountID uuid.UUID ` + "`json:\"account_id\" gorm:\"type:uuid;not null\"`" + `
	Token     string    ` + "`json:\"token\" gorm:\"unique;not null\"`" + `
	Type      TokenType ` + "`json:\"type\" gorm:\"not null\"`" + `
	ExpiresAt time.Time ` + "`json:\"expires_at\" gorm:\"not null\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
}

func (t *TokenModel) GetID() uuid.UUID        { return t.ID }
func (t *TokenModel) GetAccountID() uuid.UUID { return t.AccountID }
func (t *TokenModel) GetToken() string        { return t.Token }
func (t *TokenModel) GetType() TokenType      { return t.Type }
func (t *TokenModel) GetExpiresAt() time.Time { return t.ExpiresAt }
func (t *TokenModel) GetCreatedAt() time.Time { return t.CreatedAt }

// Claims represents JWT claims
type Claims struct {
	AccountID uuid.UUID ` + "`json:\"account_id\"`" + `
	jwt.RegisteredClaims
}

// Request/Response types

type RegisterRequest struct {
	Email    string ` + "`json:\"email\" validate:\"required,email\"`" + `
	Password string ` + "`json:\"password\" validate:\"required,min=8\"`" + `
}

type LoginRequest struct {
	Email    string ` + "`json:\"email\" validate:\"required,email\"`" + `
	Password string ` + "`json:\"password\" validate:\"required\"`" + `
}

type ForgotPasswordRequest struct {
	Email string ` + "`json:\"email\" validate:\"required,email\"`" + `
}

type ResetPasswordRequest struct {
	Token    string ` + "`json:\"token\" validate:\"required\"`" + `
	Password string ` + "`json:\"password\" validate:\"required,min=8\"`" + `
}

type RefreshTokenRequest struct {
	RefreshToken string ` + "`json:\"refresh_token\" validate:\"required\"`" + `
}

type LoginResult struct {
	Account      Account   ` + "`json:\"account\"`" + `
	Token        string    ` + "`json:\"token\"`" + `
	RefreshToken string    ` + "`json:\"refresh_token\"`" + `
	ExpiresAt    time.Time ` + "`json:\"expires_at\"`" + `
}
`

var authInterfacesTemplate = `package {{.Module.Name}}

import (
	"context"

	"github.com/google/uuid"
)

// AuthService defines the auth service interface
type AuthService interface {
	Register(ctx context.Context, email, password string) (Account, error)
	Login(ctx context.Context, email, password string) (*LoginResult, error)
	ValidateToken(token string) (*Claims, error)
	GetAccount(ctx context.Context, accountID uuid.UUID) (Account, error)
	VerifyEmail(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResult, error)
	Logout(ctx context.Context, accountID uuid.UUID) error
}

// AccountRepository defines the account repository interface
type AccountRepository interface {
	Create(ctx context.Context, account Account) error
	GetByID(ctx context.Context, id uuid.UUID) (Account, error)
	GetByEmail(ctx context.Context, email string) (Account, error)
	Update(ctx context.Context, account Account) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TokenRepository defines the token repository interface
type TokenRepository interface {
	Create(ctx context.Context, token Token) error
	GetByID(ctx context.Context, id uuid.UUID) (Token, error)
	GetByToken(ctx context.Context, token string, tokenType TokenType) (Token, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByAccountID(ctx context.Context, accountID uuid.UUID, tokenType TokenType) error
	DeleteExpired(ctx context.Context) error
}
`

// Additional template variables for other files...
var authAccountRepoTemplate = `// GORM repository implementation for Account`
var authTokenRepoTemplate = `// GORM repository implementation for Token`  
var authMigrationsTemplate = `// Database migrations for auth module`

// Core templates
var coreResponseTemplate = `// Core response utilities`
var coreErrorsTemplate = `// Core error handling`
var coreKitTemplate = `// Core kit utilities`
var coreModuleTemplate = `// Core module interface`

// Placeholder templates for other modules
var subscriptionTemplates = map[string]string{}
var teamTemplates = map[string]string{}
var notificationTemplates = map[string]string{}
var healthTemplates = map[string]string{}
var roleTemplates = map[string]string{}
var jobTemplates = map[string]string{}
var sseTemplates = map[string]string{}
var containerTemplates = map[string]string{}
`