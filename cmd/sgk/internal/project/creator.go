package project

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// createNewProject creates a new project with specified modules
func CreateNewProject(projectName string, modules []string, goModule, database string) error {
	// Validate project name
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	// Check if directory already exists
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("directory '%s' already exists", projectName)
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Store original directory
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Change to project directory
	if err := os.Chdir(projectName); err != nil {
		return err
	}
	defer os.Chdir(originalDir)

	// Set go module path
	if goModule == "" {
		goModule = projectName
	}

	// Create go.mod
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/labstack/echo/v4 v4.11.3
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
)
`, goModule)

	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Initialize project with configuration
	if err := InitProjectWithConfig(projectName, goModule, database); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	// Note: Core templates and modules will be handled by main package to avoid circular imports

	// Create main.go
	mainContent := generateMainGo(goModule, modules)
	if err := os.WriteFile("main.go", []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	return nil
}

// validateProjectName validates the project name
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check for valid Go module name characters
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("project name must contain only letters, numbers, hyphens, and underscores")
	}

	return nil
}

// generateMainGo generates the main.go file content
func generateMainGo(goModule string, modules []string) string {
	var imports []string
	var moduleRegistrations []string

	// Add core import
	imports = append(imports, fmt.Sprintf(`"%s/internal/core"`, goModule))

	// Add module imports and registrations
	for _, module := range modules {
		imports = append(imports, fmt.Sprintf(`"%s/internal/%s"`, goModule, module))
		moduleRegistrations = append(moduleRegistrations, fmt.Sprintf(`	if err := %s.RegisterModule(container); err != nil {
		log.Fatalf("Failed to register %s module: %%v", err)
	}`, module, module))
	}

	return fmt.Sprintf(`package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	%s
)

func main() {
	// Load environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost/dbname?sslmode=disable"
	}

	// Initialize database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %%v", err)
	}

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize container
	container := core.NewContainer()
	container.Set("echo", e)
	container.Set("db", db)

	// Register modules
%s

	// Start server
	log.Printf("Server starting on port %%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Server failed to start: %%v", err)
	}
}
`, strings.Join(imports, "\n\t"), strings.Join(moduleRegistrations, "\n"))
}