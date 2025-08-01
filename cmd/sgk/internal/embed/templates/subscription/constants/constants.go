package subscriptionconstants

import "time"

// Default values
const (
	// DefaultTrialDays is the default trial period
	DefaultTrialDays = 14
	
	// DefaultGracePeriodDays is the grace period after subscription ends
	DefaultGracePeriodDays = 7
	
	// DefaultMaxRetries for failed payments
	DefaultMaxRetries = 3
	
	// DefaultWebhookTimeout for webhook processing
	DefaultWebhookTimeout = 30 * time.Second
)

// Resource names for usage tracking
const (
	ResourceAPIRequests   = "api_requests"
	ResourceStorage       = "storage_gb"
	ResourceBandwidth     = "bandwidth_gb"
	ResourceTeamMembers   = "team_members"
	ResourceProjects      = "projects"
	ResourceCustomDomains = "custom_domains"
)

// Webhook events
const (
	WebhookSubscriptionCreated         = "customer.subscription.created"
	WebhookSubscriptionUpdated         = "customer.subscription.updated"
	WebhookSubscriptionDeleted         = "customer.subscription.deleted"
	WebhookSubscriptionTrialWillEnd    = "customer.subscription.trial_will_end"
	WebhookInvoicePaymentSucceeded     = "invoice.payment_succeeded"
	WebhookInvoicePaymentFailed        = "invoice.payment_failed"
	WebhookPaymentMethodAttached       = "payment_method.attached"
	WebhookPaymentMethodDetached       = "payment_method.detached"
	WebhookCheckoutSessionCompleted    = "checkout.session.completed"
	WebhookCustomerSubscriptionUpdated = "customer.subscription.updated"
)

// Context keys
const (
	// ContextKeySubscription stores the subscription in context
	ContextKeySubscription = "subscription"
	
	// ContextKeyPlan stores the plan in context
	ContextKeyPlan = "subscription_plan"
	
	// ContextKeyUsageLimits stores usage limits in context
	ContextKeyUsageLimits = "usage_limits"
)