package embed

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// templatesFS holds all embedded template files
//
//go:embed templates
var templatesFS embed.FS

type TemplateData struct {
	Project struct {
		Name     string
		GoModule string
		Database string
	}
	Module struct {
		Name string
	}
}

type CRUDTemplateData struct {
	Project struct {
		Name     string
		GoModule string
		Database string
	}
	ModuleName    string
	ModuleNameCap string
}

// CopyModuleFromEmbed copies a module template from embedded filesystem
func CopyModuleFromEmbed(moduleName string, data TemplateData) error {
	// Destination module directory
	moduleDir := filepath.Join("internal", moduleName)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return err
	}

	// Read from embedded templates
	templatePath := fmt.Sprintf("templates/%s", moduleName)

	// Walk through embedded files
	err := fs.WalkDir(templatesFS, templatePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Get relative path from module root
		relPath := strings.TrimPrefix(path, templatePath+"/")
		destPath := filepath.Join(moduleDir, relPath)

		// Create destination directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Read file content
		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Check if it's a template file
		if strings.HasSuffix(path, ".tmpl") {
			// Remove .tmpl extension from destination
			destPath = strings.TrimSuffix(destPath, ".tmpl")

			// Parse and execute template
			tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", path, err)
			}

			// Create destination file
			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			// Execute template
			if err := tmpl.Execute(destFile, data); err != nil {
				return fmt.Errorf("failed to execute template %s: %w", path, err)
			}
		} else {
			// For Go files, process template placeholders
			if strings.HasSuffix(path, ".go") {
				contentStr := string(content)
				// Replace template placeholders
				contentStr = strings.ReplaceAll(contentStr, "{{.Project.GoModule}}", data.Project.GoModule)
				contentStr = strings.ReplaceAll(contentStr, "{{.Project.Name}}", data.Project.Name)
				contentStr = strings.ReplaceAll(contentStr, "{{.Project.Database}}", data.Project.Database)
				contentStr = strings.ReplaceAll(contentStr, "{{.Module.Name}}", data.Module.Name)
				content = []byte(contentStr)
				content = fixImportPaths(content, data.Project.GoModule)
			}

			// Write file
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func CopyCRUDModuleFromEmbed(moduleName string, data CRUDTemplateData) error {
	moduleDir := filepath.Join("internal", moduleName)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return err
	}

	templatePath := "templates/crud"

	err := fs.WalkDir(templatesFS, templatePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath := strings.TrimPrefix(path, templatePath+"/")

		if strings.Contains(relPath, "entity.go") {
			relPath = strings.Replace(relPath, "entity.go", moduleName+".go", 1)
		}
		if strings.Contains(relPath, "interfaces.go") {
			relPath = strings.Replace(relPath, "interfaces.go", moduleName+".go", 1)
		}
		if strings.Contains(relPath, "repository.go") {
			relPath = strings.Replace(relPath, "repository.go", moduleName+"_repository.go", 1)
		}
		if strings.Contains(relPath, "service.go") {
			relPath = strings.Replace(relPath, "service.go", moduleName+"_service.go", 1)
		}
		if strings.Contains(relPath, "controller.go") {
			relPath = strings.Replace(relPath, "controller.go", moduleName+"_controller.go", 1)
		}

		destPath := filepath.Join(moduleDir, relPath)

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".go") {
			contentStr := string(content)
			contentStr = strings.ReplaceAll(contentStr, "{{.ModuleName}}", data.ModuleName)
			contentStr = strings.ReplaceAll(contentStr, "{{.ModuleNameCap}}", data.ModuleNameCap)
			contentStr = strings.ReplaceAll(contentStr, "{{.Project.GoModule}}", data.Project.GoModule)
			contentStr = strings.ReplaceAll(contentStr, "{{.Project.Name}}", data.Project.Name)
			contentStr = strings.ReplaceAll(contentStr, "{{.Project.Database}}", data.Project.Database)
			content = []byte(contentStr)
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return err
		}

		return nil
	})

	return err
}

// fixImportPaths replaces old saas-go-kit imports with relative project imports
func fixImportPaths(content []byte, goModule string) []byte {
	contentStr := string(content)
	coreImport := fmt.Sprintf(`"%s/internal/core"`, goModule)

	importMap := map[string]string{
		`"github.com/karurosux/saas-go-kit/core"`:         coreImport,
		`"github.com/karurosux/saas-go-kit/errors-go"`:    coreImport,
		`"github.com/karurosux/saas-go-kit/response-go"`:  coreImport,
		`"github.com/karurosux/saas-go-kit/validator-go"`: coreImport,
		`"github.com/karurosux/saas-go-kit/container-go"`: coreImport,
		`"github.com/karurosux/saas-go-kit/auth-go"`:      fmt.Sprintf(`"%s/internal/auth"`, goModule),
		`"github.com/karurosux/saas-go-kit/role-go"`:      fmt.Sprintf(`"%s/internal/role"`, goModule),
	}

	for oldImport, newImport := range importMap {
		contentStr = strings.ReplaceAll(contentStr, oldImport, newImport)
	}

	lines := strings.Split(contentStr, "\n")
	var result []string
	coreImportSeen := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == coreImport {
			if !coreImportSeen {
				result = append(result, line)
				coreImportSeen = true
			}
			// Skip duplicate core imports
		} else {
			result = append(result, line)
		}
	}

	return []byte(strings.Join(result, "\n"))
}

func CopyCoreFromEmbed() error {
	destDir := filepath.Join("internal", "core")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	corePath := "templates/core"
	entries, err := templatesFS.ReadDir(corePath)
	if err != nil {
		return fmt.Errorf("failed to read core templates: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		content, err := templatesFS.ReadFile(filepath.Join(corePath, entry.Name()))
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, entry.Name())
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return err
		}
	}

	return nil
}

func ReadEmbeddedFile(path string) (string, error) {
	content, err := templatesFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded file %s: %w", path, err)
	}
	return string(content), nil
}
