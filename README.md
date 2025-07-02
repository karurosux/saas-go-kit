# SaaS Go Kit

A modular toolkit for building SaaS applications in Go. Designed with clean architecture principles.

## 🚀 Features

- **Modular Architecture**: Plug-and-play modules with minimal coupling
- **Interface-Driven**: Easy to test, mock, and extend  
- **Echo Framework**: Built specifically for the Echo web framework
- **Universal Client Generation**: Auto-discover modules and generate type-safe TypeScript clients with zero configuration
- **Custom Module Support**: Seamlessly integrate your business logic alongside built-in modules
- **Clean Code**: Well-structured, documented, and maintainable

## 📦 Modules

### Core Modules

- **[core-go](./core-go/)** - Module management and application bootstrapping
- **[errors-go](./errors-go/)** - Structured error handling with HTTP status codes
- **[response-go](./response-go/)** - Standardized API response formatting
- **[validator-go](./validator-go/)** - Request validation with custom rules
- **[ratelimit-go](./ratelimit-go/)** - In-memory and distributed rate limiting

### Authentication & Authorization

- **[auth-go](./auth-go/)** - Complete authentication system with JWT, email verification, password reset

### Business Modules

- **[subscription-go](./subscription-go/)** - Subscription and billing management with Stripe integration
- **[team-go](./team-go/)** - Team management with role-based access control
- **[notification-go](./notification-go/)** - Multi-channel notification system (email, SMS, push)

### Development Tools

- **[Universal Client Generation](./CLIENT_GENERATION.md)** - Zero-config TypeScript client generation with auto-discovery
- **[health-go](./health-go/)** - Application health monitoring with multiple check types
- **[role-go](./role-go/)** - Role-based access control and permissions management

## 🏃‍♂️ Quick Start

### 1. Install the Library

```bash
go get github.com/karurosux/saas-go-kit
```

### 2. Simple Setup

```go
package main

import (
    "log"
    "time"
    
    "github.com/karurosux/saas-go-kit"
    "github.com/karurosux/saas-go-kit/core-go"
    "github.com/karurosux/saas-go-kit/auth-go"
    "github.com/karurosux/saas-go-kit/ratelimit-go"
)

func main() {
    // Check for route extraction flags (for client generation)
    core.CheckExtractionFlags()
    
    // Using the main library (re-exports)
    kit := saasgokit.NewKit(nil, saasgokit.KitConfig{
        Debug: true,
    })
    
    // Or use modules directly
    authService := setupAuthService()
    rateLimiter := ratelimit.New(100, time.Minute)
    
    app, err := core.NewBuilder().
        WithDebug(true).
        WithModule(auth.NewModule(auth.ModuleConfig{
            Service:     authService,
            RateLimiter: rateLimiter.EchoMiddleware(),
        })).
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    app.Start(":8080")
}
```

### 3. Advanced Setup with Custom Configuration

```go
func main() {
    e := echo.New()
    
    kit := core.NewKit(e, core.KitConfig{
        Debug:       true,
        RoutePrefix: "/api/v1",
    })
    
    // Register modules
    kit.Register(auth.NewModule(authConfig))
    kit.Register(subscription.NewModule(subConfig))
    kit.Register(team.NewModule(teamConfig))
    kit.Register(notification.NewModule(notificationConfig))
    
    // Mount all modules
    kit.Mount()
    
    e.Start(":8080")
}
```

### 4. Adding Custom Modules

SaaS Go Kit is designed to work seamlessly with your custom business modules:

```go
// internal/products/module.go
package products

import "github.com/karurosux/saas-go-kit/core-go"

type Module struct {
    service *Service
}

func NewModule(service *Service) *Module {
    return &Module{service: service}
}

func (m *Module) Routes() []core.Route {
    return []core.Route{
        {Method: "GET", Path: "/products", Handler: m.listProducts},
        {Method: "POST", Path: "/products", Handler: m.createProduct},
        {Method: "GET", Path: "/products/:id", Handler: m.getProduct},
    }
}

// main.go
func main() {
    kit := core.NewKit(e, core.KitConfig{Debug: true})
    
    // Mix built-in and custom modules
    kit.Register(auth.NewModule(authConfig))
    kit.Register(products.NewModule(productService))  // Your custom module
    kit.Register(orders.NewModule(orderService))      // Another custom module
    
    kit.Mount()
    e.Start(":8080")
}
```

### 5. Generate TypeScript Clients (Universal Auto-Discovery)

The client generation tools automatically discover **both built-in and custom modules** without configuration:

```bash
# Extract routes from your application
go run . --extract-routes --extract-output ./generated/routes.json

# Generate TypeScript clients from discovered routes
go run github.com/karurosux/saas-go-kit/cmd/generate-clients -routes=./generated/routes.json -o=./generated/clients
```

Or use the convenient make targets:

```bash
# Extract routes and generate clients in one command
make generate-clients
```

**Note:** Your application needs to call `core.CheckExtractionFlags()` early in `main()` to support direct extraction.

The tools automatically:
- ✅ **Detect saas-go-kit modules** (auth, health, role, etc.) from your imports
- ✅ **Detect custom modules** (products, orders, etc.) from your code
- ✅ **Extract TypeScript types** from your Go structs  
- ✅ **Generate type-safe clients** with authentication support

```typescript
// Use generated clients with full type safety
import { AuthClient, HealthClient, ProductsClient } from './generated/clients';

const auth = new AuthClient('http://localhost:8080', '/api/v1', {
  getToken: () => localStorage.getItem('token')
});

const products = new ProductsClient('http://localhost:8080', '/api/v1', {
  getToken: () => localStorage.getItem('token')
});

// Built-in module - fully typed!
await auth.login({ email, password });
const profile = await auth.profile();

// Your custom module - also fully typed!
const productList = await products.list();
await products.create({ name: "Widget", price: 99.99 });
```

See [Client Generation Documentation](./CLIENT_GENERATION.md) for detailed setup.

## ⚡ Quick Client Generation

Generate TypeScript clients for **any** Go project using saas-go-kit in just two commands:

```bash
# 1. Extract routes (auto-discovers all modules)
go run . --extract-routes --extract-output ./generated/routes.json

# 2. Generate TypeScript clients
go run github.com/karurosux/saas-go-kit/cmd/generate-clients -routes=./generated/routes.json -o=./generated/clients
```

**Works with any project** - detects both saas-go-kit modules and your custom business modules automatically. No configuration required! 🚀

**Prerequisites:** Add `core.CheckExtractionFlags()` to your `main()` function for direct extraction support.

## 📖 Documentation

### Module Documentation

- [Core Module](./core-go/README.md) - Application foundation and module system
- [Error Handling](./errors-go/README.md) - Structured error management
- [Response Formatting](./response-go/README.md) - Standardized API responses
- [Validation](./validator-go/README.md) - Request validation with custom rules
- [Rate Limiting](./ratelimit-go/README.md) - Request rate limiting
- [Authentication](./auth-go/README.md) - User authentication and management
- [Subscriptions](./subscription-go/README.md) - Subscription and billing management
- [Team Management](./team-go/README.md) - Team management with RBAC
- [Notifications](./notification-go/README.md) - Multi-channel notification system

### Examples

- [Basic App](./examples/basic-app/) - Complete working example with authentication
- [Microservices Guide](./MICROSERVICES.md) - Using SaaS Go Kit with microservices architecture

## 🏗️ Architecture

SaaS Go Kit follows clean architecture principles:

```
┌─────────────────────────────────────────┐
│                HTTP Layer               │
├─────────────────────────────────────────┤
│              Module Layer               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │   Auth   │ │   Sub    │ │   Team   │ │
│  └──────────┘ └──────────┘ └──────────┘ │
│  ┌──────────┐ ┌──────────┐              │
│  │  Notify  │ │  Other   │              │
│  └──────────┘ └──────────┘              │
├─────────────────────────────────────────┤
│              Service Layer              │
├─────────────────────────────────────────┤
│             Interface Layer             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │  Storage │ │  Email   │ │  Config  │ │
│  └──────────┘ └──────────┘ └──────────┘ │
├─────────────────────────────────────────┤
│           Infrastructure Layer          │
└─────────────────────────────────────────┘
```

### Key Principles

1. **Dependency Inversion**: Modules depend on interfaces, not implementations
2. **Single Responsibility**: Each module has a single, well-defined purpose
3. **Open/Closed**: Easy to extend without modifying existing code
4. **Interface Segregation**: Small, focused interfaces
5. **Dependency Injection**: All dependencies are injected, making testing easy

## 🧪 Testing

Each module includes comprehensive tests:

```bash
# Test all modules
go test ./...

# Test specific module
go test ./auth-go/...

# Test with coverage
go test -cover ./...
```

## 🔧 Configuration

### Environment Variables

```env
# Application
ENV=development
APP_NAME=My SaaS App
APP_URL=https://myapp.com

# Security
JWT_SECRET=your-jwt-secret
JWT_EXPIRATION=24h

# Database
DATABASE_URL=postgres://user:pass@localhost/db

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user@example.com
SMTP_PASS=password
```

### Configuration Providers

Implement the config interfaces for your preferred configuration method:

```go
type MyConfig struct {
    // your config fields
}

func (c *MyConfig) GetJWTSecret() string {
    return c.JWTSecret
}

// ... implement other interface methods
```

## 📝 Examples

### Authentication Flow

```bash
# Register
curl -X POST /api/auth/register \
  -d '{"email":"user@example.com","password":"secure123"}'

# Login
curl -X POST /api/auth/login \
  -d '{"email":"user@example.com","password":"secure123"}'

# Access protected endpoint
curl -H "Authorization: Bearer YOUR_TOKEN" /api/auth/me
```

### Error Handling

```go
// Errors are automatically formatted
if err := validation.Failed(); err != nil {
    return response.Error(c, errors.BadRequest("Invalid data"))
}

// Returns:
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid data"
  }
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/karurosux/saas-go-kit.git
cd saas-go-kit

# Install dependencies for all modules
find . -name "go.mod" -execdir go mod download \;

# Run tests
go test ./...

# Run example
cd examples/basic-app
go run .
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🚀 Current Status

✅ **Available Modules:**
- Core foundation (core-go, errors-go, response-go, validator-go, ratelimit-go)
- Authentication system (auth-go)
- Health monitoring (health-go)
- Role-based access control (role-go)
- Subscription management (subscription-go)
- Team management (team-go)
- Notification system (notification-go)

✅ **Universal Development Tools:**
- **Auto-discovering route extraction** (direct application flags)
- **TypeScript client generation** (`cmd/generate-clients`)
- **Zero-config setup** - works with any saas-go-kit project
- **Custom module support** - automatically detects your business modules

🔧 **Potential Future Modules:**
- Analytics and event tracking
- File storage and management
- Search functionality
- API documentation generation

## 📞 Support

- 🐛 [Issue Tracker](https://github.com/karurosux/saas-go-kit/issues)
- 💬 [Discussions](https://github.com/karurosux/saas-go-kit/discussions)