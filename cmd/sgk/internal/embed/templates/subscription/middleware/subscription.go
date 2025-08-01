package subscriptionmiddleware

import (
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/subscription/constants"
	"{{.Project.GoModule}}/internal/subscription/interface"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SubscriptionMiddleware provides subscription-related middleware
type SubscriptionMiddleware struct {
	subscriptionService subscriptioninterface.SubscriptionService
}

// NewSubscriptionMiddleware creates a new subscription middleware
func NewSubscriptionMiddleware(subscriptionService subscriptioninterface.SubscriptionService) *SubscriptionMiddleware {
	return &SubscriptionMiddleware{
		subscriptionService: subscriptionService,
	}
}

// RequireActiveSubscription middleware that requires an active subscription
func (m *SubscriptionMiddleware) RequireActiveSubscription() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return core.Error(c, core.BadRequest("account ID not provided"))
			}
			
			subscription, err := m.subscriptionService.GetSubscription(c.Request().Context(), accountID)
			if err != nil {
				return core.Error(c, core.PaymentRequired("No active subscription"))
			}
			
			// Check if subscription is active
			if subscription.GetStatus() != subscriptioninterface.StatusActive && 
			   subscription.GetStatus() != subscriptioninterface.StatusTrialing {
				return core.Error(c, core.PaymentRequired("Subscription is not active"))
			}
			
			// Store subscription in context
			c.Set(subscriptionconstants.ContextKeySubscription, subscription)
			c.Set(subscriptionconstants.ContextKeyPlan, subscription.GetPlan())
			
			return next(c)
		}
	}
}

// RequirePlan middleware that requires a specific plan or higher
func (m *SubscriptionMiddleware) RequirePlan(minPlan subscriptioninterface.PlanType) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return core.Error(c, core.BadRequest("account ID not provided"))
			}
			
			subscription, err := m.subscriptionService.GetSubscription(c.Request().Context(), accountID)
			if err != nil {
				return core.Error(c, core.PaymentRequired("No active subscription"))
			}
			
			// Check if subscription is active
			if subscription.GetStatus() != subscriptioninterface.StatusActive && 
			   subscription.GetStatus() != subscriptioninterface.StatusTrialing {
				return core.Error(c, core.PaymentRequired("Subscription is not active"))
			}
			
			// Check plan level
			if !hasSufficientPlan(subscription.GetPlan().GetType(), minPlan) {
				return core.Error(c, core.PaymentRequired("Upgrade required for this feature"))
			}
			
			// Store subscription in context
			c.Set(subscriptionconstants.ContextKeySubscription, subscription)
			c.Set(subscriptionconstants.ContextKeyPlan, subscription.GetPlan())
			
			return next(c)
		}
	}
}

// CheckUsageLimit middleware that checks if a resource is within limits
func (m *SubscriptionMiddleware) CheckUsageLimit(resource string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return core.Error(c, core.BadRequest("account ID not provided"))
			}
			
			// Check limit
			withinLimit, remaining, err := m.subscriptionService.CheckLimit(
				c.Request().Context(), 
				accountID, 
				resource,
			)
			if err != nil {
				return core.Error(c, err)
			}
			
			if !withinLimit {
				return core.Error(c, core.PaymentRequired(subscriptionconstants.ErrUsageLimitExceeded))
			}
			
			// Add remaining to response headers
			if remaining >= 0 {
				c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			}
			
			return next(c)
		}
	}
}

// TrackUsage middleware that tracks resource usage
func (m *SubscriptionMiddleware) TrackUsage(resource string, quantity int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Execute the handler first
			err := next(c)
			
			// Track usage after successful execution
			if err == nil {
				accountID, _ := GetAccountIDFromContext(c)
				if accountID != uuid.Nil {
					// Track usage asynchronously
					go m.subscriptionService.TrackUsage(
						context.Background(),
						accountID,
						resource,
						quantity,
					)
				}
			}
			
			return err
		}
	}
}

// InjectSubscription middleware that injects subscription info into context
func (m *SubscriptionMiddleware) InjectSubscription() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				// No account context, continue without injecting
				return next(c)
			}
			
			// Try to get subscription
			subscription, err := m.subscriptionService.GetSubscription(c.Request().Context(), accountID)
			if err == nil {
				c.Set(subscriptionconstants.ContextKeySubscription, subscription)
				c.Set(subscriptionconstants.ContextKeyPlan, subscription.GetPlan())
				
				// Set usage limits in context
				if subscription.GetPlan() != nil {
					c.Set(subscriptionconstants.ContextKeyUsageLimits, subscription.GetPlan().GetLimits())
				}
			}
			
			return next(c)
		}
	}
}

// hasSufficientPlan checks if a plan meets the minimum requirement
func hasSufficientPlan(userPlan, minPlan subscriptioninterface.PlanType) bool {
	planHierarchy := map[subscriptioninterface.PlanType]int{
		subscriptioninterface.PlanTypeFree:       1,
		subscriptioninterface.PlanTypeStarter:    2,
		subscriptioninterface.PlanTypePro:        3,
		subscriptioninterface.PlanTypeEnterprise: 4,
	}
	
	userLevel, ok1 := planHierarchy[userPlan]
	minLevel, ok2 := planHierarchy[minPlan]
	
	if !ok1 || !ok2 {
		return false
	}
	
	return userLevel >= minLevel
}