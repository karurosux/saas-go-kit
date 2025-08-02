package roleinterface

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RoleService interface {
	CreateRole(ctx context.Context, name, description string, permissions []string, isSystem bool) (Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (Role, error)
	GetRoleByName(ctx context.Context, name string) (Role, error)
	GetRoles(ctx context.Context, filters RoleFilters) ([]Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, updates RoleUpdates) (Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID, expiresAt *time.Time) error
	UnassignRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error)
	GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]UserRole, error)
	
	UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
	UserHasAnyPermission(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error)
	UserHasAllPermissions(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	
	CreateSystemRoles(ctx context.Context) error
	GetSystemRoles(ctx context.Context) ([]Role, error)
	
	CleanupExpiredRoles(ctx context.Context) error
}