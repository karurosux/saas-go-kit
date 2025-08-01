package subscriptioninterface

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// PlanType represents subscription plan types
type PlanType string

const (
	PlanTypeFree       PlanType = "free"
	PlanTypeStarter    PlanType = "starter"
	PlanTypePro        PlanType = "pro"
	PlanTypeEnterprise PlanType = "enterprise"
)

// BillingPeriod represents billing periods
type BillingPeriod string

const (
	BillingPeriodMonthly BillingPeriod = "monthly"
	BillingPeriodYearly  BillingPeriod = "yearly"
)

// SubscriptionStatus represents subscription status
type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusTrialing  SubscriptionStatus = "trialing"
	StatusPastDue   SubscriptionStatus = "past_due"
	StatusCanceled  SubscriptionStatus = "canceled"
	StatusPaused    SubscriptionStatus = "paused"
)

// SubscriptionPlan represents a subscription plan
type SubscriptionPlan interface {
	GetID() uuid.UUID
	GetName() string
	GetType() PlanType
	GetPriceMonthly() int64 // in cents
	GetPriceYearly() int64  // in cents
	GetFeatures() map[string]interface{}
	GetLimits() map[string]int64
	GetStripeProductID() string
	GetStripePriceMonthlyID() string
	GetStripePriceYearlyID() string
	GetTrialDays() int
	GetIsActive() bool
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// Subscription represents an account subscription
type Subscription interface {
	GetID() uuid.UUID
	GetAccountID() uuid.UUID
	GetPlanID() uuid.UUID
	GetPlan() SubscriptionPlan
	GetStatus() SubscriptionStatus
	GetBillingPeriod() BillingPeriod
	GetCurrentPeriodStart() time.Time
	GetCurrentPeriodEnd() time.Time
	GetTrialEndsAt() *time.Time
	GetCanceledAt() *time.Time
	GetCancelAtPeriodEnd() bool
	GetStripeCustomerID() string
	GetStripeSubscriptionID() string
	GetMetadata() map[string]string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetStatus(status SubscriptionStatus)
	SetCancelAtPeriodEnd(cancel bool)
	SetCanceledAt(canceledAt *time.Time)
}

// Usage represents resource usage
type Usage interface {
	GetID() uuid.UUID
	GetAccountID() uuid.UUID
	GetResource() string
	GetQuantity() int64
	GetPeriod() time.Time
	GetMetadata() map[string]interface{}
	GetCreatedAt() time.Time
	IncrementQuantity(amount int64)
}

// Invoice represents a subscription invoice
type Invoice interface {
	GetID() string
	GetAccountID() uuid.UUID
	GetSubscriptionID() uuid.UUID
	GetAmount() int64
	GetCurrency() string
	GetStatus() string
	GetPaidAt() *time.Time
	GetDueDate() time.Time
	GetStripeInvoiceID() string
	GetPDF() string
	GetCreatedAt() time.Time
}

// SubscriptionPlanRepository defines plan data access
type SubscriptionPlanRepository interface {
	Create(ctx context.Context, plan SubscriptionPlan) error
	GetByID(ctx context.Context, id uuid.UUID) (SubscriptionPlan, error)
	GetByType(ctx context.Context, planType PlanType) (SubscriptionPlan, error)
	GetAll(ctx context.Context, activeOnly bool) ([]SubscriptionPlan, error)
	Update(ctx context.Context, plan SubscriptionPlan) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SubscriptionRepository defines subscription data access
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (Subscription, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) (Subscription, error)
	GetByStripeSubscriptionID(ctx context.Context, stripeSubID string) (Subscription, error)
	Update(ctx context.Context, subscription Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetExpiringTrials(ctx context.Context, daysAhead int) ([]Subscription, error)
	GetPastDue(ctx context.Context) ([]Subscription, error)
}

// UsageRepository defines usage data access
type UsageRepository interface {
	Create(ctx context.Context, usage Usage) error
	GetByID(ctx context.Context, id uuid.UUID) (Usage, error)
	GetByAccountAndResource(ctx context.Context, accountID uuid.UUID, resource string, period time.Time) (Usage, error)
	GetByAccount(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) ([]Usage, error)
	Update(ctx context.Context, usage Usage) error
	IncrementUsage(ctx context.Context, accountID uuid.UUID, resource string, amount int64) error
	GetCurrentPeriodUsage(ctx context.Context, accountID uuid.UUID, resource string) (int64, error)
}

// PaymentProvider defines payment provider interface
type PaymentProvider interface {
	// Customer management
	CreateCustomer(ctx context.Context, email string, metadata map[string]string) (string, error)
	UpdateCustomer(ctx context.Context, customerID string, metadata map[string]string) error
	DeleteCustomer(ctx context.Context, customerID string) error
	
	// Subscription management
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*ProviderSubscription, error)
	UpdateSubscription(ctx context.Context, subscriptionID string, req UpdateSubscriptionRequest) (*ProviderSubscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string, immediately bool) error
	ResumeSubscription(ctx context.Context, subscriptionID string) error
	
	// Payment method management
	AttachPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) error
	SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error
	ListPaymentMethods(ctx context.Context, customerID string) ([]PaymentMethod, error)
	
	// Billing
	CreateInvoice(ctx context.Context, customerID string, items []InvoiceItem) (*ProviderInvoice, error)
	GetInvoice(ctx context.Context, invoiceID string) (*ProviderInvoice, error)
	ListInvoices(ctx context.Context, customerID string, limit int) ([]ProviderInvoice, error)
	
	// Checkout
	CreateCheckoutSession(ctx context.Context, req CheckoutSessionRequest) (*CheckoutSession, error)
	CreateBillingPortalSession(ctx context.Context, customerID, returnURL string) (string, error)
}

// SubscriptionService defines subscription business logic
type SubscriptionService interface {
	// Plan management
	GetPlans(ctx context.Context) ([]SubscriptionPlan, error)
	GetPlan(ctx context.Context, planID uuid.UUID) (SubscriptionPlan, error)
	
	// Subscription management
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (Subscription, error)
	GetSubscription(ctx context.Context, accountID uuid.UUID) (Subscription, error)
	UpdateSubscription(ctx context.Context, accountID uuid.UUID, req UpdateSubscriptionRequest) (Subscription, error)
	CancelSubscription(ctx context.Context, accountID uuid.UUID, immediately bool) error
	ReactivateSubscription(ctx context.Context, accountID uuid.UUID) error
	
	// Usage tracking
	TrackUsage(ctx context.Context, accountID uuid.UUID, resource string, quantity int64) error
	GetUsage(ctx context.Context, accountID uuid.UUID, resource string) (int64, error)
	GetUsageReport(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (UsageReport, error)
	CheckLimit(ctx context.Context, accountID uuid.UUID, resource string) (bool, int64, error)
	
	// Billing
	GetInvoices(ctx context.Context, accountID uuid.UUID) ([]Invoice, error)
	CreateCheckoutSession(ctx context.Context, accountID uuid.UUID, req CheckoutRequest) (string, error)
	CreateBillingPortalSession(ctx context.Context, accountID uuid.UUID) (string, error)
	
	// Webhooks
	HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error
}

// Request/Response types

type CreateSubscriptionRequest struct {
	CustomerID      string
	PriceID         string
	TrialDays       int
	Metadata        map[string]string
	PaymentMethodID string
}

type UpdateSubscriptionRequest struct {
	PriceID  string
	Quantity int64
	Metadata map[string]string
}

type CheckoutSessionRequest struct {
	PriceID         string
	Quantity        int64
	SuccessURL      string
	CancelURL       string
	CustomerID      string
	Metadata        map[string]string
	AllowPromoCodes bool
}

type CheckoutRequest struct {
	PlanID         uuid.UUID
	BillingPeriod  BillingPeriod
	SuccessURL     string
	CancelURL      string
}

type InvoiceItem struct {
	PriceID     string
	Quantity    int64
	Description string
}

type PaymentMethod struct {
	ID        string
	Type      string
	Last4     string
	Brand     string
	ExpMonth  int
	ExpYear   int
	IsDefault bool
}

type ProviderSubscription struct {
	ID                   string
	CustomerID           string
	Status               string
	CurrentPeriodStart   time.Time
	CurrentPeriodEnd     time.Time
	TrialEnd             *time.Time
	CancelAt             *time.Time
	CancelAtPeriodEnd    bool
	DefaultPaymentMethod string
	Items                []SubscriptionItem
}

type SubscriptionItem struct {
	ID       string
	PriceID  string
	Quantity int64
}

type ProviderInvoice struct {
	ID         string
	CustomerID string
	Amount     int64
	Currency   string
	Status     string
	PaidAt     *time.Time
	DueDate    time.Time
	PDF        string
}

type CheckoutSession struct {
	ID         string
	URL        string
	CustomerID string
	Status     string
	ExpiresAt  time.Time
}

type UsageReport struct {
	AccountID   uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
	Resources   map[string]ResourceUsage
	TotalCost   int64
}

type ResourceUsage struct {
	Resource string
	Quantity int64
	Limit    int64
	Cost     int64
}