# Client Generation Tools

SaaS Go Kit includes powerful tools to automatically generate type-safe TypeScript clients for your API endpoints. This enables seamless full-stack development with automatic synchronization between your Go backend and TypeScript frontend.

## Overview

The client generation system consists of two parts:
1. **Route Extractor** - A script in your project that extracts route metadata from your actual module instances
2. **Client Generator** (`github.com/karurosux/saas-go-kit/cmd/generate-clients`) - Shared tool that generates TypeScript clients from route metadata

## Quick Start

### 1. Create a Route Extractor in Your Project

Create `cmd/route-extractor/main.go` to extract routes from your modules:

```go
package main

import (
    "encoding/json"
    "log"
    "os"
    
    "github.com/karurosux/saas-go-kit/auth-go"
    "github.com/karurosux/saas-go-kit/role-go"
    "myapp/internal/products" // Your custom modules
)

func main() {
    // Initialize your modules (simplified example)
    db := setupDatabase()
    
    modules := map[string]core.Module{
        "auth": auth.NewModule(authConfig),
        "role": role.NewModule(roleConfig),
        "products": products.NewModule(productService),
    }
    
    // Extract routes
    var metadata []ModuleMetadata
    for name, module := range modules {
        metadata = append(metadata, extractModuleMetadata(name, module))
    }
    
    // Save to file
    output, _ := json.MarshalIndent(metadata, "", "  ")
    os.WriteFile("./generated/routes.json", output, 0644)
}
```

### 2. Set Up Your Makefile

```makefile
generate-clients:
	@echo "Extracting routes..."
	@go run ./cmd/route-extractor
	@echo "Generating TypeScript clients..."
	@go run github.com/karurosux/saas-go-kit/cmd/generate-clients@latest \
		-routes=./generated/routes.json \
		-o=./generated/clients
```

### 3. Generate Clients

```bash
make generate-clients
```

### 4. Use in Your Frontend

```typescript
import { createClients } from './generated/clients';

const clients = createClients({
  baseURL: 'http://localhost:8080',
  getToken: () => localStorage.getItem('token')
});

// Use the clients
await clients.auth.login({ email: 'user@example.com', password: 'password' });
await clients.role.createRole({ name: 'admin', permissions: ['read', 'write'] });
```

## Features

### ✅ Type Safety
- Automatically generated TypeScript interfaces from Go structs
- Request/response types match your Go API exactly
- Compile-time error checking for API calls

### ✅ Authentication Handling
- Automatic Bearer token injection for authenticated endpoints
- Smart detection of public vs. authenticated routes
- Configurable token retrieval function

### ✅ Multiple Usage Patterns

**Factory Pattern (Recommended)**
```typescript
const clients = createClients({
  baseURL: 'http://localhost:8080',
  getToken: () => localStorage.getItem('token')
});
```

**Individual Clients**
```typescript
import { AuthClient, RoleClient } from './generated/clients';

const authClient = new AuthClient('http://localhost:8080', '/api/v1', {
  getToken: () => localStorage.getItem('token')
});
```

### ✅ Smart Route Detection
- Public routes (login, register, health) don't require authentication
- Authenticated routes automatically include Bearer tokens
- Support for path parameters, query parameters, and request bodies

## Generated Structure

```
generated/
├── routes.json           # Route metadata
└── clients/
    ├── index.ts         # Main exports and factory function
    ├── auth.client.ts   # Auth module client
    ├── auth.types.ts    # Auth TypeScript types
    ├── role.client.ts   # Role module client
    └── role.types.ts    # Role TypeScript types
```

## Route Extraction Approaches

### Option 1: Project-specific Route Extractor (Recommended)

Create a route extractor in your project that instantiates your modules:

```go
// cmd/extract-routes/main.go
package main

import (
    "github.com/karurosux/saas-go-kit/auth-go"
    "github.com/karurosux/saas-go-kit/role-go"
    "myapp/internal/dishes"
    // ... other imports
)

func main() {
    // Create your module instances
    modules := map[string]core.Module{
        "auth": auth.NewModule(authConfig),
        "role": role.NewModule(roleConfig),
        "dishes": dishes.NewModule(dishService),
    }
    
    // Extract and save routes
    ExtractRoutes(modules, "./generated/routes.json")
}
```

### Option 2: Module Configuration File

For simpler cases, you can use a configuration file:

### Module Configuration

Each module in `saas-kit-modules.json` supports:

```json
{
  "name": "auth",                                           // Module name (used for client naming)
  "import_path": "github.com/karurosux/saas-go-kit/auth-go", // Go import path
  "module_path": "./auth"                                   // Local path for type extraction
}
```

### Including Custom Modules

The client generation tools work with **any module** that implements the saas-go-kit Module interface, not just the built-in ones:

```json
[
  {
    "name": "auth",
    "import_path": "github.com/karurosux/saas-go-kit/auth-go",
    "module_path": "./vendor/github.com/karurosux/saas-go-kit/auth-go"
  },
  {
    "name": "dishes",
    "import_path": "myapp/internal/dishes",
    "module_path": "./internal/dishes"
  },
  {
    "name": "orders", 
    "import_path": "myapp/internal/orders",
    "module_path": "./internal/orders"
  }
]
```

As long as your custom module follows the saas-go-kit patterns (implements `core.Module` interface, exports routes, uses standard Go structs), the client generator will:
- Extract all your routes
- Parse your Go structs into TypeScript types
- Generate fully typed clients
- Handle authentication automatically

### Tool Options

**Route Generator**
```bash
go run ./cmd/generate-routes \
  -config ./saas-kit-modules.json \
  -output ./generated/routes.json
```

**Client Generator**
```bash
go run ./cmd/generate-clients \
  -routes ./generated/routes.json \
  -o ./generated/clients \
  -format axios  # axios (default) or fetch
```

## Frontend Integration

### React Example

```typescript
// hooks/useApi.ts
import { createClients } from '../generated/clients';
import { useAuth } from './useAuth';

export function useApi() {
  const { getToken } = useAuth();
  
  return createClients({
    baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080',
    getToken
  });
}

// components/LoginForm.tsx
import { useApi } from '../hooks/useApi';

export function LoginForm() {
  const api = useApi();
  
  const handleLogin = async (email: string, password: string) => {
    try {
      const response = await api.auth.login({ email, password });
      console.log('Login successful:', response);
    } catch (error) {
      console.error('Login failed:', error);
    }
  };
}
```

### Vue Example

```typescript
// composables/useApi.ts
import { createClients } from '../generated/clients';

export function useApi() {
  return createClients({
    baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
    getToken: () => localStorage.getItem('token')
  });
}
```

## Advanced Usage

### Custom Authentication

```typescript
const clients = createClients({
  baseURL: 'http://localhost:8080',
  getToken: () => {
    // Custom token retrieval logic
    const token = sessionStorage.getItem('auth_token');
    return token ? `Bearer ${token}` : null;
  },
  config: {
    timeout: 5000,
    headers: {
      'X-API-Version': '1.0'
    }
  }
});
```

### Error Handling

```typescript
try {
  await clients.auth.login({ email, password });
} catch (error) {
  if (error.response?.status === 401) {
    // Handle authentication error
  } else if (error.response?.status === 422) {
    // Handle validation error
    const validationErrors = error.response.data.errors;
  }
}
```

## Best Practices

1. **Version Control**: Commit generated files to ensure team synchronization
2. **CI/CD Integration**: Regenerate clients in your build pipeline
3. **Type Safety**: Use TypeScript strict mode for maximum type safety
4. **Error Handling**: Implement consistent error handling patterns
5. **Token Management**: Use secure token storage and refresh logic

## Troubleshooting

### Common Issues

**"Cannot find module" errors**
- Ensure all dependencies are installed: `npm install axios`
- Check TypeScript configuration includes generated files

**Authentication not working**
- Verify `getToken` function returns correct format
- Check API expects `Authorization: Bearer <token>` header

**Missing types**
- Regenerate clients after Go struct changes
- Ensure module paths in config are correct

### Development Workflow

1. Modify Go structs/routes
2. Run `make generate-all`
3. Update frontend code with new types
4. Test API integration

## Contributing

To improve the client generation tools:

1. Fork the repository
2. Make changes to `cmd/generate-routes/` or `cmd/generate-clients/`
3. Test with example projects
4. Submit a pull request

## License

Same license as SaaS Go Kit.