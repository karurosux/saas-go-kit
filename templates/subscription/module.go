package subscription

import (
	"encoding/json"
	"fmt"
	"os"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/subscription/controller"
	"{{.Project.GoModule}}/internal/subscription/interface"
	"{{.Project.GoModule}}/internal/subscription/middleware"
	"{{.Project.GoModule}}/internal/subscription/model"
	"{{.Project.GoModule}}/internal/subscription/provider"
	"{{.Project.GoModule}}/internal/subscription/repository/gorm"
	"{{.Project.GoModule}}/internal/subscription/service"
	"github.com/labstack/echo/v4"
	gormdb "gorm.io/gorm"
	"gorm.io/datatypes"
)

// RegisterModule registers the subscription module with the container
func RegisterModule(c core.Container) error {
	// Get dependencies from container
	e, ok := c.Get("echo").(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance not found in container")
	}
	
	db, ok := c.Get("db").(*gormdb.DB)
	if !ok {
		return fmt.Errorf("database instance not found in container")
	}
	
	// Run migrations
	if err := gorm.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to run subscription migrations: %w", err)
	}
	
	// Create repositories
	planRepo := gorm.NewSubscriptionPlanRepository(db)
	subscriptionRepo := gorm.NewSubscriptionRepository(db)
	usageRepo := gorm.NewUsageRepository(db)
	
	// Get Stripe configuration
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY not configured")
	}
	
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return fmt.Errorf("STRIPE_WEBHOOK_SECRET not configured")
	}
	
	returnURL := os.Getenv("BASE_URL")
	if returnURL == "" {
		returnURL = "http://localhost:8080"
	}
	
	// Create payment provider
	paymentProvider := provider.NewStripeProvider(stripeKey)
	
	// Create subscription service
	subscriptionService := subscriptionservice.NewSubscriptionService(
		planRepo,
		subscriptionRepo,
		usageRepo,
		paymentProvider,
		webhookSecret,
		returnURL,
	)
	
	// Create middleware
	subscriptionMiddleware := subscriptionmiddleware.NewSubscriptionMiddleware(subscriptionService)
	
	// Create controller
	subscriptionController := subscriptioncontroller.NewSubscriptionController(subscriptionService)
	
	// Register routes
	subscriptionController.RegisterRoutes(e, "/subscriptions", subscriptionMiddleware)
	
	// Seed default plans if they don't exist
	if err := seedDefaultPlans(db); err != nil {
		return fmt.Errorf("failed to seed default plans: %w", err)
	}
	
	// Register components in container for other modules to use
	c.Set("subscription.service", subscriptionService)
	c.Set("subscription.middleware", subscriptionMiddleware)
	c.Set("subscription.planRepository", planRepo)
	c.Set("subscription.subscriptionRepository", subscriptionRepo)
	c.Set("subscription.usageRepository", usageRepo)
	
	return nil
}

// seedDefaultPlans creates default subscription plans
func seedDefaultPlans(db *gormdb.DB) error {
	plans := []subscriptionmodel.SubscriptionPlan{
		{
			Name:         "Free",
			Type:         subscriptioninterface.PlanTypeFree,
			PriceMonthly: 0,
			PriceYearly:  0,
			Features: datatypes.JSON(jsonMustMarshal(map[string]interface{}{
				"api_requests": true,
				"basic_support": true,
			})),
			Limits: datatypes.JSON(jsonMustMarshal(map[string]int64{
				"api_requests": 1000,
				"team_members": 3,
				"projects": 1,
			})),
			TrialDays: 0,
			IsActive:  true,
		},
		{
			Name:         "Starter",
			Type:         subscriptioninterface.PlanTypeStarter,
			PriceMonthly: 2900, // $29.00
			PriceYearly:  29900, // $299.00
			Features: datatypes.JSON(jsonMustMarshal(map[string]interface{}{
				"api_requests": true,
				"priority_support": true,
				"custom_domains": true,
			})),
			Limits: datatypes.JSON(jsonMustMarshal(map[string]int64{
				"api_requests": 10000,
				"team_members": 10,
				"projects": 5,
				"custom_domains": 1,
			})),
			TrialDays: 14,
			IsActive:  true,
		},
		{
			Name:         "Pro",
			Type:         subscriptioninterface.PlanTypePro,
			PriceMonthly: 9900, // $99.00
			PriceYearly:  99900, // $999.00
			Features: datatypes.JSON(jsonMustMarshal(map[string]interface{}{
				"api_requests": true,
				"priority_support": true,
				"custom_domains": true,
				"advanced_analytics": true,
				"sso": true,
			})),
			Limits: datatypes.JSON(jsonMustMarshal(map[string]int64{
				"api_requests": 100000,
				"team_members": 50,
				"projects": 25,
				"custom_domains": 10,
			})),
			TrialDays: 14,
			IsActive:  true,
		},
		{
			Name:         "Enterprise",
			Type:         subscriptioninterface.PlanTypeEnterprise,
			PriceMonthly: 49900, // $499.00
			PriceYearly:  499900, // $4999.00
			Features: datatypes.JSON(jsonMustMarshal(map[string]interface{}{
				"api_requests": true,
				"dedicated_support": true,
				"custom_domains": true,
				"advanced_analytics": true,
				"sso": true,
				"audit_logs": true,
				"custom_integrations": true,
			})),
			Limits: datatypes.JSON(jsonMustMarshal(map[string]int64{
				"api_requests": -1, // Unlimited
				"team_members": -1, // Unlimited
				"projects": -1, // Unlimited
				"custom_domains": -1, // Unlimited
			})),
			TrialDays: 30,
			IsActive:  true,
		},
	}
	
	for _, plan := range plans {
		var existing subscriptionmodel.SubscriptionPlan
		if err := db.Where("type = ?", plan.Type).First(&existing).Error; err != nil {
			if err == gormdb.ErrRecordNotFound {
				// Create plan
				if err := db.Create(&plan).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	
	return nil
}

func jsonMustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}