package subscription

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/errors-go"
	"github.com/karurosux/saas-go-kit/response-go"
	"github.com/labstack/echo/v4"
)

// MiddlewareConfig holds configuration for subscription middlewares
type MiddlewareConfig struct {
	SubscriptionService SubscriptionService
	UsageService        UsageService
	SkipPaths          []string // Paths to skip middleware checks
}

// RequireFeatureFlag creates middleware that checks if user's plan has a specific feature flag
func RequireFeatureFlag(config MiddlewareConfig, featureFlag string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if path is in skip list
			for _, path := range config.SkipPaths {
				if c.Path() == path {
					return next(c)
				}
			}

			accountIDStr := c.Get("account_id")
			if accountIDStr == nil {
				return response.Error(c, errors.Unauthorized("Authentication required"))
			}

			accountID, err := uuid.Parse(accountIDStr.(string))
			if err != nil {
				return response.Error(c, errors.BadRequest("Invalid account ID"))
			}

			subscription, err := config.SubscriptionService.GetUserSubscription(c.Request().Context(), accountID)
			if err != nil {
				return response.Error(c, errors.Forbidden("No active subscription"))
			}

			if !subscription.IsActive() {
				return response.Error(c, errors.Forbidden("Subscription is not active"))
			}

			if !subscription.Plan.Features.GetFlag(featureFlag) {
				return response.Error(c, errors.Forbidden("Feature not available in your plan"))
			}

			return next(c)
		}
	}
}

// RequireResourceLimit creates middleware that checks if user can add a resource based on plan limits
func RequireResourceLimit(config MiddlewareConfig, resourceType string, limitKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if path is in skip list
			for _, path := range config.SkipPaths {
				if c.Path() == path {
					return next(c)
				}
			}

			// Only check limits for POST/PUT requests (creation/updates)
			if c.Request().Method != http.MethodPost && c.Request().Method != http.MethodPut {
				return next(c)
			}

			accountIDStr := c.Get("account_id")
			if accountIDStr == nil {
				return response.Error(c, errors.Unauthorized("Authentication required"))
			}

			accountID, err := uuid.Parse(accountIDStr.(string))
			if err != nil {
				return response.Error(c, errors.BadRequest("Invalid account ID"))
			}

			permission, err := config.SubscriptionService.CanUserAccessResourceWithLimitKey(
				c.Request().Context(), 
				accountID, 
				resourceType, 
				limitKey,
			)
			if err != nil {
				return response.Error(c, errors.Internal("Failed to check permissions"))
			}

			if !permission.CanCreate {
				return response.Error(c, errors.Forbidden(permission.Reason))
			}

			// Store permission info in context for potential use in handlers
			c.Set("subscription_permission", permission)

			return next(c)
		}
	}
}

// RequireActiveSubscription creates middleware that ensures user has an active subscription
func RequireActiveSubscription(config MiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if path is in skip list
			for _, path := range config.SkipPaths {
				if c.Path() == path {
					return next(c)
				}
			}

			accountIDStr := c.Get("account_id")
			if accountIDStr == nil {
				return response.Error(c, errors.Unauthorized("Authentication required"))
			}

			accountID, err := uuid.Parse(accountIDStr.(string))
			if err != nil {
				return response.Error(c, errors.BadRequest("Invalid account ID"))
			}

			subscription, err := config.SubscriptionService.GetUserSubscription(c.Request().Context(), accountID)
			if err != nil {
				return response.Error(c, errors.Forbidden("No subscription found"))
			}

			if !subscription.IsActive() {
				return response.Error(c, errors.Forbidden("Subscription is not active"))
			}

			// Store subscription in context for potential use in handlers
			c.Set("user_subscription", subscription)

			return next(c)
		}
	}
}

// RequirePlanTier creates middleware that checks if user's plan meets minimum tier requirements
func RequirePlanTier(config MiddlewareConfig, allowedPlanCodes []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if path is in skip list
			for _, path := range config.SkipPaths {
				if c.Path() == path {
					return next(c)
				}
			}

			accountIDStr := c.Get("account_id")
			if accountIDStr == nil {
				return response.Error(c, errors.Unauthorized("Authentication required"))
			}

			accountID, err := uuid.Parse(accountIDStr.(string))
			if err != nil {
				return response.Error(c, errors.BadRequest("Invalid account ID"))
			}

			subscription, err := config.SubscriptionService.GetUserSubscription(c.Request().Context(), accountID)
			if err != nil {
				return response.Error(c, errors.Forbidden("No active subscription"))
			}

			if !subscription.IsActive() {
				return response.Error(c, errors.Forbidden("Subscription is not active"))
			}

			// Check if plan code is in allowed list
			planAllowed := false
			for _, allowedCode := range allowedPlanCodes {
				if subscription.Plan.Code == allowedCode {
					planAllowed = true
					break
				}
			}

			if !planAllowed {
				return response.Error(c, errors.Forbidden("Plan tier not sufficient for this feature"))
			}

			return next(c)
		}
	}
}

// UsageTrackingMiddleware creates middleware that automatically tracks resource usage
func UsageTrackingMiddleware(config MiddlewareConfig, resourceType string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Execute the handler first
			err := next(c)
			
			// If handler succeeded and this was a POST request, track usage
			if err == nil && c.Request().Method == http.MethodPost {
				accountIDStr := c.Get("account_id")
				if accountIDStr != nil {
					if accountID, parseErr := uuid.Parse(accountIDStr.(string)); parseErr == nil {
						if subscription, subErr := config.SubscriptionService.GetUserSubscription(c.Request().Context(), accountID); subErr == nil {
							// Track usage asynchronously to avoid impacting response time
							go func() {
								config.UsageService.TrackUsage(c.Request().Context(), subscription.ID, resourceType, 1)
							}()
						}
					}
				}
			}

			return err
		}
	}
}