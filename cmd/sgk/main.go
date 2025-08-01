package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/karurosux/saas-go-kit/cmd/sgk/commands"
	"github.com/karurosux/saas-go-kit/cmd/sgk/internal/embed"
	"github.com/karurosux/saas-go-kit/cmd/sgk/internal/modules"
	"github.com/karurosux/saas-go-kit/cmd/sgk/internal/project"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "sgk",
		Short: "SaaS Go Kit - Copy-paste modular components for Go SaaS applications",
		Long: `SaaS Go Kit CLI allows you to add pre-built, customizable modules to your Go application.
Similar to shadcn/ui, you own the code and can modify it as needed.

Available modules: auth, subscription, team, notification, health, role, job, sse, container`,
	}

	// Add all commands with dependency injection to avoid circular imports
	rootCmd.AddCommand(commands.NewCmd(createNewProjectWithModules))
	rootCmd.AddCommand(commands.InitCmd(project.InitProject))
	rootCmd.AddCommand(commands.AddCmd(addModuleWithAllDeps))
	rootCmd.AddCommand(commands.ListCmd(modules.ListAvailableModules, listInstalledModulesFromConfig))
	rootCmd.AddCommand(commands.RemoveCmd(modules.RemoveModule))
	rootCmd.AddCommand(commands.UpdateCmd(modules.UpdateModule))
	rootCmd.AddCommand(commands.GenerateCmd(embed.GenerateClients))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// createNewProjectWithModules creates a new project with all dependencies properly handled
func createNewProjectWithModules(projectName string, moduleList []string, goModule, database string) error {
	// Create basic project structure
	if err := project.CreateNewProject(projectName, moduleList, goModule, database); err != nil {
		return err
	}

	// Store current directory and change to project directory
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(projectName); err != nil {
		return err
	}

	// Copy core templates
	if err := embed.CopyCoreFromEmbed(); err != nil {
		return fmt.Errorf("failed to copy core templates: %w", err)
	}

	// Add specified modules
	for _, moduleName := range moduleList {
		options := map[string]interface{}{
			"database": database,
		}
		
		if err := addModuleWithAllDeps(moduleName, options); err != nil {
			return fmt.Errorf("failed to add module %s: %w", moduleName, err)
		}
	}

	return nil
}

// addModuleWithAllDeps adds a single module with all dependencies handled
func addModuleWithAllDeps(moduleName string, options map[string]interface{}) error {
	// Validate module exists
	if !modules.IsModuleAvailable(moduleName) {
		return fmt.Errorf("unknown module '%s'. Run 'sgk list' to see available modules", moduleName)
	}

	// Load project config for template data
	config, err := project.LoadProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	// Check if already installed
	if config.Modules == nil {
		config.Modules = make(map[string]project.ModuleInfo)
	}
	if _, exists := config.Modules[moduleName]; exists {
		return fmt.Errorf("module '%s' is already installed", moduleName)
	}

	// Get module definition
	moduleDef, err := modules.GetModule(moduleName)
	if err != nil {
		return err
	}

	// Set default route prefix
	if options["route_prefix"] == "" {
		options["route_prefix"] = fmt.Sprintf("/api/%s", moduleName)
	}

	// Create template data
	templateData := embed.TemplateData{
		Project: struct {
			Name     string
			GoModule string
			Database string
		}{
			Name:     config.Project.Name,
			GoModule: config.Project.GoModule,
			Database: getStringOption(options, "database", "postgres"),
		},
		Module: struct {
			Name string
		}{
			Name: moduleName,
		},
	}

	// Copy module files
	if err := embed.CopyModuleFromEmbed(moduleName, templateData); err != nil {
		return err
	}

	// Convert options to interface{} map for storage
	configMap := make(map[string]interface{})
	for k, v := range options {
		configMap[k] = v
	}

	// Add module to project config
	config.Modules[moduleName] = project.ModuleInfo{
		Version:              moduleDef.Version,
		InstalledAt:          time.Now(),
		Config:               configMap,
		InternalDependencies: moduleDef.InternalDependencies,
		ExternalDependencies: moduleDef.Dependencies,
	}
	
	// Update dependencies
	for _, dep := range moduleDef.Dependencies {
		found := false
		for _, existing := range config.Dependencies {
			if existing == dep {
				found = true
				break
			}
		}
		if !found {
			config.Dependencies = append(config.Dependencies, dep)
		}
	}

	// Save project config
	if err := project.SaveProjectConfig(config); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	// Update main.go to register the new module
	if err := updateMainGoWithModule(config.Project.GoModule, moduleName); err != nil {
		return fmt.Errorf("failed to update main.go: %w", err)
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

// listInstalledModulesFromConfig lists installed modules using project config
func listInstalledModulesFromConfig() error {
	config, err := project.LoadProjectConfig()
	if err != nil {
		// If config doesn't exist, no modules are installed
		fmt.Println("📦 No modules installed yet.")
		fmt.Println("Run 'sgk init' to initialize a project, then 'sgk add <module>' to install modules.")
		return nil
	}

	if len(config.Modules) == 0 {
		fmt.Println("📦 No modules installed yet.")
		fmt.Println("Run 'sgk add <module>' to install a module.")
		return nil
	}

	fmt.Println("📦 Installed modules:")
	fmt.Println()

	for name, module := range config.Modules {
		fmt.Printf("  %s (v%s)\n", name, module.Version)
		fmt.Printf("    Installed: %s\n", module.InstalledAt.Format("2006-01-02 15:04:05"))
		if len(module.Config) > 0 {
			fmt.Printf("    Config: %v\n", module.Config)
		}
		if len(module.InternalDependencies) > 0 {
			fmt.Printf("    Dependencies: %v\n", module.InternalDependencies)
		}
		fmt.Println()
	}

	return nil
}

// updateMainGoWithModule updates the main.go file to register a new module
func updateMainGoWithModule(goModule, moduleName string) error {
	// Read existing main.go
	mainPath := "main.go"
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	mainContent := string(content)

	// Check if module is already imported
	moduleImport := fmt.Sprintf(`"%s/internal/%s"`, goModule, moduleName)
	if strings.Contains(mainContent, moduleImport) {
		// Module already registered
		return nil
	}

	// Find the imports section and add module import in the right place
	// Look for the pattern of external imports followed by internal imports
	importPattern := `"gorm.io/gorm"`
	
	// Check if there are already internal module imports
	coreImport := fmt.Sprintf(`"%s/internal/core"`, goModule)
	if strings.Contains(mainContent, coreImport) {
		// Add after the last internal import
		internalImports := []string{}
		lines := strings.Split(mainContent, "\n")
		for _, line := range lines {
			if strings.Contains(line, goModule+"/internal/") {
				internalImports = append(internalImports, strings.TrimSpace(line))
			}
		}
		if len(internalImports) > 0 {
			lastImport := internalImports[len(internalImports)-1]
			mainContent = strings.Replace(mainContent, lastImport, lastImport+"\n\t"+moduleImport, 1)
		}
	} else {
		// Add after gorm import with proper spacing
		importReplacement := fmt.Sprintf(`"gorm.io/gorm"
	
	"%s/internal/core"
	%s`, goModule, moduleImport)
		mainContent = strings.Replace(mainContent, importPattern, importReplacement, 1)
	}

	// Add module registration
	registrationMarker := "// Register modules"
	moduleRegistration := fmt.Sprintf(`	if err := %s.RegisterModule(container); err != nil {
		log.Fatalf("Failed to register %s module: %%v", err)
	}`, moduleName, moduleName)
	
	// Find the position after "// Register modules" and any existing registrations
	markerPos := strings.Index(mainContent, registrationMarker)
	if markerPos == -1 {
		return fmt.Errorf("could not find '// Register modules' marker in main.go")
	}
	
	// Find where to insert the new registration
	afterMarker := mainContent[markerPos:]
	lines := strings.Split(afterMarker, "\n")
	insertLine := 1 // Default to line after marker
	
	// Find existing registrations
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, ".RegisterModule(container)") {
			insertLine = i + 1
		} else if line == "" && insertLine > 1 {
			// Found empty line after registrations
			break
		}
	}
	
	// Insert the new registration
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertLine]...)
	newLines = append(newLines, moduleRegistration)
	newLines = append(newLines, lines[insertLine:]...)
	
	mainContent = mainContent[:markerPos] + strings.Join(newLines, "\n")

	// Write updated main.go
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	return nil
}