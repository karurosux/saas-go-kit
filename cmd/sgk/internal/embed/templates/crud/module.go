package {{.ModuleName}}

import (
	"fmt"

	"github.com/samber/do"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"{{.Project.GoModule}}/internal/core"
	{{.ModuleName}}controller "{{.Project.GoModule}}/internal/{{.ModuleName}}/controller"
	{{.ModuleName}}interface "{{.Project.GoModule}}/internal/{{.ModuleName}}/interface"
	{{.ModuleName}}gorm "{{.Project.GoModule}}/internal/{{.ModuleName}}/repository/gorm"
	{{.ModuleName}}service "{{.Project.GoModule}}/internal/{{.ModuleName}}/service"
)

func Provide{{.ModuleNameCap}}Repository(i *do.Injector) ({{.ModuleName}}interface.{{.ModuleNameCap}}Repository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	
	if err := {{.ModuleName}}gorm.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run {{.ModuleName}} migrations: %w", err)
	}
	
	return {{.ModuleName}}gorm.New{{.ModuleNameCap}}Repository(db), nil
}

func Provide{{.ModuleNameCap}}Service(i *do.Injector) ({{.ModuleName}}interface.{{.ModuleNameCap}}Service, error) {
	repo := do.MustInvoke[{{.ModuleName}}interface.{{.ModuleNameCap}}Repository](i)
	return {{.ModuleName}}service.New{{.ModuleNameCap}}Service(repo), nil
}

func Provide{{.ModuleNameCap}}Controller(i *do.Injector) (*{{.ModuleName}}controller.{{.ModuleNameCap}}Controller, error) {
	service := do.MustInvoke[{{.ModuleName}}interface.{{.ModuleNameCap}}Service](i)
	return {{.ModuleName}}controller.New{{.ModuleNameCap}}Controller(service), nil
}

func RegisterModule(container *core.Container) error {
	do.Provide(container, Provide{{.ModuleNameCap}}Repository)
	do.Provide(container, Provide{{.ModuleNameCap}}Service)
	do.Provide(container, Provide{{.ModuleNameCap}}Controller)
	
	e := do.MustInvoke[*echo.Echo](container)
	controller := do.MustInvoke[*{{.ModuleName}}controller.{{.ModuleNameCap}}Controller](container)
	
	controller.RegisterRoutes(e, "/api/v1/{{.ModuleName}}s")
	
	return nil
}