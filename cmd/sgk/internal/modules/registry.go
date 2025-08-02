package modules

import (
	"fmt"
)

type ModuleDefinition struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Description          string            `json:"description"`
	Dependencies         []string          `json:"dependencies"`
	InternalDependencies []string          `json:"internal_dependencies"`
	ContainerServices    map[string]string `json:"container_services"`
	Files                []string          `json:"files"`
	Database             string            `json:"database"`
	Options              map[string]string `json:"options"`
}

func GetAvailableModules() map[string]ModuleDefinition {
	return availableModules
}

func GetModule(name string) (ModuleDefinition, error) {
	module, exists := availableModules[name]
	if !exists {
		return ModuleDefinition{}, fmt.Errorf("module '%s' not found", name)
	}
	return module, nil
}

func IsModuleAvailable(name string) bool {
	_, exists := availableModules[name]
	return exists
}

var availableModules = map[string]ModuleDefinition{
	"auth": {
		Name:        "auth",
		Version:     "1.0.0",
		Description: "Complete authentication system with JWT, email verification, password reset",
		Dependencies: []string{
			"github.com/golang-jwt/jwt/v5",
			"golang.org/x/crypto",
		},
		InternalDependencies: []string{
			"core",
			"email",
		},
		ContainerServices: map[string]string{
			"echo":         "echo.Echo",
			"db":           "gorm.DB",
			"notification": "notification.Service",
		},
		Files: []string{
			"handlers.go",
			"service.go",
			"models.go",
			"interfaces.go",
			"module.go",
			"repositories/gorm/account_repository.go",
			"repositories/gorm/token_repository.go",
			"repositories/gorm/migrations.go",
		},
	},
	"subscription": {
		Name:        "subscription",
		Version:     "1.0.0",
		Description: "Subscription and billing management with Stripe integration",
		Dependencies: []string{
			"github.com/stripe/stripe-go/v76",
		},
		InternalDependencies: []string{
			"core",
		},
		ContainerServices: map[string]string{
			"echo": "echo.Echo",
			"db":   "gorm.DB",
		},
		Files: []string{
			"handlers.go",
			"service.go",
			"models.go",
			"interfaces.go",
			"module.go",
			"middleware.go",
			"stripe_provider.go",
			"repositories/gorm/subscription_repository.go",
			"repositories/gorm/subscription_plan_repository.go",
			"repositories/gorm/usage_repository.go",
			"repositories/gorm/migrations.go",
		},
	},
	"team": {
		Name:        "team",
		Version:     "1.0.0",
		Description: "Team management with role-based access control",
		Dependencies: []string{},
		InternalDependencies: []string{
			"core",
			"auth",
			"role",
		},
		ContainerServices: map[string]string{
			"echo":         "echo.Echo",
			"db":           "gorm.DB",
			"auth":         "auth.Service",
			"notification": "notification.Service",
		},
		Files: []string{
			"handlers.go",
			"service.go",
			"models.go",
			"interfaces.go",
			"module.go",
			"repositories/gorm/team_member_repository.go",
			"repositories/gorm/user_repository.go",
			"repositories/gorm/invitation_token_repository.go",
			"repositories/gorm/migrations.go",
		},
	},
	"notification": {
		Name:        "notification",
		Version:     "1.0.0",
		Description: "Multi-channel notification system (email, SMS, push)",
		Dependencies: []string{
			"gopkg.in/gomail.v2",
		},
		InternalDependencies: []string{
			"core",
		},
		ContainerServices: map[string]string{
			"echo": "echo.Echo",
		},
		Files: []string{
			"handlers.go",
			"service.go",
			"interfaces.go",
			"smtp_provider.go",
			"module.go",
		},
	},
	"health": {
		Name:        "health",
		Version:     "1.0.0",
		Description: "Application health monitoring with multiple check types",
		Dependencies: []string{},
		InternalDependencies: []string{
			"core",
		},
		ContainerServices: map[string]string{
			"echo": "echo.Echo",
			"db":   "gorm.DB",
		},
		Files: []string{
			"service.go",
			"interfaces.go",
			"checkers.go",
			"gorm_checker.go",
			"module.go",
		},
	},
	"role": {
		Name:        "role",
		Version:     "1.0.0",
		Description: "Role-based access control and permissions management",
		Dependencies: []string{},
		InternalDependencies: []string{
			"core",
		},
		ContainerServices: map[string]string{
			"echo": "echo.Echo",
			"db":   "gorm.DB",
		},
		Files: []string{
			"interfaces.go",
			"models.go",
			"permissions.go",
			"middleware.go",
			"service.go",
			"module.go",
			"repositories/gorm/role_repository.go",
			"repositories/gorm/user_role_repository.go",
		},
	},
	"job": {
		Name:        "job",
		Version:     "1.0.0",
		Description: "Background job processing system",
		Dependencies: []string{},
		InternalDependencies: []string{
			"core",
		},
		ContainerServices: map[string]string{
			"echo": "echo.Echo",
			"db":   "gorm.DB",
		},
		Files: []string{
			"interfaces.go",
			"models.go",
			"handlers.go",
			"service.go",
			"queue.go",
			"errors.go",
			"module.go",
			"repositories/gorm/job_repository.go",
			"repositories/gorm/job_result_repository.go",
			"repositories/gorm/migrations.go",
		},
	},
	"sse": {
		Name:        "sse",
		Version:     "1.0.0",
		Description: "Server-sent events for real-time communication",
		Dependencies: []string{},
		InternalDependencies: []string{
			"core",
		},
		ContainerServices: map[string]string{
			"echo": "echo.Echo",
		},
		Files: []string{
			"sse.go",
			"module.go",
		},
	},
	"container": {
		Name:        "container",
		Version:     "1.0.0",
		Description: "Dependency injection container",
		Dependencies: []string{},
		InternalDependencies: []string{
			"core",
		},
		ContainerServices: map[string]string{},
		Files: []string{
			"container.go",
			"interfaces.go",
		},
	},
	"email": {
		Name:        "email",
		Version:     "1.0.0",
		Description: "Email service with SMTP support, template management, and queue processing",
		Dependencies: []string{},
		InternalDependencies: []string{
			"core",
		},
		ContainerServices: map[string]string{
			"echo": "echo.Echo",
			"db":   "gorm.DB",
		},
		Files: []string{
			"interface/interfaces.go",
			"service/email_service.go",
			"service/smtp_sender.go", 
			"service/mock_sender.go",
			"service/template_manager.go",
			"controller/email_controller.go",
			"repository/gorm/email_queue.go",
			"repository/gorm/template_repository.go",
			"repository/gorm/migrations.go",
			"module.go",
		},
	},
}

func ListAvailableModules() {
	fmt.Println("📦 Available modules:")
	fmt.Println()

	for _, module := range availableModules {
		fmt.Printf("  %s\n", module.Name)
		fmt.Printf("    %s\n", module.Description)
		if len(module.Dependencies) > 0 {
			fmt.Printf("    Dependencies: %v\n", module.Dependencies)
		}
		fmt.Println()
	}

	fmt.Println("Usage: sgk add <module>")
}
