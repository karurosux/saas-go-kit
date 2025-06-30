package team

import "errors"

// Repository errors
var (
	ErrUserNotFound            = errors.New("user not found")
	ErrTeamMemberNotFound      = errors.New("team member not found")
	ErrInvitationTokenNotFound = errors.New("invitation token not found")
)

// Service errors
var (
	ErrUserAlreadyExists       = errors.New("user with this email already exists")
	ErrTeamMemberAlreadyExists = errors.New("user is already a team member")
	ErrInvalidRole             = errors.New("invalid member role")
	ErrInvitationExpired       = errors.New("invitation has expired")
	ErrInvitationAlreadyUsed   = errors.New("invitation has already been used")
	ErrCannotRemoveOwner       = errors.New("cannot remove the owner from the team")
	ErrCannotChangeOwnerRole   = errors.New("cannot change the role of the owner")
	ErrInsufficientPermissions = errors.New("insufficient permissions for this action")
	ErrCannotInviteSelf        = errors.New("cannot invite yourself to the team")
	ErrMaxMembersReached       = errors.New("maximum number of team members reached")
)