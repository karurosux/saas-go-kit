package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/errors-go"
)

// Service implements the AuthService interface
type Service struct {
	accountStore      AccountStore
	tokenStore        TokenStore
	emailProvider     EmailProvider
	configProvider    ConfigProvider
	passwordValidator PasswordValidator
	rateLimiter       RateLimiter
	eventListener     EventListener
	auditLogger       AuditLogger
}

// ServiceOption configures the service
type ServiceOption func(*Service)

// WithPasswordValidator sets a custom password validator
func WithPasswordValidator(validator PasswordValidator) ServiceOption {
	return func(s *Service) {
		s.passwordValidator = validator
	}
}

// WithRateLimiter sets a rate limiter
func WithRateLimiter(limiter RateLimiter) ServiceOption {
	return func(s *Service) {
		s.rateLimiter = limiter
	}
}

// WithEventListener sets an event listener
func WithEventListener(listener EventListener) ServiceOption {
	return func(s *Service) {
		s.eventListener = listener
	}
}

// WithAuditLogger sets an audit logger
func WithAuditLogger(logger AuditLogger) ServiceOption {
	return func(s *Service) {
		s.auditLogger = logger
	}
}

// NewService creates a new auth service
func NewService(
	accountStore AccountStore,
	tokenStore TokenStore,
	emailProvider EmailProvider,
	configProvider ConfigProvider,
	opts ...ServiceOption,
) *Service {
	s := &Service{
		accountStore:      accountStore,
		tokenStore:        tokenStore,
		emailProvider:     emailProvider,
		configProvider:    configProvider,
		passwordValidator: &DefaultPasswordValidator{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Register creates a new account
func (s *Service) Register(ctx context.Context, email, password string, metadata map[string]interface{}) (Account, error) {
	// Check rate limit
	if s.rateLimiter != nil {
		if err := s.rateLimiter.CheckLimit(ctx, email, "register"); err != nil {
			return nil, errors.ErrRateLimitExceeded
		}
	}

	// Validate password
	if err := s.passwordValidator.ValidatePassword(password); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// Check if email already exists
	existing, _ := s.accountStore.FindByEmail(ctx, email)
	if existing != nil {
		return nil, errors.Conflict("Email already registered")
	}

	// Create account
	account := NewAccount(email)
	if err := account.SetPassword(password); err != nil {
		return nil, errors.Internal("Failed to hash password")
	}

	// Set metadata
	if metadata != nil {
		account.Metadata = metadata
	}

	// Save account
	if err := s.accountStore.Create(ctx, account); err != nil {
		return nil, errors.Internal("Failed to create account")
	}

	// Create verification token
	token := NewVerificationToken(account.ID, TokenTypeEmailVerification, 24*time.Hour)
	if err := s.tokenStore.Create(ctx, token); err != nil {
		return nil, errors.Internal("Failed to create verification token")
	}

	// Send verification email
	if err := s.emailProvider.SendVerificationEmail(ctx, email, token.Token); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	// Fire event
	if s.eventListener != nil {
		s.eventListener.OnRegister(ctx, account)
	}

	// Record attempt
	if s.rateLimiter != nil {
		s.rateLimiter.RecordAttempt(ctx, email, "register")
	}

	return account, nil
}

// Login authenticates an account and returns a JWT token
func (s *Service) Login(ctx context.Context, email, password string) (string, Account, error) {
	// Check rate limit
	if s.rateLimiter != nil {
		if err := s.rateLimiter.CheckLimit(ctx, email, "login"); err != nil {
			return "", nil, errors.ErrRateLimitExceeded
		}
	}

	// Find account
	account, err := s.accountStore.FindByEmail(ctx, email)
	if err != nil {
		if s.rateLimiter != nil {
			s.rateLimiter.RecordAttempt(ctx, email, "login")
		}
		return "", nil, errors.ErrInvalidCredentials
	}

	// Check password
	if !account.CheckPassword(password) {
		if s.rateLimiter != nil {
			s.rateLimiter.RecordAttempt(ctx, email, "login")
		}
		if s.auditLogger != nil {
			s.auditLogger.LogLogin(ctx, account.GetID(), "", "", false)
		}
		return "", nil, errors.ErrInvalidCredentials
	}

	// Check if account is active
	if !account.IsActive() {
		return "", nil, errors.Forbidden("Account is deactivated")
	}

	// Generate token
	token, err := s.GenerateToken(account)
	if err != nil {
		return "", nil, errors.Internal("Failed to generate token")
	}

	// Fire event
	if s.eventListener != nil {
		s.eventListener.OnLogin(ctx, account)
	}

	// Log successful login
	if s.auditLogger != nil {
		s.auditLogger.LogLogin(ctx, account.GetID(), "", "", true)
	}

	return token, account, nil
}

// GenerateToken generates a JWT token for an account
func (s *Service) GenerateToken(account Account) (string, error) {
	claims := jwt.MapClaims{
		"account_id": account.GetID().String(),
		"email":      account.GetEmail(),
		"type":       "access",
		"exp":        time.Now().Add(s.configProvider.GetJWTExpiration()).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.configProvider.GetJWTSecret()))
}

// ValidateToken validates a JWT token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.configProvider.GetJWTSecret()), nil
	})

	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		accountID, err := uuid.Parse(claims["account_id"].(string))
		if err != nil {
			return nil, errors.ErrTokenInvalid
		}

		return &Claims{
			AccountID: accountID,
			Email:     claims["email"].(string),
			Type:      claims["type"].(string),
		}, nil
	}

	return nil, errors.ErrTokenInvalid
}

// RefreshToken refreshes a JWT token
func (s *Service) RefreshToken(ctx context.Context, oldToken string) (string, error) {
	// Validate old token
	claims, err := s.ValidateToken(oldToken)
	if err != nil {
		return "", err
	}

	// Get account
	account, err := s.accountStore.FindByID(ctx, claims.AccountID)
	if err != nil {
		return "", errors.ErrTokenInvalid
	}

	// Check if account is active
	if !account.IsActive() {
		return "", errors.Forbidden("Account is deactivated")
	}

	// Generate new token
	return s.GenerateToken(account)
}

// SendVerificationEmail sends a verification email
func (s *Service) SendVerificationEmail(ctx context.Context, accountID uuid.UUID) error {
	// Get account
	account, err := s.accountStore.FindByID(ctx, accountID)
	if err != nil {
		return errors.NotFound("Account")
	}

	// Check if already verified
	if account.IsEmailVerified() {
		return errors.BadRequest("Email already verified")
	}

	// Delete existing tokens
	s.tokenStore.DeleteByAccountID(ctx, accountID, TokenTypeEmailVerification)

	// Create new token
	token := NewVerificationToken(accountID, TokenTypeEmailVerification, 24*time.Hour)
	if err := s.tokenStore.Create(ctx, token); err != nil {
		return errors.Internal("Failed to create verification token")
	}

	// Send email
	return s.emailProvider.SendVerificationEmail(ctx, account.GetEmail(), token.Token)
}

// VerifyEmail verifies an email address
func (s *Service) VerifyEmail(ctx context.Context, tokenString string) error {
	// Find token
	token, err := s.tokenStore.FindByToken(ctx, tokenString)
	if err != nil {
		return errors.BadRequest("Invalid or expired token")
	}

	// Validate token
	if !token.IsValid() {
		return errors.BadRequest("Invalid or expired token")
	}

	// Get account
	account, err := s.accountStore.FindByID(ctx, token.GetAccountID())
	if err != nil {
		return errors.NotFound("Account")
	}

	// Update account
	defaultAccount, ok := account.(*DefaultAccount)
	if !ok {
		return errors.Internal("Invalid account type")
	}

	now := time.Now()
	defaultAccount.EmailVerified = true
	defaultAccount.EmailVerifiedAt = &now

	if err := s.accountStore.Update(ctx, defaultAccount); err != nil {
		return errors.Internal("Failed to update account")
	}

	// Mark token as used
	if err := s.tokenStore.MarkAsUsed(ctx, token.GetAccountID()); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to mark token as used: %v\n", err)
	}

	// Fire event
	if s.eventListener != nil {
		s.eventListener.OnEmailVerified(ctx, account)
	}

	// Send welcome email
	s.emailProvider.SendWelcomeEmail(ctx, account.GetEmail())

	return nil
}

// Logout logs out an account (placeholder for session-based auth)
func (s *Service) Logout(ctx context.Context, accountID uuid.UUID) error {
	// Get account
	account, err := s.accountStore.FindByID(ctx, accountID)
	if err != nil {
		return errors.NotFound("Account")
	}

	// Fire event
	if s.eventListener != nil {
		s.eventListener.OnLogout(ctx, account)
	}

	return nil
}

// RequestPasswordReset requests a password reset
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	// Find account
	account, err := s.accountStore.FindByEmail(ctx, email)
	if err != nil {
		// Don't reveal if account exists
		return nil
	}

	// Delete existing tokens
	s.tokenStore.DeleteByAccountID(ctx, account.GetID(), TokenTypePasswordReset)

	// Create reset token
	token := NewVerificationToken(account.GetID(), TokenTypePasswordReset, 1*time.Hour)
	if err := s.tokenStore.Create(ctx, token); err != nil {
		return errors.Internal("Failed to create reset token")
	}

	// Send email
	return s.emailProvider.SendPasswordResetEmail(ctx, email, token.Token)
}

// ResetPassword resets password with token
func (s *Service) ResetPassword(ctx context.Context, tokenString, newPassword string) error {
	// Find token
	token, err := s.tokenStore.FindByToken(ctx, tokenString)
	if err != nil {
		return errors.BadRequest("Invalid or expired token")
	}

	// Validate token
	if !token.IsValid() || token.GetType() != TokenTypePasswordReset {
		return errors.BadRequest("Invalid or expired token")
	}

	// Validate password
	if err := s.passwordValidator.ValidatePassword(newPassword); err != nil {
		return errors.BadRequest(err.Error())
	}

	// Get account
	account, err := s.accountStore.FindByID(ctx, token.GetAccountID())
	if err != nil {
		return errors.NotFound("Account")
	}

	// Update password
	if err := account.SetPassword(newPassword); err != nil {
		return errors.Internal("Failed to hash password")
	}

	if err := s.accountStore.Update(ctx, account); err != nil {
		return errors.Internal("Failed to update account")
	}

	// Mark token as used
	s.tokenStore.MarkAsUsed(ctx, token.GetAccountID())

	// Fire event
	if s.eventListener != nil {
		s.eventListener.OnPasswordReset(ctx, account)
	}

	// Log audit
	if s.auditLogger != nil {
		s.auditLogger.LogPasswordChange(ctx, account.GetID(), "")
	}

	return nil
}

// ChangePassword changes account password
func (s *Service) ChangePassword(ctx context.Context, accountID uuid.UUID, oldPassword, newPassword string) error {
	// Get account
	account, err := s.accountStore.FindByID(ctx, accountID)
	if err != nil {
		return errors.NotFound("Account")
	}

	// Verify old password
	if !account.CheckPassword(oldPassword) {
		return errors.BadRequest("Current password is incorrect")
	}

	// Validate new password
	if err := s.passwordValidator.ValidatePassword(newPassword); err != nil {
		return errors.BadRequest(err.Error())
	}

	// Update password
	if err := account.SetPassword(newPassword); err != nil {
		return errors.Internal("Failed to hash password")
	}

	if err := s.accountStore.Update(ctx, account); err != nil {
		return errors.Internal("Failed to update account")
	}

	// Log audit
	if s.auditLogger != nil {
		s.auditLogger.LogPasswordChange(ctx, accountID, "")
	}

	return nil
}

// GetAccount retrieves an account by ID
func (s *Service) GetAccount(ctx context.Context, accountID uuid.UUID) (Account, error) {
	return s.accountStore.FindByID(ctx, accountID)
}

// UpdateAccount updates account information
func (s *Service) UpdateAccount(ctx context.Context, accountID uuid.UUID, updates map[string]interface{}) (Account, error) {
	// Get account
	account, err := s.accountStore.FindByID(ctx, accountID)
	if err != nil {
		return nil, errors.NotFound("Account")
	}

	// Apply updates to DefaultAccount
	if defAccount, ok := account.(*DefaultAccount); ok {
		if companyName, ok := updates["company_name"].(string); ok {
			defAccount.CompanyName = companyName
		}
		if phone, ok := updates["phone"].(string); ok {
			defAccount.Phone = phone
		}
		// Update metadata
		if defAccount.Metadata == nil {
			defAccount.Metadata = make(map[string]interface{})
		}
		for k, v := range updates {
			if k != "company_name" && k != "phone" {
				defAccount.Metadata[k] = v
			}
		}
		defAccount.UpdatedAt = time.Now()
	}

	// Save updates
	if err := s.accountStore.Update(ctx, account); err != nil {
		return nil, errors.Internal("Failed to update account")
	}

	return account, nil
}

// RequestEmailChange requests an email change
func (s *Service) RequestEmailChange(ctx context.Context, accountID uuid.UUID, newEmail string) error {
	// Check if new email already exists
	existing, _ := s.accountStore.FindByEmail(ctx, newEmail)
	if existing != nil {
		return errors.Conflict("Email already in use")
	}

	// Get account
	account, err := s.accountStore.FindByID(ctx, accountID)
	if err != nil {
		return errors.NotFound("Account")
	}

	// Delete existing tokens
	s.tokenStore.DeleteByAccountID(ctx, accountID, TokenTypeEmailChange)

	// Create token with new email in metadata
	defToken := NewVerificationToken(accountID, TokenTypeEmailChange, 24*time.Hour)
	defToken.Token = defToken.Token + ":" + newEmail // Store new email in token

	if err := s.tokenStore.Create(ctx, defToken); err != nil {
		return errors.Internal("Failed to create email change token")
	}

	// Send confirmation email
	return s.emailProvider.SendEmailChangeConfirmation(ctx, account.GetEmail(), newEmail, defToken.Token)
}

// ConfirmEmailChange confirms an email change
func (s *Service) ConfirmEmailChange(ctx context.Context, tokenString string) error {
	// Find token
	token, err := s.tokenStore.FindByToken(ctx, tokenString)
	if err != nil {
		return errors.BadRequest("Invalid or expired token")
	}

	// Validate token
	if !token.IsValid() || token.GetType() != TokenTypeEmailChange {
		return errors.BadRequest("Invalid or expired token")
	}

	// Extract new email from token (simplified for example)
	// In production, store this in token metadata
	tokenParts := len(token.GetToken())
	newEmail := token.GetToken()[tokenParts-20:] // This is simplified

	// Get account
	account, err := s.accountStore.FindByID(ctx, token.GetAccountID())
	if err != nil {
		return errors.NotFound("Account")
	}

	oldEmail := account.GetEmail()

	// Update email
	if defAccount, ok := account.(*DefaultAccount); ok {
		defAccount.Email = newEmail
		defAccount.UpdatedAt = time.Now()
	}

	if err := s.accountStore.Update(ctx, account); err != nil {
		return errors.Internal("Failed to update account")
	}

	// Mark token as used
	s.tokenStore.MarkAsUsed(ctx, token.GetAccountID())

	// Fire event
	if s.eventListener != nil {
		s.eventListener.OnEmailChanged(ctx, account, oldEmail)
	}

	// Log audit
	if s.auditLogger != nil {
		s.auditLogger.LogEmailChange(ctx, account.GetID(), oldEmail, newEmail)
	}

	return nil
}

// DeactivateAccount deactivates an account
func (s *Service) DeactivateAccount(ctx context.Context, accountID uuid.UUID) error {
	// Get account
	account, err := s.accountStore.FindByID(ctx, accountID)
	if err != nil {
		return errors.NotFound("Account")
	}

	// Deactivate
	if defAccount, ok := account.(*DefaultAccount); ok {
		now := time.Now()
		defAccount.Active = false
		defAccount.DeactivatedAt = &now
		scheduledDeletion := now.Add(15 * 24 * time.Hour) // 15 days
		defAccount.ScheduledDeletionAt = &scheduledDeletion
		defAccount.UpdatedAt = now
	}

	if err := s.accountStore.Update(ctx, account); err != nil {
		return errors.Internal("Failed to deactivate account")
	}

	// Fire event
	if s.eventListener != nil {
		s.eventListener.OnAccountDeactivated(ctx, account)
	}

	// Log audit
	if s.auditLogger != nil {
		s.auditLogger.LogAccountDeactivation(ctx, accountID, "User requested")
	}

	return nil
}

// ReactivateAccount reactivates an account
func (s *Service) ReactivateAccount(ctx context.Context, accountID uuid.UUID) error {
	// Get account
	account, err := s.accountStore.FindByID(ctx, accountID)
	if err != nil {
		return errors.NotFound("Account")
	}

	// Reactivate
	if defAccount, ok := account.(*DefaultAccount); ok {
		defAccount.Active = true
		defAccount.DeactivatedAt = nil
		defAccount.ScheduledDeletionAt = nil
		defAccount.UpdatedAt = time.Now()
	}

	if err := s.accountStore.Update(ctx, account); err != nil {
		return errors.Internal("Failed to reactivate account")
	}

	return nil
}

// DefaultPasswordValidator provides basic password validation
type DefaultPasswordValidator struct{}

func (v *DefaultPasswordValidator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}
	return nil
}