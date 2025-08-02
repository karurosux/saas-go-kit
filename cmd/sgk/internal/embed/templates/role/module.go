package role

import (
	"fmt"
	
	"github.com/samber/do"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	
	"{{.Project.GoModule}}/internal/core"
	rolecontroller "{{.Project.GoModule}}/internal/role/controller"
	roleinterface "{{.Project.GoModule}}/internal/role/interface"
	rolemiddleware "{{.Project.GoModule}}/internal/role/middleware"
	rolegorm "{{.Project.GoModule}}/internal/role/repository/gorm"
	roleservice "{{.Project.GoModule}}/internal/role/service"
)


func ProvideRoleRepository(i *do.Injector) (roleinterface.RoleRepository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	
	if err := rolegorm.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run role migrations: %w", err)
	}
	
	return rolegorm.NewRoleRepository(db), nil
}

func ProvideUserRoleRepository(i *do.Injector) (roleinterface.UserRoleRepository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	return rolegorm.NewUserRoleRepository(db), nil
}

func ProvideRoleService(i *do.Injector) (roleinterface.RoleService, error) {
	roleRepo := do.MustInvoke[roleinterface.RoleRepository](i)
	userRoleRepo := do.MustInvoke[roleinterface.UserRoleRepository](i)
	
	roleService := roleservice.NewRoleService(roleRepo, userRoleRepo)
	
	
	return roleService, nil
}

func ProvideRBACMiddleware(i *do.Injector) (*rolemiddleware.RBACMiddleware, error) {
	roleService := do.MustInvoke[roleinterface.RoleService](i)
	return rolemiddleware.NewRBACMiddleware(roleService), nil
}

func ProvideRoleController(i *do.Injector) (*rolecontroller.RoleController, error) {
	roleService := do.MustInvoke[roleinterface.RoleService](i)
	return rolecontroller.NewRoleController(roleService), nil
}

func RegisterModule(container *core.Container) error {
	do.Provide(container, ProvideRoleRepository)
	do.Provide(container, ProvideUserRoleRepository)
	do.Provide(container, ProvideRoleService)
	do.Provide(container, ProvideRBACMiddleware)
	do.Provide(container, ProvideRoleController)
	
	e := do.MustInvoke[*echo.Echo](container)
	roleController := do.MustInvoke[*rolecontroller.RoleController](container)
	rbacMiddleware := do.MustInvoke[*rolemiddleware.RBACMiddleware](container)
	
	roleController.RegisterRoutes(e, "/api/v1/roles", rbacMiddleware)
	
	return nil
}