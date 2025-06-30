package subscription

import (
	"github.com/karurosux/saas-go-kit/core-go"
	"github.com/labstack/echo/v4"
)

type Module struct {
	config   ModuleConfig
	handlers *Handlers
}

type ModuleConfig struct {
	SubscriptionService SubscriptionService
	UsageService        UsageService
	PaymentService      PaymentService
	RoutePrefix         string
	AdminOnly           []string // Routes that require admin access
}

func NewModule(config ModuleConfig) core.Module {
	if config.RoutePrefix == "" {
		config.RoutePrefix = "/subscription"
	}

	handlers := NewHandlers(
		config.SubscriptionService,
		config.UsageService,
		config.PaymentService,
	)

	return &Module{
		config:   config,
		handlers: handlers,
	}
}

func (m *Module) Name() string {
	return "subscription"
}

func (m *Module) Mount(e *echo.Echo) error {
	g := e.Group(m.config.RoutePrefix)

	// Public routes
	g.GET("/plans", m.handlers.GetAvailablePlans)
	g.GET("/features", m.handlers.GetFeatureRegistry)
	g.GET("/features/category/:category", m.handlers.GetFeaturesByCategory)

	// User routes (require authentication)
	userGroup := g.Group("")
	// Note: Authentication middleware should be added by the application
	userGroup.GET("/me", m.handlers.GetUserSubscription)
	userGroup.GET("/usage", m.handlers.GetCurrentUsage)
	userGroup.GET("/permissions/:resourceType", m.handlers.CheckResourcePermission)
	userGroup.POST("/checkout", m.handlers.CreateCheckoutSession)
	userGroup.POST("/cancel", m.handlers.CancelSubscription)
	userGroup.POST("/portal", m.handlers.CreatePortalSession)
	userGroup.GET("/payment-methods", m.handlers.GetPaymentMethods)
	userGroup.GET("/invoices", m.handlers.GetInvoiceHistory)

	// Admin routes (require admin privileges)
	adminGroup := g.Group("/admin")
	// Note: Admin middleware should be added by the application
	adminGroup.GET("/plans", m.handlers.GetAllPlans)
	adminGroup.POST("/assign/:accountId/:planCode", m.handlers.AssignCustomPlan)

	// Webhook routes
	g.POST("/webhooks/stripe", m.handlers.HandleWebhook)

	return nil
}

func (m *Module) Dependencies() []string {
	return []string{"auth"} // Typically depends on auth module for user context
}

func (m *Module) Priority() int {
	return 100 // Standard priority
}