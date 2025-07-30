package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProjectConfig represents the saas-kit.json configuration
type ProjectConfig struct {
	Version string `json:"version"`
	Project struct {
		Name     string `json:"name"`
		GoModule string `json:"go_module"`
	} `json:"project"`
	Modules      map[string]ModuleInfo `json:"modules"`
	Dependencies []string              `json:"dependencies"`
}

// ModuleInfo stores information about an installed module
type ModuleInfo struct {
	Version     string                 `json:"version"`
	InstalledAt time.Time              `json:"installed_at"`
	Config      map[string]interface{} `json:"config"`
}

// TemplateData holds data for template rendering
type TemplateData struct {
	Project   ProjectInfo            `json:"project"`
	Module    ModuleTemplateInfo     `json:"module"`
	Options   map[string]interface{} `json:"options"`
	Timestamp time.Time              `json:"timestamp"`
}

// ProjectInfo contains project-level information
type ProjectInfo struct {
	Name     string `json:"name"`
	GoModule string `json:"go_module"`
	Database string `json:"database"`
}

// ModuleTemplateInfo contains module-specific template information
type ModuleTemplateInfo struct {
	Name        string                 `json:"name"`
	RoutePrefix string                 `json:"route_prefix"`
	Database    string                 `json:"database"`
	Options     map[string]interface{} `json:"options"`
}

const configFileName = "saas-kit.json"

// initProject initializes a new saas-kit project
func initProject() error {
	// Check if already initialized
	if _, err := os.Stat(configFileName); err == nil {
		return fmt.Errorf("project already initialized (saas-kit.json exists)")
	}

	// Get current directory name as project name
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectName := filepath.Base(pwd)

	// Create basic project config
	config := ProjectConfig{
		Version: "1.0.0",
		Modules: make(map[string]ModuleInfo),
		Dependencies: []string{
			"github.com/labstack/echo/v4",
			"github.com/golang-jwt/jwt/v5",
		},
	}
	
	config.Project.Name = projectName
	config.Project.GoModule = fmt.Sprintf("github.com/user/%s", projectName)

	// Create directories
	dirs := []string{"internal", "config", "migrations", "docs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Copy core utilities
	if err := copyCore(); err != nil {
		return err
	}

	// Save config
	return saveConfig(config)
}

// loadConfig loads the project configuration
func loadConfig() (*ProjectConfig, error) {
	file, err := os.Open(configFileName)
	if err != nil {
		return nil, fmt.Errorf("project not initialized (run 'saas-kit init')")
	}
	defer file.Close()

	var config ProjectConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveConfig saves the project configuration
func saveConfig(config ProjectConfig) error {
	file, err := os.Create(configFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

// addModule adds a module to the project
func addModule(moduleName string, options map[string]interface{}) error {
	// Load existing config
	config, err := loadConfig()
	if err != nil {
		return err
	}

	// Check if module already exists
	if _, exists := config.Modules[moduleName]; exists {
		return fmt.Errorf("module '%s' is already installed", moduleName)
	}

	// Validate module exists
	if !isValidModule(moduleName) {
		return fmt.Errorf("unknown module '%s'. Run 'saas-kit list' to see available modules", moduleName)
	}

	// Set default route prefix
	if options["route_prefix"] == "" {
		options["route_prefix"] = fmt.Sprintf("/api/%s", moduleName)
	}

	// Create template data
	templateData := TemplateData{
		Project: ProjectInfo{
			Name:     config.Project.Name,
			GoModule: config.Project.GoModule,
			Database: getStringOption(options, "database", "postgres"),
		},
		Module: ModuleTemplateInfo{
			Name:        moduleName,
			RoutePrefix: getStringOption(options, "route_prefix", fmt.Sprintf("/api/%s", moduleName)),
			Database:    getStringOption(options, "database", "postgres"),
			Options:     options,
		},
		Options:   options,
		Timestamp: time.Now(),
	}

	// Copy module files
	if err := copyModuleTemplate(moduleName, templateData); err != nil {
		return err
	}

	// Update config
	config.Modules[moduleName] = ModuleInfo{
		Version:     "1.0.0", // TODO: Get from template metadata
		InstalledAt: time.Now(),
		Config:      options,
	}

	// Add module dependencies
	moduleDeps := getModuleDependencies(moduleName)
	config.Dependencies = append(config.Dependencies, moduleDeps...)
	config.Dependencies = removeDuplicates(config.Dependencies)

	return saveConfig(*config)
}

// removeModule removes a module from the project
func removeModule(moduleName string) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	// Check if module exists
	if _, exists := config.Modules[moduleName]; !exists {
		return fmt.Errorf("module '%s' is not installed", moduleName)
	}

	// Remove module files
	if err := removeModuleFiles(moduleName); err != nil {
		return err
	}

	// Update config
	delete(config.Modules, moduleName)

	return saveConfig(*config)
}

// updateModule updates a module to the latest version
func updateModule(moduleName string) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	// Check if module exists
	moduleInfo, exists := config.Modules[moduleName]
	if !exists {
		return fmt.Errorf("module '%s' is not installed", moduleName)
	}

	// Re-copy module files with existing options
	templateData := TemplateData{
		Project: ProjectInfo{
			Name:     config.Project.Name,
			GoModule: config.Project.GoModule,
		},
		Module: ModuleTemplateInfo{
			Name:    moduleName,
			Options: moduleInfo.Config,
		},
		Options:   moduleInfo.Config,
		Timestamp: time.Now(),
	}

	if err := copyModuleTemplate(moduleName, templateData); err != nil {
		return err
	}

	// Update version and timestamp
	moduleInfo.Version = "1.0.0" // TODO: Get latest version
	moduleInfo.InstalledAt = time.Now()
	config.Modules[moduleName] = moduleInfo

	return saveConfig(*config)
}

// Helper functions

func getStringOption(options map[string]interface{}, key, defaultValue string) string {
	if val, ok := options[key].(string); ok && val != "" {
		return val
	}
	return defaultValue
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}