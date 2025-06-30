# Subscription Module

A comprehensive subscription and billing management module for SaaS applications, featuring Stripe integration, usage tracking, and flexible plan management.

## Features

- **Subscription Management**: Complete subscription lifecycle management
- **Flexible Plan System**: Feature-based plans with limits and flags
- **Usage Tracking**: Real-time usage monitoring and limits enforcement
- **Stripe Integration**: Full Stripe payment processing support
- **Webhook Handling**: Automated webhook processing for payment events
- **Customer Portal**: Self-service billing portal integration
- **Permission System**: Resource access control based on subscription limits

## Installation

```bash
go get github.com/karurosux/saas-go-kit/subscription-go
```

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "github.com/karurosux/saas-go-kit/core-go"
    "github.com/karurosux/saas-go-kit/subscription-go"
    "gorm.io/gorm"
)

func main() {
    // Setup your database and repositories
    var db *gorm.DB // your database connection
    
    // Initialize repositories
    subscriptionRepo := NewSubscriptionRepository(db)
    planRepo := NewSubscriptionPlanRepository(db)
    usageRepo := NewUsageRepository(db)
    
    // Initialize services
    subscriptionService := subscription.NewSubscriptionService(
        subscriptionRepo, planRepo, usageRepo,
    )
    usageService := subscription.NewUsageService(usageRepo, subscriptionRepo)
    
    // Setup Stripe provider
    stripeProvider := subscription.NewStripeProvider()
    stripeProvider.Initialize(subscription.PaymentConfig{
        SecretKey:     "sk_test_...",
        WebhookSecret: "whsec_...",
    })
    
    paymentService := subscription.NewPaymentService(
        stripeProvider, subscriptionRepo, planRepo,
    )
    
    // Create and mount module
    module := subscription.NewModule(subscription.ModuleConfig{
        SubscriptionService: subscriptionService,
        UsageService:        usageService,
        PaymentService:      paymentService,
        RoutePrefix:         "/api/subscription",
    })
    
    app := core.NewKit(nil, core.KitConfig{})
    app.Register(module)
    app.Mount()
}
```

### 2. Define Your Plans

```go
// Create subscription plans
plans := []subscription.SubscriptionPlan{
    {
        Name:        "Starter",
        Code:        "starter",
        Description: "Perfect for small businesses",
        Price:       29.99,
        Currency:    "USD",
        Interval:    "month",
        Features: subscription.PlanFeatures{
            Limits: map[string]int64{
                subscription.LimitRestaurants:      1,
                subscription.LimitFeedbacksPerMonth: 100,
                subscription.LimitTeamMembers:      3,
            },
            Flags: map[string]bool{
                subscription.FlagAdvancedAnalytics: false,
                subscription.FlagCustomBranding:    false,
            },
        },
        StripePriceID: "price_1234567890",
    },
    {
        Name:        "Professional",
        Code:        "pro",
        Description: "For growing businesses",
        Price:       99.99,
        Currency:    "USD",
        Interval:    "month",
        Features: subscription.PlanFeatures{
            Limits: map[string]int64{
                subscription.LimitRestaurants:      5,
                subscription.LimitFeedbacksPerMonth: -1, // unlimited
                subscription.LimitTeamMembers:      10,
            },
            Flags: map[string]bool{
                subscription.FlagAdvancedAnalytics: true,
                subscription.FlagCustomBranding:    true,
                subscription.FlagAPIAccess:         true,
            },
        },
        StripePriceID: "price_0987654321",
    },
}
```

### 3. Track Usage

```go
// Track resource usage
err := usageService.TrackUsage(ctx, subscriptionID, subscription.ResourceTypeRestaurant, 1)
if err != nil {
    log.Printf("Failed to track usage: %v", err)
}

// Check if user can add more resources
canAdd, reason, err := usageService.CanAddResource(ctx, subscriptionID, subscription.ResourceTypeRestaurant)
if err != nil {
    log.Printf("Failed to check permissions: %v", err)
} else if !canAdd {
    log.Printf("Cannot add resource: %s", reason)
}

// Record usage events for auditing
event := &subscription.UsageEvent{
    SubscriptionID: subscriptionID,
    EventType:      subscription.EventTypeCreate,
    ResourceType:   subscription.ResourceTypeRestaurant,
    ResourceID:     restaurantID,
    Metadata:       `{"name": "My Restaurant"}`,
}
err = usageService.RecordUsageEvent(ctx, event)
```

## API Endpoints

### Public Endpoints

- `GET /subscription/plans` - Get available subscription plans
- `GET /subscription/features` - Get feature registry
- `GET /subscription/features/category/:category` - Get features by category

### User Endpoints (Require Authentication)

- `GET /subscription/me` - Get user's current subscription
- `GET /subscription/usage` - Get current usage statistics
- `GET /subscription/permissions/:resourceType` - Check resource permissions
- `POST /subscription/checkout` - Create Stripe checkout session
- `POST /subscription/cancel` - Cancel subscription
- `POST /subscription/portal` - Create customer portal session
- `GET /subscription/payment-methods` - Get payment methods
- `GET /subscription/invoices` - Get invoice history

### Admin Endpoints

- `GET /subscription/admin/plans` - Get all plans (including hidden)
- `POST /subscription/admin/assign/:accountId/:planCode` - Assign custom plan

### Webhook Endpoints

- `POST /subscription/webhooks/stripe` - Stripe webhook handler

## Plan Features System

The module uses a flexible feature system with three types:

### Limits
Numerical limits for resources:
```go
const (
    LimitRestaurants            = "max_restaurants"
    LimitFeedbacksPerMonth      = "max_feedbacks_per_month"
    LimitTeamMembers            = "max_team_members"
    LimitStorageGB              = "max_storage_gb"
    LimitAPICallsPerHour        = "max_api_calls_per_hour"
)
```

### Feature Flags
Boolean flags for features:
```go
const (
    FlagAdvancedAnalytics = "advanced_analytics"
    FlagCustomBranding    = "custom_branding"
    FlagAPIAccess         = "api_access"
    FlagPrioritySupport   = "priority_support"
    FlagWhiteLabel        = "white_label"
    FlagCustomDomain      = "custom_domain"
)
```

### Usage Example

```go
// Check if user can add a restaurant
subscription, _ := subscriptionService.GetUserSubscription(ctx, accountID)
canAdd := subscription.CanAddResource("restaurant", currentRestaurantCount)

// Check if user has advanced analytics
hasAnalytics := subscription.Plan.Features.GetFlag(subscription.FlagAdvancedAnalytics)

// Get restaurant limit
limit := subscription.Plan.Features.GetLimit(subscription.LimitRestaurants)
isUnlimited := subscription.Plan.Features.IsUnlimited(subscription.LimitRestaurants) // limit == -1
```

## Stripe Integration

### Webhook Configuration

Configure your Stripe webhook endpoint to point to `/subscription/webhooks/stripe` and select these events:

- `checkout.session.completed`
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`
- `invoice.payment_succeeded`
- `invoice.payment_failed`

### Environment Variables

```env
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

## Database Schema

The module requires these database tables:

```sql
-- Subscription plans
CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR NOT NULL,
    code VARCHAR UNIQUE NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    interval VARCHAR DEFAULT 'month',
    features JSONB,
    is_active BOOLEAN DEFAULT true,
    is_visible BOOLEAN DEFAULT true,
    trial_days INTEGER DEFAULT 0,
    stripe_price_id VARCHAR,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- User subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL,
    plan_id UUID REFERENCES subscription_plans(id),
    status VARCHAR NOT NULL,
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    cancel_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    stripe_customer_id VARCHAR,
    stripe_subscription_id VARCHAR,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Usage tracking
CREATE TABLE subscription_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID REFERENCES subscriptions(id),
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    feedbacks_count INTEGER DEFAULT 0,
    restaurants_count INTEGER DEFAULT 0,
    locations_count INTEGER DEFAULT 0,
    qr_codes_count INTEGER DEFAULT 0,
    team_members_count INTEGER DEFAULT 0,
    last_updated_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Usage events (for auditing)
CREATE TABLE usage_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID REFERENCES subscriptions(id),
    event_type VARCHAR NOT NULL,
    resource_type VARCHAR NOT NULL,
    resource_id UUID,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

## Configuration

### Module Configuration

```go
type ModuleConfig struct {
    SubscriptionService SubscriptionService
    UsageService        UsageService
    PaymentService      PaymentService
    RoutePrefix         string            // Default: "/subscription"
    AdminOnly           []string          // Routes requiring admin access
}
```

### Payment Provider Configuration

```go
type PaymentConfig struct {
    SecretKey      string                 // Stripe secret key
    PublishableKey string                 // Stripe publishable key
    WebhookSecret  string                 // Stripe webhook secret
    Extra          map[string]interface{} // Provider-specific config
}
```

## Testing

```bash
# Run tests for the subscription module
cd subscription-go
go test ./...

# Run with coverage
go test -cover ./...
```

## Dependencies

- `github.com/google/uuid` - UUID support
- `github.com/stripe/stripe-go/v76` - Stripe API client
- `github.com/labstack/echo/v4` - HTTP framework
- `gorm.io/gorm` - ORM for database operations
- `github.com/karurosux/saas-go-kit/core-go` - Core module system
- `github.com/karurosux/saas-go-kit/errors-go` - Error handling
- `github.com/karurosux/saas-go-kit/response-go` - Response formatting

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.