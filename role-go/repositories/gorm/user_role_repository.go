package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/role-go"
	"gorm.io/gorm"
)

type UserRoleRepository struct {
	db *gorm.DB
}

// NewUserRoleRepository creates a new GORM user role repository
func NewUserRoleRepository(db *gorm.DB) role.UserRoleRepository {
	return &UserRoleRepository{db: db}
}

func (r *UserRoleRepository) AssignRole(ctx context.Context, userRole role.UserRole) error {
	defaultUserRole, ok := userRole.(*role.DefaultUserRole)
	if !ok {
		return errors.New("userRole must be of type *DefaultUserRole")
	}

	if err := r.db.WithContext(ctx).Create(defaultUserRole).Error; err != nil {
		return err
	}

	return nil
}

func (r *UserRoleRepository) UnassignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&role.DefaultUserRole{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserRoleRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]role.UserRole, error) {
	var userRoles []role.DefaultUserRole
	if err := r.db.WithContext(ctx).Preload("Role").Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	// Convert to interface slice
	result := make([]role.UserRole, len(userRoles))
	for i, ur := range userRoles {
		result[i] = &ur
	}

	return result, nil
}

func (r *UserRoleRepository) FindByRoleID(ctx context.Context, roleID uuid.UUID) ([]role.UserRole, error) {
	var userRoles []role.DefaultUserRole
	if err := r.db.WithContext(ctx).Preload("Role").Where("role_id = ?", roleID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	// Convert to interface slice
	result := make([]role.UserRole, len(userRoles))
	for i, ur := range userRoles {
		result[i] = &ur
	}

	return result, nil
}

func (r *UserRoleRepository) FindUserRole(ctx context.Context, userID, roleID uuid.UUID) (role.UserRole, error) {
	var userRole role.DefaultUserRole
	if err := r.db.WithContext(ctx).Preload("Role").Where("user_id = ? AND role_id = ?", userID, roleID).First(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user role assignment not found")
		}
		return nil, err
	}
	return &userRole, nil
}

func (r *UserRoleRepository) FindActiveUserRoles(ctx context.Context, userID uuid.UUID) ([]role.UserRole, error) {
	var userRoles []role.DefaultUserRole
	now := time.Now()
	
	query := r.db.WithContext(ctx).Preload("Role").Where("user_id = ?", userID)
	query = query.Where("(expires_at IS NULL OR expires_at > ?)", now)
	
	if err := query.Find(&userRoles).Error; err != nil {
		return nil, err
	}

	// Convert to interface slice
	result := make([]role.UserRole, len(userRoles))
	for i, ur := range userRoles {
		result[i] = &ur
	}

	return result, nil
}

func (r *UserRoleRepository) CleanupExpiredRoles(ctx context.Context) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at <= ?", now).Delete(&role.DefaultUserRole{}).Error; err != nil {
		return err
	}
	return nil
}