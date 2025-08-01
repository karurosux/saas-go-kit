package authmodel

import (
	"time"
	
	"github.com/google/uuid"
)

// Account represents a user account
type Account struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string    `json:"email" gorm:"uniqueIndex;not null"`
	Phone         string    `json:"phone,omitempty" gorm:"uniqueIndex"`
	PasswordHash  string    `json:"-" gorm:"not null"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	PhoneVerified bool      `json:"phone_verified" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (a *Account) GetID() uuid.UUID {
	return a.ID
}

func (a *Account) GetEmail() string {
	return a.Email
}

func (a *Account) GetPhone() string {
	return a.Phone
}

func (a *Account) GetPasswordHash() string {
	return a.PasswordHash
}

func (a *Account) GetEmailVerified() bool {
	return a.EmailVerified
}

func (a *Account) GetPhoneVerified() bool {
	return a.PhoneVerified
}

func (a *Account) GetCreatedAt() time.Time {
	return a.CreatedAt
}

func (a *Account) GetUpdatedAt() time.Time {
	return a.UpdatedAt
}

func (a *Account) SetPasswordHash(hash string) {
	a.PasswordHash = hash
}

func (a *Account) SetEmailVerified(verified bool) {
	a.EmailVerified = verified
}

func (a *Account) SetPhoneVerified(verified bool) {
	a.PhoneVerified = verified
}