package emailgorm

import (
	"context"

	"gorm.io/gorm"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

// TemplateRepository implements template storage using GORM
type TemplateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// GetTemplate retrieves a template by name
func (r *TemplateRepository) GetTemplate(ctx context.Context, name string) (*emailinterface.EmailTemplate, error) {
	var template emailinterface.EmailTemplate
	err := r.db.WithContext(ctx).
		Where("name = ? AND active = ?", name, true).
		First(&template).Error
	
	if err != nil {
		return nil, err
	}
	
	return &template, nil
}

// CreateTemplate creates a new template
func (r *TemplateRepository) CreateTemplate(ctx context.Context, template *emailinterface.EmailTemplate) error {
	if template.Active == false {
		template.Active = true
	}
	return r.db.WithContext(ctx).Create(template).Error
}

// UpdateTemplate updates an existing template
func (r *TemplateRepository) UpdateTemplate(ctx context.Context, name string, template *emailinterface.EmailTemplate) error {
	return r.db.WithContext(ctx).
		Model(&emailinterface.EmailTemplate{}).
		Where("name = ?", name).
		Updates(template).Error
}

// DeleteTemplate soft deletes a template
func (r *TemplateRepository) DeleteTemplate(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).
		Model(&emailinterface.EmailTemplate{}).
		Where("name = ?", name).
		Update("active", false).Error
}

// ListTemplates lists all active templates
func (r *TemplateRepository) ListTemplates(ctx context.Context) ([]*emailinterface.EmailTemplate, error) {
	var templates []*emailinterface.EmailTemplate
	err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Order("name ASC").
		Find(&templates).Error
	
	return templates, err
}