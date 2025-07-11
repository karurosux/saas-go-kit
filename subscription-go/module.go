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
	MiddlewareConfig    *MiddlewareConfig // Optional middleware configuration
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

func (m *Module) Routes() []core.Route {
	prefix := m.config.RoutePrefix
	return []core.Route{
		{Method: "GET", Path: prefix + "/plans", Handler: m.handlers.GetAvailablePlans},
		{Method: "GET", Path: prefix + "/features", Handler: m.handlers.GetFeatureRegistry},
		{Method: "GET", Path: prefix + "/features/category/:category", Handler: m.handlers.GetFeaturesByCategory},
		{Method: "GET", Path: prefix + "/me", Handler: m.handlers.GetUserSubscription},
		{Method: "GET", Path: prefix + "/usage", Handler: m.handlers.GetCurrentUsage},
		{Method: "GET", Path: prefix + "/permissions/:resourceType", Handler: m.handlers.CheckResourcePermission},
		{Method: "POST", Path: prefix + "/checkout", Handler: m.handlers.CreateCheckoutSession},
		{Method: "POST", Path: prefix + "/cancel", Handler: m.handlers.CancelSubscription},
		{Method: "POST", Path: prefix + "/portal", Handler: m.handlers.CreatePortalSession},
		{Method: "GET", Path: prefix + "/payment-methods", Handler: m.handlers.GetPaymentMethods},
		{Method: "GET", Path: prefix + "/invoices", Handler: m.handlers.GetInvoiceHistory},
		{Method: "GET", Path: prefix + "/admin/plans", Handler: m.handlers.GetAllPlans},
		{Method: "POST", Path: prefix + "/admin/assign/:accountId/:planCode", Handler: m.handlers.AssignCustomPlan},
		{Method: "POST", Path: prefix + "/webhooks/stripe", Handler: m.handlers.HandleWebhook},
	}
}

func (m *Module) Middleware() []echo.MiddlewareFunc {
	middlewares := []echo.MiddlewareFunc{}
	
	// Add default active subscription middleware if config is provided
	if m.config.MiddlewareConfig != nil {
		middlewares = append(middlewares, RequireActiveSubscription(*m.config.MiddlewareConfig))
	}
	
	return middlewares
}

func (m *Module) Dependencies() []string {
	return []string{"auth"}
}

func (m *Module) Init(deps map[string]core.Module) error {
	return nil
}

// Middleware helper methods for easy access

// GetFeatureFlagMiddleware returns middleware that requires a specific feature flag
func (m *Module) GetFeatureFlagMiddleware(featureFlag string) echo.MiddlewareFunc {
	if m.config.MiddlewareConfig == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}
	return RequireFeatureFlag(*m.config.MiddlewareConfig, featureFlag)
}

// GetResourceLimitMiddleware returns middleware that checks resource limits
func (m *Module) GetResourceLimitMiddleware(resourceType string, limitKey string) echo.MiddlewareFunc {
	if m.config.MiddlewareConfig == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}
	return RequireResourceLimit(*m.config.MiddlewareConfig, resourceType, limitKey)
}

// GetPlanTierMiddleware returns middleware that requires specific plan tiers
func (m *Module) GetPlanTierMiddleware(allowedPlanCodes []string) echo.MiddlewareFunc {
	if m.config.MiddlewareConfig == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}
	return RequirePlanTier(*m.config.MiddlewareConfig, allowedPlanCodes)
}

// GetUsageTrackingMiddleware returns middleware that tracks resource usage
func (m *Module) GetUsageTrackingMiddleware(resourceType string) echo.MiddlewareFunc {
	if m.config.MiddlewareConfig == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}
	return UsageTrackingMiddleware(*m.config.MiddlewareConfig, resourceType)
}