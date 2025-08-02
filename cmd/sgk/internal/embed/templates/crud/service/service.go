package {{.ModuleName}}service

import (
	"context"
	"fmt"

	{{.ModuleName}}interface "{{.Project.GoModule}}/internal/{{.ModuleName}}/interface"
	{{.ModuleName}}model "{{.Project.GoModule}}/internal/{{.ModuleName}}/model"
)

type {{.ModuleNameCap}}Service struct {
	repo {{.ModuleName}}interface.{{.ModuleNameCap}}Repository
}

func New{{.ModuleNameCap}}Service(repo {{.ModuleName}}interface.{{.ModuleNameCap}}Repository) {{.ModuleName}}interface.{{.ModuleNameCap}}Service {
	return &{{.ModuleNameCap}}Service{
		repo: repo,
	}
}

func (s *{{.ModuleNameCap}}Service) Create(ctx context.Context, req {{.ModuleName}}model.Create{{.ModuleNameCap}}Request) (*{{.ModuleName}}model.{{.ModuleNameCap}}, error) {
	{{.ModuleName}} := &{{.ModuleName}}model.{{.ModuleNameCap}}{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}

	if req.IsActive != nil {
		{{.ModuleName}}.IsActive = *req.IsActive
	}

	if err := s.repo.Create(ctx, {{.ModuleName}}); err != nil {
		return nil, fmt.Errorf("failed to create {{.ModuleName}}: %w", err)
	}

	return {{.ModuleName}}, nil
}

func (s *{{.ModuleNameCap}}Service) GetByID(ctx context.Context, id uint) (*{{.ModuleName}}model.{{.ModuleNameCap}}, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *{{.ModuleNameCap}}Service) List(ctx context.Context, query {{.ModuleName}}model.{{.ModuleNameCap}}Query) ([]*{{.ModuleName}}model.{{.ModuleNameCap}}, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}

	return s.repo.List(ctx, query)
}

func (s *{{.ModuleNameCap}}Service) Update(ctx context.Context, id uint, req {{.ModuleName}}model.Update{{.ModuleNameCap}}Request) (*{{.ModuleName}}model.{{.ModuleNameCap}}, error) {
	if err := s.repo.Update(ctx, id, req); err != nil {
		return nil, fmt.Errorf("failed to update {{.ModuleName}}: %w", err)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *{{.ModuleNameCap}}Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}