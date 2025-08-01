package authservice

import (
	"context"
	"fmt"
	"time"
	
	"{{.Project.GoModule}}/internal/auth/constants"
	"{{.Project.GoModule}}/internal/auth/interface"
	"{{.Project.GoModule}}/internal/auth/model"
	"{{.Project.GoModule}}/internal/core"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthService implements the authentication service
type AuthService struct {
	accountRepo    authinterface.AccountRepository
	tokenRepo      authinterface.TokenRepository
	sessionStore   authinterface.SessionStore
	passwordHasher authinterface.PasswordHasher
	tokenGenerator authinterface.TokenGenerator
	emailSender    authinterface.EmailSender
	smsSender      authinterface.SMSSender
	config         authinterface.AuthConfig
}

// NewAuthService creates a new auth service
func NewAuthService(
	accountRepo authinterface.AccountRepository,
	tokenRepo authinterface.TokenRepository,
	sessionStore authinterface.SessionStore,
	passwordHasher authinterface.PasswordHasher,
	tokenGenerator authinterface.TokenGenerator,
	emailSender authinterface.EmailSender,
	smsSender authinterface.SMSSender,
	config authinterface.AuthConfig,
) *AuthService {
	return &AuthService{
		accountRepo:    accountRepo,
		tokenRepo:      tokenRepo,
		sessionStore:   sessionStore,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
		emailSender:    emailSender,
		smsSender:      smsSender,
		config:         config,
	}
}

// Register creates a new account
func (s *AuthService) Register(ctx context.Context, req authinterface.RegisterRequest) (authinterface.Account, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Check if account already exists
	exists, err := s.accountRepo.ExistsByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, core.InternalServerError("failed to check account existence")
	}
	if exists {
		return nil, core.BadRequest(authconstants.ErrAccountAlreadyExists)
	}
	
	// Hash password
	passwordHash, err := s.passwordHasher.Hash(req.GetPassword())
	if err != nil {
		return nil, core.InternalServerError("failed to hash password")
	}
	
	// Create account
	account := &authmodel.Account{
		ID:           uuid.New(),
		Email:        req.GetEmail(),
		Phone:        req.GetPhone(),
		PasswordHash: passwordHash,
	}
	
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, core.InternalServerError("failed to create account")
	}
	
	// Send verification email if required
	if s.config.IsEmailVerificationRequired() {
		if err := s.SendEmailVerification(ctx, account.ID); err != nil {
			// Log error but don't fail registration
			fmt.Printf("Failed to send verification email: %v\n", err)
		}
	}
	
	// Send welcome email
	if err := s.emailSender.SendWelcomeEmail(account.Email); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to send welcome email: %v\n", err)
	}
	
	return account, nil
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(ctx context.Context, req authinterface.LoginRequest) (authinterface.Session, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Get account
	account, err := s.accountRepo.GetByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, core.Unauthorized(authconstants.ErrInvalidCredentials)
	}
	
	// Verify password
	if err := s.passwordHasher.Verify(req.GetPassword(), account.GetPasswordHash()); err != nil {
		return nil, core.Unauthorized(authconstants.ErrInvalidCredentials)
	}
	
	// Check email verification if required
	if s.config.IsEmailVerificationRequired() && !account.GetEmailVerified() {
		return nil, core.Forbidden(authconstants.ErrEmailNotVerified)
	}
	
	// Create session
	session := s.createSession(account.GetID())
	
	// Store session
	if err := s.sessionStore.Store(ctx, session); err != nil {
		return nil, core.InternalServerError("failed to create session")
	}
	
	return session, nil
}

// RefreshSession refreshes an existing session
func (s *AuthService) RefreshSession(ctx context.Context, refreshToken string) (authinterface.Session, error) {
	// Get existing session
	oldSession, err := s.sessionStore.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, core.Unauthorized(authconstants.ErrInvalidToken)
	}
	
	// Check if refresh token is expired
	if oldSession.IsRefreshExpired() {
		return nil, core.Unauthorized(authconstants.ErrTokenExpired)
	}
	
	// Delete old session
	if err := s.sessionStore.DeleteByToken(ctx, oldSession.GetToken()); err != nil {
		// Log error but continue
		fmt.Printf("Failed to delete old session: %v\n", err)
	}
	
	// Create new session
	newSession := s.createSession(oldSession.GetUserID())
	
	// Store new session
	if err := s.sessionStore.Store(ctx, newSession); err != nil {
		return nil, core.InternalServerError("failed to create session")
	}
	
	return newSession, nil
}

// Logout destroys a user's session
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.sessionStore.Delete(ctx, userID)
}

// SendEmailVerification sends an email verification token
func (s *AuthService) SendEmailVerification(ctx context.Context, accountID uuid.UUID) error {
	// Get account
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return core.NotFound(authconstants.ErrAccountNotFound)
	}
	
	// Check if already verified
	if account.GetEmailVerified() {
		return nil
	}
	
	// Create verification token
	token := &authmodel.Token{
		ID:        uuid.New(),
		AccountID: accountID,
		Token:     s.tokenGenerator.GenerateSecureToken(),
		Type:      authinterface.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(s.config.GetVerificationTokenExpiration()),
	}
	
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return core.InternalServerError("failed to create verification token")
	}
	
	// Send email
	return s.emailSender.SendVerificationEmail(account.GetEmail(), token.Token)
}

// VerifyEmail verifies an email with a token
func (s *AuthService) VerifyEmail(ctx context.Context, tokenStr string) error {
	// Get token
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return core.BadRequest(authconstants.ErrInvalidToken)
	}
	
	// Check token type
	if token.GetType() != authinterface.TokenTypeEmailVerification {
		return core.BadRequest(authconstants.ErrInvalidToken)
	}
	
	// Check if used
	if token.GetUsed() {
		return core.BadRequest(authconstants.ErrTokenAlreadyUsed)
	}
	
	// Check if expired
	if token.IsExpired() {
		return core.BadRequest(authconstants.ErrTokenExpired)
	}
	
	// Get account
	account, err := s.accountRepo.GetByID(ctx, token.GetAccountID())
	if err != nil {
		return core.NotFound(authconstants.ErrAccountNotFound)
	}
	
	// Update account
	account.SetEmailVerified(true)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return core.InternalServerError("failed to update account")
	}
	
	// Mark token as used
	if err := s.tokenRepo.MarkAsUsed(ctx, token.GetID()); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to mark token as used: %v\n", err)
	}
	
	return nil
}

// SendPhoneVerification sends a phone verification code
func (s *AuthService) SendPhoneVerification(ctx context.Context, accountID uuid.UUID) error {
	// Get account
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return core.NotFound(authconstants.ErrAccountNotFound)
	}
	
	// Check if already verified
	if account.GetPhoneVerified() {
		return nil
	}
	
	// Check if phone exists
	if account.GetPhone() == "" {
		return core.BadRequest("no phone number associated with account")
	}
	
	// Generate 6-digit code
	code := s.tokenGenerator.GenerateToken()[:6]
	
	// Create verification token
	token := &authmodel.Token{
		ID:        uuid.New(),
		AccountID: accountID,
		Token:     code,
		Type:      authinterface.TokenTypePhoneVerification,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return core.InternalServerError("failed to create verification token")
	}
	
	// Send SMS
	return s.smsSender.SendVerificationSMS(account.GetPhone(), code)
}

// VerifyPhone verifies a phone with a code
func (s *AuthService) VerifyPhone(ctx context.Context, accountID uuid.UUID, code string) error {
	// Get tokens
	tokens, err := s.tokenRepo.GetByAccountAndType(ctx, accountID, authinterface.TokenTypePhoneVerification)
	if err != nil {
		return core.InternalServerError("failed to get verification tokens")
	}
	
	// Find valid token
	var validToken authinterface.Token
	for _, token := range tokens {
		if token.GetToken() == code && !token.GetUsed() && !token.IsExpired() {
			validToken = token
			break
		}
	}
	
	if validToken == nil {
		return core.BadRequest(authconstants.ErrInvalidToken)
	}
	
	// Get account
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return core.NotFound(authconstants.ErrAccountNotFound)
	}
	
	// Update account
	account.SetPhoneVerified(true)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return core.InternalServerError("failed to update account")
	}
	
	// Mark token as used
	if err := s.tokenRepo.MarkAsUsed(ctx, validToken.GetID()); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to mark token as used: %v\n", err)
	}
	
	return nil
}

// SendPasswordReset sends a password reset email
func (s *AuthService) SendPasswordReset(ctx context.Context, email string) error {
	// Get account
	account, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if account exists
		return nil
	}
	
	// Create reset token
	token := &authmodel.Token{
		ID:        uuid.New(),
		AccountID: account.GetID(),
		Token:     s.tokenGenerator.GenerateSecureToken(),
		Type:      authinterface.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(s.config.GetPasswordResetTokenExpiration()),
	}
	
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return core.InternalServerError("failed to create reset token")
	}
	
	// Send email
	return s.emailSender.SendPasswordResetEmail(account.GetEmail(), token.Token)
}

// ResetPassword resets a password using a token
func (s *AuthService) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	// Validate password
	if len(newPassword) < 8 {
		return core.BadRequest(authconstants.ErrPasswordTooWeak)
	}
	
	// Get token
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return core.BadRequest(authconstants.ErrInvalidToken)
	}
	
	// Check token type
	if token.GetType() != authinterface.TokenTypePasswordReset {
		return core.BadRequest(authconstants.ErrInvalidToken)
	}
	
	// Check if used
	if token.GetUsed() {
		return core.BadRequest(authconstants.ErrTokenAlreadyUsed)
	}
	
	// Check if expired
	if token.IsExpired() {
		return core.BadRequest(authconstants.ErrTokenExpired)
	}
	
	// Get account
	account, err := s.accountRepo.GetByID(ctx, token.GetAccountID())
	if err != nil {
		return core.NotFound(authconstants.ErrAccountNotFound)
	}
	
	// Hash new password
	passwordHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return core.InternalServerError("failed to hash password")
	}
	
	// Update account
	account.SetPasswordHash(passwordHash)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return core.InternalServerError("failed to update account")
	}
	
	// Mark token as used
	if err := s.tokenRepo.MarkAsUsed(ctx, token.GetID()); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to mark token as used: %v\n", err)
	}
	
	// Delete all sessions for this user
	if err := s.sessionStore.Delete(ctx, account.GetID()); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to delete sessions: %v\n", err)
	}
	
	return nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, accountID uuid.UUID, oldPassword, newPassword string) error {
	// Validate new password
	if len(newPassword) < 8 {
		return core.BadRequest(authconstants.ErrPasswordTooWeak)
	}
	
	// Get account
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return core.NotFound(authconstants.ErrAccountNotFound)
	}
	
	// Verify old password
	if err := s.passwordHasher.Verify(oldPassword, account.GetPasswordHash()); err != nil {
		return core.BadRequest(authconstants.ErrInvalidPassword)
	}
	
	// Hash new password
	passwordHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return core.InternalServerError("failed to hash password")
	}
	
	// Update account
	account.SetPasswordHash(passwordHash)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return core.InternalServerError("failed to update account")
	}
	
	// Delete all sessions for this user
	if err := s.sessionStore.Delete(ctx, accountID); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to delete sessions: %v\n", err)
	}
	
	return nil
}

// GetAccount gets an account by ID
func (s *AuthService) GetAccount(ctx context.Context, accountID uuid.UUID) (authinterface.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, core.NotFound(authconstants.ErrAccountNotFound)
	}
	return account, nil
}

// GetAccountByEmail gets an account by email
func (s *AuthService) GetAccountByEmail(ctx context.Context, email string) (authinterface.Account, error) {
	account, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, core.NotFound(authconstants.ErrAccountNotFound)
	}
	return account, nil
}

// UpdateAccount updates an account
func (s *AuthService) UpdateAccount(ctx context.Context, accountID uuid.UUID, updates authinterface.AccountUpdates) (authinterface.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, core.NotFound(authconstants.ErrAccountNotFound)
	}
	
	// Apply updates
	if updates.Email != nil {
		// Check if new email already exists
		if *updates.Email != account.GetEmail() {
			exists, err := s.accountRepo.ExistsByEmail(ctx, *updates.Email)
			if err != nil {
				return nil, core.InternalServerError("failed to check email existence")
			}
			if exists {
				return nil, core.BadRequest("email already in use")
			}
		}
		if acc, ok := account.(*authmodel.Account); ok {
			acc.Email = *updates.Email
			acc.EmailVerified = false // Reset verification
		}
	}
	
	if updates.Phone != nil {
		// Check if new phone already exists
		if *updates.Phone != account.GetPhone() && *updates.Phone != "" {
			exists, err := s.accountRepo.ExistsByPhone(ctx, *updates.Phone)
			if err != nil {
				return nil, core.InternalServerError("failed to check phone existence")
			}
			if exists {
				return nil, core.BadRequest("phone already in use")
			}
		}
		if acc, ok := account.(*authmodel.Account); ok {
			acc.Phone = *updates.Phone
			acc.PhoneVerified = false // Reset verification
		}
	}
	
	if updates.EmailVerified != nil {
		account.SetEmailVerified(*updates.EmailVerified)
	}
	
	if updates.PhoneVerified != nil {
		account.SetPhoneVerified(*updates.PhoneVerified)
	}
	
	// Save updates
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, core.InternalServerError("failed to update account")
	}
	
	return account, nil
}

// ValidateSession validates a session token and returns the associated account
func (s *AuthService) ValidateSession(ctx context.Context, token string) (authinterface.Account, error) {
	// Parse JWT token
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.GetJWTSecret()), nil
	})
	
	if err != nil || !jwtToken.Valid {
		return nil, core.Unauthorized(authconstants.ErrInvalidToken)
	}
	
	// Get user ID from claims
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, core.Unauthorized(authconstants.ErrInvalidToken)
	}
	
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, core.Unauthorized(authconstants.ErrInvalidToken)
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, core.Unauthorized(authconstants.ErrInvalidToken)
	}
	
	// Get session from store
	session, err := s.sessionStore.Get(ctx, token)
	if err != nil {
		return nil, core.Unauthorized(authconstants.ErrSessionExpired)
	}
	
	// Check if expired
	if session.IsExpired() {
		return nil, core.Unauthorized(authconstants.ErrSessionExpired)
	}
	
	// Get account
	return s.GetAccount(ctx, userID)
}

// createSession creates a new session for a user
func (s *AuthService) createSession(userID uuid.UUID) authinterface.Session {
	now := time.Now()
	
	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     now.Add(s.config.GetJWTExpiration()).Unix(),
		"iat":     now.Unix(),
	})
	
	tokenString, _ := token.SignedString([]byte(s.config.GetJWTSecret()))
	
	// Create refresh token
	refreshToken := s.tokenGenerator.GenerateSecureToken()
	
	return &authmodel.Session{
		UserID:           userID,
		Token:            tokenString,
		RefreshToken:     refreshToken,
		ExpiresAt:        now.Add(s.config.GetJWTExpiration()),
		RefreshExpiresAt: now.Add(s.config.GetRefreshTokenExpiration()),
	}
}