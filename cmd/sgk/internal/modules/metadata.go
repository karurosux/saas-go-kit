package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ModuleMetadata struct {
	Modules      map[string]InstalledModule `json:"modules"`
	Dependencies []string                   `json:"dependencies"`
	UpdatedAt    time.Time                  `json:"updated_at"`
}

type InstalledModule struct {
	Version              string            `json:"version"`
	InstalledAt          time.Time         `json:"installed_at"`
	InternalDependencies []string          `json:"internal_dependencies"`
	ExternalDependencies []string          `json:"external_dependencies"`
	Configuration        map[string]string `json:"configuration"`
}

type ModuleRegistry struct {
	Modules map[string]ModuleVersions `json:"modules"`
}

type ModuleVersions struct {
	Latest   string                      `json:"latest"`
	Versions map[string]ModuleDefinition `json:"versions"`
}

func LoadModuleMetadata() (*ModuleMetadata, error) {
	path := "sgk.json"
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
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

func SaveModuleMetadata(metadata *ModuleMetadata) error {
	metadata.UpdatedAt = time.Now()
	
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("sgk.json", data, 0644)
}

func (m *ModuleMetadata) AddModule(name string, def ModuleDefinition, config map[string]string) {
	if m.Modules == nil {
		m.Modules = make(map[string]InstalledModule)
	}
	if m.Dependencies == nil {
		m.Dependencies = []string{}
	}

	m.Modules[name] = InstalledModule{
		Version:              def.Version,
		InstalledAt:          time.Now(),
		InternalDependencies: def.InternalDependencies,
		ExternalDependencies: def.Dependencies,
		Configuration:        config,
	}

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

func (m *ModuleMetadata) RemoveModule(name string) {
	delete(m.Modules, name)
	
}

func (m *ModuleMetadata) HasModule(name string) bool {
	_, exists := m.Modules[name]
	return exists
}

func (m *ModuleMetadata) GetModuleDependencyGraph() map[string][]string {
	graph := make(map[string][]string)
	
	for name, module := range m.Modules {
		graph[name] = module.InternalDependencies
	}
	
	return graph
}

func (m *ModuleMetadata) CheckDependencies() error {
	for moduleName, module := range m.Modules {
		for _, dep := range module.InternalDependencies {
			if dep == "core" {
				continue
			}
			
			if !m.HasModule(dep) {
				return fmt.Errorf("module %s requires %s, but it's not installed", moduleName, dep)
			}
		}
	}
	
	return nil
}

func (m *ModuleMetadata) GenerateGoModRequires() []string {
	requires := []string{
		"github.com/labstack/echo/v4 v4.11.3",
		"gorm.io/gorm v1.25.5",
		"gorm.io/driver/postgres v1.5.4",
		"gorm.io/driver/mysql v1.5.2",
		"gorm.io/driver/sqlite v1.5.4",
	}
	
	for _, dep := range m.Dependencies {
		requires = append(requires, dep)
	}
	
	return requires
}

func WriteModuleVersionFile(moduleName string, version string) error {
	versionFile := filepath.Join("internal", moduleName, ".version")
	return os.WriteFile(versionFile, []byte(version), 0644)
}

func ReadModuleVersionFile(moduleName string) (string, error) {
	versionFile := filepath.Join("internal", moduleName, ".version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}