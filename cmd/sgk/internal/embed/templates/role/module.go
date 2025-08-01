package role

import (
	"fmt"
	
	"{{.Project.GoModule}}/internal/core"
	rolecontroller "{{.Project.GoModule}}/internal/role/controller"
	rolemiddleware "{{.Project.GoModule}}/internal/role/middleware"
	rolegorm "{{.Project.GoModule}}/internal/role/repository/gorm"
	roleservice "{{.Project.GoModule}}/internal/role/service"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RegisterModule registers the role module with the container
func RegisterModule(c *core.Container) error {
	// Get dependencies from container
	eInt, err := c.Get("echo")
	if err != nil {
		return fmt.Errorf("echo instance not found in container: %w", err)
	}
	e, ok := eInt.(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance has invalid type")
	}
	
	dbInt, err := c.Get("db")
	if err != nil {
		return fmt.Errorf("database instance not found in container: %w", err)
	}
	db, ok := dbInt.(*gorm.DB)
	if !ok {
		return fmt.Errorf("database instance has invalid type")
	}
	
	// Run migrations
	if err := rolegorm.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to run role migrations: %w", err)
	}
	
	// Create repositories
	roleRepo := rolegorm.NewRoleRepository(db)
	userRoleRepo := rolegorm.NewUserRoleRepository(db)
	
	// Create service
	roleService := roleservice.NewRoleService(roleRepo, userRoleRepo)
	
	// TODO: Implement SeedDefaultRoles method in role service if needed
	// if err := roleService.SeedDefaultRoles(context.Background()); err != nil {
	//	return fmt.Errorf("failed to seed default roles: %w", err)
	// }
	
	// Create middleware
	rbacMiddleware := rolemiddleware.NewRBACMiddleware(roleService)
	
	// Create controller
	roleController := rolecontroller.NewRoleController(roleService)
	
	// Register routes
	roleController.RegisterRoutes(e, "/roles", rbacMiddleware)
	
	// Register components in container for other modules to use
	c.Set("role.service", roleService)
	c.Set("role.middleware", rbacMiddleware)
	c.Set("role.roleRepository", roleRepo)
	c.Set("role.userRoleRepository", userRoleRepo)
	
	return nil
}