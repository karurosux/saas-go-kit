package authmodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/auth/interface"
	"github.com/google/uuid"
)

// Token represents an authentication or verification token
type Token struct {
	ID        uuid.UUID                `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AccountID uuid.UUID                `json:"account_id" gorm:"type:uuid;not null;index"`
	Token     string                   `json:"token" gorm:"uniqueIndex;not null"`
	Type      authinterface.TokenType  `json:"type" gorm:"not null;index"`
	Used      bool                     `json:"used" gorm:"default:false;index"`
	ExpiresAt time.Time                `json:"expires_at" gorm:"not null;index"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// GetID returns the token ID
func (t *Token) GetID() uuid.UUID {
	return t.ID
}

// GetAccountID returns the associated account ID
func (t *Token) GetAccountID() uuid.UUID {
	return t.AccountID
}

// GetToken returns the token string
func (t *Token) GetToken() string {
	return t.Token
}

// GetType returns the token type
func (t *Token) GetType() authinterface.TokenType {
	return t.Type
}

// GetUsed returns whether the token has been used
func (t *Token) GetUsed() bool {
	return t.Used
}

// GetExpiresAt returns the expiration time
func (t *Token) GetExpiresAt() time.Time {
	return t.ExpiresAt
}

// GetCreatedAt returns creation time
func (t *Token) GetCreatedAt() time.Time {
	return t.CreatedAt
}

// GetUpdatedAt returns last update time
func (t *Token) GetUpdatedAt() time.Time {
	return t.UpdatedAt
}

// SetUsed marks the token as used
func (t *Token) SetUsed(used bool) {
	t.Used = used
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}