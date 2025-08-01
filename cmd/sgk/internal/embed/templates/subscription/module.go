package subscription

import (
	"encoding/json"
	"fmt"
	"os"
	"{{.Project.GoModule}}/internal/core"

	subscriptioncontroller "{{.Project.GoModule}}/internal/subscription/controller"
	subscriptioninterface "{{.Project.GoModule}}/internal/subscription/interface"
	subscriptionmiddleware "{{.Project.GoModule}}/internal/subscription/middleware"
	subscriptionmodel "{{.Project.GoModule}}/internal/subscription/model"
	subscriptionprovider "{{.Project.GoModule}}/internal/subscription/provider"
	subscriptiongorm "{{.Project.GoModule}}/internal/subscription/repository/gorm"
	subscriptionservice "{{.Project.GoModule}}/internal/subscription/service"

	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func RegisterModule(c *core.Container) error {
	eInt, err := c.Get("echo")
	if err != nil {
		return fmt.Errorf("echo instance not found in container: %w", err)
	}
	e, ok := eInt.(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance has invalid type")
	}

	dbInt, err := c.Get("db")
	if err != nil {
		return fmt.Errorf("database instance not found in container: %w", err)
	}
	db, ok := dbInt.(*gorm.DB)
	if !ok {
		return fmt.Errorf("database instance has invalid type")
	}

	if err := subscriptiongorm.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to run subscription migrations: %w", err)
	}

	planRepo := subscriptiongorm.NewSubscriptionPlanRepository(db)
	subscriptionRepo := subscriptiongorm.NewSubscriptionRepository(db)
	usageRepo := subscriptiongorm.NewUsageRepository(db)

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

	paymentProvider := subscriptionprovider.NewStripeProvider(stripeKey)

	subscriptionService := subscriptionservice.NewSubscriptionService(
		planRepo,
		subscriptionRepo,
		usageRepo,
		paymentProvider,
		webhookSecret,
		returnURL,
	)

	subscriptionMiddleware := subscriptionmiddleware.NewSubscriptionMiddleware(subscriptionService)

	subscriptionController := subscriptioncontroller.NewSubscriptionController(subscriptionService)

	subscriptionController.RegisterRoutes(e, "/subscriptions", subscriptionMiddleware)

	if err := seedDefaultPlans(db); err != nil {
		return fmt.Errorf("failed to seed default plans: %w", err)
	}

	c.Set("subscription.service", subscriptionService)
	c.Set("subscription.middleware", subscriptionMiddleware)
	c.Set("subscription.planRepository", planRepo)
	c.Set("subscription.subscriptionRepository", subscriptionRepo)
	c.Set("subscription.usageRepository", usageRepo)

	return nil
}

func seedDefaultPlans(db *gorm.DB) error {
	plans := []subscriptionmodel.SubscriptionPlan{
		{
			Name:         "Free",
			Type:         subscriptioninterface.PlanTypeFree,
			PriceMonthly: 0,
			PriceYearly:  0,
			Features: datatypes.JSON(jsonMustMarshal(map[string]any{
				"api_requests":  true,
				"basic_support": true,
			})),
			Limits: datatypes.JSON(jsonMustMarshal(map[string]int64{
				"api_requests": 1000,
				"team_members": 3,
				"projects":     1,
			})),
			TrialDays: 0,
			IsActive:  true,
		},
		{
			Name:         "Starter",
			Type:         subscriptioninterface.PlanTypeStarter,
			PriceMonthly: 2900,  // $29.00
			PriceYearly:  29900, // $299.00
			Features: datatypes.JSON(jsonMustMarshal(map[string]any{
				"api_requests":     true,
				"priority_support": true,
				"custom_domains":   true,
			})),
			Limits: datatypes.JSON(jsonMustMarshal(map[string]int64{
				"api_requests":   10000,
				"team_members":   10,
				"projects":       5,
				"custom_domains": 1,
			})),
			TrialDays: 14,
			IsActive:  true,
		},
		{
			Name:         "Pro",
			Type:         subscriptioninterface.PlanTypePro,
			PriceMonthly: 9900,  // $99.00
			PriceYearly:  99900, // $999.00
			Features: datatypes.JSON(jsonMustMarshal(map[string]any{
				"api_requests":       true,
				"priority_support":   true,
				"custom_domains":     true,
				"advanced_analytics": true,
				"sso":                true,
			})),
			Limits: datatypes.JSON(jsonMustMarshal(map[string]int64{
				"api_requests":   100000,
				"team_members":   50,
				"projects":       25,
				"custom_domains": 10,
			})),
			TrialDays: 14,
			IsActive:  true,
		},
		{
			Name:         "Enterprise",
			Type:         subscriptioninterface.PlanTypeEnterprise,
			PriceMonthly: 49900,  // $499.00
			PriceYearly:  499900, // $4999.00
			Features: datatypes.JSON(jsonMustMarshal(map[string]any{
				"api_requests":        true,
				"dedicated_support":   true,
				"custom_domains":      true,
				"advanced_analytics":  true,
				"sso":                 true,
				"audit_logs":          true,
				"custom_integrations": true,
			})),
			Limits: datatypes.JSON(jsonMustMarshal(map[string]int64{
				"api_requests":   -1, // Unlimited
				"team_members":   -1, // Unlimited
				"projects":       -1, // Unlimited
				"custom_domains": -1, // Unlimited
			})),
			TrialDays: 30,
			IsActive:  true,
		},
	}

	for _, plan := range plans {
		var existing subscriptionmodel.SubscriptionPlan
		if err := db.Where("type = ?", plan.Type).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
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

func jsonMustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
