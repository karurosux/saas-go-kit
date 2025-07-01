package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/role-go"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new GORM role repository
func NewRoleRepository(db *gorm.DB) role.RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(ctx context.Context, roleObj role.Role) error {
	defaultRole, ok := roleObj.(*role.DefaultRole)
	if !ok {
		return errors.New("role must be of type *DefaultRole")
	}

	if err := r.db.WithContext(ctx).Create(defaultRole).Error; err != nil {
		return err
	}

	return nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id uuid.UUID) (role.Role, error) {
	var roleObj role.DefaultRole
	if err := r.db.WithContext(ctx).First(&roleObj, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &roleObj, nil
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (role.Role, error) {
	var roleObj role.DefaultRole
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&roleObj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &roleObj, nil
}

func (r *RoleRepository) FindAll(ctx context.Context, filters role.RoleFilters) ([]role.Role, error) {
	var roles []role.DefaultRole
	query := r.db.WithContext(ctx)

	// Apply filters
	if filters.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filters.Name+"%")
	}

	if filters.IsSystem != nil {
		query = query.Where("is_system = ?", *filters.IsSystem)
	}

	if filters.HasPermission != "" {
		query = query.Where("permissions @> ?", `["`+filters.HasPermission+`"]`)
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	if err := query.Find(&roles).Error; err != nil {
		return nil, err
	}

	// Convert to interface slice
	result := make([]role.Role, len(roles))
	for i, r := range roles {
		result[i] = &r
	}

	return result, nil
}

func (r *RoleRepository) Update(ctx context.Context, roleObj role.Role) error {
	defaultRole, ok := roleObj.(*role.DefaultRole)
	if !ok {
		return errors.New("role must be of type *DefaultRole")
	}

	if err := r.db.WithContext(ctx).Save(defaultRole).Error; err != nil {
		return err
	}

	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&role.DefaultRole{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleRepository) FindSystemRoles(ctx context.Context) ([]role.Role, error) {
	var roles []role.DefaultRole
	if err := r.db.WithContext(ctx).Where("is_system = ?", true).Find(&roles).Error; err != nil {
		return nil, err
	}

	// Convert to interface slice
	result := make([]role.Role, len(roles))
	for i, r := range roles {
		result[i] = &r
	}

	return result, nil
}