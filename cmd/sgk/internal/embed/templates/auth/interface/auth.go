package authinterface

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

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

type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePhoneVerification TokenType = "phone_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
	TokenTypeRefresh           TokenType = "refresh"
)

type Session interface {
	GetUserID() uuid.UUID
	GetToken() string
	GetRefreshToken() string
	GetExpiresAt() time.Time
	GetRefreshExpiresAt() time.Time
	IsExpired() bool
	IsRefreshExpired() bool
}

type LoginRequest interface {
	GetEmail() string
	GetPassword() string
	Validate() error
}

type RegisterRequest interface {
	GetEmail() string
	GetPhone() string
	GetPassword() string
	Validate() error
}

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (Account, error)
	Login(ctx context.Context, req LoginRequest) (Session, error)
	RefreshSession(ctx context.Context, refreshToken string) (Session, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	
	SendEmailVerification(ctx context.Context, accountID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error
	
	SendPhoneVerification(ctx context.Context, accountID uuid.UUID) error
	VerifyPhone(ctx context.Context, accountID uuid.UUID, code string) error
	
	SendPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, accountID uuid.UUID, oldPassword, newPassword string) error
	
	GetAccount(ctx context.Context, accountID uuid.UUID) (Account, error)
	GetAccountByEmail(ctx context.Context, email string) (Account, error)
	UpdateAccount(ctx context.Context, accountID uuid.UUID, updates AccountUpdates) (Account, error)
	
	ValidateSession(ctx context.Context, token string) (Account, error)
}

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

type TokenRepository interface {
	Create(ctx context.Context, token Token) error
	GetByToken(ctx context.Context, token string) (Token, error)
	GetByAccountAndType(ctx context.Context, accountID uuid.UUID, tokenType TokenType) ([]Token, error)
	Update(ctx context.Context, token Token) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
}

type SessionStore interface {
	Store(ctx context.Context, session Session) error
	Get(ctx context.Context, token string) (Session, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (Session, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	DeleteByToken(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

type TokenGenerator interface {
	GenerateToken() string
	GenerateSecureToken() string
}

type EmailSender interface {
	SendVerificationEmail(email, token string) error
	SendPasswordResetEmail(email, token string) error
	SendWelcomeEmail(email string) error
}

type SMSSender interface {
	SendVerificationSMS(phone, code string) error
}

type AccountUpdates struct {
	Email         *string
	Phone         *string
	EmailVerified *bool
	PhoneVerified *bool
}

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