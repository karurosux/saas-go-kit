# SaaS Go Kit CLI

A personal CLI tool I built to generate Go applications with common SaaS components. Includes modules for auth, subscriptions, teams, and other features I frequently need.

## Quick Start

### Installation

```bash
go install github.com/karurosux/saas-go-kit/cmd/sgk@latest
```

Or build from source:

```bash
git clone https://github.com/karurosux/saas-go-kit.git
cd saas-go-kit/cmd/sgk
make install
```

### Create Your First Project

```bash
# Create a new project with auth module
sgk new myapp --modules auth

# Navigate to your project
cd myapp

# Install dependencies and run
go mod tidy
go run main.go
```

## Commands

### Project Management

```bash
# Create a new project
sgk new [project-name] [flags]

# Available flags:
--modules stringSlice    Modules to include (auth,subscription,team,etc)
--go-module string      Go module path (defaults to project name)
--database string       Database type (postgres, mysql, sqlite) (default "postgres")
```

### Module Management

```bash
# List available modules
sgk list

# List installed modules in current project
sgk list --installed

# Add a module to existing project
sgk add [module-name]

# Add CRUD operations for a model
sgk crud [model-name]
```

### Other Commands

```bash
# Initialize sgk in existing project
sgk init

# Update existing modules
sgk update [module-name]

# Show version information
sgk version
```

## Available Modules

### Core Modules

- **core** - Basic utilities, validation, HTTP helpers, database config
- **auth** - Authentication with JWT, password reset, user management
- **subscription** - Basic subscription management
- **team** - Team functionality with roles
- **email** - SMTP email service with templates

### Other Modules

- **role** - Role-based access control
- **invitation** - User invitations
- **notification** - In-app notifications
- **analytics** - Basic event tracking
- **webhook** - Webhook handling
- **file** - File uploads

## Project Structure

Generated projects follow this structure:

```
myapp/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── docker-compose.yml      # Development database
├── internal/
│   ├── core/              # Core utilities and configuration
│   ├── auth/              # Authentication module (if selected)
│   └── [other-modules]/   # Additional selected modules
└── cmd/
    └── migrate/           # Database migration tool
```

## Module Dependencies

Modules automatically handle their dependencies:

- **auth** → requires **email**
- **subscription** → requires **auth**
- **team** → requires **auth**
- **invitation** → requires **auth**, **email**

Dependencies are installed automatically when adding modules.

## Database Support

Supported databases:

- **PostgreSQL** (default)
- **MySQL**
- **SQLite**

Database configuration is handled through environment variables:

```env
DATABASE_URL=postgres://user:pass@localhost/dbname?sslmode=disable
```

## Development

### Building from Source

```bash
cd cmd/sgk
make build
```

### Running Tests

```bash
make test
```

### Version Information

```bash
make version
```

## Examples

### Basic SaaS with Authentication

```bash
sgk new saas-app --modules auth
cd saas-app
go mod tidy
go run main.go
```

### Full-Featured SaaS Platform

```bash
sgk new platform \
  --modules auth,subscription,team,email,notification,analytics \
  --database postgres
```

### Add CRUD Operations

```bash
# In your project directory
sgk crud product
sgk crud order
sgk crud customer
```


## Features

- Modular - add only what you need
- Basic logging, validation, and error handling
- PostgreSQL, MySQL, SQLite support
- Automatic module dependency installation
- Docker compose for local development
- Database migrations
- Standard Go patterns

## Contributing

This is a personal tool I built for my own projects. If you find it useful and want to contribute or suggest improvements, feel free to open an issue or pull request.

## License

MIT License - see LICENSE file.

