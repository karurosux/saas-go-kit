package roleinterface

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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

type RoleFilters struct {
	Name       string
	IsSystem   *bool
	HasPermission string
	Limit      int
	Offset     int
}

type RoleUpdates struct {
	Name        *string
	Description *string
	Permissions *[]string
}

type PermissionChecker interface {
	Check(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

type PermissionUtils interface {
	ParsePermission(permission string) (resource, action string)
	MatchesPattern(permission, pattern string) bool
	BuildPermission(resource, action string) string
	IsValidPermission(permission string) bool
}