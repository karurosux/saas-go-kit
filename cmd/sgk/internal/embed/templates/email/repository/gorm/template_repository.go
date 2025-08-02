package emailgorm

import (
	"context"

	"gorm.io/gorm"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

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

func (r *TemplateRepository) CreateTemplate(ctx context.Context, template *emailinterface.EmailTemplate) error {
	if template.Active == false {
		template.Active = true
	}
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *TemplateRepository) UpdateTemplate(ctx context.Context, name string, template *emailinterface.EmailTemplate) error {
	return r.db.WithContext(ctx).
		Model(&emailinterface.EmailTemplate{}).
		Where("name = ?", name).
		Updates(template).Error
}

func (r *TemplateRepository) DeleteTemplate(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).
		Model(&emailinterface.EmailTemplate{}).
		Where("name = ?", name).
		Update("active", false).Error
}

func (r *TemplateRepository) ListTemplates(ctx context.Context) ([]*emailinterface.EmailTemplate, error) {
	var templates []*emailinterface.EmailTemplate
	err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Order("name ASC").
		Find(&templates).Error
	
	return templates, err
}