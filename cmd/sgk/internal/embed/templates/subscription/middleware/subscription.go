package subscriptionmiddleware

import (
	"context"
	"fmt"
	"{{.Project.GoModule}}/internal/core"
	subscriptionconstants "{{.Project.GoModule}}/internal/subscription/constants"
	subscriptioninterface "{{.Project.GoModule}}/internal/subscription/interface"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SubscriptionMiddleware struct {
	subscriptionService subscriptioninterface.SubscriptionService
}

func NewSubscriptionMiddleware(subscriptionService subscriptioninterface.SubscriptionService) *SubscriptionMiddleware {
	return &SubscriptionMiddleware{
		subscriptionService: subscriptionService,
	}
}

func (m *SubscriptionMiddleware) RequireActiveSubscription() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return core.BadRequest(c, fmt.Errorf("account ID not provided"))
			}
			
			subscription, err := m.subscriptionService.GetSubscription(c.Request().Context(), accountID)
			if err != nil {
				return core.PaymentRequired(c, fmt.Errorf("No active subscription"))
			}
			
			if subscription.GetStatus() != subscriptioninterface.StatusActive && 
			   subscription.GetStatus() != subscriptioninterface.StatusTrialing {
				return core.PaymentRequired(c, fmt.Errorf("Subscription is not active"))
			}
			
			c.Set(subscriptionconstants.ContextKeySubscription, subscription)
			c.Set(subscriptionconstants.ContextKeyPlan, subscription.GetPlan())
			
			return next(c)
		}
	}
}

func (m *SubscriptionMiddleware) RequirePlan(minPlan subscriptioninterface.PlanType) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return core.BadRequest(c, fmt.Errorf("account ID not provided"))
			}
			
			subscription, err := m.subscriptionService.GetSubscription(c.Request().Context(), accountID)
			if err != nil {
				return core.PaymentRequired(c, fmt.Errorf("No active subscription"))
			}
			
			if subscription.GetStatus() != subscriptioninterface.StatusActive && 
			   subscription.GetStatus() != subscriptioninterface.StatusTrialing {
				return core.PaymentRequired(c, fmt.Errorf("Subscription is not active"))
			}
			
			if !hasSufficientPlan(subscription.GetPlan().GetType(), minPlan) {
				return core.PaymentRequired(c, fmt.Errorf("Upgrade required for this feature"))
			}
			
			c.Set(subscriptionconstants.ContextKeySubscription, subscription)
			c.Set(subscriptionconstants.ContextKeyPlan, subscription.GetPlan())
			
			return next(c)
		}
	}
}

func (m *SubscriptionMiddleware) CheckUsageLimit(resource string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return core.BadRequest(c, fmt.Errorf("account ID not provided"))
			}
			
			withinLimit, remaining, err := m.subscriptionService.CheckLimit(
				c.Request().Context(), 
				accountID, 
				resource,
			)
			if err != nil {
				return core.InternalServerError(c, err)
			}
			
			if !withinLimit {
				return core.PaymentRequired(c, fmt.Errorf(subscriptionconstants.ErrUsageLimitExceeded))
			}
			
			if remaining >= 0 {
				c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			}
			
			return next(c)
		}
	}
}

func (m *SubscriptionMiddleware) TrackUsage(resource string, quantity int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			
			if err == nil {
				accountID, _ := GetAccountIDFromContext(c)
				if accountID != uuid.Nil {
					go func() {
						_ = m.subscriptionService.TrackUsage(
							context.Background(),
							accountID,
							resource,
							quantity,
						)
					}()
				}
			}
			
			return err
		}
	}
}

func (m *SubscriptionMiddleware) InjectSubscription() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accountID, err := GetAccountIDFromContext(c)
			if err != nil {
				return next(c)
			}
			
			subscription, err := m.subscriptionService.GetSubscription(c.Request().Context(), accountID)
			if err == nil {
				c.Set(subscriptionconstants.ContextKeySubscription, subscription)
				c.Set(subscriptionconstants.ContextKeyPlan, subscription.GetPlan())
				
				if subscription.GetPlan() != nil {
					c.Set(subscriptionconstants.ContextKeyUsageLimits, subscription.GetPlan().GetLimits())
				}
			}
			
			return next(c)
		}
	}
}

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