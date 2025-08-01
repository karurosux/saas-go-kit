package subscriptionmiddleware

import (
	"context"
	"fmt"
	
	"{{.Project.GoModule}}/internal/subscription/constants"
	"{{.Project.GoModule}}/internal/subscription/interface"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// GetAccountIDFromContext gets account ID from context
func GetAccountIDFromContext(c echo.Context) (uuid.UUID, error) {
	// Try from team context first
	if accountID := c.Get("account_id"); accountID != nil {
		switch v := accountID.(type) {
		case string:
			return uuid.Parse(v)
		case uuid.UUID:
			return v, nil
		}
	}
	
	// Try from path parameter
	if accountIDStr := c.Param("accountId"); accountIDStr != "" {
		return uuid.Parse(accountIDStr)
	}
	
	// Try from query parameter
	if accountIDStr := c.QueryParam("account_id"); accountIDStr != "" {
		return uuid.Parse(accountIDStr)
	}
	
	return uuid.Nil, echo.NewHTTPError(400, "account ID not provided")
}

// GetSubscriptionFromContext gets subscription from context
func GetSubscriptionFromContext(c echo.Context) subscriptioninterface.Subscription {
	if subscription := c.Get(subscriptionconstants.ContextKeySubscription); subscription != nil {
		if sub, ok := subscription.(subscriptioninterface.Subscription); ok {
			return sub
		}
	}
	return nil
}

// GetPlanFromContext gets plan from context
func GetPlanFromContext(c echo.Context) subscriptioninterface.SubscriptionPlan {
	if plan := c.Get(subscriptionconstants.ContextKeyPlan); plan != nil {
		if p, ok := plan.(subscriptioninterface.SubscriptionPlan); ok {
			return p
		}
	}
	return nil
}

// GetUsageLimitsFromContext gets usage limits from context
func GetUsageLimitsFromContext(c echo.Context) map[string]int64 {
	if limits := c.Get(subscriptionconstants.ContextKeyUsageLimits); limits != nil {
		if l, ok := limits.(map[string]int64); ok {
			return l
		}
	}
	return nil
}

// HasActiveSubscription checks if account has active subscription
func HasActiveSubscription(c echo.Context) bool {
	subscription := GetSubscriptionFromContext(c)
	if subscription == nil {
		return false
	}
	
	status := subscription.GetStatus()
	return status == subscriptioninterface.StatusActive || status == subscriptioninterface.StatusTrialing
}

// GetPlanType gets the current plan type
func GetPlanType(c echo.Context) subscriptioninterface.PlanType {
	plan := GetPlanFromContext(c)
	if plan == nil {
		return subscriptioninterface.PlanTypeFree
	}
	return plan.GetType()
}

// CheckResourceLimit checks if a resource is within limits
func CheckResourceLimit(c echo.Context, resource string, requestedQuantity int64) (bool, int64) {
	limits := GetUsageLimitsFromContext(c)
	if limits == nil {
		return true, -1 // No limits
	}
	
	limit, exists := limits[resource]
	if !exists {
		return true, -1 // No limit for this resource
	}
	
	// Get current usage from context or service
	// This is a simplified version - in production you'd query the service
	currentUsage := int64(0)
	
	remaining := limit - currentUsage - requestedQuantity
	return remaining >= 0, remaining
}

// SetAccountIDInContext sets account ID in context
func SetAccountIDInContext(c echo.Context, accountID uuid.UUID) {
	c.Set("account_id", accountID)
}