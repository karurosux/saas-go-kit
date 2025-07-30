# SaaS Go Kit CLI - Copy-Paste Module System Design

## Overview

Transform SaaS Go Kit from a dependency-based framework to a shadcn/ui-style copy-paste system where developers own and customize their business modules.

## CLI Architecture

### Command Structure
```bash
saas-kit init                           # Initialize project with saas-kit support
saas-kit add auth                       # Add authentication module
saas-kit add subscription               # Add subscription/billing module
saas-kit add team                       # Add team management module
saas-kit add notification               # Add notification system
saas-kit add health                     # Add health monitoring
saas-kit add role                       # Add role management
saas-kit add job                        # Add background jobs
saas-kit add sse                        # Add server-sent events
saas-kit add container                  # Add dependency container

saas-kit list                           # List available modules
saas-kit list --installed               # List installed modules
saas-kit remove auth                    # Remove module from project
saas-kit update auth                    # Update module to latest template
saas-kit generate routes                # Generate TypeScript clients
```

### Project Structure After Module Installation

```
my-saas-app/
├── go.mod
├── main.go
├── internal/
│   ├── auth/                          # Copied from saas-kit/templates/auth
│   │   ├── handlers.go
│   │   ├── service.go
│   │   ├── models.go
│   │   ├── interfaces.go
│   │   ├── module.go
│   │   └── repositories/
│   │       └── gorm/
│   │           ├── account_repository.go
│   │           ├── token_repository.go
│   │           └── migrations.go
│   ├── subscription/                  # Copied from saas-kit/templates/subscription
│   │   ├── handlers.go
│   │   ├── service.go
│   │   ├── models.go
│   │   ├── interfaces.go
│   │   ├── module.go
│   │   ├── stripe_provider.go
│   │   └── repositories/
│   └── core/                          # Core utilities (always copied)
│       ├── kit.go
│       ├── module.go
│       ├── response.go
│       ├── errors.go
│       └── validator.go
├── migrations/                        # Database migrations from modules
│   ├── 001_auth_tables.sql
│   └── 002_subscription_tables.sql
├── config/
│   ├── auth.go                        # Module-specific config
│   ├── subscription.go
│   └── database.go
├── saas-kit.json                      # Project configuration
└── docs/
    ├── auth.md                        # Module documentation
    └── subscription.md
```

## CLI Implementation Details

### 1. Project Initialization
```bash
saas-kit init
```
- Creates `saas-kit.json` configuration file
- Copies core utilities (response, errors, validator)
- Sets up basic project structure
- Adds core dependencies to go.mod

### 2. Module Addition
```bash
saas-kit add auth --database=postgres --email-provider=smtp
```
- Copies module template files to `internal/{module}/`
- Copies database migrations to `migrations/`
- Updates `saas-kit.json` with module info
- Adds module dependencies to go.mod
- Creates module-specific config in `config/`
- Generates documentation in `docs/`

### 3. Configuration System
Each module can be customized during installation:

```bash
# Auth module options
saas-kit add auth \
  --database=postgres \
  --email-provider=smtp \
  --jwt-expiration=24h \
  --require-verification=true \
  --route-prefix=/api/auth

# Subscription module options  
saas-kit add subscription \
  --payment-provider=stripe \
  --database=postgres \
  --route-prefix=/api/subscription \
  --webhook-endpoint=/webhooks/stripe
```

### 4. Template System
Each module has a template directory structure:

```
templates/
├── auth/
│   ├── template.json              # Module metadata and options
│   ├── files/                     # Files to copy
│   │   ├── handlers.go.tmpl
│   │   ├── service.go.tmpl
│   │   ├── models.go.tmpl
│   │   └── ...
│   ├── migrations/                # Database migrations
│   │   └── auth_tables.sql.tmpl
│   ├── config/                    # Configuration templates
│   │   └── auth.go.tmpl
│   └── docs/                      # Documentation
│       └── README.md.tmpl
```

### 5. saas-kit.json Configuration
```json
{
  "version": "1.0.0",
  "project": {
    "name": "my-saas-app",
    "go_module": "github.com/user/my-saas-app"
  },
  "modules": {
    "auth": {
      "version": "1.2.0",
      "installed_at": "2024-01-15T10:30:00Z",
      "config": {
        "database": "postgres",
        "email_provider": "smtp",
        "route_prefix": "/api/auth",
        "require_verification": true
      }
    },
    "subscription": {
      "version": "1.1.0", 
      "installed_at": "2024-01-15T11:00:00Z",
      "config": {
        "payment_provider": "stripe",
        "database": "postgres",
        "route_prefix": "/api/subscription"
      }
    }
  },
  "dependencies": [
    "github.com/labstack/echo/v4",
    "gorm.io/gorm",
    "github.com/golang-jwt/jwt/v5"
  ]
}
```

## Benefits of This Approach

1. **Full Ownership** - Developers own all the code, can modify anything
2. **No Dependency Hell** - No version conflicts or breaking changes
3. **Easy Customization** - Modify business logic, database schemas, routes
4. **Selective Installation** - Only install modules you need
5. **Zero Lock-in** - Remove saas-kit entirely, code still works
6. **Better Debugging** - All code is in your project
7. **Custom Business Logic** - Easy to add domain-specific functionality

## Migration Path

### Phase 1: CLI Tool Development
- Create `cmd/saas-kit` CLI application
- Implement `init`, `add`, `list`, `remove` commands  
- Create template system with Go templates
- Build configuration system

### Phase 2: Template Creation
- Convert each existing module to template format
- Add configuration options for customization
- Create migration templates
- Write module documentation

### Phase 3: Integration Features
- TypeScript client generation (reuse existing)
- Module dependency resolution
- Update/sync capabilities
- Integration with existing projects

### Phase 4: Advanced Features
- Custom module creation tools
- Module marketplace/registry
- IDE integrations
- Testing utilities

## Template System Details

### Template Variables
```go
type TemplateData struct {
    Project    ProjectConfig
    Module     ModuleConfig  
    Options    map[string]interface{}
    Timestamp  time.Time
}

type ProjectConfig struct {
    Name       string
    GoModule   string
    Database   string
}

type ModuleConfig struct {
    Name         string
    RoutePrefix  string
    Database     string
    Options      map[string]interface{}
}
```

### Template Functions
```go
// Custom template functions
funcMap := template.FuncMap{
    "title":     strings.Title,
    "camel":     toCamelCase,
    "snake":     toSnakeCase,
    "pascal":    toPascalCase,
    "plural":    toPlural,
    "singular":  toSingular,
}
```

This design provides the flexibility of shadcn/ui while maintaining the power and structure of SaaS Go Kit's modular architecture.