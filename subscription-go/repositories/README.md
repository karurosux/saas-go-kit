# Repository Implementations

This directory contains ready-to-use repository implementations for different ORMs and databases, while maintaining the clean architecture interfaces defined in the main module.

## 🚀 Quick Start

### Using GORM Implementation

```go
package main

import (
    "log"
    
    "github.com/karurosux/saas-go-kit/subscription-go"
    gormrepo "github.com/karurosux/saas-go-kit/subscription-go/repositories/gorm"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // Database connection
    db, err := gorm.Open(postgres.Open("your-dsn"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    // Auto-migrate models with indexes
    if err := gormrepo.AutoMigrate(db); err != nil {
        log.Fatal("Migration failed:", err)
    }
    
    // Create repository instances
    subscriptionRepo := gormrepo.NewSubscriptionRepository(db)
    planRepo := gormrepo.NewSubscriptionPlanRepository(db)
    usageRepo := gormrepo.NewUsageRepository(db)
    
    // Create service with repositories
    subscriptionService := subscription.NewSubscriptionService(
        subscriptionRepo,
        planRepo,
        usageRepo,
    )
    
    // Use the service
    plans, err := subscriptionService.GetAvailablePlans(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d available plans\n", len(plans))
}
```

## 📋 Available Implementations

### GORM Implementation (`./gorm/`)

**Supports:**
- PostgreSQL, MySQL, SQLite, SQL Server
- Automatic preloading with relationships
- Transaction support
- Efficient queries with proper indexing
- Soft deletes
- Context cancellation

**Files:**
- `subscription_repository.go` - Subscription CRUD operations
- `subscription_plan_repository.go` - Plan management
- `usage_repository.go` - Usage tracking and events

**Features:**
- ✅ All interface methods implemented
- ✅ Proper error handling (converts GORM errors)
- ✅ Context support for cancellation
- ✅ Optimized queries with preloading
- ✅ Transaction-safe operations

## 🛠️ Creating Custom Implementations

### Example: Redis Cache Repository

Create a cache layer that implements the same interfaces:

```go
package redis

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
    "github.com/google/uuid"
    "github.com/karurosux/saas-go-kit/subscription-go"
)

type CachedSubscriptionRepository struct {
    primary subscription.SubscriptionRepository
    redis   *redis.Client
    ttl     time.Duration
}

func NewCachedSubscriptionRepository(
    primary subscription.SubscriptionRepository,
    redis *redis.Client,
    ttl time.Duration,
) subscription.SubscriptionRepository {
    return &CachedSubscriptionRepository{
        primary: primary,
        redis:   redis,
        ttl:     ttl,
    }
}

func (r *CachedSubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*subscription.Subscription, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("subscription:%s", id)
    cached, err := r.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var sub subscription.Subscription
        if json.Unmarshal([]byte(cached), &sub) == nil {
            return &sub, nil
        }
    }
    
    // Fall back to primary repository
    sub, err := r.primary.FindByID(ctx, id, preloads...)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    if data, err := json.Marshal(sub); err == nil {
        r.redis.Set(ctx, cacheKey, data, r.ttl)
    }
    
    return sub, nil
}

func (r *CachedSubscriptionRepository) Create(ctx context.Context, sub *subscription.Subscription) error {
    err := r.primary.Create(ctx, sub)
    if err != nil {
        return err
    }
    
    // Cache the new subscription
    cacheKey := fmt.Sprintf("subscription:%s", sub.ID)
    if data, err := json.Marshal(sub); err == nil {
        r.redis.Set(ctx, cacheKey, data, r.ttl)
    }
    
    return nil
}

// Implement other methods...
```

### Example: MongoDB Implementation

```go
package mongo

import (
    "context"
    
    "github.com/google/uuid"
    "github.com/karurosux/saas-go-kit/subscription-go"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
)

type MongoSubscriptionRepository struct {
    collection *mongo.Collection
}

func NewMongoSubscriptionRepository(db *mongo.Database) subscription.SubscriptionRepository {
    return &MongoSubscriptionRepository{
        collection: db.Collection("subscriptions"),
    }
}

func (r *MongoSubscriptionRepository) Create(ctx context.Context, sub *subscription.Subscription) error {
    _, err := r.collection.InsertOne(ctx, sub)
    return err
}

func (r *MongoSubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*subscription.Subscription, error) {
    var sub subscription.Subscription
    err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&sub)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, subscription.ErrSubscriptionNotFound
        }
        return nil, err
    }
    return &sub, nil
}

// Implement other methods...
```

### Example: In-Memory Repository for Testing

```go
package memory

import (
    "context"
    "sync"
    
    "github.com/google/uuid"
    "github.com/karurosux/saas-go-kit/subscription-go"
)

type InMemorySubscriptionRepository struct {
    mu            sync.RWMutex
    subscriptions map[uuid.UUID]*subscription.Subscription
}

func NewInMemorySubscriptionRepository() subscription.SubscriptionRepository {
    return &InMemorySubscriptionRepository{
        subscriptions: make(map[uuid.UUID]*subscription.Subscription),
    }
}

func (r *InMemorySubscriptionRepository) Create(ctx context.Context, sub *subscription.Subscription) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if sub.ID == uuid.Nil {
        sub.ID = uuid.New()
    }
    
    r.subscriptions[sub.ID] = sub
    return nil
}

func (r *InMemorySubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*subscription.Subscription, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    sub, exists := r.subscriptions[id]
    if !exists {
        return nil, subscription.ErrSubscriptionNotFound
    }
    
    // Return a copy to prevent modifications
    subCopy := *sub
    return &subCopy, nil
}

// Implement other methods...
```

## 🏗️ Repository Interface Design

All repository implementations must satisfy these interfaces:

```go
// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
    Create(ctx context.Context, subscription *Subscription) error
    FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*Subscription, error)
    FindByAccountID(ctx context.Context, accountID uuid.UUID) (*Subscription, error)
    FindByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*Subscription, error)
    Update(ctx context.Context, subscription *Subscription) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// SubscriptionPlanRepository defines the interface for subscription plan data access
type SubscriptionPlanRepository interface {
    FindAll(ctx context.Context) ([]SubscriptionPlan, error)
    FindAllIncludingHidden(ctx context.Context) ([]SubscriptionPlan, error)
    FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*SubscriptionPlan, error)
    FindByCode(ctx context.Context, code string) (*SubscriptionPlan, error)
    Create(ctx context.Context, plan *SubscriptionPlan) error
    Update(ctx context.Context, plan *SubscriptionPlan) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// UsageRepository defines the interface for usage tracking data access
type UsageRepository interface {
    Create(ctx context.Context, usage *SubscriptionUsage) error
    Update(ctx context.Context, usage *SubscriptionUsage) error
    FindByID(ctx context.Context, id uuid.UUID) (*SubscriptionUsage, error)
    FindBySubscriptionAndPeriod(ctx context.Context, subscriptionID uuid.UUID, periodStart, periodEnd time.Time) (*SubscriptionUsage, error)
    FindBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]*SubscriptionUsage, error)
    CreateEvent(ctx context.Context, event *UsageEvent) error
    FindEventsBySubscription(ctx context.Context, subscriptionID uuid.UUID, limit int) ([]*UsageEvent, error)
}
```

## ⚡ Performance Considerations

### GORM Optimizations

1. **Use Preloading Wisely**
```go
// Good: Only preload what you need
sub, err := repo.FindByID(ctx, id, "Plan", "Plan.Features")

// Bad: Preloading everything
sub, err := repo.FindByID(ctx, id, "Plan", "Usage", "Events")
```

2. **Batch Operations**
```go
// Create multiple plans at once
plans := []subscription.SubscriptionPlan{plan1, plan2, plan3}
err := db.Create(&plans).Error
```

3. **Use Transactions for Consistency**
```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&subscription).Error; err != nil {
        return err
    }
    if err := tx.Create(&usage).Error; err != nil {
        return err
    }
    return nil
})
```

### Caching Strategies

1. **Cache Frequently Accessed Data**
   - Subscription plans (rarely change)
   - Active subscriptions
   - Feature definitions

2. **Cache Invalidation**
   - Update cache when data changes
   - Use TTL for automatic expiration
   - Consider cache-aside pattern

## 🔧 Configuration

### GORM with PostgreSQL

```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

dsn := "host=localhost user=username password=password dbname=mydb port=5432 sslmode=disable"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // Log all SQL
    DryRun: false, // Set to true for testing queries without execution
})
```

### GORM with MySQL

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
```

## 🧪 Testing

### Unit Testing with In-Memory Repository

```go
func TestSubscriptionService(t *testing.T) {
    // Use in-memory repository for testing
    subscriptionRepo := memory.NewInMemorySubscriptionRepository()
    planRepo := memory.NewInMemoryPlanRepository()
    usageRepo := memory.NewInMemoryUsageRepository()
    
    service := subscription.NewSubscriptionService(
        subscriptionRepo,
        planRepo,
        usageRepo,
    )
    
    // Test service methods
    plans, err := service.GetAvailablePlans(context.Background())
    assert.NoError(t, err)
    assert.Empty(t, plans) // Should be empty initially
}
```

### Integration Testing with Test Database

```go
func TestSubscriptionRepository(t *testing.T) {
    // Set up test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := gormrepo.NewSubscriptionRepository(db)
    
    // Test repository methods
    sub := &subscription.Subscription{
        AccountID: uuid.New(),
        PlanID:    uuid.New(),
        Status:    subscription.SubscriptionActive,
    }
    
    err := repo.Create(context.Background(), sub)
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, sub.ID)
}
```

## 📚 Best Practices

### 1. Always Use Context
```go
// Good
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Model, error) {
    return r.db.WithContext(ctx).First(&model, id).Error
}

// Bad
func (r *Repository) FindByID(id uuid.UUID) (*Model, error) {
    return r.db.First(&model, id).Error
}
```

### 2. Handle Errors Consistently
```go
// Convert GORM errors to domain errors
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, subscription.ErrSubscriptionNotFound
}
```

### 3. Use Preloading for Related Data
```go
// Load subscription with plan details
sub, err := repo.FindByID(ctx, id, "Plan", "Plan.Features")
```

### 4. Implement Soft Deletes
```go
type BaseModel struct {
    ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    CreatedAt time.Time  `gorm:"autoCreateTime"`
    UpdatedAt time.Time  `gorm:"autoUpdateTime"`
    DeletedAt *time.Time `gorm:"index"` // Enables soft delete
}
```

### 5. Use Transactions for Multiple Operations
```go
func (s *Service) CreateSubscriptionWithUsage(ctx context.Context, req *CreateRequest) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Use tx instead of s.db for all operations
        if err := tx.Create(&subscription).Error; err != nil {
            return err
        }
        return tx.Create(&usage).Error
    })
}
```

## 🔗 Integration Examples

### With Echo Web Framework

```go
func setupSubscriptionHandlers(e *echo.Echo, service subscription.SubscriptionService) {
    g := e.Group("/api/subscriptions")
    
    g.GET("/plans", func(c echo.Context) error {
        plans, err := service.GetAvailablePlans(c.Request().Context())
        if err != nil {
            return err
        }
        return c.JSON(200, plans)
    })
}
```

### With Dependency Injection

```go
type Container struct {
    DB                   *gorm.DB
    SubscriptionRepo     subscription.SubscriptionRepository
    SubscriptionService  subscription.SubscriptionService
}

func NewContainer(db *gorm.DB) *Container {
    subscriptionRepo := gormrepo.NewSubscriptionRepository(db)
    planRepo := gormrepo.NewSubscriptionPlanRepository(db)
    usageRepo := gormrepo.NewUsageRepository(db)
    
    subscriptionService := subscription.NewSubscriptionService(
        subscriptionRepo,
        planRepo,
        usageRepo,
    )
    
    return &Container{
        DB:                  db,
        SubscriptionRepo:    subscriptionRepo,
        SubscriptionService: subscriptionService,
    }
}
```

This repository pattern provides the perfect balance between convenience and flexibility - ready-to-use implementations for rapid development, while maintaining clean interfaces for custom implementations when needed.