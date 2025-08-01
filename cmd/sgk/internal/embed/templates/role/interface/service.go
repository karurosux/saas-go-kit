package roleinterface

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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