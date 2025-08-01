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

// GetID returns the account ID
func (a *Account) GetID() uuid.UUID {
	return a.ID
}

// GetEmail returns the account email
func (a *Account) GetEmail() string {
	return a.Email
}

// GetPhone returns the account phone
func (a *Account) GetPhone() string {
	return a.Phone
}

// GetPasswordHash returns the password hash
func (a *Account) GetPasswordHash() string {
	return a.PasswordHash
}

// GetEmailVerified returns whether email is verified
func (a *Account) GetEmailVerified() bool {
	return a.EmailVerified
}

// GetPhoneVerified returns whether phone is verified
func (a *Account) GetPhoneVerified() bool {
	return a.PhoneVerified
}

// GetCreatedAt returns creation time
func (a *Account) GetCreatedAt() time.Time {
	return a.CreatedAt
}

// GetUpdatedAt returns last update time
func (a *Account) GetUpdatedAt() time.Time {
	return a.UpdatedAt
}

// SetPasswordHash sets the password hash
func (a *Account) SetPasswordHash(hash string) {
	a.PasswordHash = hash
}

// SetEmailVerified sets email verification status
func (a *Account) SetEmailVerified(verified bool) {
	a.EmailVerified = verified
}

// SetPhoneVerified sets phone verification status
func (a *Account) SetPhoneVerified(verified bool) {
	a.PhoneVerified = verified
}