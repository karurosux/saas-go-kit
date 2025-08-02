package emailservice

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

// TemplateManager manages email templates from files and database
type TemplateManager struct {
	templateFS   fs.FS
	templatePath string
	dbTemplates  emailinterface.TemplateRepository
}


// NewTemplateManager creates a new template manager
func NewTemplateManager(templateFS fs.FS, templatePath string, dbTemplates emailinterface.TemplateRepository) *TemplateManager {
	return &TemplateManager{
		templateFS:   templateFS,
		templatePath: templatePath,
		dbTemplates:  dbTemplates,
	}
}

// GetTemplate retrieves a template by name (first from DB, then from files)
func (tm *TemplateManager) GetTemplate(ctx context.Context, name string) (*emailinterface.EmailTemplate, error) {
	// First check database
	if tm.dbTemplates != nil {
		tmpl, err := tm.dbTemplates.GetTemplate(ctx, name)
		if err == nil && tmpl != nil {
			return tmpl, nil
		}
	}

	// Then check file system
	return tm.getFileTemplate(name)
}

// CreateTemplate creates a new template in the database
func (tm *TemplateManager) CreateTemplate(ctx context.Context, template *emailinterface.EmailTemplate) error {
	if tm.dbTemplates == nil {
		return fmt.Errorf("database templates not available")
	}
	return tm.dbTemplates.CreateTemplate(ctx, template)
}

// UpdateTemplate updates an existing template
func (tm *TemplateManager) UpdateTemplate(ctx context.Context, name string, template *emailinterface.EmailTemplate) error {
	if tm.dbTemplates == nil {
		return fmt.Errorf("database templates not available")
	}
	return tm.dbTemplates.UpdateTemplate(ctx, name, template)
}

// DeleteTemplate deletes a template
func (tm *TemplateManager) DeleteTemplate(ctx context.Context, name string) error {
	if tm.dbTemplates == nil {
		return fmt.Errorf("database templates not available")
	}
	return tm.dbTemplates.DeleteTemplate(ctx, name)
}

// ListTemplates lists all available templates
func (tm *TemplateManager) ListTemplates(ctx context.Context) ([]*emailinterface.EmailTemplate, error) {
	templates := make([]*emailinterface.EmailTemplate, 0)

	// Get database templates
	if tm.dbTemplates != nil {
		dbTemplates, err := tm.dbTemplates.ListTemplates(ctx)
		if err == nil {
			templates = append(templates, dbTemplates...)
		}
	}

	// Get file templates
	if tm.templateFS != nil {
		fileTemplates, err := tm.listFileTemplates()
		if err == nil {
			templates = append(templates, fileTemplates...)
		}
	}

	return templates, nil
}

// RenderTemplate renders a template with the provided data
func (tm *TemplateManager) RenderTemplate(ctx context.Context, name string, data map[string]interface{}) (subject, body, html string, err error) {
	tmpl, err := tm.GetTemplate(ctx, name)
	if err != nil {
		return "", "", "", fmt.Errorf("template not found: %w", err)
	}

	// Render subject
	subjectBuf := new(bytes.Buffer)
	subjectTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse subject template: %w", err)
	}
	if err := subjectTmpl.Execute(subjectBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to render subject: %w", err)
	}
	subject = subjectBuf.String()

	// Render body
	bodyBuf := new(bytes.Buffer)
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse body template: %w", err)
	}
	if err := bodyTmpl.Execute(bodyBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to render body: %w", err)
	}
	body = bodyBuf.String()

	// Render HTML if available
	if tmpl.HTML != "" {
		htmlBuf := new(bytes.Buffer)
		htmlTmpl, err := template.New("html").Parse(tmpl.HTML)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse HTML template: %w", err)
		}
		if err := htmlTmpl.Execute(htmlBuf, data); err != nil {
			return "", "", "", fmt.Errorf("failed to render HTML: %w", err)
		}
		html = htmlBuf.String()
	}

	return subject, body, html, nil
}

// getFileTemplate reads a template from the file system
func (tm *TemplateManager) getFileTemplate(name string) (*emailinterface.EmailTemplate, error) {
	basePath := filepath.Join(tm.templatePath, name)
	
	// Read subject file
	subject, err := fs.ReadFile(tm.templateFS, basePath+".subject.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read subject file: %w", err)
	}

	// Read body file
	body, err := fs.ReadFile(tm.templateFS, basePath+".body.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read body file: %w", err)
	}

	// Try to read HTML file (optional)
	html, _ := fs.ReadFile(tm.templateFS, basePath+".html")

	// Extract variables from templates
	variables := extractVariables(string(subject), string(body), string(html))

	return &emailinterface.EmailTemplate{
		Name:      name,
		Subject:   string(subject),
		Body:      string(body),
		HTML:      string(html),
		Variables: variables,
		Active:    true,
	}, nil
}

// listFileTemplates lists all templates in the file system
func (tm *TemplateManager) listFileTemplates() ([]*emailinterface.EmailTemplate, error) {
	templates := make([]*emailinterface.EmailTemplate, 0)
	seen := make(map[string]bool)

	err := fs.WalkDir(tm.templateFS, tm.templatePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Extract template name from filename
		if strings.HasSuffix(path, ".subject.txt") {
			name := strings.TrimSuffix(filepath.Base(path), ".subject.txt")
			if !seen[name] {
				seen[name] = true
				tmpl, err := tm.getFileTemplate(name)
				if err == nil {
					templates = append(templates, tmpl)
				}
			}
		}

		return nil
	})

	return templates, err
}

// extractVariables extracts template variables from content
func extractVariables(contents ...string) []string {
	variables := make(map[string]bool)
	
	for _, content := range contents {
		// Simple regex to find {{.Variable}} patterns
		matches := templateVarRegex.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				variables[match[1]] = true
			}
		}
	}

	result := make([]string, 0, len(variables))
	for v := range variables {
		result = append(result, v)
	}
	return result
}

var templateVarRegex = regexp.MustCompile(`\{\{\.(\w+)\}\}`)