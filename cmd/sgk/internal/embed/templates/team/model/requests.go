package teammodel

import (
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/team/constants"
	"{{.Project.GoModule}}/internal/team/interface"
	"github.com/google/uuid"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// InviteMemberRequest represents a request to invite a team member
type InviteMemberRequest struct {
	AccountID uuid.UUID                `json:"account_id"`
	InviterID uuid.UUID                `json:"inviter_id"`
	Email     string                   `json:"email" validate:"required,email"`
	Role      teaminterface.MemberRole `json:"role" validate:"required"`
}

// GetAccountID returns the account ID
func (r *InviteMemberRequest) GetAccountID() uuid.UUID {
	return r.AccountID
}

// GetInviterID returns the inviter ID
func (r *InviteMemberRequest) GetInviterID() uuid.UUID {
	return r.InviterID
}

// GetEmail returns the email
func (r *InviteMemberRequest) GetEmail() string {
	return r.Email
}

// GetRole returns the role
func (r *InviteMemberRequest) GetRole() teaminterface.MemberRole {
	return r.Role
}

// Validate validates the request
func (r *InviteMemberRequest) Validate() error {
	if r.AccountID == uuid.Nil {
		return core.BadRequest("account ID is required")
	}
	if r.InviterID == uuid.Nil {
		return core.BadRequest("inviter ID is required")
	}
	if r.Email == "" {
		return core.BadRequest("email is required")
	}
	if !emailRegex.MatchString(r.Email) {
		return core.BadRequest("invalid email format")
	}
	if !isValidRole(r.Role) {
		return core.BadRequest(teamconstants.ErrInvalidRole)
	}
	return nil
}

// UpdateMemberRoleRequest represents a request to update member role
type UpdateMemberRoleRequest struct {
	AccountID   uuid.UUID                `json:"account_id"`
	MemberID    uuid.UUID                `json:"member_id"`
	NewRole     teaminterface.MemberRole `json:"new_role" validate:"required"`
	UpdatedByID uuid.UUID                `json:"updated_by_id"`
}

// GetAccountID returns the account ID
func (r *UpdateMemberRoleRequest) GetAccountID() uuid.UUID {
	return r.AccountID
}

// GetMemberID returns the member ID
func (r *UpdateMemberRoleRequest) GetMemberID() uuid.UUID {
	return r.MemberID
}

// GetNewRole returns the new role
func (r *UpdateMemberRoleRequest) GetNewRole() teaminterface.MemberRole {
	return r.NewRole
}

// GetUpdatedByID returns who is updating
func (r *UpdateMemberRoleRequest) GetUpdatedByID() uuid.UUID {
	return r.UpdatedByID
}

// Validate validates the request
func (r *UpdateMemberRoleRequest) Validate() error {
	if r.AccountID == uuid.Nil {
		return core.BadRequest("account ID is required")
	}
	if r.MemberID == uuid.Nil {
		return core.BadRequest("member ID is required")
	}
	if r.UpdatedByID == uuid.Nil {
		return core.BadRequest("updater ID is required")
	}
	if !isValidRole(r.NewRole) {
		return core.BadRequest(teamconstants.ErrInvalidRole)
	}
	return nil
}

// RemoveMemberRequest represents a request to remove a member
type RemoveMemberRequest struct {
	AccountID   uuid.UUID `json:"account_id"`
	MemberID    uuid.UUID `json:"member_id"`
	RemovedByID uuid.UUID `json:"removed_by_id"`
}

// GetAccountID returns the account ID
func (r *RemoveMemberRequest) GetAccountID() uuid.UUID {
	return r.AccountID
}

// GetMemberID returns the member ID
func (r *RemoveMemberRequest) GetMemberID() uuid.UUID {
	return r.MemberID
}

// GetRemovedByID returns who is removing
func (r *RemoveMemberRequest) GetRemovedByID() uuid.UUID {
	return r.RemovedByID
}

// Validate validates the request
func (r *RemoveMemberRequest) Validate() error {
	if r.AccountID == uuid.Nil {
		return core.BadRequest("account ID is required")
	}
	if r.MemberID == uuid.Nil {
		return core.BadRequest("member ID is required")
	}
	if r.RemovedByID == uuid.Nil {
		return core.BadRequest("remover ID is required")
	}
	return nil
}

// AcceptInvitationRequest represents a request to accept invitation
type AcceptInvitationRequest struct {
	Password  string `json:"password,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// GetPassword returns the password
func (r *AcceptInvitationRequest) GetPassword() string {
	return r.Password
}

// GetFirstName returns the first name
func (r *AcceptInvitationRequest) GetFirstName() string {
	return r.FirstName
}

// GetLastName returns the last name
func (r *AcceptInvitationRequest) GetLastName() string {
	return r.LastName
}

// Validate validates the request
func (r *AcceptInvitationRequest) Validate() error {
	// Password is only required for new users
	// This will be checked in the service layer
	return nil
}

// Helper function to validate role
func isValidRole(role teaminterface.MemberRole) bool {
	switch role {
	case teaminterface.RoleOwner, teaminterface.RoleAdmin, 
	     teaminterface.RoleMember, teaminterface.RoleViewer:
		return true
	default:
		return false
	}
}