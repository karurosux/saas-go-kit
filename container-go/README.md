# Container Module for SaaS Go Kit

A simple dependency injection container to manage service dependencies and reduce coupling in your applications.

## Why Use This?

When your `main.go` becomes complex with many interdependent services, this container helps by:
- Centralizing service registration
- Managing initialization order
- Handling graceful shutdown
- Providing type-safe service retrieval

## Installation

```go
import "github.com/karurosux/saas-go-kit/container-go"
```

## Basic Usage

### 1. Define Service Keys

```go
// internal/container/keys.go
package container

import "github.com/karurosux/saas-go-kit/container-go"

const (
    KeyDatabase          container.ServiceKey = "database"
    KeyAuthService       container.ServiceKey = "auth"
    KeyBadgeService      container.ServiceKey = "badge"
    KeyDishService       container.ServiceKey = "dish"
    KeyUserService       container.ServiceKey = "user"
    KeyRestaurantService container.ServiceKey = "restaurant"
)
```

### 2. Implement Service Interface (Optional)

For services that need initialization or cleanup:

```go
type UserService struct {
    db    *gorm.DB
    badge badges.BadgeService
}

func (s *UserService) Name() string {
    return "UserService"
}

func (s *UserService) Initialize(ctx context.Context, c container.Container) error {
    // Retrieve dependencies from container
    var err error
    s.db, err = container.GetTyped[*gorm.DB](c, KeyDatabase)
    if err != nil {
        return err
    }
    
    s.badge, err = container.GetTyped[badges.BadgeService](c, KeyBadgeService)
    if err != nil {
        return err
    }
    
    return nil
}

func (s *UserService) Shutdown(ctx context.Context) error {
    // Cleanup if needed
    return nil
}
```

### 3. Simplify Your main.go

Before:
```go
func main() {
    // Complex manual wiring
    db := setupDB()
    
    badgeRepo := badges.NewRepository(db)
    badgeService := badges.NewService(badgeRepo)
    
    dishRepo := dishes.NewRepository(db)
    dishService := dishes.NewService(dishRepo, badgeService)
    
    userRepo := users.NewRepository(db)
    userService := users.NewService(userRepo, badgeService)
    
    restaurantRepo := restaurants.NewGormRepository(db)
    restaurantService := restaurants.NewService(restaurantRepo)
    
    // Connect services with adapters
    restaurantAdapter := dishes.NewRestaurantServiceAdapter(restaurantService)
    dishService.SetRestaurantAdapter(restaurantAdapter)
    
    // ... more complex wiring
}
```

After:
```go
func main() {
    // Create container
    c := container.New()
    
    // Register services
    c.Register(KeyDatabase, setupDB())
    c.Register(KeyBadgeService, badges.NewService(badges.NewRepository(db)))
    c.Register(KeyDishService, &dishes.Service{}) // Will initialize itself
    c.Register(KeyUserService, &users.Service{})  // Will initialize itself
    c.Register(KeyRestaurantService, restaurants.NewService(restaurants.NewGormRepository(db)))
    
    // Initialize all services
    ctx := context.Background()
    if err := c.InitializeAll(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Retrieve services when needed
    dishService := container.MustGetTyped[*dishes.Service](c, KeyDishService)
    
    // Graceful shutdown
    defer c.ShutdownAll(ctx)
}
```

### 4. Service Factory Pattern

For even cleaner code, use factory functions:

```go
// internal/container/factories.go
func RegisterServices(c container.Container, db *gorm.DB) error {
    // Register database first
    c.Register(KeyDatabase, db)
    
    // Register services with factories
    c.Register(KeyBadgeService, func() badges.BadgeService {
        return badges.NewService(badges.NewRepository(db))
    })
    
    c.Register(KeyRestaurantService, func() restaurants.RestaurantService {
        return restaurants.NewService(restaurants.NewGormRepository(db))
    })
    
    // Services that need initialization
    c.Register(KeyDishService, &dishes.Service{})
    c.Register(KeyUserService, &users.Service{})
    
    return nil
}
```

## Best Practices

1. **Define constants for service keys** to avoid typos
2. **Use the Service interface** for complex services that need initialization
3. **Register in dependency order** (database first, then repositories, then services)
4. **Use GetTyped/MustGetTyped** for type safety
5. **Handle errors properly** during initialization

## Testing

The container makes testing easier:

```go
func TestUserService(t *testing.T) {
    // Create test container
    c := container.New()
    
    // Register mocks
    c.Register(KeyDatabase, mockDB)
    c.Register(KeyBadgeService, mockBadgeService)
    
    // Test your service
    userService := &UserService{}
    err := userService.Initialize(context.Background(), c)
    assert.NoError(t, err)
    
    // Run tests...
}
```

## Future Enhancements

When you're ready for microservices, we can extend this with:
- Service discovery
- Remote service providers
- Circuit breakers
- Health checks
- Metrics collection

But for now, this simple container will clean up your dependency management significantly!