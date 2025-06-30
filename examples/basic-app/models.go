package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/saas-go-kit/auth-go"
	"gorm.io/gorm"
)

// AccountModel is the GORM model for accounts
type AccountModel struct {
	ID                  uuid.UUID              `gorm:"type:uuid;primary_key"`
	Email               string                 `gorm:"uniqueIndex;not null"`
	PasswordHash        string                 `gorm:"not null"`
	CompanyName         string
	Phone               string
	IsActive            bool                   `gorm:"default:true"`
	EmailVerified       bool                   `gorm:"default:false"`
	EmailVerifiedAt     *time.Time
	DeactivatedAt       *time.Time
	ScheduledDeletionAt *time.Time
	Metadata            map[string]interface{} `gorm:"serializer:json"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt         `gorm:"index"`
}

// TableName specifies the table name
func (AccountModel) TableName() string {
	return "accounts"
}

// ToAuthAccount converts to auth.Account interface
func (a *AccountModel) ToAuthAccount() auth.Account {
	return &auth.DefaultAccount{
		ID:                  a.ID,
		Email:               a.Email,
		PasswordHash:        a.PasswordHash,
		CompanyName:         a.CompanyName,
		Phone:               a.Phone,
		Active:              a.IsActive,
		EmailVerified:       a.EmailVerified,
		EmailVerifiedAt:     a.EmailVerifiedAt,
		DeactivatedAt:       a.DeactivatedAt,
		ScheduledDeletionAt: a.ScheduledDeletionAt,
		Metadata:            a.Metadata,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}
}

// FromAuthAccount updates from auth.Account interface
func (a *AccountModel) FromAuthAccount(account auth.Account) {
	if defAccount, ok := account.(*auth.DefaultAccount); ok {
		a.ID = defAccount.ID
		a.Email = defAccount.Email
		a.PasswordHash = defAccount.PasswordHash
		a.CompanyName = defAccount.CompanyName
		a.Phone = defAccount.Phone
		a.IsActive = defAccount.Active
		a.EmailVerified = defAccount.EmailVerified
		a.EmailVerifiedAt = defAccount.EmailVerifiedAt
		a.DeactivatedAt = defAccount.DeactivatedAt
		a.ScheduledDeletionAt = defAccount.ScheduledDeletionAt
		a.Metadata = defAccount.Metadata
		a.CreatedAt = defAccount.CreatedAt
		a.UpdatedAt = defAccount.UpdatedAt
	}
}

// TokenModel is the GORM model for tokens
type TokenModel struct {
	ID        uuid.UUID          `gorm:"type:uuid;primary_key"`
	AccountID uuid.UUID          `gorm:"type:uuid;not null;index"`
	Token     string             `gorm:"uniqueIndex;not null"`
	Type      auth.TokenType     `gorm:"not null"`
	ExpiresAt time.Time          `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt     `gorm:"index"`
}

// TableName specifies the table name
func (TokenModel) TableName() string {
	return "tokens"
}

// ToAuthToken converts to auth.VerificationToken interface
func (t *TokenModel) ToAuthToken() auth.VerificationToken {
	return &auth.DefaultVerificationToken{
		ID:        t.ID,
		AccountID: t.AccountID,
		Token:     t.Token,
		Type:      t.Type,
		ExpiresAt: t.ExpiresAt,
		UsedAt:    t.UsedAt,
		CreatedAt: t.CreatedAt,
	}
}

// FromAuthToken updates from auth.VerificationToken interface
func (t *TokenModel) FromAuthToken(token auth.VerificationToken) {
	if defToken, ok := token.(*auth.DefaultVerificationToken); ok {
		t.ID = defToken.ID
		t.AccountID = defToken.AccountID
		t.Token = defToken.Token
		t.Type = defToken.Type
		t.ExpiresAt = defToken.ExpiresAt
		t.UsedAt = defToken.UsedAt
		t.CreatedAt = defToken.CreatedAt
	}
}