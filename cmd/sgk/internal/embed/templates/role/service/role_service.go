package roleservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"{{.Project.GoModule}}/internal/role/constants"
	"{{.Project.GoModule}}/internal/role/interface"
	"{{.Project.GoModule}}/internal/role/model"
	"github.com/google/uuid"
)

// RoleService implements the role management business logic
type RoleService struct {
	roleRepo     roleinterface.RoleRepository
	userRoleRepo roleinterface.UserRoleRepository
}

// NewRoleService creates a new role service
func NewRoleService(
	roleRepo roleinterface.RoleRepository,
	userRoleRepo roleinterface.UserRoleRepository,
) roleinterface.RoleService {
	return &RoleService{
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
	}
}

// Role management

func (s *RoleService) CreateRole(ctx context.Context, name, description string, permissions []string, isSystem bool) (roleinterface.Role, error) {
	// Check if role already exists
	existing, err := s.roleRepo.FindByName(ctx, name)
	if err == nil && existing != nil {
		return nil, errors.New(roleconstants.ErrRoleAlreadyExists)
	}

	role := &rolemodel.DefaultRole{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Permissions: rolemodel.PermissionList(permissions),
		IsSystem:    isSystem,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (s *RoleService) GetRole(ctx context.Context, id uuid.UUID) (roleinterface.Role, error) {
	return s.roleRepo.FindByID(ctx, id)
}

func (s *RoleService) GetRoleByName(ctx context.Context, name string) (roleinterface.Role, error) {
	return s.roleRepo.FindByName(ctx, name)
}

func (s *RoleService) GetRoles(ctx context.Context, filters roleinterface.RoleFilters) ([]roleinterface.Role, error) {
	return s.roleRepo.FindAll(ctx, filters)
}

func (s *RoleService) UpdateRole(ctx context.Context, id uuid.UUID, updates roleinterface.RoleUpdates) (roleinterface.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Don't allow updating system roles
	if role.IsSystemRole() {
		return nil, errors.New("cannot update system role")
	}

	// Apply updates
	defaultRole := role.(*rolemodel.DefaultRole)
	if updates.Name != nil {
		defaultRole.Name = *updates.Name
	}
	if updates.Description != nil {
		defaultRole.Description = *updates.Description
	}
	if updates.Permissions != nil {
		defaultRole.Permissions = rolemodel.PermissionList(*updates.Permissions)
	}

	if err := s.roleRepo.Update(ctx, defaultRole); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return defaultRole, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Don't allow deleting system roles
	if role.IsSystemRole() {
		return errors.New(roleconstants.ErrSystemRoleDelete)
	}

	// Check if any users have this role
	userRoles, err := s.userRoleRepo.FindByRoleID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}
	if len(userRoles) > 0 {
		return errors.New("cannot delete role that is assigned to users")
	}

	return s.roleRepo.Delete(ctx, id)
}

// User role assignment

func (s *RoleService) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID, expiresAt *time.Time) error {
	// Check if role exists
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return err
	}

	// Check if already assigned
	existing, _ := s.userRoleRepo.FindUserRole(ctx, userID, roleID)
	if existing != nil {
		return errors.New(roleconstants.ErrRoleAlreadyAssigned)
	}

	userRole := &rolemodel.DefaultUserRole{
		ID:         uuid.New(),
		UserID:     userID,
		RoleID:     roleID,
		Role:       role.(*rolemodel.DefaultRole),
		AssignedBy: assignedBy,
		AssignedAt: time.Now(),
		ExpiresAt:  expiresAt,
	}

	return s.userRoleRepo.AssignRole(ctx, userRole)
}

func (s *RoleService) UnassignRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.userRoleRepo.UnassignRole(ctx, userID, roleID)
}

func (s *RoleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]roleinterface.Role, error) {
	userRoles, err := s.userRoleRepo.FindActiveUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]roleinterface.Role, len(userRoles))
	for i, ur := range userRoles {
		roles[i] = ur.GetRole()
	}

	return roles, nil
}

func (s *RoleService) GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]roleinterface.UserRole, error) {
	return s.userRoleRepo.FindByRoleID(ctx, roleID)
}

// Permission checking

func (s *RoleService) UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.HasPermission(permission) {
			return true, nil
		}
	}

	return false, nil
}

func (s *RoleService) UserHasAnyPermission(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.HasAnyPermission(permissions) {
			return true, nil
		}
	}

	return false, nil
}

func (s *RoleService) UserHasAllPermissions(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	// Collect all permissions from all roles
	allPerms := make(map[string]bool)
	for _, role := range roles {
		for _, perm := range role.GetPermissions() {
			allPerms[perm] = true
		}
	}

	// Check if all required permissions are present
	for _, perm := range permissions {
		found := false
		for _, role := range roles {
			if role.HasPermission(perm) {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}

func (s *RoleService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Collect unique permissions
	permMap := make(map[string]bool)
	for _, role := range roles {
		for _, perm := range role.GetPermissions() {
			permMap[perm] = true
		}
	}

	// Convert to slice
	permissions := make([]string, 0, len(permMap))
	for perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// System roles

func (s *RoleService) CreateSystemRoles(ctx context.Context) error {
	systemRoles := []struct {
		name        string
		description string
		permissions []string
	}{
		{
			name:        roleconstants.RoleAdmin,
			description: roleconstants.RoleAdminDesc,
			permissions: []string{roleconstants.PermissionAll},
		},
		{
			name:        roleconstants.RoleUser,
			description: roleconstants.RoleUserDesc,
			permissions: []string{
				"profile:read",
				"profile:update",
			},
		},
		{
			name:        roleconstants.RoleModerator,
			description: roleconstants.RoleModeratorDesc,
			permissions: []string{
				"content:read",
				"content:update",
				"content:delete",
				"users:list",
				"users:read",
			},
		},
	}

	for _, sr := range systemRoles {
		// Check if already exists
		existing, _ := s.roleRepo.FindByName(ctx, sr.name)
		if existing != nil {
			continue
		}

		_, err := s.CreateRole(ctx, sr.name, sr.description, sr.permissions, true)
		if err != nil {
			return fmt.Errorf("failed to create system role %s: %w", sr.name, err)
		}
	}

	return nil
}

func (s *RoleService) GetSystemRoles(ctx context.Context) ([]roleinterface.Role, error) {
	return s.roleRepo.FindSystemRoles(ctx)
}

// Maintenance

func (s *RoleService) CleanupExpiredRoles(ctx context.Context) error {
	return s.userRoleRepo.CleanupExpiredRoles(ctx)
}