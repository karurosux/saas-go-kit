package project

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// createNewProject creates a new project with specified modules
func CreateNewProject(projectName string, modules []string, goModule, database string) error {
	// Validate project name
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	// Check if directory already exists
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("directory '%s' already exists", projectName)
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Store original directory
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Change to project directory
	if err := os.Chdir(projectName); err != nil {
		return err
	}
	defer os.Chdir(originalDir)

	// Set go module path
	if goModule == "" {
		goModule = projectName
	}

	// Create go.mod
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

	// Initialize project with configuration
	if err := InitProjectWithConfig(projectName, goModule, database); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	// Note: Core templates and modules will be handled by main package to avoid circular imports

	// Create main.go
	mainContent := generateMainGo(goModule, modules)
	if err := os.WriteFile("main.go", []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	// Generate .env.example
	if err := generateEnvExample(database); err != nil {
		return fmt.Errorf("failed to generate .env.example: %w", err)
	}

	// Generate docker-compose.yml
	if err := generateDockerCompose(database); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
	}

	// Generate Makefile
	if err := generateMakefile(projectName, goModule); err != nil {
		return fmt.Errorf("failed to generate Makefile: %w", err)
	}

	return nil
}

// validateProjectName validates the project name
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check for valid Go module name characters
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("project name must contain only letters, numbers, hyphens, and underscores")
	}

	return nil
}

// generateMainGo generates the main.go file content
func generateMainGo(goModule string, modules []string) string {
	var imports []string
	var moduleRegistrations []string

	// Add core import
	imports = append(imports, fmt.Sprintf(`"%s/internal/core"`, goModule))

	// Add module imports and registrations
	for _, module := range modules {
		imports = append(imports, fmt.Sprintf(`"%s/internal/%s"`, goModule, module))
		moduleRegistrations = append(moduleRegistrations, fmt.Sprintf(`	if err := %s.RegisterModule(container); err != nil {
		log.Fatalf("Failed to register %s module: %%v", err)
	}`, module, module))
	}

	return fmt.Sprintf(`package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	%s
)

func main() {
	// Initialize configuration
	config := loadConfig()
	
	// Initialize database
	db := initDatabase(config.DatabaseURL)
	
	// Initialize web server
	e := initServer()
	
	// Initialize dependency container
	container := initContainer(e, db)
	
	// Register modules
%s
	
	// Start server
	startServer(e, config.Port)
}

func loadConfig() *Config {
	return &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		DatabaseURL: getEnvOrDefault("DATABASE_URL", "postgres://user:password@localhost/dbname?sslmode=disable"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDatabase(dbURL string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %%v", err)
	}
	return db
}

func initServer() *echo.Echo {
	e := echo.New()
	
	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	
	return e
}

func initContainer(e *echo.Echo, db *gorm.DB) *core.Container {
	container := core.NewContainer()
	container.Set("echo", e)
	container.Set("db", db)
	return container
}

func startServer(e *echo.Echo, port string) {
	log.Printf("Server starting on port %%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Server failed to start: %%v", err)
	}
}

type Config struct {
	Port        string
	DatabaseURL string
}
`, strings.Join(imports, "\n\t"), strings.Join(moduleRegistrations, "\n"))
}

// generateEnvExample generates the .env.example file
func generateEnvExample(database string) error {
	var dbURL string
	var dbPort string
	
	switch database {
	case "postgres":
		dbURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
		dbPort = "5432"
	case "mysql":
		dbURL = "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		dbPort = "3306"
	default:
		dbURL = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
		dbPort = "5432"
	}

	envContent := fmt.Sprintf(`# Application Configuration
PORT=8080

# Database Configuration
DATABASE_URL=%s

# Database Connection Details (for docker-compose)
DB_HOST=localhost
DB_PORT=%s
DB_NAME=dbname
DB_USER=user
DB_PASSWORD=password

# JWT Configuration (generate your own secret in production)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Redis Configuration (if using Redis)
REDIS_URL=redis://localhost:6379

# Email Configuration (if using email features)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Environment
ENV=development
`, dbURL, dbPort)

	return os.WriteFile(".env.example", []byte(envContent), 0644)
}

// generateDockerCompose generates the docker-compose.yml file
func generateDockerCompose(database string) error {
	var dbService string
	
	switch database {
	case "postgres":
		dbService = `  postgres:
    image: postgres:15-alpine
    container_name: saas-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: dbname
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d dbname"]
      interval: 30s
      timeout: 10s
      retries: 5`
	case "mysql":
		dbService = `  mysql:
    image: mysql:8.0
    container_name: saas-mysql
    restart: unless-stopped
    environment:
      MYSQL_DATABASE: dbname
      MYSQL_USER: user
      MYSQL_PASSWORD: password
      MYSQL_ROOT_PASSWORD: rootpassword
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 30s
      timeout: 10s
      retries: 5`
	default:
		dbService = `  postgres:
    image: postgres:15-alpine
    container_name: saas-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: dbname
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d dbname"]
      interval: 30s
      timeout: 10s
      retries: 5`
	}
	
	var volumeName string
	if database == "mysql" {
		volumeName = "mysql_data"
	} else {
		volumeName = "postgres_data"
	}

	dockerComposeContent := fmt.Sprintf(`version: '3.8'

services:
%s

  redis:
    image: redis:7-alpine
    container_name: saas-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 5

volumes:
  %s:
  redis_data:

networks:
  default:
    name: saas-network
`, dbService, volumeName)

	return os.WriteFile("docker-compose.yml", []byte(dockerComposeContent), 0644)
}

// generateMakefile generates the Makefile with common commands
func generateMakefile(projectName, goModule string) error {
	makefileContent := fmt.Sprintf(`# %s Makefile
# Auto-generated by saas-go-kit

.PHONY: help setup build run clean test lint fmt deps-up deps-down migrate-up migrate-down

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %%s\n", $$1, $$2}'

# Setup commands
setup: ## Set up the development environment
	@echo "🚀 Setting up development environment..."
	@if [ ! -f .env ]; then cp .env.example .env && echo "📝 Created .env from .env.example - please review and update"; fi
	@make deps-up
	@make deps
	@echo "✅ Setup complete! Run 'make run' to start the application"

deps: ## Download and tidy Go dependencies
	@echo "📦 Installing Go dependencies..."
	@go mod download
	@go mod tidy

# Docker commands
deps-up: ## Start database and Redis services
	@echo "🐳 Starting dependencies..."
	@docker compose up -d

deps-down: ## Stop database and Redis services
	@echo "🛑 Stopping dependencies..."
	@docker compose down

deps-logs: ## Show logs from dependencies
	@docker compose logs -f

# Build commands
build: ## Build the application binary
	@echo "🔨 Building application..."
	@go build -o bin/%s main.go
	@echo "✅ Built binary: bin/%s"

build-linux: ## Build the application binary for Linux
	@echo "🔨 Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o bin/%s-linux main.go
	@echo "✅ Built Linux binary: bin/%s-linux"

# Run commands
run: ## Run the application in development mode
	@echo "🏃 Starting %s..."
	@go run main.go

run-bin: build ## Build and run the binary
	@echo "🏃 Running built binary..."
	@./bin/%s

# Development commands
test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

lint: ## Run linter (requires golangci-lint)
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format Go code
	@echo "✨ Formatting code..."
	@go fmt ./...

# Database migration commands (if using migrations)
migrate-up: ## Run database migrations up
	@echo "⬆️  Running migrations up..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path db/migrations -database "$(DATABASE_URL)" up; \
	else \
		echo "⚠️  migrate tool not found. Install from: https://github.com/golang-migrate/migrate"; \
	fi

migrate-down: ## Run database migrations down
	@echo "⬇️  Running migrations down..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path db/migrations -database "$(DATABASE_URL)" down; \
	else \
		echo "⚠️  migrate tool not found. Install from: https://github.com/golang-migrate/migrate"; \
	fi

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Please provide NAME: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "📝 Creating migration: $(NAME)..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate create -ext sql -dir db/migrations $(NAME); \
	else \
		echo "⚠️  migrate tool not found. Install from: https://github.com/golang-migrate/migrate"; \
	fi

# Cleanup commands
clean: ## Clean build artifacts and dependencies
	@echo "🧹 Cleaning up..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean
	@make deps-down

clean-all: clean ## Clean everything including Docker volumes
	@echo "🧹 Deep cleaning..."
	@docker compose down -v
	@docker system prune -f

# Docker deployment commands
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t %s .

docker-run: docker-build ## Build and run Docker container
	@echo "🐳 Running Docker container..."
	@docker run --rm -p 8080:8080 --env-file .env %s

# Development workflow shortcuts
dev: deps-up run ## Quick start development (start deps + run app)

restart: ## Restart the application
	@echo "🔄 Restarting application..."
	@make deps-down
	@make deps-up
	@make run

# Git hooks (optional)
install-hooks: ## Install Git pre-commit hooks
	@echo "🪝 Installing Git hooks..."
	@cp -f scripts/pre-commit .git/hooks/pre-commit 2>/dev/null || echo "⚠️  No pre-commit script found in scripts/"
	@chmod +x .git/hooks/pre-commit 2>/dev/null || true
`, projectName, projectName, projectName, projectName, projectName, projectName, projectName, projectName)

	return os.WriteFile("Makefile", []byte(makefileContent), 0644)
}