package authinterface

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// Account represents a user account
type Account interface {
	GetID() uuid.UUID
	GetEmail() string
	GetPhone() string
	GetPasswordHash() string
	GetEmailVerified() bool
	GetPhoneVerified() bool
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetPasswordHash(hash string)
	SetEmailVerified(verified bool)
	SetPhoneVerified(verified bool)
}

// Token represents an authentication or verification token
type Token interface {
	GetID() uuid.UUID
	GetAccountID() uuid.UUID
	GetToken() string
	GetType() TokenType
	GetUsed() bool
	GetExpiresAt() time.Time
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetUsed(used bool)
	IsExpired() bool
}

// TokenType represents the type of token
type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePhoneVerification TokenType = "phone_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
	TokenTypeRefresh           TokenType = "refresh"
)

// Session represents a user session
type Session interface {
	GetUserID() uuid.UUID
	GetToken() string
	GetRefreshToken() string
	GetExpiresAt() time.Time
	GetRefreshExpiresAt() time.Time
	IsExpired() bool
	IsRefreshExpired() bool
}

// LoginRequest represents a login request
type LoginRequest interface {
	GetEmail() string
	GetPassword() string
	Validate() error
}

// RegisterRequest represents a registration request
type RegisterRequest interface {
	GetEmail() string
	GetPhone() string
	GetPassword() string
	Validate() error
}

// AuthService defines the authentication service interface
type AuthService interface {
	// Account management
	Register(ctx context.Context, req RegisterRequest) (Account, error)
	Login(ctx context.Context, req LoginRequest) (Session, error)
	RefreshSession(ctx context.Context, refreshToken string) (Session, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	
	// Email verification
	SendEmailVerification(ctx context.Context, accountID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error
	
	// Phone verification
	SendPhoneVerification(ctx context.Context, accountID uuid.UUID) error
	VerifyPhone(ctx context.Context, accountID uuid.UUID, code string) error
	
	// Password management
	SendPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, accountID uuid.UUID, oldPassword, newPassword string) error
	
	// Account queries
	GetAccount(ctx context.Context, accountID uuid.UUID) (Account, error)
	GetAccountByEmail(ctx context.Context, email string) (Account, error)
	UpdateAccount(ctx context.Context, accountID uuid.UUID, updates AccountUpdates) (Account, error)
	
	// Session validation
	ValidateSession(ctx context.Context, token string) (Account, error)
}

// AccountRepository defines the account repository interface
type AccountRepository interface {
	Create(ctx context.Context, account Account) error
	GetByID(ctx context.Context, id uuid.UUID) (Account, error)
	GetByEmail(ctx context.Context, email string) (Account, error)
	GetByPhone(ctx context.Context, phone string) (Account, error)
	Update(ctx context.Context, account Account) error
	Delete(ctx context.Context, id uuid.UUID) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
}

// TokenRepository defines the token repository interface
type TokenRepository interface {
	Create(ctx context.Context, token Token) error
	GetByToken(ctx context.Context, token string) (Token, error)
	GetByAccountAndType(ctx context.Context, accountID uuid.UUID, tokenType TokenType) ([]Token, error)
	Update(ctx context.Context, token Token) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
}

// SessionStore defines the session storage interface
type SessionStore interface {
	Store(ctx context.Context, session Session) error
	Get(ctx context.Context, token string) (Session, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (Session, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	DeleteByToken(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

// PasswordHasher defines the password hashing interface
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// TokenGenerator defines the token generation interface
type TokenGenerator interface {
	GenerateToken() string
	GenerateSecureToken() string
}

// EmailSender defines the email sending interface
type EmailSender interface {
	SendVerificationEmail(email, token string) error
	SendPasswordResetEmail(email, token string) error
	SendWelcomeEmail(email string) error
}

// SMSSender defines the SMS sending interface
type SMSSender interface {
	SendVerificationSMS(phone, code string) error
}

// AccountUpdates represents fields that can be updated on an account
type AccountUpdates struct {
	Email         *string
	Phone         *string
	EmailVerified *bool
	PhoneVerified *bool
}

// AuthConfig represents authentication configuration
type AuthConfig interface {
	GetJWTSecret() string
	GetJWTExpiration() time.Duration
	GetRefreshTokenExpiration() time.Duration
	GetVerificationTokenExpiration() time.Duration
	GetPasswordResetTokenExpiration() time.Duration
	GetBcryptCost() int
	IsEmailVerificationRequired() bool
	IsPhoneVerificationRequired() bool
}