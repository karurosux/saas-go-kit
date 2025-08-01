package subscriptionmiddleware

import (
	subscriptionconstants "{{.Project.GoModule}}/internal/subscription/constants"
	subscriptioninterface "{{.Project.GoModule}}/internal/subscription/interface"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

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

func GetSubscriptionFromContext(c echo.Context) subscriptioninterface.Subscription {
	if subscription := c.Get(subscriptionconstants.ContextKeySubscription); subscription != nil {
		if sub, ok := subscription.(subscriptioninterface.Subscription); ok {
			return sub
		}
	}
	return nil
}

func GetPlanFromContext(c echo.Context) subscriptioninterface.SubscriptionPlan {
	if plan := c.Get(subscriptionconstants.ContextKeyPlan); plan != nil {
		if p, ok := plan.(subscriptioninterface.SubscriptionPlan); ok {
			return p
		}
	}
	return nil
}

func GetUsageLimitsFromContext(c echo.Context) map[string]int64 {
	if limits := c.Get(subscriptionconstants.ContextKeyUsageLimits); limits != nil {
		if l, ok := limits.(map[string]int64); ok {
			return l
		}
	}
	return nil
}

func HasActiveSubscription(c echo.Context) bool {
	subscription := GetSubscriptionFromContext(c)
	if subscription == nil {
		return false
	}
	
	status := subscription.GetStatus()
	return status == subscriptioninterface.StatusActive || status == subscriptioninterface.StatusTrialing
}

func GetPlanType(c echo.Context) subscriptioninterface.PlanType {
	plan := GetPlanFromContext(c)
	if plan == nil {
		return subscriptioninterface.PlanTypeFree
	}
	return plan.GetType()
}

func CheckResourceLimit(c echo.Context, resource string, requestedQuantity int64) (bool, int64) {
	limits := GetUsageLimitsFromContext(c)
	if limits == nil {
		return true, -1
	}
	
	limit, exists := limits[resource]
	if !exists {
		return true, -1
	}
	
	currentUsage := int64(0)
	
	remaining := limit - currentUsage - requestedQuantity
	return remaining >= 0, remaining
}

func SetAccountIDInContext(c echo.Context, accountID uuid.UUID) {
	c.Set("account_id", accountID)
}