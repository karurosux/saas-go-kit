package modules

import (
	"fmt"
	"os"
	"time"
)

// AddModule adds a module to the project
func AddModule(moduleName string, options map[string]interface{}) error {
	return AddModuleWithOptions(moduleName, options, true)
}

// AddModuleWithOptions adds a module with control over main.go updates
func AddModuleWithOptions(moduleName string, options map[string]interface{}, updateMainGo bool) error {
	// Validate module exists
	if !IsModuleAvailable(moduleName) {
		return fmt.Errorf("unknown module '%s'. Run 'sgk list' to see available modules", moduleName)
	}

	// Load module metadata
	metadata, err := LoadModuleMetadata()
	if err != nil {
		return fmt.Errorf("failed to load module metadata: %w", err)
	}

	// Check if already installed
	if metadata.HasModule(moduleName) {
		return fmt.Errorf("module '%s' is already installed", moduleName)
	}

	// Get module definition
	moduleDef, err := GetModule(moduleName)
	if err != nil {
		return err
	}

	// Check internal dependencies
	for _, dep := range moduleDef.InternalDependencies {
		// Skip core modules (always available)
		if dep == "core-go" || dep == "errors-go" || dep == "response-go" || 
		   dep == "validator-go" || dep == "container-go" {
			continue
		}
		
		// Check if dependency is installed
		if !metadata.HasModule(dep) && dep != moduleName {
			fmt.Printf("⚠️  Module %s requires %s. Installing it first...\n", moduleName, dep)
			// Recursively install dependency
			if err := AddModule(dep, map[string]interface{}{}); err != nil {
				return fmt.Errorf("failed to install dependency %s: %w", dep, err)
			}
			// Reload metadata after installing dependency
			metadata, _ = LoadModuleMetadata()
		}
	}

	// Load project config (simplified to avoid circular imports for now)
	// TODO: Integrate with project package properly
	_ = "default" // projectName
	_ = "default" // goModule

	// Set default route prefix
	if options["route_prefix"] == "" {
		options["route_prefix"] = fmt.Sprintf("/api/%s", moduleName)
	}

	// Module copying will be handled by main package to avoid circular imports
	// TODO: Fix circular import issue and enable module copying

	// Convert options to string map for storage
	configMap := make(map[string]string)
	for k, v := range options {
		configMap[k] = fmt.Sprint(v)
	}

	// Update metadata
	metadata.AddModule(moduleName, moduleDef, configMap)

	// Save metadata
	if err := SaveModuleMetadata(metadata); err != nil {
		return err
	}

	return nil
}

// RemoveModule removes a module from the project
func RemoveModule(moduleName string) error {
	// Load module metadata
	metadata, err := LoadModuleMetadata()
	if err != nil {
		return fmt.Errorf("failed to load module metadata: %w", err)
	}

	// Check if module is installed
	if !metadata.HasModule(moduleName) {
		return fmt.Errorf("module '%s' is not installed", moduleName)
	}

	// Check if other modules depend on this one
	graph := metadata.GetModuleDependencyGraph()
	for name, deps := range graph {
		if name == moduleName {
			continue
		}
		for _, dep := range deps {
			if dep == moduleName {
				return fmt.Errorf("cannot remove module '%s': module '%s' depends on it", moduleName, name)
			}
		}
	}

	// Remove module directory
	moduleDir := fmt.Sprintf("internal/%s", moduleName)
	if err := os.RemoveAll(moduleDir); err != nil {
		return fmt.Errorf("failed to remove module directory: %w", err)
	}

	// Remove from metadata
	metadata.RemoveModule(moduleName)

	// Save metadata
	if err := SaveModuleMetadata(metadata); err != nil {
		return err
	}

	fmt.Printf("✅ Module '%s' removed successfully!\n", moduleName)
	return nil
}

// UpdateModule updates a module to the latest version
func UpdateModule(moduleName string) error {
	// Load module metadata
	metadata, err := LoadModuleMetadata()
	if err != nil {
		return fmt.Errorf("failed to load module metadata: %w", err)
	}

	// Check if module is installed
	if !metadata.HasModule(moduleName) {
		return fmt.Errorf("module '%s' is not installed", moduleName)
	}

	// Get current and latest versions
	installedModule := metadata.Modules[moduleName]
	latestDef, err := GetModule(moduleName)
	if err != nil {
		return err
	}

	if installedModule.Version == latestDef.Version {
		fmt.Printf("Module '%s' is already up to date (v%s)\n", moduleName, installedModule.Version)
		return nil
	}

	fmt.Printf("Updating module '%s' from v%s to v%s...\n", moduleName, installedModule.Version, latestDef.Version)

	// Create backup of existing module
	backupDir := fmt.Sprintf("internal/%s.backup.%d", moduleName, time.Now().Unix())
	moduleDir := fmt.Sprintf("internal/%s", moduleName)
	
	if err := os.Rename(moduleDir, backupDir); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Reinstall module with latest version
	options := make(map[string]interface{})
	for k, v := range installedModule.Configuration {
		options[k] = v
	}

	if err := AddModuleWithOptions(moduleName, options, false); err != nil {
		// Restore backup on failure
		os.RemoveAll(moduleDir)
		os.Rename(backupDir, moduleDir)
		return fmt.Errorf("failed to update module: %w", err)
	}

	// Remove backup on success
	os.RemoveAll(backupDir)

	fmt.Printf("✅ Module '%s' updated successfully!\n", moduleName)
	return nil
}

// ListInstalledModules prints installed modules to stdout
func ListInstalledModules() error {
	// For now, try to use the old metadata format for backward compatibility
	// TODO: Update to use project config
	metadata, err := LoadModuleMetadata()
	if err != nil {
		// If metadata doesn't exist, create empty one
		if os.IsNotExist(err) {
			fmt.Println("📦 No modules installed yet.")
			fmt.Println("Run 'sgk add <module>' to install a module.")
			return nil
		}
		return fmt.Errorf("failed to load module metadata: %w", err)
	}

	if len(metadata.Modules) == 0 {
		fmt.Println("📦 No modules installed yet.")
		fmt.Println("Run 'sgk add <module>' to install a module.")
		return nil
	}

	fmt.Println("📦 Installed modules:")
	fmt.Println()

	for name, module := range metadata.Modules {
		fmt.Printf("  %s (v%s)\n", name, module.Version)
		fmt.Printf("    Installed: %s\n", module.InstalledAt.Format("2006-01-02 15:04:05"))
		if len(module.Configuration) > 0 {
			fmt.Printf("    Config: %v\n", module.Configuration)
		}
		fmt.Println()
	}

	return nil
}

// getStringOption safely gets a string option with a default value
func getStringOption(options map[string]interface{}, key, defaultValue string) string {
	if val, ok := options[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}