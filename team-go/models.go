package team

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// User represents an individual user in the system
type User struct {
	BaseModel
	Email        string       `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string       `gorm:"not null" json:"-"`
	FirstName    string       `json:"first_name"`
	LastName     string       `json:"last_name"`
	IsActive     bool         `gorm:"default:true" json:"is_active"`
	TeamMembers  []TeamMember `json:"team_members,omitempty"`
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	return u.FirstName + " " + u.LastName
}

// TeamMember represents a user's membership in a team/account
type TeamMember struct {
	BaseModel
	AccountID  uuid.UUID  `gorm:"not null" json:"account_id"`
	UserID     uuid.UUID  `gorm:"not null" json:"user_id"`
	User       User       `json:"user,omitempty"`
	Role       MemberRole `gorm:"not null" json:"role"`
	InvitedBy  uuid.UUID  `json:"invited_by"`
	InvitedAt  time.Time  `json:"invited_at"`
	AcceptedAt *time.Time `json:"accepted_at"`
}

// IsActive returns true if the member has accepted the invitation
func (tm *TeamMember) IsActive() bool {
	return tm.AcceptedAt != nil
}

// IsPending returns true if the member hasn't accepted the invitation yet
func (tm *TeamMember) IsPending() bool {
	return tm.AcceptedAt == nil
}

// MemberRole defines the role of a team member
type MemberRole string

const (
	RoleOwner   MemberRole = "OWNER"
	RoleAdmin   MemberRole = "ADMIN"
	RoleManager MemberRole = "MANAGER"
	RoleViewer  MemberRole = "VIEWER"
)

// String returns the string representation of the role
func (r MemberRole) String() string {
	return string(r)
}

// IsValid checks if the role is valid
func (r MemberRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleManager, RoleViewer:
		return true
	default:
		return false
	}
}

// CanInviteMembers returns true if this role can invite other members
func (r MemberRole) CanInviteMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

// CanManageMembers returns true if this role can manage other members
func (r MemberRole) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

// CanViewMembers returns true if this role can view team members
func (r MemberRole) CanViewMembers() bool {
	return true // All roles can view team members
}

// GetPermissionLevel returns a numeric permission level for comparison
func (r MemberRole) GetPermissionLevel() int {
	switch r {
	case RoleOwner:
		return 100
	case RoleAdmin:
		return 80
	case RoleManager:
		return 60
	case RoleViewer:
		return 40
	default:
		return 0
	}
}

// InvitationToken represents a token for team invitations
type InvitationToken struct {
	BaseModel
	AccountID  uuid.UUID `gorm:"not null" json:"account_id"`
	MemberID   uuid.UUID `gorm:"not null" json:"member_id"`
	Token      string    `gorm:"uniqueIndex;not null" json:"token"`
	Email      string    `gorm:"not null" json:"email"`
	Role       MemberRole `gorm:"not null" json:"role"`
	InvitedBy  uuid.UUID `json:"invited_by"`
	ExpiresAt  time.Time `json:"expires_at"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
}

// IsValid checks if the token is still valid and unused
func (it *InvitationToken) IsValid() bool {
	return it.UsedAt == nil && time.Now().Before(it.ExpiresAt)
}

// IsExpired checks if the token has expired
func (it *InvitationToken) IsExpired() bool {
	return time.Now().After(it.ExpiresAt)
}

// IsUsed checks if the token has been used
func (it *InvitationToken) IsUsed() bool {
	return it.UsedAt != nil
}

// MarkAsUsed marks the token as used
func (it *InvitationToken) MarkAsUsed() {
	now := time.Now()
	it.UsedAt = &now
}

// GenerateInvitationToken generates a secure random token for invitations
func GenerateInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// TeamStats provides statistics about a team
type TeamStats struct {
	TotalMembers   int            `json:"total_members"`
	ActiveMembers  int            `json:"active_members"`
	PendingMembers int            `json:"pending_members"`
	RoleBreakdown  map[string]int `json:"role_breakdown"`
}

// Permission represents a specific permission
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// RolePermissions defines what each role can do
var RolePermissions = map[MemberRole][]Permission{
	RoleOwner: {
		{Name: "manage_account", Description: "Manage account settings", Category: "account"},
		{Name: "invite_members", Description: "Invite team members", Category: "team"},
		{Name: "remove_members", Description: "Remove team members", Category: "team"},
		{Name: "change_roles", Description: "Change member roles", Category: "team"},
		{Name: "view_billing", Description: "View billing information", Category: "billing"},
		{Name: "manage_billing", Description: "Manage billing and subscriptions", Category: "billing"},
		{Name: "view_analytics", Description: "View analytics and reports", Category: "analytics"},
		{Name: "manage_integrations", Description: "Manage integrations", Category: "integrations"},
	},
	RoleAdmin: {
		{Name: "invite_members", Description: "Invite team members", Category: "team"},
		{Name: "remove_members", Description: "Remove team members (except owners)", Category: "team"},
		{Name: "change_roles", Description: "Change member roles (except owners)", Category: "team"},
		{Name: "view_billing", Description: "View billing information", Category: "billing"},
		{Name: "view_analytics", Description: "View analytics and reports", Category: "analytics"},
		{Name: "manage_integrations", Description: "Manage integrations", Category: "integrations"},
	},
	RoleManager: {
		{Name: "view_team", Description: "View team members", Category: "team"},
		{Name: "view_analytics", Description: "View analytics and reports", Category: "analytics"},
	},
	RoleViewer: {
		{Name: "view_team", Description: "View team members", Category: "team"},
		{Name: "view_basic_analytics", Description: "View basic analytics", Category: "analytics"},
	},
}

// HasPermission checks if a role has a specific permission
func (r MemberRole) HasPermission(permissionName string) bool {
	permissions, exists := RolePermissions[r]
	if !exists {
		return false
	}

	for _, permission := range permissions {
		if permission.Name == permissionName {
			return true
		}
	}
	return false
}

// GetPermissions returns all permissions for a role
func (r MemberRole) GetPermissions() []Permission {
	return RolePermissions[r]
}