package roleinterface

import (
	"context"

	"github.com/google/uuid"
)

type RoleRepository interface {
	Create(ctx context.Context, role Role) error
	FindByID(ctx context.Context, id uuid.UUID) (Role, error)
	FindByName(ctx context.Context, name string) (Role, error)
	FindAll(ctx context.Context, filters RoleFilters) ([]Role, error)
	Update(ctx context.Context, role Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindSystemRoles(ctx context.Context) ([]Role, error)
}

type UserRoleRepository interface {
	AssignRole(ctx context.Context, userRole UserRole) error
	UnassignRole(ctx context.Context, userID, roleID uuid.UUID) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]UserRole, error)
	FindByRoleID(ctx context.Context, roleID uuid.UUID) ([]UserRole, error)
	FindUserRole(ctx context.Context, userID, roleID uuid.UUID) (UserRole, error)
	FindActiveUserRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error)
	CleanupExpiredRoles(ctx context.Context) error
}