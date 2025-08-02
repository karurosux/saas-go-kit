package authservice

import (
	"context"
	"fmt"
	"time"
	
	authconstants "{{.Project.GoModule}}/internal/auth/constants"
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmodel "{{.Project.GoModule}}/internal/auth/model"
	"{{.Project.GoModule}}/internal/core"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

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

func (s *AuthService) Register(ctx context.Context, req authinterface.RegisterRequest) (authinterface.Account, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	exists, err := s.accountRepo.ExistsByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, fmt.Errorf("failed to check account existence")
	}
	if exists {
		return nil, fmt.Errorf(authconstants.ErrAccountAlreadyExists)
	}
	
	passwordHash, err := s.passwordHasher.Hash(req.GetPassword())
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	account := &authmodel.Account{
		ID:           uuid.New(),
		Email:        req.GetEmail(),
		Phone:        req.GetPhone(),
		PasswordHash: passwordHash,
	}
	
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	
	if s.config.IsEmailVerificationRequired() {
		if err := s.SendEmailVerification(ctx, account.ID); err != nil {
			fmt.Printf("Failed to send verification email: %v\n", err)
		}
	}
	
	if err := s.emailSender.SendWelcomeEmail(account.Email); err != nil {
		fmt.Printf("Failed to send welcome email: %v\n", err)
	}
	
	return account, nil
}

func (s *AuthService) Login(ctx context.Context, req authinterface.LoginRequest) (authinterface.Session, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	account, err := s.accountRepo.GetByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrInvalidCredentials)
	}
	
	if err := s.passwordHasher.Verify(req.GetPassword(), account.GetPasswordHash()); err != nil {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrInvalidCredentials)
	}
	
	if s.config.IsEmailVerificationRequired() && !account.GetEmailVerified() {
		return nil, core.NewAppError(core.ErrCodeForbidden, authconstants.ErrEmailNotVerified)
	}
	
	session := s.createSession(account.GetID())
	
	if err := s.sessionStore.Store(ctx, session); err != nil {
		return nil, core.NewAppError(core.ErrCodeInternalServer, "failed to create session")
	}
	
	return session, nil
}

func (s *AuthService) RefreshSession(ctx context.Context, refreshToken string) (authinterface.Session, error) {
	oldSession, err := s.sessionStore.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrInvalidToken)
	}
	
	if oldSession.IsRefreshExpired() {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrTokenExpired)
	}
	
	if err := s.sessionStore.DeleteByToken(ctx, oldSession.GetToken()); err != nil {
		fmt.Printf("Failed to delete old session: %v\n", err)
	}
	
	newSession := s.createSession(oldSession.GetUserID())
	
	if err := s.sessionStore.Store(ctx, newSession); err != nil {
		return nil, core.NewAppError(core.ErrCodeInternalServer, "failed to create session")
	}
	
	return newSession, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.sessionStore.Delete(ctx, userID)
}

func (s *AuthService) SendEmailVerification(ctx context.Context, accountID uuid.UUID) error {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return core.NewAppError(core.ErrCodeNotFound, authconstants.ErrAccountNotFound)
	}
	
	if account.GetEmailVerified() {
		return nil
	}
	
	token := &authmodel.Token{
		ID:        uuid.New(),
		AccountID: accountID,
		Token:     s.tokenGenerator.GenerateSecureToken(),
		Type:      authinterface.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(s.config.GetVerificationTokenExpiration()),
	}
	
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to create verification token")
	}
	
	return s.emailSender.SendVerificationEmail(account.GetEmail(), token.Token)
}

func (s *AuthService) VerifyEmail(ctx context.Context, tokenStr string) error {
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return fmt.Errorf(authconstants.ErrInvalidToken)
	}
	
	if token.GetType() != authinterface.TokenTypeEmailVerification {
		return fmt.Errorf(authconstants.ErrInvalidToken)
	}
	
	if token.GetUsed() {
		return fmt.Errorf(authconstants.ErrTokenAlreadyUsed)
	}
	
	if token.IsExpired() {
		return fmt.Errorf(authconstants.ErrTokenExpired)
	}
	
	account, err := s.accountRepo.GetByID(ctx, token.GetAccountID())
	if err != nil {
		return core.NewAppError(core.ErrCodeNotFound, authconstants.ErrAccountNotFound)
	}
	
	account.SetEmailVerified(true)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to update account")
	}
	
	if err := s.tokenRepo.MarkAsUsed(ctx, token.GetID()); err != nil {
		fmt.Printf("Failed to mark token as used: %v\n", err)
	}
	
	return nil
}

func (s *AuthService) SendPhoneVerification(ctx context.Context, accountID uuid.UUID) error {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return core.NewAppError(core.ErrCodeNotFound, authconstants.ErrAccountNotFound)
	}
	
	if account.GetPhoneVerified() {
		return nil
	}
	
	if account.GetPhone() == "" {
		return fmt.Errorf("no phone number associated with account")
	}
	
	code := s.tokenGenerator.GenerateToken()[:6]
	
	token := &authmodel.Token{
		ID:        uuid.New(),
		AccountID: accountID,
		Token:     code,
		Type:      authinterface.TokenTypePhoneVerification,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to create verification token")
	}
	
	return s.smsSender.SendVerificationSMS(account.GetPhone(), code)
}

func (s *AuthService) VerifyPhone(ctx context.Context, accountID uuid.UUID, code string) error {
	tokens, err := s.tokenRepo.GetByAccountAndType(ctx, accountID, authinterface.TokenTypePhoneVerification)
	if err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to get verification tokens")
	}
	
	var validToken authinterface.Token
	for _, token := range tokens {
		if token.GetToken() == code && !token.GetUsed() && !token.IsExpired() {
			validToken = token
			break
		}
	}
	
	if validToken == nil {
		return fmt.Errorf(authconstants.ErrInvalidToken)
	}
	
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return core.NewAppError(core.ErrCodeNotFound, authconstants.ErrAccountNotFound)
	}
	
	account.SetPhoneVerified(true)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to update account")
	}
	
	if err := s.tokenRepo.MarkAsUsed(ctx, validToken.GetID()); err != nil {
		fmt.Printf("Failed to mark token as used: %v\n", err)
	}
	
	return nil
}

func (s *AuthService) SendPasswordReset(ctx context.Context, email string) error {
	account, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}
	
	token := &authmodel.Token{
		ID:        uuid.New(),
		AccountID: account.GetID(),
		Token:     s.tokenGenerator.GenerateSecureToken(),
		Type:      authinterface.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(s.config.GetPasswordResetTokenExpiration()),
	}
	
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to create reset token")
	}
	
	return s.emailSender.SendPasswordResetEmail(account.GetEmail(), token.Token)
}

func (s *AuthService) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf(authconstants.ErrPasswordTooWeak)
	}
	
	token, err := s.tokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return fmt.Errorf(authconstants.ErrInvalidToken)
	}
	
	if token.GetType() != authinterface.TokenTypePasswordReset {
		return fmt.Errorf(authconstants.ErrInvalidToken)
	}
	
	if token.GetUsed() {
		return fmt.Errorf(authconstants.ErrTokenAlreadyUsed)
	}
	
	if token.IsExpired() {
		return fmt.Errorf(authconstants.ErrTokenExpired)
	}
	
	account, err := s.accountRepo.GetByID(ctx, token.GetAccountID())
	if err != nil {
		return core.NewAppError(core.ErrCodeNotFound, authconstants.ErrAccountNotFound)
	}
	
	passwordHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to hash password")
	}
	
	account.SetPasswordHash(passwordHash)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to update account")
	}
	
	if err := s.tokenRepo.MarkAsUsed(ctx, token.GetID()); err != nil {
		fmt.Printf("Failed to mark token as used: %v\n", err)
	}
	
	if err := s.sessionStore.Delete(ctx, account.GetID()); err != nil {
		fmt.Printf("Failed to delete sessions: %v\n", err)
	}
	
	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, accountID uuid.UUID, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf(authconstants.ErrPasswordTooWeak)
	}
	
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return core.NewAppError(core.ErrCodeNotFound, authconstants.ErrAccountNotFound)
	}
	
	if err := s.passwordHasher.Verify(oldPassword, account.GetPasswordHash()); err != nil {
		return core.NewAppError(core.ErrCodeBadRequest, authconstants.ErrInvalidPassword)
	}
	
	passwordHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to hash password")
	}
	
	account.SetPasswordHash(passwordHash)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return core.NewAppError(core.ErrCodeInternalServer, "failed to update account")
	}
	
	if err := s.sessionStore.Delete(ctx, accountID); err != nil {
		fmt.Printf("Failed to delete sessions: %v\n", err)
	}
	
	return nil
}

func (s *AuthService) GetAccount(ctx context.Context, accountID uuid.UUID) (authinterface.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeNotFound, authconstants.ErrAccountNotFound)
	}
	return account, nil
}

func (s *AuthService) GetAccountByEmail(ctx context.Context, email string) (authinterface.Account, error) {
	account, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeNotFound, authconstants.ErrAccountNotFound)
	}
	return account, nil
}

func (s *AuthService) UpdateAccount(ctx context.Context, accountID uuid.UUID, updates authinterface.AccountUpdates) (authinterface.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeNotFound, authconstants.ErrAccountNotFound)
	}
	
	if updates.Email != nil {
		if *updates.Email != account.GetEmail() {
			exists, err := s.accountRepo.ExistsByEmail(ctx, *updates.Email)
			if err != nil {
				return nil, core.NewAppError(core.ErrCodeInternalServer, "failed to check email existence")
			}
			if exists {
				return nil, fmt.Errorf("email already in use")
			}
		}
		if acc, ok := account.(*authmodel.Account); ok {
			acc.Email = *updates.Email
			acc.EmailVerified = false
		}
	}
	
	if updates.Phone != nil {
		if *updates.Phone != account.GetPhone() && *updates.Phone != "" {
			exists, err := s.accountRepo.ExistsByPhone(ctx, *updates.Phone)
			if err != nil {
				return nil, core.NewAppError(core.ErrCodeInternalServer, "failed to check phone existence")
			}
			if exists {
				return nil, fmt.Errorf("phone already in use")
			}
		}
		if acc, ok := account.(*authmodel.Account); ok {
			acc.Phone = *updates.Phone
			acc.PhoneVerified = false
		}
	}
	
	if updates.EmailVerified != nil {
		account.SetEmailVerified(*updates.EmailVerified)
	}
	
	if updates.PhoneVerified != nil {
		account.SetPhoneVerified(*updates.PhoneVerified)
	}
	
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, core.NewAppError(core.ErrCodeInternalServer, "failed to update account")
	}
	
	return account, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (authinterface.Account, error) {
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.GetJWTSecret()), nil
	})
	
	if err != nil || !jwtToken.Valid {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrInvalidToken)
	}
	
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrInvalidToken)
	}
	
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrInvalidToken)
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrInvalidToken)
	}
	
	session, err := s.sessionStore.Get(ctx, token)
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrSessionExpired)
	}
	
	if session.IsExpired() {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrSessionExpired)
	}
	
	return s.GetAccount(ctx, userID)
}

func (s *AuthService) createSession(userID uuid.UUID) authinterface.Session {
	now := time.Now()
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     now.Add(s.config.GetJWTExpiration()).Unix(),
		"iat":     now.Unix(),
	})
	
	tokenString, _ := token.SignedString([]byte(s.config.GetJWTSecret()))
	
	refreshToken := s.tokenGenerator.GenerateSecureToken()
	
	return &authmodel.Session{
		UserID:           userID,
		Token:            tokenString,
		RefreshToken:     refreshToken,
		ExpiresAt:        now.Add(s.config.GetJWTExpiration()),
		RefreshExpiresAt: now.Add(s.config.GetRefreshTokenExpiration()),
	}
}
