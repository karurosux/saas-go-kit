package teammodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/team/interface"
	"github.com/google/uuid"
)

// TeamMember represents a team member
type TeamMember struct {
	ID          uuid.UUID                 `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AccountID   uuid.UUID                 `json:"account_id" gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID                 `json:"user_id" gorm:"type:uuid;not null;index"`
	User        *User                     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Role        teaminterface.MemberRole  `json:"role" gorm:"not null"`
	IsActive    bool                      `json:"is_active" gorm:"default:true"`
	IsPending   bool                      `json:"is_pending" gorm:"default:true"`
	InvitedAt   time.Time                 `json:"invited_at" gorm:"not null"`
	AcceptedAt  *time.Time                `json:"accepted_at,omitempty"`
	InvitedByID uuid.UUID                 `json:"invited_by_id" gorm:"type:uuid;not null"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// GetID returns the member ID
func (tm *TeamMember) GetID() uuid.UUID {
	return tm.ID
}

// GetAccountID returns the account ID
func (tm *TeamMember) GetAccountID() uuid.UUID {
	return tm.AccountID
}

// GetUserID returns the user ID
func (tm *TeamMember) GetUserID() uuid.UUID {
	return tm.UserID
}

// GetUser returns the user
func (tm *TeamMember) GetUser() teaminterface.User {
	if tm.User == nil {
		return nil
	}
	return tm.User
}

// GetRole returns the member role
func (tm *TeamMember) GetRole() teaminterface.MemberRole {
	return tm.Role
}

// GetIsActive returns if member is active
func (tm *TeamMember) GetIsActive() bool {
	return tm.IsActive
}

// GetIsPending returns if member is pending
func (tm *TeamMember) GetIsPending() bool {
	return tm.IsPending
}

// GetInvitedAt returns when member was invited
func (tm *TeamMember) GetInvitedAt() time.Time {
	return tm.InvitedAt
}

// GetAcceptedAt returns when invitation was accepted
func (tm *TeamMember) GetAcceptedAt() *time.Time {
	return tm.AcceptedAt
}

// GetInvitedByID returns who invited the member
func (tm *TeamMember) GetInvitedByID() uuid.UUID {
	return tm.InvitedByID
}

// GetCreatedAt returns creation time
func (tm *TeamMember) GetCreatedAt() time.Time {
	return tm.CreatedAt
}

// GetUpdatedAt returns last update time
func (tm *TeamMember) GetUpdatedAt() time.Time {
	return tm.UpdatedAt
}

// SetRole sets the member role
func (tm *TeamMember) SetRole(role teaminterface.MemberRole) {
	tm.Role = role
}

// SetIsActive sets if member is active
func (tm *TeamMember) SetIsActive(active bool) {
	tm.IsActive = active
}

// SetAcceptedAt sets when invitation was accepted
func (tm *TeamMember) SetAcceptedAt(acceptedAt *time.Time) {
	tm.AcceptedAt = acceptedAt
	if acceptedAt != nil {
		tm.IsPending = false
	}
}