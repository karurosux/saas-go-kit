package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/karurosux/saas-go-kit/cmd/sgk/commands"
	"github.com/karurosux/saas-go-kit/cmd/sgk/internal/crud"
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

	rootCmd.AddCommand(commands.NewCmd(createNewProjectWithModules))
	rootCmd.AddCommand(commands.InitCmd(project.InitProject))
	rootCmd.AddCommand(commands.AddCmd(addModuleWithAllDeps))
	rootCmd.AddCommand(commands.ListCmd(modules.ListAvailableModules, listInstalledModulesFromConfig))
	rootCmd.AddCommand(commands.UpdateCmd(modules.UpdateModule))
	rootCmd.AddCommand(commands.CrudCmd(crud.GenerateCRUDModule))
	rootCmd.AddCommand(commands.VersionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createNewProjectWithModules(projectName string, moduleList []string, goModule, database string) error {
	if err := project.CreateNewProject(projectName, moduleList, goModule, database); err != nil {
		return err
	}

	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(projectName); err != nil {
		return err
	}

	if err := embed.CopyCoreFromEmbed(); err != nil {
		return fmt.Errorf("failed to copy core templates: %w", err)
	}

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

func addModuleWithAllDeps(moduleName string, options map[string]interface{}) error {
	if !modules.IsModuleAvailable(moduleName) {
		return fmt.Errorf("unknown module '%s'. Run 'sgk list' to see available modules", moduleName)
	}

	config, err := project.LoadProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	if config.Modules == nil {
		config.Modules = make(map[string]project.ModuleInfo)
	}
	if _, exists := config.Modules[moduleName]; exists {
		return fmt.Errorf("module '%s' is already installed", moduleName)
	}

	moduleDef, err := modules.GetModule(moduleName)
	if err != nil {
		return err
	}

	for _, dep := range moduleDef.InternalDependencies {
		if dep == "core" {
			continue
		}
		
		if _, exists := config.Modules[dep]; !exists {
			fmt.Printf("⚠️  Module %s requires %s. Installing it first...\n", moduleName, dep)
			if err := addModuleWithAllDeps(dep, map[string]interface{}{}); err != nil {
				return fmt.Errorf("failed to install dependency %s: %w", dep, err)
			}
			config, err = project.LoadProjectConfig()
			if err != nil {
				return fmt.Errorf("failed to reload project config: %w", err)
			}
		} else {
			fmt.Printf("✅ Module %s dependency %s is already installed\n", moduleName, dep)
		}
	}

	if options["route_prefix"] == "" {
		options["route_prefix"] = fmt.Sprintf("/api/%s", moduleName)
	}

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

	if err := embed.CopyModuleFromEmbed(moduleName, templateData); err != nil {
		return err
	}

	config.Modules[moduleName] = project.ModuleInfo{
		Version:     moduleDef.Version,
		InstalledAt: time.Now(),
	}

	if err := project.SaveProjectConfig(config); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	if err := updateMainGoWithModule(config.Project.GoModule, moduleName); err != nil {
		return fmt.Errorf("failed to update main.go: %w", err)
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

func listInstalledModulesFromConfig() error {
	config, err := project.LoadProjectConfig()
	if err != nil {
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
		fmt.Println()
	}

	return nil
}

func updateMainGoWithModule(goModule, moduleName string) error {
	mainPath := "main.go"
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	mainContent := string(content)

	moduleImport := fmt.Sprintf(`"%s/internal/%s"`, goModule, moduleName)
	if strings.Contains(mainContent, moduleImport) {
		return nil
	}

	importPattern := `"gorm.io/gorm"`
	
	coreImport := fmt.Sprintf(`"%s/internal/core"`, goModule)
	if strings.Contains(mainContent, coreImport) {
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
		importReplacement := fmt.Sprintf(`"gorm.io/gorm"
	
	"%s/internal/core"
	%s`, goModule, moduleImport)
		mainContent = strings.Replace(mainContent, importPattern, importReplacement, 1)
	}

	registrationMarker := "// Register modules"
	moduleRegistration := fmt.Sprintf(`	if err := %s.RegisterModule(container); err != nil {
		log.Fatalf(core.ErrMsgModuleRegistration, "%s", err)
	}`, moduleName, moduleName)
	
	markerPos := strings.Index(mainContent, registrationMarker)
	if markerPos == -1 {
		return fmt.Errorf("could not find '// Register modules' marker in main.go")
	}
	
	afterMarker := mainContent[markerPos:]
	lines := strings.Split(afterMarker, "\n")
	insertLine := 1 // Default to line after marker
	
	inRegistrationBlock := false
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		
		if strings.Contains(line, "if err := ") && strings.Contains(line, ".RegisterModule(container)") {
			inRegistrationBlock = true
			continue
		}
		
		if inRegistrationBlock && line == "}" {
			insertLine = i + 1
			inRegistrationBlock = false
			continue
		}
		
		if !inRegistrationBlock && line != "" && !strings.Contains(line, "log.Fatalf") && insertLine > 1 {
			break
		}
	}
	
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertLine]...)
	newLines = append(newLines, moduleRegistration)
	newLines = append(newLines, lines[insertLine:]...)
	
	mainContent = mainContent[:markerPos] + strings.Join(newLines, "\n")

	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	return nil
}