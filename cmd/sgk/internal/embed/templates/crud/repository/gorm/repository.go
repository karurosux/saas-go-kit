package {{.ModuleName}}gorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	{{.ModuleName}}interface "{{.Project.GoModule}}/internal/{{.ModuleName}}/interface"
	{{.ModuleName}}model "{{.Project.GoModule}}/internal/{{.ModuleName}}/model"
)

type {{.ModuleNameCap}}Repository struct {
	db *gorm.DB
}

func New{{.ModuleNameCap}}Repository(db *gorm.DB) {{.ModuleName}}interface.{{.ModuleNameCap}}Repository {
	return &{{.ModuleNameCap}}Repository{db: db}
}

func (r *{{.ModuleNameCap}}Repository) Create(ctx context.Context, {{.ModuleName}} *{{.ModuleName}}model.{{.ModuleNameCap}}) error {
	if err := r.db.WithContext(ctx).Create({{.ModuleName}}).Error; err != nil {
		return fmt.Errorf("failed to create {{.ModuleName}}: %w", err)
	}
	return nil
}

func (r *{{.ModuleNameCap}}Repository) GetByID(ctx context.Context, id uint) (*{{.ModuleName}}model.{{.ModuleNameCap}}, error) {
	var {{.ModuleName}} {{.ModuleName}}model.{{.ModuleNameCap}}
	if err := r.db.WithContext(ctx).First(&{{.ModuleName}}, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("{{.ModuleName}} not found")
		}
		return nil, fmt.Errorf("failed to get {{.ModuleName}}: %w", err)
	}
	return &{{.ModuleName}}, nil
}

func (r *{{.ModuleNameCap}}Repository) List(ctx context.Context, query {{.ModuleName}}model.{{.ModuleNameCap}}Query) ([]*{{.ModuleName}}model.{{.ModuleNameCap}}, int64, error) {
	var {{.ModuleName}}s []*{{.ModuleName}}model.{{.ModuleNameCap}}
	var total int64

	db := r.db.WithContext(ctx).Model(&{{.ModuleName}}model.{{.ModuleNameCap}}{})

	if query.Name != nil {
		db = db.Where("name ILIKE ?", "%"+*query.Name+"%")
	}
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count {{.ModuleName}}s: %w", err)
	}

	if query.Page > 0 && query.Limit > 0 {
		offset := (query.Page - 1) * query.Limit
		db = db.Offset(offset).Limit(query.Limit)
	}

	if err := db.Find(&{{.ModuleName}}s).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list {{.ModuleName}}s: %w", err)
	}

	return {{.ModuleName}}s, total, nil
}

func (r *{{.ModuleNameCap}}Repository) Update(ctx context.Context, id uint, updates {{.ModuleName}}model.Update{{.ModuleNameCap}}Request) error {
	updateData := make(map[string]interface{})
	
	if updates.Name != nil {
		updateData["name"] = *updates.Name
	}
	if updates.Description != nil {
		updateData["description"] = *updates.Description
	}
	if updates.IsActive != nil {
		updateData["is_active"] = *updates.IsActive
	}

	if len(updateData) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).Model(&{{.ModuleName}}model.{{.ModuleNameCap}}{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		return fmt.Errorf("failed to update {{.ModuleName}}: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("{{.ModuleName}} not found")
	}

	return nil
}

func (r *{{.ModuleNameCap}}Repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&{{.ModuleName}}model.{{.ModuleNameCap}}{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete {{.ModuleName}}: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("{{.ModuleName}} not found")
	}
	return nil
}