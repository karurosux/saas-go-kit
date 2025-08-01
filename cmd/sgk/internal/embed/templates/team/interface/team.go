package teaminterface

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// MemberRole represents team member roles
type MemberRole string

const (
	RoleOwner   MemberRole = "owner"
	RoleAdmin   MemberRole = "admin"
	RoleMember  MemberRole = "member"
	RoleViewer  MemberRole = "viewer"
)

// User represents a user
type User interface {
	GetID() uuid.UUID
	GetEmail() string
	GetFirstName() string
	GetLastName() string
	GetFullName() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// TeamMember represents a team member
type TeamMember interface {
	GetID() uuid.UUID
	GetAccountID() uuid.UUID
	GetUserID() uuid.UUID
	GetUser() User
	GetRole() MemberRole
	GetIsActive() bool
	GetIsPending() bool
	GetInvitedAt() time.Time
	GetAcceptedAt() *time.Time
	GetInvitedByID() uuid.UUID
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetRole(role MemberRole)
	SetIsActive(active bool)
	SetAcceptedAt(acceptedAt *time.Time)
}

// InvitationToken represents a team invitation token
type InvitationToken interface {
	GetID() uuid.UUID
	GetToken() string
	GetMemberID() uuid.UUID
	GetExpiresAt() time.Time
	GetUsed() bool
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	IsExpired() bool
	SetUsed(used bool)
}

// TeamStats represents team statistics
type TeamStats interface {
	GetTotalMembers() int64
	GetActiveMembers() int64
	GetPendingInvitations() int64
	GetMembersByRole() map[MemberRole]int64
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id uuid.UUID) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// TeamMemberRepository defines the interface for team member data access
type TeamMemberRepository interface {
	Create(ctx context.Context, member TeamMember) error
	GetByID(ctx context.Context, id uuid.UUID) (TeamMember, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]TeamMember, error)
	GetByUserAndAccount(ctx context.Context, userID, accountID uuid.UUID) (TeamMember, error)
	Update(ctx context.Context, member TeamMember) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByAccountID(ctx context.Context, accountID uuid.UUID) (int64, error)
	GetTeamStats(ctx context.Context, accountID uuid.UUID) (TeamStats, error)
}

// InvitationTokenRepository defines the interface for invitation token data access
type InvitationTokenRepository interface {
	Create(ctx context.Context, token InvitationToken) error
	GetByToken(ctx context.Context, token string) (InvitationToken, error)
	GetByMemberID(ctx context.Context, memberID uuid.UUID) (InvitationToken, error)
	Update(ctx context.Context, token InvitationToken) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// TeamService defines the interface for team management business logic
type TeamService interface {
	// Member management
	ListMembers(ctx context.Context, accountID uuid.UUID) ([]TeamMember, error)
	GetMember(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) (TeamMember, error)
	InviteMember(ctx context.Context, req InviteMemberRequest) (TeamMember, error)
	UpdateMemberRole(ctx context.Context, req UpdateMemberRoleRequest) error
	RemoveMember(ctx context.Context, req RemoveMemberRequest) error
	
	// Invitation management
	AcceptInvitation(ctx context.Context, token string, acceptReq AcceptInvitationRequest) error
	ResendInvitation(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) error
	CancelInvitation(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) error
	
	// Team statistics
	GetTeamStats(ctx context.Context, accountID uuid.UUID) (TeamStats, error)
	
	// Permission checks
	CheckPermission(ctx context.Context, accountID uuid.UUID, userID uuid.UUID, permission string) (bool, error)
	GetMemberRole(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (MemberRole, error)
	IsOwner(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) bool
	IsAdmin(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) bool
	CanManageTeam(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) bool
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendTeamInvitation(ctx context.Context, email, inviterName, teamName, role, token string, expiresAt time.Time) error
	SendRoleChanged(ctx context.Context, email, userName, teamName, oldRole, newRole, changedBy string) error
	SendMemberRemoved(ctx context.Context, email, userName, teamName, removedBy string) error
}

// EmailSender defines the interface for sending emails
type EmailSender interface {
	SendInvitationEmail(email, inviterName, teamName, role, inviteLink string, expiresAt time.Time) error
	SendRoleChangedEmail(email, userName, teamName, oldRole, newRole, changedBy string) error
	SendMemberRemovedEmail(email, userName, teamName, removedBy string) error
}

// Request DTOs

type InviteMemberRequest interface {
	GetAccountID() uuid.UUID
	GetInviterID() uuid.UUID
	GetEmail() string
	GetRole() MemberRole
	Validate() error
}

type UpdateMemberRoleRequest interface {
	GetAccountID() uuid.UUID
	GetMemberID() uuid.UUID
	GetNewRole() MemberRole
	GetUpdatedByID() uuid.UUID
	Validate() error
}

type RemoveMemberRequest interface {
	GetAccountID() uuid.UUID
	GetMemberID() uuid.UUID
	GetRemovedByID() uuid.UUID
	Validate() error
}

type AcceptInvitationRequest interface {
	GetPassword() string
	GetFirstName() string
	GetLastName() string
	Validate() error
}