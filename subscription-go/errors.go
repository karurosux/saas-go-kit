package subscription

import "errors"

// Repository errors
var (
	ErrSubscriptionNotFound     = errors.New("subscription not found")
	ErrSubscriptionPlanNotFound = errors.New("subscription plan not found")
	ErrUsageNotFound            = errors.New("usage record not found")
	ErrUsageEventNotFound       = errors.New("usage event not found")
)

// Service errors
var (
	ErrInvalidPlan               = errors.New("invalid subscription plan")
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists for account")
	ErrInactiveSubscription      = errors.New("subscription is not active")
	ErrLimitExceeded             = errors.New("subscription limit exceeded")
	ErrInvalidUsageData          = errors.New("invalid usage data")
)

// Payment errors
var (
	ErrPaymentFailed           = errors.New("payment failed")
	ErrInvalidPaymentProvider  = errors.New("invalid payment provider")
	ErrWebhookValidationFailed = errors.New("webhook validation failed")
	ErrCustomerNotFound        = errors.New("customer not found")
	ErrCheckoutSessionExpired  = errors.New("checkout session expired")
)