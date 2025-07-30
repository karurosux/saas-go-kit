package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// availableModules defines all available modules
var availableModules = map[string]ModuleDefinition{
	"auth": {
		Name:        "auth",
		Description: "Complete authentication system with JWT, email verification, password reset",
		Dependencies: []string{
			"github.com/golang-jwt/jwt/v5",
			"golang.org/x/crypto",
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
		Description: "Subscription and billing management with Stripe integration",
		Dependencies: []string{
			"github.com/stripe/stripe-go/v76",
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
		Description: "Team management with role-based access control",
		Dependencies: []string{},
		Files: []string{
			"handlers.go",
			"service.go",
			"models.go",
			"interfaces.go",
			"module.go",
			"repositories/gorm/team_member_repository.go",
			"repositories/gorm/invitation_token_repository.go",
			"repositories/gorm/user_repository.go",
			"repositories/gorm/migrations.go",
		},
	},
	"notification": {
		Name:        "notification",
		Description: "Multi-channel notification system (email, SMS, push)",
		Dependencies: []string{
			"gopkg.in/gomail.v2",
		},
		Files: []string{
			"handlers.go",
			"service.go",
			"interfaces.go",
			"module.go",
			"smtp_provider.go",
		},
	},
	"health": {
		Name:        "health",
		Description: "Application health monitoring with multiple check types",
		Dependencies: []string{},
		Files: []string{
			"service.go",
			"interfaces.go",
			"module.go",
			"checkers.go",
			"gorm_checker.go",
		},
	},
	"role": {
		Name:        "role",
		Description: "Role-based access control and permissions management",
		Dependencies: []string{},
		Files: []string{
			"service.go",
			"interfaces.go",
			"module.go",
			"models.go",
			"middleware.go",
			"permissions.go",
			"repositories/gorm/role_repository.go",
			"repositories/gorm/user_role_repository.go",
		},
	},
	"job": {
		Name:        "job",
		Description: "Background job processing system",
		Dependencies: []string{},
		Files: []string{
			"handlers.go",
			"service.go",
			"interfaces.go",
			"module.go",
			"models.go",
			"queue.go",
			"repositories/gorm/job_repository.go",
			"repositories/gorm/job_result_repository.go",
			"repositories/gorm/migrations.go",
		},
	},
	"sse": {
		Name:        "sse",
		Description: "Server-sent events for real-time communication",
		Dependencies: []string{},
		Files: []string{
			"module.go",
			"sse.go",
		},
	},
	"container": {
		Name:        "container",
		Description: "Dependency injection container",
		Dependencies: []string{},
		Files: []string{
			"container.go",
			"interfaces.go",
			"module.go",
		},
	},
}

// ModuleDefinition defines a module's metadata
type ModuleDefinition struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Files        []string `json:"files"`
}

// isValidModule checks if a module name is valid
func isValidModule(name string) bool {
	_, exists := availableModules[name]
	return exists
}

// getModuleDependencies returns the dependencies for a module
func getModuleDependencies(name string) []string {
	if module, exists := availableModules[name]; exists {
		return module.Dependencies
	}
	return []string{}
}

// listAvailableModules prints all available modules
func listAvailableModules() {
	fmt.Println("📦 Available modules:")
	fmt.Println()
	
	for name, module := range availableModules {
		fmt.Printf("  %s\n", name)
		fmt.Printf("    %s\n", module.Description)
		if len(module.Dependencies) > 0 {
			fmt.Printf("    Dependencies: %v\n", module.Dependencies)
		}
		fmt.Println()
	}
	
	fmt.Println("Usage: saas-kit add <module>")
}

// listInstalledModules prints installed modules
func listInstalledModules() error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	if len(config.Modules) == 0 {
		fmt.Println("No modules installed.")
		fmt.Println("Run 'saas-kit add <module>' to install a module.")
		return nil
	}

	fmt.Println("📦 Installed modules:")
	fmt.Println()
	
	for name, info := range config.Modules {
		module := availableModules[name]
		fmt.Printf("  %s (v%s)\n", name, info.Version)
		fmt.Printf("    %s\n", module.Description)
		fmt.Printf("    Installed: %s\n", info.InstalledAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
	
	return nil
}

// removeModuleFiles removes module files from the project
func removeModuleFiles(moduleName string) error {
	modulePath := filepath.Join("internal", moduleName)
	
	// Remove module directory
	if err := os.RemoveAll(modulePath); err != nil {
		return fmt.Errorf("failed to remove module directory: %v", err)
	}
	
	// Remove config file
	configPath := filepath.Join("config", moduleName+".go")
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			return fmt.Errorf("failed to remove config file: %v", err)
		}
	}
	
	// Remove documentation
	docPath := filepath.Join("docs", moduleName+".md")
	if _, err := os.Stat(docPath); err == nil {
		if err := os.Remove(docPath); err != nil {
			return fmt.Errorf("failed to remove documentation: %v", err)
		}
	}
	
	// TODO: Remove migrations (need to be more careful here)
	// TODO: Remove from main.go integration
	
	return nil
}