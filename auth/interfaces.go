package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Account represents an authenticated account
type Account interface {
	GetID() uuid.UUID
	GetEmail() string
	SetPassword(password string) error
	CheckPassword(password string) bool
	IsActive() bool
	IsEmailVerified() bool
	GetMetadata() map[string]interface{}
}

// User represents an individual user (for multi-user accounts)
type User interface {
	GetID() uuid.UUID
	GetEmail() string
	GetFirstName() string
	GetLastName() string
	SetPassword(password string) error
	CheckPassword(password string) bool
}

// TokenType represents the type of verification token
type TokenType string

const (
	TokenTypeEmailVerification TokenType = "EMAIL_VERIFICATION"
	TokenTypePasswordReset     TokenType = "PASSWORD_RESET"
	TokenTypeTeamInvite        TokenType = "TEAM_INVITE"
	TokenTypeEmailChange       TokenType = "EMAIL_CHANGE"
)

// VerificationToken represents a verification token
type VerificationToken interface {
	GetToken() string
	GetType() TokenType
	GetAccountID() uuid.UUID
	GetExpiresAt() time.Time
	IsValid() bool
	MarkAsUsed()
}

// AccountStore handles account persistence
type AccountStore interface {
	Create(ctx context.Context, account Account) error
	FindByID(ctx context.Context, id uuid.UUID) (Account, error)
	FindByEmail(ctx context.Context, email string) (Account, error)
	Update(ctx context.Context, account Account) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// TokenStore handles token persistence
type TokenStore interface {
	Create(ctx context.Context, token VerificationToken) error
	FindByToken(ctx context.Context, token string) (VerificationToken, error)
	MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	DeleteByAccountID(ctx context.Context, accountID uuid.UUID, tokenType TokenType) error
}

// EmailProvider handles email sending
type EmailProvider interface {
	SendVerificationEmail(ctx context.Context, email, token string) error
	SendPasswordResetEmail(ctx context.Context, email, token string) error
	SendEmailChangeConfirmation(ctx context.Context, oldEmail, newEmail, token string) error
	SendWelcomeEmail(ctx context.Context, email string) error
}

// ConfigProvider provides configuration values
type ConfigProvider interface {
	GetJWTSecret() string
	GetJWTExpiration() time.Duration
	GetRefreshExpiration() time.Duration
	GetAppURL() string
	GetAppName() string
	IsDevMode() bool
}

// PasswordValidator validates password strength
type PasswordValidator interface {
	ValidatePassword(password string) error
}

// RateLimiter provides rate limiting for auth operations
type RateLimiter interface {
	CheckLimit(ctx context.Context, identifier string, action string) error
	RecordAttempt(ctx context.Context, identifier string, action string) error
}

// EventListener handles auth events
type EventListener interface {
	OnRegister(ctx context.Context, account Account) error
	OnLogin(ctx context.Context, account Account) error
	OnLogout(ctx context.Context, account Account) error
	OnPasswordReset(ctx context.Context, account Account) error
	OnEmailVerified(ctx context.Context, account Account) error
	OnEmailChanged(ctx context.Context, account Account, oldEmail string) error
	OnAccountDeactivated(ctx context.Context, account Account) error
}

// Claims represents JWT claims
type Claims struct {
	AccountID uuid.UUID              `json:"account_id"`
	Email     string                 `json:"email"`
	Type      string                 `json:"type"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AuthService provides authentication operations
type AuthService interface {
	// Account management
	Register(ctx context.Context, email, password string, metadata map[string]interface{}) (Account, error)
	Login(ctx context.Context, email, password string) (string, Account, error)
	Logout(ctx context.Context, accountID uuid.UUID) error
	
	// Token operations
	GenerateToken(account Account) (string, error)
	ValidateToken(token string) (*Claims, error)
	RefreshToken(ctx context.Context, oldToken string) (string, error)
	
	// Email verification
	SendVerificationEmail(ctx context.Context, accountID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error
	
	// Password operations
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, accountID uuid.UUID, oldPassword, newPassword string) error
	
	// Email change
	RequestEmailChange(ctx context.Context, accountID uuid.UUID, newEmail string) error
	ConfirmEmailChange(ctx context.Context, token string) error
	
	// Account operations
	GetAccount(ctx context.Context, accountID uuid.UUID) (Account, error)
	UpdateAccount(ctx context.Context, accountID uuid.UUID, updates map[string]interface{}) (Account, error)
	DeactivateAccount(ctx context.Context, accountID uuid.UUID) error
	ReactivateAccount(ctx context.Context, accountID uuid.UUID) error
}

// SessionStore handles session management (optional)
type SessionStore interface {
	Create(ctx context.Context, accountID uuid.UUID, token string, expiresAt time.Time) error
	Get(ctx context.Context, token string) (uuid.UUID, error)
	Delete(ctx context.Context, token string) error
	DeleteByAccountID(ctx context.Context, accountID uuid.UUID) error
}

// AuditLogger logs security events
type AuditLogger interface {
	LogLogin(ctx context.Context, accountID uuid.UUID, ip string, userAgent string, success bool) error
	LogPasswordChange(ctx context.Context, accountID uuid.UUID, ip string) error
	LogEmailChange(ctx context.Context, accountID uuid.UUID, oldEmail, newEmail string) error
	LogAccountDeactivation(ctx context.Context, accountID uuid.UUID, reason string) error
}