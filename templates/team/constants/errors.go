package teamconstants

// Error messages
const (
	ErrUserNotFound           = "user not found"
	ErrMemberNotFound         = "team member not found"
	ErrInvitationNotFound     = "invitation not found"
	ErrInvalidToken           = "invalid invitation token"
	ErrTokenExpired           = "invitation token expired"
	ErrTokenAlreadyUsed       = "invitation token already used"
	ErrAlreadyMember          = "user is already a team member"
	ErrCannotRemoveOwner      = "cannot remove team owner"
	ErrCannotChangeOwnerRole  = "cannot change owner role"
	ErrMustHaveOwner          = "team must have at least one owner"
	ErrInsufficientPermission = "insufficient permission for this operation"
	ErrTeamLimitReached       = "team member limit reached"
	ErrInvalidRole            = "invalid member role"
	ErrSelfRemoval            = "cannot remove yourself from the team"
	ErrSelfRoleChange         = "cannot change your own role"
)