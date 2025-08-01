package role

import (
	"fmt"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/role/controller"
	"{{.Project.GoModule}}/internal/role/middleware"
	"{{.Project.GoModule}}/internal/role/repository/gorm"
	"{{.Project.GoModule}}/internal/role/service"
	"github.com/labstack/echo/v4"
	gormdb "gorm.io/gorm"
)

// RegisterModule registers the role module with the container
func RegisterModule(c core.Container) error {
	// Get dependencies from container
	e, ok := c.Get("echo").(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance not found in container")
	}
	
	db, ok := c.Get("db").(*gormdb.DB)
	if !ok {
		return fmt.Errorf("database instance not found in container")
	}
	
	// Run migrations
	if err := gorm.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to run role migrations: %w", err)
	}
	
	// Create repositories
	roleRepo := gorm.NewRoleRepository(db)
	userRoleRepo := gorm.NewUserRoleRepository(db)
	
	// Create service
	roleService := roleservice.NewRoleService(roleRepo, userRoleRepo)
	
	// Seed default roles
	if err := roleService.SeedDefaultRoles(context.Background()); err != nil {
		return fmt.Errorf("failed to seed default roles: %w", err)
	}
	
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