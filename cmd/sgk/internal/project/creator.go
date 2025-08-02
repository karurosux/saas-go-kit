package project

import (
	"fmt"
	"os"
	"regexp"
	"text/template"
	
	"github.com/karurosux/saas-go-kit/cmd/sgk/internal/embed"
)

type TemplateData struct {
	Project             ProjectInfo
	Modules            []string
	DefaultDatabaseURL string
	DatabaseURL        string
	DatabasePort       string
	DatabaseService    string
	VolumeNames        string
	ProjectName        string
}

type ProjectInfo struct {
	Name     string
	GoModule string
	Database string
}

func CreateNewProject(projectName string, modules []string, goModule, database string) error {
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("directory '%s' already exists", projectName)
	}

	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.Chdir(projectName); err != nil {
		return err
	}
	defer os.Chdir(originalDir)

	if goModule == "" {
		goModule = projectName
	}

	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/labstack/echo/v4 v4.11.3
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
)
`, goModule)

	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	if err := InitProjectWithConfig(projectName, goModule, database); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}


	templateData := prepareTemplateData(projectName, goModule, database, modules)

	if err := generateFromEmbeddedTemplate("main.go", "templates/project/main.tmpl", templateData); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	if err := generateFromEmbeddedTemplate(".env.example", "templates/project/env.example.tmpl", templateData); err != nil {
		return fmt.Errorf("failed to generate .env.example: %w", err)
	}

	if err := generateFromEmbeddedTemplate("docker-compose.yml", "templates/project/dockercompose.tmpl", templateData); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
	}

	if err := generateFromEmbeddedTemplate("Makefile", "templates/project/makefile.tmpl", templateData); err != nil {
		return fmt.Errorf("failed to generate Makefile: %w", err)
	}

	return nil
}

func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("project name must contain only letters, numbers, hyphens, and underscores")
	}

	return nil
}

func prepareTemplateData(projectName, goModule, database string, modules []string) TemplateData {
	var dbURL, dbPort, dbServiceTemplate, volumeName string
	
	switch database {
	case "postgres":
		dbURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
		dbPort = "5432"
		dbServiceTemplate = "templates/project/postgresservice.tmpl"
		volumeName = "postgres_data"
	case "mysql":
		dbURL = "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		dbPort = "3306"
		dbServiceTemplate = "templates/project/mysqlservice.tmpl"
		volumeName = "mysql_data"
	default:
		dbURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
		dbPort = "5432"
		dbServiceTemplate = "templates/project/postgresservice.tmpl"
		volumeName = "postgres_data"
	}

	dbService, err := embed.ReadEmbeddedFile(dbServiceTemplate)
	if err != nil {
		dbService = "# Error loading database service template"
	}

	return TemplateData{
		Project: ProjectInfo{
			Name:     projectName,
			GoModule: goModule,
			Database: database,
		},
		Modules:            modules,
		DefaultDatabaseURL: dbURL,
		DatabaseURL:        dbURL,
		DatabasePort:       dbPort,
		DatabaseService:    dbService,
		VolumeNames:        volumeName,
		ProjectName:        projectName,
	}
}

func generateFromEmbeddedTemplate(filename, templatePath string, data TemplateData) error {
	templateContent, err := embed.ReadEmbeddedFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(filename).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse template for %s: %w", filename, err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template for %s: %w", filename, err)
	}

	return nil
}

