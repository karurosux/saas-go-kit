# Example: SaaS App with Client Generation

This example demonstrates how to use SaaS Go Kit's client generation tools to create type-safe TypeScript clients for your API.

## Project Structure

```
.
├── main.go                    # Go backend server
├── saas-kit-modules.json     # Module configuration
├── Makefile                  # Build targets
├── package.json              # Frontend dependencies
├── generated/                # Generated files (auto-created)
│   ├── routes.json          # Route metadata
│   └── clients/             # TypeScript clients
└── frontend/                # Example frontend usage
    └── example.ts
```

## Setup

1. **Install Go dependencies:**
   ```bash
   go mod tidy
   ```

2. **Install frontend dependencies:**
   ```bash
   npm install
   ```

3. **Generate clients:**
   ```bash
   make generate-clients
   ```

## Usage

### Backend (main.go)

```go
package main

import (
    "github.com/karurosux/saas-go-kit"
    "github.com/karurosux/saas-go-kit/auth-go"
    "github.com/karurosux/saas-go-kit/role-go"
    "github.com/karurosux/saas-go-kit/health-go"
)

func main() {
    kit := saasgokit.New()
    
    // Add modules
    kit.AddModule(auth.NewModule())
    kit.AddModule(role.NewModule())
    kit.AddModule(health.NewModule())
    
    // Start server
    kit.Start(":8080")
}
```

### Frontend (frontend/example.ts)

```typescript
import { createClients } from '../generated/clients';

// Create API clients
const api = createClients({
  baseURL: 'http://localhost:8080',
  getToken: () => localStorage.getItem('token')
});

// Type-safe API calls
async function example() {
  try {
    // Login (public endpoint - no token required)
    const loginResponse = await api.auth.login({
      email: 'user@example.com',
      password: 'password'
    });
    
    // Store token
    localStorage.setItem('token', loginResponse.token);
    
    // Get profile (authenticated endpoint - token automatically included)
    const profile = await api.auth.me();
    console.log('User profile:', profile);
    
    // Create role (authenticated endpoint)
    const role = await api.role.createRole({
      name: 'admin',
      permissions: ['read', 'write']
    });
    
    console.log('Role created:', role);
  } catch (error) {
    console.error('API Error:', error);
  }
}
```

## Key Features Demonstrated

1. **Type Safety**: All API calls are fully typed
2. **Authentication**: Automatic token handling for protected routes
3. **Error Handling**: Proper error types and handling
4. **Module Discovery**: Automatically discovers all saas-go-kit modules

## Development Workflow

1. Modify Go structs or add new endpoints
2. Run `make generate-clients`
3. Use updated types in frontend
4. Compile-time errors catch API mismatches

## Generated Files

After running `make generate-clients`, you'll get:

- `generated/routes.json` - Route metadata extracted from Go code
- `generated/clients/index.ts` - Main client exports
- `generated/clients/auth.client.ts` - Auth module client
- `generated/clients/auth.types.ts` - Auth TypeScript types
- `generated/clients/role.client.ts` - Role module client
- `generated/clients/role.types.ts` - Role TypeScript types

## Integration Examples

### React Hook

```typescript
// hooks/useApi.ts
import { createClients } from '../generated/clients';
import { useAuthToken } from './useAuth';

export function useApi() {
  const { getToken } = useAuthToken();
  
  return useMemo(() => createClients({
    baseURL: process.env.REACT_APP_API_URL,
    getToken
  }), [getToken]);
}
```

### Vue Composable

```typescript
// composables/useApi.ts
import { createClients } from '../generated/clients';

export function useApi() {
  return createClients({
    baseURL: import.meta.env.VITE_API_URL,
    getToken: () => localStorage.getItem('token')
  });
}
```

This example shows how SaaS Go Kit's client generation creates a seamless bridge between your Go backend and TypeScript frontend, ensuring type safety and reducing API integration errors.