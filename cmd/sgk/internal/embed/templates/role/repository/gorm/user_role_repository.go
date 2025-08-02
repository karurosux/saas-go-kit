package gorm

import (
	"context"
	"errors"
	"time"

	roleconstants "{{.Project.GoModule}}/internal/role/constants"
	roleinterface "{{.Project.GoModule}}/internal/role/interface"
	rolemodel "{{.Project.GoModule}}/internal/role/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) roleinterface.UserRoleRepository {
	return &UserRoleRepository{db: db}
}

func (r *UserRoleRepository) AssignRole(ctx context.Context, userRole roleinterface.UserRole) error {
	defaultUserRole, ok := userRole.(*rolemodel.DefaultUserRole)
	if !ok {
		return errors.New("userRole must be of type *rolemodel.DefaultUserRole")
	}
	
	return r.db.WithContext(ctx).Create(defaultUserRole).Error
}

func (r *UserRoleRepository) UnassignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&rolemodel.DefaultUserRole{})
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New(roleconstants.ErrUserRoleNotFound)
	}
	
	return nil
}

func (r *UserRoleRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]roleinterface.UserRole, error) {
	var userRoles []rolemodel.DefaultUserRole
	if err := r.db.WithContext(ctx).
		Preload("Role").
		Where("user_id = ?", userID).
		Find(&userRoles).Error; err != nil {
		return nil, err
	}

	result := make([]roleinterface.UserRole, len(userRoles))
	for i := range userRoles {
		result[i] = &userRoles[i]
	}
	return result, nil
}

func (r *UserRoleRepository) FindByRoleID(ctx context.Context, roleID uuid.UUID) ([]roleinterface.UserRole, error) {
	var userRoles []rolemodel.DefaultUserRole
	if err := r.db.WithContext(ctx).
		Preload("Role").
		Where("role_id = ?", roleID).
		Find(&userRoles).Error; err != nil {
		return nil, err
	}

	result := make([]roleinterface.UserRole, len(userRoles))
	for i := range userRoles {
		result[i] = &userRoles[i]
	}
	return result, nil
}

func (r *UserRoleRepository) FindUserRole(ctx context.Context, userID, roleID uuid.UUID) (roleinterface.UserRole, error) {
	var userRole rolemodel.DefaultUserRole
	if err := r.db.WithContext(ctx).
		Preload("Role").
		Where("user_id = ? AND role_id = ?", userID, roleID).
		First(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(roleconstants.ErrUserRoleNotFound)
		}
		return nil, err
	}
	return &userRole, nil
}

func (r *UserRoleRepository) FindActiveUserRoles(ctx context.Context, userID uuid.UUID) ([]roleinterface.UserRole, error) {
	var userRoles []rolemodel.DefaultUserRole
	query := r.db.WithContext(ctx).
		Preload("Role").
		Where("user_id = ?", userID)
	
	query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())
	
	if err := query.Find(&userRoles).Error; err != nil {
		return nil, err
	}

	result := make([]roleinterface.UserRole, len(userRoles))
	for i := range userRoles {
		result[i] = &userRoles[i]
	}
	return result, nil
}

func (r *UserRoleRepository) CleanupExpiredRoles(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).
		Delete(&rolemodel.DefaultUserRole{}).Error
}