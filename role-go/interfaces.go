package role

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

// RoleRepository handles role persistence
type RoleRepository interface {
	Create(ctx context.Context, role Role) error
	FindByID(ctx context.Context, id uuid.UUID) (Role, error)
	FindByName(ctx context.Context, name string) (Role, error)
	FindAll(ctx context.Context, filters RoleFilters) ([]Role, error)
	Update(ctx context.Context, role Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindSystemRoles(ctx context.Context) ([]Role, error)
}

// UserRoleRepository handles user role assignments
type UserRoleRepository interface {
	AssignRole(ctx context.Context, userRole UserRole) error
	UnassignRole(ctx context.Context, userID, roleID uuid.UUID) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]UserRole, error)
	FindByRoleID(ctx context.Context, roleID uuid.UUID) ([]UserRole, error)
	FindUserRole(ctx context.Context, userID, roleID uuid.UUID) (UserRole, error)
	FindActiveUserRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error)
	CleanupExpiredRoles(ctx context.Context) error
}

// RoleService provides role management operations
type RoleService interface {
	// Role management
	CreateRole(ctx context.Context, name, description string, permissions []string, isSystem bool) (Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (Role, error)
	GetRoleByName(ctx context.Context, name string) (Role, error)
	GetRoles(ctx context.Context, filters RoleFilters) ([]Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, updates RoleUpdates) (Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	
	// User role assignment
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID, expiresAt *time.Time) error
	UnassignRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error)
	GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]UserRole, error)
	
	// Permission checking
	UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
	UserHasAnyPermission(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error)
	UserHasAllPermissions(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	
	// System roles
	CreateSystemRoles(ctx context.Context) error
	GetSystemRoles(ctx context.Context) ([]Role, error)
	
	// Maintenance
	CleanupExpiredRoles(ctx context.Context) error
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

// Common permission patterns
const (
	// Resource actions
	ActionRead   = "read"
	ActionWrite  = "write"
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionList   = "list"
	
	// Special permissions
	PermissionAll = "*"
	PermissionAdmin = "admin:*"
)

// Permission utility functions
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