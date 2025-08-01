# SaaS Go Kit Templates

This directory contains module templates that are copied to your project when you run `saas-kit add <module>`.

## Structure

Each module directory contains:
- **Go files**: Copied as-is to your project
- **`.tmpl` files**: Processed with Go templates, variables replaced based on your project config
- **Nested directories**: Maintain the same structure in your project

## Available Modules

- **auth** - Authentication with JWT, email verification, password reset
- **subscription** - Stripe integration for billing and subscriptions  
- **team** - Team management with invitations and roles
- **notification** - Multi-channel notifications (email, SMS, push)
- **health** - Health checks for database, external services
- **role** - Role-based access control and permissions
- **job** - Background job processing with queues
- **sse** - Server-sent events for real-time communication
- **container** - Dependency injection container

## How It Works

1. Run `saas-kit add auth` 
2. CLI copies files from `templates/auth/` to `internal/auth/`
3. Template files (`.tmpl`) get processed with your project info
4. Regular files are copied as-is
5. Your module is ready to use with `RegisterModule(container)`

## Template Variables

Available in `.tmpl` files:
- `{{.Project.Name}}` - Your project name
- `{{.Project.GoModule}}` - Your Go module path  
- `{{.Module.Name}}` - Module being installed
- `{{.Module.RoutePrefix}}` - API route prefix
- `{{.Options.jwt_secret}}` - CLI options passed

## Module Pattern

Each module follows the same pattern:

```go
package mymodule

func RegisterModule(c container.Container) error {
    // Get dependencies from container
    echo := container.MustGetTyped[*echo.Echo](c, "echo")
    db := container.MustGetTyped[*gorm.DB](c, "db")
    
    // Create service
    service := NewService(ServiceConfig{...})
    
    // Register service in container
    c.Register("mymodule.service", service)
    
    // Register routes
    registerRoutes(echo, service)
    
    return nil
}
```

Simple and clean!