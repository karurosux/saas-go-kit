package teammodel

import (
	"time"
	
	"github.com/google/uuid"
)

// InvitationToken represents a team invitation token
type InvitationToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	MemberID  uuid.UUID `json:"member_id" gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	Used      bool      `json:"used" gorm:"default:false;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetID returns the token ID
func (it *InvitationToken) GetID() uuid.UUID {
	return it.ID
}

// GetToken returns the token string
func (it *InvitationToken) GetToken() string {
	return it.Token
}

// GetMemberID returns the associated member ID
func (it *InvitationToken) GetMemberID() uuid.UUID {
	return it.MemberID
}

// GetExpiresAt returns the expiration time
func (it *InvitationToken) GetExpiresAt() time.Time {
	return it.ExpiresAt
}

// GetUsed returns whether the token has been used
func (it *InvitationToken) GetUsed() bool {
	return it.Used
}

// GetCreatedAt returns creation time
func (it *InvitationToken) GetCreatedAt() time.Time {
	return it.CreatedAt
}

// GetUpdatedAt returns last update time
func (it *InvitationToken) GetUpdatedAt() time.Time {
	return it.UpdatedAt
}

// IsExpired checks if the token has expired
func (it *InvitationToken) IsExpired() bool {
	return time.Now().After(it.ExpiresAt)
}

// SetUsed marks the token as used
func (it *InvitationToken) SetUsed(used bool) {
	it.Used = used
}