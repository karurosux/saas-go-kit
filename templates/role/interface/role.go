package roleinterface

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Role represents a role with tag-based permissions
type Role interface {
	GetID() uuid.UUID
	GetName() string
	GetDescription() string
	GetPermissions() []string
	IsSystemRole() bool
	HasPermission(permission string) bool
	HasAnyPermission(permissions []string) bool
	HasAllPermissions(permissions []string) bool
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// UserRole represents the assignment of a role to a user
type UserRole interface {
	GetID() uuid.UUID
	GetUserID() uuid.UUID
	GetRoleID() uuid.UUID
	GetRole() Role
	GetAssignedBy() uuid.UUID
	GetAssignedAt() time.Time
	GetExpiresAt() *time.Time
	IsActive() bool
}

// RoleFilters for querying roles
type RoleFilters struct {
	Name       string
	IsSystem   *bool
	HasPermission string
	Limit      int
	Offset     int
}

// RoleUpdates for updating roles
type RoleUpdates struct {
	Name        *string
	Description *string
	Permissions *[]string
}

// PermissionChecker interface for middleware
type PermissionChecker interface {
	Check(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

// PermissionUtils interface for permission utilities
type PermissionUtils interface {
	// Parse permission tag: "users:read" -> ("users", "read")
	ParsePermission(permission string) (resource, action string)
	// Check if permission matches pattern (supports wildcards)
	MatchesPattern(permission, pattern string) bool
	// Build permission tag: ("users", "read") -> "users:read"
	BuildPermission(resource, action string) string
	// Validate permission format
	IsValidPermission(permission string) bool
}