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

// Service providers for dependency injection

// ProvideRoleRepository provides the role repository
func ProvideRoleRepository(i *do.Injector) (roleinterface.RoleRepository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	
	// Run migrations
	if err := rolegorm.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run role migrations: %w", err)
	}
	
	return rolegorm.NewRoleRepository(db), nil
}

// ProvideUserRoleRepository provides the user role repository
func ProvideUserRoleRepository(i *do.Injector) (roleinterface.UserRoleRepository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	return rolegorm.NewUserRoleRepository(db), nil
}

// ProvideRoleService provides the role service
func ProvideRoleService(i *do.Injector) (roleinterface.RoleService, error) {
	roleRepo := do.MustInvoke[roleinterface.RoleRepository](i)
	userRoleRepo := do.MustInvoke[roleinterface.UserRoleRepository](i)
	
	roleService := roleservice.NewRoleService(roleRepo, userRoleRepo)
	
	// TODO: Implement SeedDefaultRoles method in role service if needed
	// if err := roleService.SeedDefaultRoles(context.Background()); err != nil {
	//	return nil, fmt.Errorf("failed to seed default roles: %w", err)
	// }
	
	return roleService, nil
}

// ProvideRBACMiddleware provides the RBAC middleware
func ProvideRBACMiddleware(i *do.Injector) (*rolemiddleware.RBACMiddleware, error) {
	roleService := do.MustInvoke[roleinterface.RoleService](i)
	return rolemiddleware.NewRBACMiddleware(roleService), nil
}

// ProvideRoleController provides the role controller
func ProvideRoleController(i *do.Injector) (*rolecontroller.RoleController, error) {
	roleService := do.MustInvoke[roleinterface.RoleService](i)
	return rolecontroller.NewRoleController(roleService), nil
}

// RegisterModule registers the role module with the container
func RegisterModule(container *core.Container) error {
	// Register all role services
	do.Provide(container, ProvideRoleRepository)
	do.Provide(container, ProvideUserRoleRepository)
	do.Provide(container, ProvideRoleService)
	do.Provide(container, ProvideRBACMiddleware)
	do.Provide(container, ProvideRoleController)
	
	// Register routes
	e := do.MustInvoke[*echo.Echo](container)
	roleController := do.MustInvoke[*rolecontroller.RoleController](container)
	rbacMiddleware := do.MustInvoke[*rolemiddleware.RBACMiddleware](container)
	
	roleController.RegisterRoutes(e, "/api/v1/roles", rbacMiddleware)
	
	return nil
}