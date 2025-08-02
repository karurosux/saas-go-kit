package authmodel

import (
	"time"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	"github.com/google/uuid"
)

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

func (t *Token) GetID() uuid.UUID {
	return t.ID
}

func (t *Token) GetAccountID() uuid.UUID {
	return t.AccountID
}

func (t *Token) GetToken() string {
	return t.Token
}

func (t *Token) GetType() authinterface.TokenType {
	return t.Type
}

func (t *Token) GetUsed() bool {
	return t.Used
}

func (t *Token) GetExpiresAt() time.Time {
	return t.ExpiresAt
}

func (t *Token) GetCreatedAt() time.Time {
	return t.CreatedAt
}

func (t *Token) GetUpdatedAt() time.Time {
	return t.UpdatedAt
}

func (t *Token) SetUsed(used bool) {
	t.Used = used
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}