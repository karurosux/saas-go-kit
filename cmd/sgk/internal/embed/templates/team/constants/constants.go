package teamconstants

import "time"

// Default values
const (
	// DefaultInvitationExpiration is the default expiration time for invitations
	DefaultInvitationExpiration = 7 * 24 * time.Hour // 7 days
	
	// DefaultMaxTeamSize is the default maximum team size
	DefaultMaxTeamSize = 100
	
	// MinPasswordLength is the minimum password length for new users
	MinPasswordLength = 8
)

// Context keys
const (
	// ContextKeyTeamMember stores the team member in context
	ContextKeyTeamMember = "team_member"
	
	// ContextKeyMemberRole stores the member role in context
	ContextKeyMemberRole = "member_role"
	
	// ContextKeyAccountID stores the account ID in context
	ContextKeyAccountID = "account_id"
)