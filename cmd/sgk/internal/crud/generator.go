package crud

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/karurosux/saas-go-kit/cmd/sgk/internal/embed"
	"github.com/karurosux/saas-go-kit/cmd/sgk/internal/project"
)

func GenerateCRUDModule(moduleName string) error {
	config, err := project.LoadProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	if config.Modules == nil {
		config.Modules = make(map[string]project.ModuleInfo)
	}
	if _, exists := config.Modules[moduleName]; exists {
		return fmt.Errorf("module '%s' already exists", moduleName)
	}

	templateData := embed.TemplateData{
		Project: struct {
			Name     string
			GoModule string
			Database string
		}{
			Name:     config.Project.Name,
			GoModule: config.Project.GoModule,
			Database: config.Project.Database,
		},
		Module: struct {
			Name string
		}{
			Name: moduleName,
		},
	}

	if err := copyCRUDTemplate(moduleName, templateData); err != nil {
		return fmt.Errorf("failed to copy CRUD template: %w", err)
	}

	if err := updateMainGoWithCRUDModule(config.Project.GoModule, moduleName); err != nil {
		return fmt.Errorf("failed to update main.go: %w", err)
	}

	config.Modules[moduleName] = project.ModuleInfo{
		Version:     "1.0.0",
		InstalledAt: time.Now(),
	}

	if err := project.SaveProjectConfig(config); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	return nil
}

func copyCRUDTemplate(moduleName string, data embed.TemplateData) error {
	return embed.CopyCRUDModuleFromEmbed(moduleName, embed.CRUDTemplateData{
		Project: data.Project,
		ModuleName: moduleName,
		ModuleNameCap: strings.Title(moduleName),
	})
}

func updateMainGoWithCRUDModule(goModule, moduleName string) error {
	mainPath := "main.go"
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	mainContent := string(content)

	moduleImport := fmt.Sprintf(`%s "%s/internal/%s"`, moduleName, goModule, moduleName)
	if strings.Contains(mainContent, moduleImport) {
		return nil
	}

	lines := strings.Split(mainContent, "\n")
	
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		if strings.Contains(trimmedLine, fmt.Sprintf(`"%s/internal/core"`, goModule)) {
			lines = append(lines[:i+1], append([]string{"\t" + moduleImport}, lines[i+1:]...)...)
			break
		}
	}
	
	lastModuleRegistrationLine := -1
	for i, line := range lines {
		if strings.Contains(line, ".RegisterModule(container)") {
			lastModuleRegistrationLine = i + 2
		}
	}
	
	if lastModuleRegistrationLine > 0 {
		registrationLine := fmt.Sprintf("\tif err := %s.RegisterModule(container); err != nil {", moduleName)
		logLine := fmt.Sprintf("\t\tlog.Fatalf(core.ErrMsgModuleRegistration, \"%s\", err)", moduleName)
		closeLine := "\t}"
		
		newLines := []string{registrationLine, logLine, closeLine}
		lines = append(lines[:lastModuleRegistrationLine+1], append(newLines, lines[lastModuleRegistrationLine+1:]...)...)
	}

	mainContent = strings.Join(lines, "\n")
	return os.WriteFile(mainPath, []byte(mainContent), 0644)
}