package role

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type roleService struct {
	roleRepo     RoleRepository
	userRoleRepo UserRoleRepository
}

// NewRoleService creates a new role service
func NewRoleService(roleRepo RoleRepository, userRoleRepo UserRoleRepository) RoleService {
	return &roleService{
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
	}
}

// Role management
func (s *roleService) CreateRole(ctx context.Context, name, description string, permissions []string, isSystem bool) (Role, error) {
	role := &DefaultRole{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Permissions: PermissionList(permissions),
		IsSystem:    isSystem,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) GetRole(ctx context.Context, id uuid.UUID) (Role, error) {
	return s.roleRepo.FindByID(ctx, id)
}

func (s *roleService) GetRoleByName(ctx context.Context, name string) (Role, error) {
	return s.roleRepo.FindByName(ctx, name)
}

func (s *roleService) GetRoles(ctx context.Context, filters RoleFilters) ([]Role, error) {
	return s.roleRepo.FindAll(ctx, filters)
}

func (s *roleService) UpdateRole(ctx context.Context, id uuid.UUID, updates RoleUpdates) (Role, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	defaultRole, ok := role.(*DefaultRole)
	if !ok {
		return nil, errors.New("role must be of type *DefaultRole")
	}

	// Apply updates
	if updates.Name != nil {
		defaultRole.Name = *updates.Name
	}
	if updates.Description != nil {
		defaultRole.Description = *updates.Description
	}
	if updates.Permissions != nil {
		defaultRole.Permissions = PermissionList(*updates.Permissions)
	}
	defaultRole.UpdatedAt = time.Now()

	if err := s.roleRepo.Update(ctx, defaultRole); err != nil {
		return nil, err
	}

	return defaultRole, nil
}

func (s *roleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	// Check if role is a system role
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if role.IsSystemRole() {
		return errors.New("cannot delete system role")
	}

	return s.roleRepo.Delete(ctx, id)
}

// User role assignment
func (s *roleService) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID, expiresAt *time.Time) error {
	// Check if role exists
	_, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return err
	}

	// Check if assignment already exists
	_, err = s.userRoleRepo.FindUserRole(ctx, userID, roleID)
	if err == nil {
		return errors.New("user already has this role assigned")
	}

	userRole := &DefaultUserRole{
		ID:         uuid.New(),
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: assignedBy,
		AssignedAt: time.Now(),
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return s.userRoleRepo.AssignRole(ctx, userRole)
}

func (s *roleService) UnassignRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.userRoleRepo.UnassignRole(ctx, userID, roleID)
}

func (s *roleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error) {
	userRoles, err := s.userRoleRepo.FindActiveUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]Role, len(userRoles))
	for i, ur := range userRoles {
		roles[i] = ur.GetRole()
	}

	return roles, nil
}

func (s *roleService) GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]UserRole, error) {
	return s.userRoleRepo.FindByRoleID(ctx, roleID)
}

// Permission checking
func (s *roleService) UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
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

func (s *roleService) UserHasAnyPermission(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error) {
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

func (s *roleService) UserHasAllPermissions(ctx context.Context, userID uuid.UUID, permissions []string) (bool, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	// Collect all user permissions
	allPermissions := make([]string, 0)
	for _, role := range roles {
		allPermissions = append(allPermissions, role.GetPermissions()...)
	}

	// Create a temporary role to check permissions
	tempRole := &DefaultRole{
		Permissions: PermissionList(allPermissions),
	}

	return tempRole.HasAllPermissions(permissions), nil
}

func (s *roleService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissionSet := make(map[string]bool)
	for _, role := range roles {
		for _, permission := range role.GetPermissions() {
			permissionSet[permission] = true
		}
	}

	permissions := make([]string, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// System roles
func (s *roleService) CreateSystemRoles(ctx context.Context) error {
	// This method is intentionally empty as per user request to not include seeds
	// Applications can implement their own system role creation logic
	return nil
}

func (s *roleService) GetSystemRoles(ctx context.Context) ([]Role, error) {
	return s.roleRepo.FindSystemRoles(ctx)
}

// Maintenance
func (s *roleService) CleanupExpiredRoles(ctx context.Context) error {
	return s.userRoleRepo.CleanupExpiredRoles(ctx)
}