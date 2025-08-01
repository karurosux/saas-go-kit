package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ModuleMetadata tracks module information in the project
type ModuleMetadata struct {
	Modules      map[string]InstalledModule `json:"modules"`
	Dependencies []string                   `json:"dependencies"`
	UpdatedAt    time.Time                  `json:"updated_at"`
}

// InstalledModule represents an installed module's metadata
type InstalledModule struct {
	Version              string            `json:"version"`
	InstalledAt          time.Time         `json:"installed_at"`
	InternalDependencies []string          `json:"internal_dependencies"`
	ExternalDependencies []string          `json:"external_dependencies"`
	Configuration        map[string]string `json:"configuration"`
}

// ModuleRegistry tracks all available module versions
type ModuleRegistry struct {
	Modules map[string]ModuleVersions `json:"modules"`
}

// ModuleVersions tracks versions for a module
type ModuleVersions struct {
	Latest   string                      `json:"latest"`
	Versions map[string]ModuleDefinition `json:"versions"`
}

// LoadModuleMetadata loads the module metadata from sgk.json
func LoadModuleMetadata() (*ModuleMetadata, error) {
	path := "sgk.json"
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty metadata
			return &ModuleMetadata{
				Modules:      make(map[string]InstalledModule),
				Dependencies: []string{},
				UpdatedAt:    time.Now(),
			}, nil
		}
		return nil, err
	}

	var metadata ModuleMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// SaveModuleMetadata saves the module metadata to sgk.json
func SaveModuleMetadata(metadata *ModuleMetadata) error {
	metadata.UpdatedAt = time.Now()
	
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("sgk.json", data, 0644)
}

// AddModule adds a module to the metadata
func (m *ModuleMetadata) AddModule(name string, def ModuleDefinition, config map[string]string) {
	if m.Modules == nil {
		m.Modules = make(map[string]InstalledModule)
	}
	if m.Dependencies == nil {
		m.Dependencies = []string{}
	}

	// Add module
	m.Modules[name] = InstalledModule{
		Version:              def.Version,
		InstalledAt:          time.Now(),
		InternalDependencies: def.InternalDependencies,
		ExternalDependencies: def.Dependencies,
		Configuration:        config,
	}

	// Add external dependencies (avoid duplicates)
	for _, dep := range def.Dependencies {
		found := false
		for _, existing := range m.Dependencies {
			if existing == dep {
				found = true
				break
			}
		}
		if !found {
			m.Dependencies = append(m.Dependencies, dep)
		}
	}
}

// RemoveModule removes a module from the metadata
func (m *ModuleMetadata) RemoveModule(name string) {
	delete(m.Modules, name)
	
	// TODO: Clean up dependencies that are no longer needed
}

// HasModule checks if a module is installed
func (m *ModuleMetadata) HasModule(name string) bool {
	_, exists := m.Modules[name]
	return exists
}

// GetModuleDependencyGraph builds a dependency graph for installed modules
func (m *ModuleMetadata) GetModuleDependencyGraph() map[string][]string {
	graph := make(map[string][]string)
	
	for name, module := range m.Modules {
		graph[name] = module.InternalDependencies
	}
	
	return graph
}

// CheckDependencies verifies all module dependencies are satisfied
func (m *ModuleMetadata) CheckDependencies() error {
	for moduleName, module := range m.Modules {
		for _, dep := range module.InternalDependencies {
			// Skip core dependencies (they're always available)
			if dep == "core-go" || dep == "errors-go" || dep == "response-go" || 
			   dep == "validator-go" || dep == "container-go" {
				continue
			}
			
			// Check if dependency is installed
			if !m.HasModule(dep) {
				return fmt.Errorf("module %s requires %s, but it's not installed", moduleName, dep)
			}
		}
	}
	
	return nil
}

// GenerateGoModRequires generates go.mod require statements
func (m *ModuleMetadata) GenerateGoModRequires() []string {
	requires := []string{
		"github.com/labstack/echo/v4 v4.11.3",
		"gorm.io/gorm v1.25.5",
		"gorm.io/driver/postgres v1.5.4",
		"gorm.io/driver/mysql v1.5.2",
		"gorm.io/driver/sqlite v1.5.4",
	}
	
	// Add module-specific dependencies
	for _, dep := range m.Dependencies {
		requires = append(requires, dep)
	}
	
	return requires
}

// WriteModuleVersionFile writes a version file for tracking
func WriteModuleVersionFile(moduleName string, version string) error {
	versionFile := filepath.Join("internal", moduleName, ".version")
	return os.WriteFile(versionFile, []byte(version), 0644)
}

// ReadModuleVersionFile reads the version file for a module
func ReadModuleVersionFile(moduleName string) (string, error) {
	versionFile := filepath.Join("internal", moduleName, ".version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}