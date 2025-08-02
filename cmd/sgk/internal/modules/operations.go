package modules

import (
	"fmt"
	"os"
	"time"
)

func AddModule(moduleName string, options map[string]interface{}) error {
	return AddModuleWithOptions(moduleName, options, true)
}

func AddModuleWithOptions(moduleName string, options map[string]interface{}, updateMainGo bool) error {
	if !IsModuleAvailable(moduleName) {
		return fmt.Errorf("unknown module '%s'. Run 'sgk list' to see available modules", moduleName)
	}

	// Load module metadata
	metadata, err := LoadModuleMetadata()
	if err != nil {
		return fmt.Errorf("failed to load module metadata: %w", err)
	}

	if metadata.HasModule(moduleName) {
		return fmt.Errorf("module '%s' is already installed", moduleName)
	}

	moduleDef, err := GetModule(moduleName)
	if err != nil {
		return err
	}

	for _, dep := range moduleDef.InternalDependencies {
		if dep == "core" {
			continue
		}
		
		if !metadata.HasModule(dep) && dep != moduleName {
			fmt.Printf("⚠️  Module %s requires %s. Installing it first...\n", moduleName, dep)
			if err := AddModule(dep, map[string]interface{}{}); err != nil {
				return fmt.Errorf("failed to install dependency %s: %w", dep, err)
			}
			metadata, _ = LoadModuleMetadata()
		}
	}

	_ = "default" // projectName
	_ = "default" // goModule

	if options["route_prefix"] == "" {
		options["route_prefix"] = fmt.Sprintf("/api/%s", moduleName)
	}


	configMap := make(map[string]string)
	for k, v := range options {
		configMap[k] = fmt.Sprint(v)
	}

	metadata.AddModule(moduleName, moduleDef, configMap)

	// Save metadata
	if err := SaveModuleMetadata(metadata); err != nil {
		return err
	}

	return nil
}

func RemoveModule(moduleName string) error {
	// Load module metadata
	metadata, err := LoadModuleMetadata()
	if err != nil {
		return fmt.Errorf("failed to load module metadata: %w", err)
	}

	if !metadata.HasModule(moduleName) {
		return fmt.Errorf("module '%s' is not installed", moduleName)
	}

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

	moduleDir := fmt.Sprintf("internal/%s", moduleName)
	if err := os.RemoveAll(moduleDir); err != nil {
		return fmt.Errorf("failed to remove module directory: %w", err)
	}

	metadata.RemoveModule(moduleName)

	// Save metadata
	if err := SaveModuleMetadata(metadata); err != nil {
		return err
	}

	fmt.Printf("✅ Module '%s' removed successfully!\n", moduleName)
	return nil
}

func UpdateModule(moduleName string) error {
	// Load module metadata
	metadata, err := LoadModuleMetadata()
	if err != nil {
		return fmt.Errorf("failed to load module metadata: %w", err)
	}

	if !metadata.HasModule(moduleName) {
		return fmt.Errorf("module '%s' is not installed", moduleName)
	}

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

	backupDir := fmt.Sprintf("internal/%s.backup.%d", moduleName, time.Now().Unix())
	moduleDir := fmt.Sprintf("internal/%s", moduleName)
	
	if err := os.Rename(moduleDir, backupDir); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	options := make(map[string]interface{})
	for k, v := range installedModule.Configuration {
		options[k] = v
	}

	if err := AddModuleWithOptions(moduleName, options, false); err != nil {
		os.RemoveAll(moduleDir)
		os.Rename(backupDir, moduleDir)
		return fmt.Errorf("failed to update module: %w", err)
	}

	os.RemoveAll(backupDir)

	fmt.Printf("✅ Module '%s' updated successfully!\n", moduleName)
	return nil
}

func ListInstalledModules() error {
	metadata, err := LoadModuleMetadata()
	if err != nil {
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

func getStringOption(options map[string]interface{}, key, defaultValue string) string {
	if val, ok := options[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}