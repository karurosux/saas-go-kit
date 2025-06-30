package subscription

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *Subscription) error
	FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*Subscription, error)
	FindByAccountID(ctx context.Context, accountID uuid.UUID) (*Subscription, error)
	FindByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*Subscription, error)
	Update(ctx context.Context, subscription *Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SubscriptionPlanRepository defines the interface for subscription plan data access
type SubscriptionPlanRepository interface {
	FindAll(ctx context.Context) ([]SubscriptionPlan, error)
	FindAllIncludingHidden(ctx context.Context) ([]SubscriptionPlan, error)
	FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*SubscriptionPlan, error)
	FindByCode(ctx context.Context, code string) (*SubscriptionPlan, error)
	Create(ctx context.Context, plan *SubscriptionPlan) error
	Update(ctx context.Context, plan *SubscriptionPlan) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UsageRepository defines the interface for usage tracking data access
type UsageRepository interface {
	Create(ctx context.Context, usage *SubscriptionUsage) error
	Update(ctx context.Context, usage *SubscriptionUsage) error
	FindByID(ctx context.Context, id uuid.UUID) (*SubscriptionUsage, error)
	FindBySubscriptionAndPeriod(ctx context.Context, subscriptionID uuid.UUID, periodStart, periodEnd time.Time) (*SubscriptionUsage, error)
	FindBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]*SubscriptionUsage, error)
	CreateEvent(ctx context.Context, event *UsageEvent) error
	FindEventsBySubscription(ctx context.Context, subscriptionID uuid.UUID, limit int) ([]*UsageEvent, error)
}

// SubscriptionService defines the interface for subscription business logic
type SubscriptionService interface {
	GetUserSubscription(ctx context.Context, accountID uuid.UUID) (*Subscription, error)
	GetAvailablePlans(ctx context.Context) ([]SubscriptionPlan, error)
	GetAllPlans(ctx context.Context) ([]SubscriptionPlan, error)
	CreateSubscription(ctx context.Context, req *CreateSubscriptionRequest) (*Subscription, error)
	AssignCustomPlan(ctx context.Context, accountID uuid.UUID, planCode string) error
	CancelSubscription(ctx context.Context, accountID uuid.UUID) error
	CanUserAccessResource(ctx context.Context, accountID uuid.UUID, resourceType string) (*PermissionResponse, error)
}

// UsageService defines the interface for usage tracking business logic
type UsageService interface {
	TrackUsage(ctx context.Context, subscriptionID uuid.UUID, resourceType string, delta int) error
	RecordUsageEvent(ctx context.Context, event *UsageEvent) error
	CanAddResource(ctx context.Context, subscriptionID uuid.UUID, resourceType string) (bool, string, error)
	GetCurrentUsage(ctx context.Context, subscriptionID uuid.UUID) (*SubscriptionUsage, error)
	GetUsageForPeriod(ctx context.Context, subscriptionID uuid.UUID, start, end time.Time) (*SubscriptionUsage, error)
	InitializeUsagePeriod(ctx context.Context, subscriptionID uuid.UUID, periodStart, periodEnd time.Time) error
	ResetMonthlyUsage(ctx context.Context) error
}

// PaymentProvider defines the interface for payment processing
type PaymentProvider interface {
	Initialize(config PaymentConfig) error
	GetProviderName() string
	CreateCheckoutSession(ctx context.Context, options CheckoutOptions) (*CheckoutSession, error)
	GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSession, error)
	CreateCustomer(ctx context.Context, customer CustomerInfo) (*Customer, error)
	UpdateCustomer(ctx context.Context, customerID string, updates CustomerInfo) (*Customer, error)
	GetCustomer(ctx context.Context, customerID string) (*Customer, error)
	CreateSubscription(ctx context.Context, options SubscriptionOptions) (*PaymentSubscription, error)
	UpdateSubscription(ctx context.Context, subscriptionID string, options UpdateSubscriptionOptions) (*PaymentSubscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string, immediately bool) error
	GetSubscription(ctx context.Context, subscriptionID string) (*PaymentSubscription, error)
	AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) error
	ListPaymentMethods(ctx context.Context, customerID string) ([]*PaymentMethod, error)
	SetDefaultPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error
	CreatePortalSession(ctx context.Context, customerID string, returnURL string) (*PortalSession, error)
	ConstructWebhookEvent(payload []byte, signature string) (*WebhookEvent, error)
	HandleWebhookEvent(ctx context.Context, event *WebhookEvent) error
	GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error)
	ListInvoices(ctx context.Context, customerID string, limit int) ([]*Invoice, error)
}

// PaymentService defines the interface for payment-related business logic
type PaymentService interface {
	CreateCheckoutSession(ctx context.Context, req *CreateCheckoutRequest) (*CheckoutSession, error)
	HandleWebhookEvent(ctx context.Context, payload []byte, signature string) error
	CreateCustomerPortalSession(ctx context.Context, accountID uuid.UUID, returnURL string) (*PortalSession, error)
	GetPaymentMethods(ctx context.Context, accountID uuid.UUID) ([]*PaymentMethod, error)
	GetInvoiceHistory(ctx context.Context, accountID uuid.UUID, limit int) ([]*Invoice, error)
}

// DTOs and Request/Response types

type CreateSubscriptionRequest struct {
	AccountID uuid.UUID `json:"account_id"`
	PlanID    uuid.UUID `json:"plan_id"`
}

type PermissionResponse struct {
	CanCreate          bool   `json:"can_create"`
	Reason             string `json:"reason"`
	CurrentCount       int    `json:"current_count"`
	MaxAllowed         int    `json:"max_allowed"`
	SubscriptionStatus string `json:"subscription_status"`
}

type CreateCheckoutRequest struct {
	AccountID           uuid.UUID `json:"account_id"`
	PlanID              uuid.UUID `json:"plan_id"`
	SuccessURL          string    `json:"success_url"`
	CancelURL           string    `json:"cancel_url"`
	AllowPromotionCodes bool      `json:"allow_promotion_codes"`
}

// Payment Provider Types

type PaymentConfig struct {
	SecretKey      string
	PublishableKey string
	WebhookSecret  string
	Extra          map[string]interface{}
}

type CheckoutOptions struct {
	CustomerID          string
	PriceID             string
	SuccessURL          string
	CancelURL           string
	TrialPeriodDays     int
	AllowPromotionCodes bool
	Metadata            map[string]string
}

type CheckoutSession struct {
	ID                string
	URL               string
	Status            string
	CustomerID        string
	SubscriptionID    string
	PaymentIntentID   string
	AmountTotal       int64
	Currency          string
	ExpiresAt         time.Time
	Metadata          map[string]string
}

type CustomerInfo struct {
	Email       string
	Name        string
	Phone       string
	Description string
	Metadata    map[string]string
}

type Customer struct {
	ID               string
	Email            string
	Name             string
	Phone            string
	Description      string
	DefaultPaymentID string
	Metadata         map[string]string
	CreatedAt        time.Time
}

type SubscriptionOptions struct {
	CustomerID       string
	PriceID          string
	TrialPeriodDays  int
	DefaultPaymentID string
	Metadata         map[string]string
}

type UpdateSubscriptionOptions struct {
	PriceID           string
	ProrationBehavior string
	CancelAtPeriodEnd bool
	DefaultPaymentID  string
	Metadata          map[string]string
}

type PaymentSubscription struct {
	ID                   string
	CustomerID           string
	Status               string
	CurrentPeriodStart   time.Time
	CurrentPeriodEnd     time.Time
	CancelAt             *time.Time
	CanceledAt           *time.Time
	EndedAt              *time.Time
	TrialStart           *time.Time
	TrialEnd             *time.Time
	Items                []SubscriptionItem
	DefaultPaymentID     string
	LatestInvoiceID      string
	Metadata             map[string]string
}

type SubscriptionItem struct {
	ID       string
	PriceID  string
	Quantity int64
}

type PaymentMethod struct {
	ID        string
	Type      string
	IsDefault bool
	Card      *CardDetails
	CreatedAt time.Time
}

type CardDetails struct {
	Brand    string
	Last4    string
	ExpMonth int
	ExpYear  int
	Country  string
}

type PortalSession struct {
	ID        string
	URL       string
	ReturnURL string
	CreatedAt time.Time
}

type WebhookEvent struct {
	ID        string
	Type      string
	Data      interface{}
	CreatedAt time.Time
}

const (
	WebhookCheckoutCompleted       = "checkout.session.completed"
	WebhookSubscriptionCreated     = "customer.subscription.created"
	WebhookSubscriptionUpdated     = "customer.subscription.updated"
	WebhookSubscriptionDeleted     = "customer.subscription.deleted"
	WebhookInvoicePaymentSucceeded = "invoice.payment_succeeded"
	WebhookInvoicePaymentFailed    = "invoice.payment_failed"
	WebhookPaymentMethodAttached   = "payment_method.attached"
	WebhookPaymentMethodDetached   = "payment_method.detached"
)

type Invoice struct {
	ID                string
	Number            string
	CustomerID        string
	SubscriptionID    string
	Status            string
	AmountDue         int64
	AmountPaid        int64
	Currency          string
	InvoicePDF        string
	HostedInvoiceURL  string
	CreatedAt         time.Time
	PaidAt            *time.Time
	DueDate           *time.Time
	Lines             []InvoiceLine
}

type InvoiceLine struct {
	Description string
	Quantity    int64
	UnitAmount  int64
	Amount      int64
}