package subscriptionconstants

// Error messages
const (
	ErrPlanNotFound              = "subscription plan not found"
	ErrSubscriptionNotFound      = "subscription not found"
	ErrAlreadySubscribed         = "account already has an active subscription"
	ErrInvalidPlan               = "invalid subscription plan"
	ErrPaymentMethodRequired     = "payment method required for paid plans"
	ErrUsageLimitExceeded        = "usage limit exceeded for resource"
	ErrCannotDowngrade           = "cannot downgrade to a plan with lower limits while exceeding them"
	ErrTrialNotAvailable         = "trial not available for this plan"
	ErrInvalidBillingPeriod      = "invalid billing period"
	ErrSubscriptionNotActive     = "subscription is not active"
	ErrCannotCancelFreePlan      = "cannot cancel free plan subscription"
	ErrWebhookSignatureInvalid   = "webhook signature validation failed"
	ErrCustomerNotFound          = "stripe customer not found"
	ErrPaymentFailed             = "payment failed"
	ErrInvalidWebhookEvent       = "invalid webhook event"
	ErrSubscriptionAlreadyCanceled = "subscription is already canceled"
)