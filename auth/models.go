package auth

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// DefaultAccount provides a basic implementation of the Account interface
type DefaultAccount struct {
	ID                   uuid.UUID              `json:"id"`
	Email                string                 `json:"email"`
	PasswordHash         string                 `json:"-"`
	CompanyName          string                 `json:"company_name,omitempty"`
	Phone                string                 `json:"phone,omitempty"`
	Active               bool                   `json:"is_active"`
	EmailVerified        bool                   `json:"email_verified"`
	EmailVerifiedAt      *time.Time             `json:"email_verified_at,omitempty"`
	DeactivatedAt        *time.Time             `json:"deactivated_at,omitempty"`
	ScheduledDeletionAt  *time.Time             `json:"scheduled_deletion_at,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (DefaultAccount) TableName() string {
	return "auth.users"
}

// GetID returns the account ID
func (a *DefaultAccount) GetID() uuid.UUID {
	return a.ID
}

// GetEmail returns the account email
func (a *DefaultAccount) GetEmail() string {
	return a.Email
}

// SetPassword hashes and sets the password
func (a *DefaultAccount) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password
func (a *DefaultAccount) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password))
	return err == nil
}

// IsActive returns whether the account is active
func (a *DefaultAccount) IsActive() bool {
	return a.Active && a.DeactivatedAt == nil
}

// IsEmailVerified returns whether the email is verified
func (a *DefaultAccount) IsEmailVerified() bool {
	return a.EmailVerified
}

// GetMetadata returns the account metadata
func (a *DefaultAccount) GetMetadata() map[string]interface{} {
	if a.Metadata == nil {
		return make(map[string]interface{})
	}
	return a.Metadata
}

// DefaultUser provides a basic implementation of the User interface
type DefaultUser struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetID returns the user ID
func (u *DefaultUser) GetID() uuid.UUID {
	return u.ID
}

// GetEmail returns the user email
func (u *DefaultUser) GetEmail() string {
	return u.Email
}

// GetFirstName returns the user's first name
func (u *DefaultUser) GetFirstName() string {
	return u.FirstName
}

// GetLastName returns the user's last name
func (u *DefaultUser) GetLastName() string {
	return u.LastName
}

// SetPassword hashes and sets the password
func (u *DefaultUser) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password
func (u *DefaultUser) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// DefaultVerificationToken provides a basic implementation of VerificationToken
type DefaultVerificationToken struct {
	ID        uuid.UUID  `json:"id"`
	AccountID uuid.UUID  `json:"account_id"`
	Token     string     `json:"token"`
	Type      TokenType  `json:"type"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// TableName specifies the table name for GORM
func (DefaultVerificationToken) TableName() string {
	return "auth.verification_tokens"
}

// GetToken returns the token string
func (t *DefaultVerificationToken) GetToken() string {
	return t.Token
}

// GetType returns the token type
func (t *DefaultVerificationToken) GetType() TokenType {
	return t.Type
}

// GetAccountID returns the associated account ID
func (t *DefaultVerificationToken) GetAccountID() uuid.UUID {
	return t.AccountID
}

// GetExpiresAt returns the expiration time
func (t *DefaultVerificationToken) GetExpiresAt() time.Time {
	return t.ExpiresAt
}

// IsValid checks if the token is valid
func (t *DefaultVerificationToken) IsValid() bool {
	return t.UsedAt == nil && time.Now().Before(t.ExpiresAt)
}

// MarkAsUsed marks the token as used
func (t *DefaultVerificationToken) MarkAsUsed() {
	now := time.Now()
	t.UsedAt = &now
}

// NewAccount creates a new default account
func NewAccount(email string) *DefaultAccount {
	now := time.Now()
	return &DefaultAccount{
		ID:            uuid.New(),
		Email:         email,
		Active:        true,
		EmailVerified: false,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// NewUser creates a new default user
func NewUser(email, firstName, lastName string) *DefaultUser {
	now := time.Now()
	return &DefaultUser{
		ID:        uuid.New(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewVerificationToken creates a new verification token
func NewVerificationToken(accountID uuid.UUID, tokenType TokenType, duration time.Duration) *DefaultVerificationToken {
	return &DefaultVerificationToken{
		ID:        uuid.New(),
		AccountID: accountID,
		Token:     generateSecureToken(),
		Type:      tokenType,
		ExpiresAt: time.Now().Add(duration),
		CreatedAt: time.Now(),
	}
}

// generateSecureToken generates a secure random token
func generateSecureToken() string {
	return uuid.New().String() + uuid.New().String()
}