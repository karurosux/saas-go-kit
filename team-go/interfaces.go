package team

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TeamMemberRepository defines the interface for team member data access
type TeamMemberRepository interface {
	Create(ctx context.Context, member *TeamMember) error
	FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*TeamMember, error)
	FindByAccountID(ctx context.Context, accountID uuid.UUID) ([]TeamMember, error)
	FindByUserAndAccount(ctx context.Context, userID, accountID uuid.UUID) (*TeamMember, error)
	Update(ctx context.Context, member *TeamMember) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByAccountID(ctx context.Context, accountID uuid.UUID) (int64, error)
	GetTeamStats(ctx context.Context, accountID uuid.UUID) (*TeamStats, error)
}

// InvitationTokenRepository defines the interface for invitation token data access
type InvitationTokenRepository interface {
	Create(ctx context.Context, token *InvitationToken) error
	FindByToken(ctx context.Context, token string) (*InvitationToken, error)
	FindByMemberID(ctx context.Context, memberID uuid.UUID) (*InvitationToken, error)
	Update(ctx context.Context, token *InvitationToken) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// TeamService defines the interface for team management business logic
type TeamService interface {
	// Member management
	ListMembers(ctx context.Context, accountID uuid.UUID) ([]TeamMember, error)
	GetMember(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) (*TeamMember, error)
	InviteMember(ctx context.Context, req *InviteMemberRequest) (*TeamMember, error)
	UpdateMemberRole(ctx context.Context, req *UpdateMemberRoleRequest) error
	RemoveMember(ctx context.Context, req *RemoveMemberRequest) error
	
	// Invitation management
	AcceptInvitation(ctx context.Context, token string) error
	ResendInvitation(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) error
	CancelInvitation(ctx context.Context, accountID uuid.UUID, memberID uuid.UUID) error
	
	// Team statistics
	GetTeamStats(ctx context.Context, accountID uuid.UUID) (*TeamStats, error)
	
	// Permission checks
	CheckPermission(ctx context.Context, accountID uuid.UUID, userID uuid.UUID, permission string) (bool, error)
	GetMemberRole(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) (MemberRole, error)
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendTeamInvitation(ctx context.Context, req *TeamInvitationNotification) error
	SendRoleChanged(ctx context.Context, req *RoleChangedNotification) error
	SendMemberRemoved(ctx context.Context, req *MemberRemovedNotification) error
}

// UsageService defines the interface for tracking team usage
type UsageService interface {
	TrackMemberAdded(ctx context.Context, accountID uuid.UUID) error
	TrackMemberRemoved(ctx context.Context, accountID uuid.UUID) error
	CanAddMember(ctx context.Context, accountID uuid.UUID) (bool, string, error)
}

// Request/Response DTOs

type InviteMemberRequest struct {
	AccountID uuid.UUID  `json:"account_id"`
	InviterID uuid.UUID  `json:"inviter_id"`
	Email     string     `json:"email" validate:"required,email"`
	Role      MemberRole `json:"role" validate:"required"`
}

type UpdateMemberRoleRequest struct {
	AccountID   uuid.UUID  `json:"account_id"`
	MemberID    uuid.UUID  `json:"member_id"`
	NewRole     MemberRole `json:"new_role" validate:"required"`
	UpdatedByID uuid.UUID  `json:"updated_by_id"`
}

type RemoveMemberRequest struct {
	AccountID   uuid.UUID `json:"account_id"`
	MemberID    uuid.UUID `json:"member_id"`
	RemovedByID uuid.UUID `json:"removed_by_id"`
}

type AcceptInvitationRequest struct {
	Token     string `json:"token" validate:"required"`
	Password  string `json:"password,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Notification DTOs

type TeamInvitationNotification struct {
	Email       string `json:"email"`
	InviterName string `json:"inviter_name"`
	TeamName    string `json:"team_name"`
	Role        string `json:"role"`
	Token       string `json:"token"`
	ExpiresAt   string `json:"expires_at"`
}

type RoleChangedNotification struct {
	Email       string `json:"email"`
	UserName    string `json:"user_name"`
	TeamName    string `json:"team_name"`
	OldRole     string `json:"old_role"`
	NewRole     string `json:"new_role"`
	ChangedBy   string `json:"changed_by"`
}

type MemberRemovedNotification struct {
	Email     string `json:"email"`
	UserName  string `json:"user_name"`
	TeamName  string `json:"team_name"`
	RemovedBy string `json:"removed_by"`
}

// Response DTOs

type TeamMemberResponse struct {
	ID         uuid.UUID  `json:"id"`
	User       UserResponse `json:"user"`
	Role       MemberRole `json:"role"`
	IsActive   bool       `json:"is_active"`
	IsPending  bool       `json:"is_pending"`
	InvitedAt  string     `json:"invited_at"`
	AcceptedAt *string    `json:"accepted_at,omitempty"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	FullName  string    `json:"full_name"`
}

type InvitationResponse struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Role      MemberRole `json:"role"`
	ExpiresAt string     `json:"expires_at"`
	IsExpired bool       `json:"is_expired"`
}

type PermissionCheckResponse struct {
	HasPermission bool   `json:"has_permission"`
	Role          string `json:"role"`
	Reason        string `json:"reason,omitempty"`
}