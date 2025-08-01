package gorm

import (
	"context"
	"errors"

	"{{.Project.GoModule}}/internal/role/constants"
	"{{.Project.GoModule}}/internal/role/interface"
	"{{.Project.GoModule}}/internal/role/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new GORM role repository
func NewRoleRepository(db *gorm.DB) roleinterface.RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(ctx context.Context, role roleinterface.Role) error {
	defaultRole, ok := role.(*rolemodel.DefaultRole)
	if !ok {
		return errors.New("role must be of type *rolemodel.DefaultRole")
	}

	if err := r.db.WithContext(ctx).Create(defaultRole).Error; err != nil {
		return err
	}

	return nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id uuid.UUID) (roleinterface.Role, error) {
	var role rolemodel.DefaultRole
	if err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(roleconstants.ErrRoleNotFound)
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (roleinterface.Role, error) {
	var role rolemodel.DefaultRole
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(roleconstants.ErrRoleNotFound)
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) FindAll(ctx context.Context, filters roleinterface.RoleFilters) ([]roleinterface.Role, error) {
	var roles []rolemodel.DefaultRole
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
	result := make([]roleinterface.Role, len(roles))
	for i := range roles {
		result[i] = &roles[i]
	}

	return result, nil
}

func (r *RoleRepository) Update(ctx context.Context, role roleinterface.Role) error {
	defaultRole, ok := role.(*rolemodel.DefaultRole)
	if !ok {
		return errors.New("role must be of type *rolemodel.DefaultRole")
	}

	if err := r.db.WithContext(ctx).Save(defaultRole).Error; err != nil {
		return err
	}

	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&rolemodel.DefaultRole{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *RoleRepository) FindSystemRoles(ctx context.Context) ([]roleinterface.Role, error) {
	var roles []rolemodel.DefaultRole
	if err := r.db.WithContext(ctx).Where("is_system = ?", true).Find(&roles).Error; err != nil {
		return nil, err
	}

	// Convert to interface slice
	result := make([]roleinterface.Role, len(roles))
	for i := range roles {
		result[i] = &roles[i]
	}

	return result, nil
}