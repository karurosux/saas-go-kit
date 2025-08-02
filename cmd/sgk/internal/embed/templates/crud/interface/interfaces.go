package {{.ModuleName}}interface

import (
	"context"
	{{.ModuleName}}model "{{.Project.GoModule}}/internal/{{.ModuleName}}/model"
)

type {{.ModuleNameCap}}Repository interface {
	Create(ctx context.Context, {{.ModuleName}} *{{.ModuleName}}model.{{.ModuleNameCap}}) error
	GetByID(ctx context.Context, id uint) (*{{.ModuleName}}model.{{.ModuleNameCap}}, error)
	List(ctx context.Context, query {{.ModuleName}}model.{{.ModuleNameCap}}Query) ([]*{{.ModuleName}}model.{{.ModuleNameCap}}, int64, error)
	Update(ctx context.Context, id uint, updates {{.ModuleName}}model.Update{{.ModuleNameCap}}Request) error
	Delete(ctx context.Context, id uint) error
}

type {{.ModuleNameCap}}Service interface {
	Create(ctx context.Context, req {{.ModuleName}}model.Create{{.ModuleNameCap}}Request) (*{{.ModuleName}}model.{{.ModuleNameCap}}, error)
	GetByID(ctx context.Context, id uint) (*{{.ModuleName}}model.{{.ModuleNameCap}}, error)
	List(ctx context.Context, query {{.ModuleName}}model.{{.ModuleNameCap}}Query) ([]*{{.ModuleName}}model.{{.ModuleNameCap}}, int64, error)
	Update(ctx context.Context, id uint, req {{.ModuleName}}model.Update{{.ModuleNameCap}}Request) (*{{.ModuleName}}model.{{.ModuleNameCap}}, error)
	Delete(ctx context.Context, id uint) error
}