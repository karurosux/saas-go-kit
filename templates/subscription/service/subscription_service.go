package subscriptionservice

import (
	"context"
	"fmt"
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/subscription/constants"
	"{{.Project.GoModule}}/internal/subscription/interface"
	"{{.Project.GoModule}}/internal/subscription/model"
	"github.com/google/uuid"
)

// SubscriptionService implements subscription management
type SubscriptionService struct {
	planRepo        subscriptioninterface.SubscriptionPlanRepository
	subscriptionRepo subscriptioninterface.SubscriptionRepository
	usageRepo       subscriptioninterface.UsageRepository
	paymentProvider subscriptioninterface.PaymentProvider
	webhookSecret   string
	returnURL       string
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	planRepo subscriptioninterface.SubscriptionPlanRepository,
	subscriptionRepo subscriptioninterface.SubscriptionRepository,
	usageRepo subscriptioninterface.UsageRepository,
	paymentProvider subscriptioninterface.PaymentProvider,
	webhookSecret string,
	returnURL string,
) subscriptioninterface.SubscriptionService {
	return &SubscriptionService{
		planRepo:        planRepo,
		subscriptionRepo: subscriptionRepo,
		usageRepo:       usageRepo,
		paymentProvider: paymentProvider,
		webhookSecret:   webhookSecret,
		returnURL:       returnURL,
	}
}

// GetPlans returns all active subscription plans
func (s *SubscriptionService) GetPlans(ctx context.Context) ([]subscriptioninterface.SubscriptionPlan, error) {
	return s.planRepo.GetAll(ctx, true)
}

// GetPlan returns a specific subscription plan
func (s *SubscriptionService) GetPlan(ctx context.Context, planID uuid.UUID) (subscriptioninterface.SubscriptionPlan, error) {
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, core.NotFound(subscriptionconstants.ErrPlanNotFound)
	}
	return plan, nil
}

// CreateSubscription creates a new subscription
func (s *SubscriptionService) CreateSubscription(ctx context.Context, req subscriptioninterface.CreateSubscriptionRequest) (subscriptioninterface.Subscription, error) {
	// Check if already subscribed
	existing, _ := s.subscriptionRepo.GetByAccountID(ctx, req.AccountID)
	if existing != nil && existing.GetStatus() == subscriptioninterface.StatusActive {
		return nil, core.BadRequest(subscriptionconstants.ErrAlreadySubscribed)
	}
	
	// Get plan
	plan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		return nil, core.NotFound(subscriptionconstants.ErrPlanNotFound)
	}
	
	if !plan.GetIsActive() {
		return nil, core.BadRequest(subscriptionconstants.ErrInvalidPlan)
	}
	
	// Get or create Stripe customer
	var customerID string
	if existing != nil && existing.GetStripeCustomerID() != "" {
		customerID = existing.GetStripeCustomerID()
	} else {
		// Create new customer
		metadata := map[string]string{
			"account_id": req.AccountID.String(),
		}
		customerID, err = s.paymentProvider.CreateCustomer(ctx, req.CustomerEmail, metadata)
		if err != nil {
			return nil, core.InternalServerError("failed to create customer")
		}
	}
	
	// Determine price ID based on billing period
	var priceID string
	if req.BillingPeriod == subscriptioninterface.BillingPeriodYearly {
		priceID = plan.GetStripePriceYearlyID()
	} else {
		priceID = plan.GetStripePriceMonthlyID()
	}
	
	// For free plans, create subscription directly
	if plan.GetType() == subscriptioninterface.PlanTypeFree {
		now := time.Now()
		subscription := &subscriptionmodel.Subscription{
			ID:                 uuid.New(),
			AccountID:          req.AccountID,
			PlanID:             plan.GetID(),
			Status:             subscriptioninterface.StatusActive,
			BillingPeriod:      req.BillingPeriod,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:   now.AddDate(0, 1, 0), // 1 month for free plan
			StripeCustomerID:   customerID,
		}
		
		if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
			return nil, core.InternalServerError("failed to create subscription")
		}
		
		return subscription, nil
	}
	
	// For paid plans, create Stripe subscription
	stripeReq := subscriptioninterface.CreateSubscriptionRequest{
		CustomerID:      customerID,
		PriceID:         priceID,
		TrialDays:       plan.GetTrialDays(),
		PaymentMethodID: req.PaymentMethodID,
		Metadata: map[string]string{
			"account_id": req.AccountID.String(),
			"plan_id":    plan.GetID().String(),
		},
	}
	
	providerSub, err := s.paymentProvider.CreateSubscription(ctx, stripeReq)
	if err != nil {
		return nil, core.InternalServerError("failed to create subscription: " + err.Error())
	}
	
	// Create local subscription record
	subscription := &subscriptionmodel.Subscription{
		ID:                   uuid.New(),
		AccountID:            req.AccountID,
		PlanID:               plan.GetID(),
		Status:               s.mapStripeStatus(providerSub.Status),
		BillingPeriod:        req.BillingPeriod,
		CurrentPeriodStart:   providerSub.CurrentPeriodStart,
		CurrentPeriodEnd:     providerSub.CurrentPeriodEnd,
		TrialEndsAt:          providerSub.TrialEnd,
		StripeCustomerID:     customerID,
		StripeSubscriptionID: providerSub.ID,
	}
	
	if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
		// Cancel Stripe subscription if local creation fails
		s.paymentProvider.CancelSubscription(ctx, providerSub.ID, true)
		return nil, core.InternalServerError("failed to create subscription")
	}
	
	return subscription, nil
}

// GetSubscription returns the subscription for an account
func (s *SubscriptionService) GetSubscription(ctx context.Context, accountID uuid.UUID) (subscriptioninterface.Subscription, error) {
	subscription, err := s.subscriptionRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, core.NotFound(subscriptionconstants.ErrSubscriptionNotFound)
	}
	return subscription, nil
}

// UpdateSubscription updates a subscription (change plan)
func (s *SubscriptionService) UpdateSubscription(ctx context.Context, accountID uuid.UUID, req subscriptioninterface.UpdateSubscriptionRequest) (subscriptioninterface.Subscription, error) {
	// Get current subscription
	subscription, err := s.subscriptionRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, core.NotFound(subscriptionconstants.ErrSubscriptionNotFound)
	}
	
	if subscription.GetStatus() != subscriptioninterface.StatusActive {
		return nil, core.BadRequest(subscriptionconstants.ErrSubscriptionNotActive)
	}
	
	// Get new plan
	newPlan, err := s.planRepo.GetByID(ctx, req.NewPlanID)
	if err != nil {
		return nil, core.NotFound(subscriptionconstants.ErrPlanNotFound)
	}
	
	if !newPlan.GetIsActive() {
		return nil, core.BadRequest(subscriptionconstants.ErrInvalidPlan)
	}
	
	// Check if downgrading and if current usage exceeds new limits
	if err := s.checkDowngradeLimits(ctx, accountID, subscription.GetPlan(), newPlan); err != nil {
		return nil, err
	}
	
	// Determine new price ID
	var priceID string
	if req.BillingPeriod == subscriptioninterface.BillingPeriodYearly {
		priceID = newPlan.GetStripePriceYearlyID()
	} else {
		priceID = newPlan.GetStripePriceMonthlyID()
	}
	
	// Update Stripe subscription
	if subscription.GetStripeSubscriptionID() != "" {
		stripeReq := subscriptioninterface.UpdateSubscriptionRequest{
			PriceID: priceID,
			Metadata: map[string]string{
				"plan_id": newPlan.GetID().String(),
			},
		}
		
		providerSub, err := s.paymentProvider.UpdateSubscription(ctx, subscription.GetStripeSubscriptionID(), stripeReq)
		if err != nil {
			return nil, core.InternalServerError("failed to update subscription")
		}
		
		// Update local subscription
		if sub, ok := subscription.(*subscriptionmodel.Subscription); ok {
			sub.PlanID = newPlan.GetID()
			sub.BillingPeriod = req.BillingPeriod
			sub.CurrentPeriodStart = providerSub.CurrentPeriodStart
			sub.CurrentPeriodEnd = providerSub.CurrentPeriodEnd
			sub.Status = s.mapStripeStatus(providerSub.Status)
			
			if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
				return nil, core.InternalServerError("failed to update subscription")
			}
		}
	} else {
		// Handle free plan changes
		if sub, ok := subscription.(*subscriptionmodel.Subscription); ok {
			sub.PlanID = newPlan.GetID()
			sub.BillingPeriod = req.BillingPeriod
			
			if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
				return nil, core.InternalServerError("failed to update subscription")
			}
		}
	}
	
	return subscription, nil
}

// CancelSubscription cancels a subscription
func (s *SubscriptionService) CancelSubscription(ctx context.Context, accountID uuid.UUID, immediately bool) error {
	subscription, err := s.subscriptionRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return core.NotFound(subscriptionconstants.ErrSubscriptionNotFound)
	}
	
	if subscription.GetStatus() == subscriptioninterface.StatusCanceled {
		return core.BadRequest(subscriptionconstants.ErrSubscriptionAlreadyCanceled)
	}
	
	// Cannot cancel free plan
	if subscription.GetPlan() != nil && subscription.GetPlan().GetType() == subscriptioninterface.PlanTypeFree {
		return core.BadRequest(subscriptionconstants.ErrCannotCancelFreePlan)
	}
	
	// Cancel in Stripe
	if subscription.GetStripeSubscriptionID() != "" {
		if err := s.paymentProvider.CancelSubscription(ctx, subscription.GetStripeSubscriptionID(), immediately); err != nil {
			return core.InternalServerError("failed to cancel subscription")
		}
	}
	
	// Update local subscription
	if sub, ok := subscription.(*subscriptionmodel.Subscription); ok {
		now := time.Now()
		sub.SetCanceledAt(&now)
		
		if immediately {
			sub.SetStatus(subscriptioninterface.StatusCanceled)
		} else {
			sub.SetCancelAtPeriodEnd(true)
		}
		
		if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
			return core.InternalServerError("failed to update subscription")
		}
	}
	
	return nil
}

// ReactivateSubscription reactivates a canceled subscription
func (s *SubscriptionService) ReactivateSubscription(ctx context.Context, accountID uuid.UUID) error {
	subscription, err := s.subscriptionRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return core.NotFound(subscriptionconstants.ErrSubscriptionNotFound)
	}
	
	if !subscription.GetCancelAtPeriodEnd() {
		return core.BadRequest("subscription is not scheduled for cancellation")
	}
	
	// Resume in Stripe
	if subscription.GetStripeSubscriptionID() != "" {
		if err := s.paymentProvider.ResumeSubscription(ctx, subscription.GetStripeSubscriptionID()); err != nil {
			return core.InternalServerError("failed to reactivate subscription")
		}
	}
	
	// Update local subscription
	if sub, ok := subscription.(*subscriptionmodel.Subscription); ok {
		sub.SetCancelAtPeriodEnd(false)
		sub.SetCanceledAt(nil)
		
		if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
			return core.InternalServerError("failed to update subscription")
		}
	}
	
	return nil
}

// TrackUsage tracks resource usage
func (s *SubscriptionService) TrackUsage(ctx context.Context, accountID uuid.UUID, resource string, quantity int64) error {
	// Get current period
	period := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	
	// Get or create usage record
	usage, err := s.usageRepo.GetByAccountAndResource(ctx, accountID, resource, period)
	if err != nil {
		// Create new usage record
		usage = &subscriptionmodel.Usage{
			ID:        uuid.New(),
			AccountID: accountID,
			Resource:  resource,
			Quantity:  quantity,
			Period:    period,
		}
		return s.usageRepo.Create(ctx, usage)
	}
	
	// Update existing usage
	return s.usageRepo.IncrementUsage(ctx, accountID, resource, quantity)
}

// GetUsage returns current usage for a resource
func (s *SubscriptionService) GetUsage(ctx context.Context, accountID uuid.UUID, resource string) (int64, error) {
	return s.usageRepo.GetCurrentPeriodUsage(ctx, accountID, resource)
}

// GetUsageReport returns a usage report
func (s *SubscriptionService) GetUsageReport(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) (subscriptioninterface.UsageReport, error) {
	// Get subscription to check limits
	subscription, err := s.subscriptionRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return subscriptioninterface.UsageReport{}, core.NotFound(subscriptionconstants.ErrSubscriptionNotFound)
	}
	
	// Get usage records
	usageRecords, err := s.usageRepo.GetByAccount(ctx, accountID, startDate, endDate)
	if err != nil {
		return subscriptioninterface.UsageReport{}, core.InternalServerError("failed to get usage records")
	}
	
	// Get plan limits
	limits := subscription.GetPlan().GetLimits()
	
	// Build report
	report := subscriptioninterface.UsageReport{
		AccountID: accountID,
		StartDate: startDate,
		EndDate:   endDate,
		Resources: make(map[string]subscriptioninterface.ResourceUsage),
	}
	
	// Aggregate usage by resource
	for _, usage := range usageRecords {
		resource := usage.GetResource()
		if _, exists := report.Resources[resource]; !exists {
			report.Resources[resource] = subscriptioninterface.ResourceUsage{
				Resource: resource,
				Quantity: 0,
				Limit:    limits[resource],
			}
		}
		
		ru := report.Resources[resource]
		ru.Quantity += usage.GetQuantity()
		report.Resources[resource] = ru
	}
	
	return report, nil
}

// CheckLimit checks if a resource is within limits
func (s *SubscriptionService) CheckLimit(ctx context.Context, accountID uuid.UUID, resource string) (bool, int64, error) {
	// Get subscription
	subscription, err := s.subscriptionRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return false, 0, core.NotFound(subscriptionconstants.ErrSubscriptionNotFound)
	}
	
	// Get plan limits
	limits := subscription.GetPlan().GetLimits()
	limit, hasLimit := limits[resource]
	if !hasLimit {
		// No limit for this resource
		return true, -1, nil
	}
	
	// Get current usage
	usage, err := s.usageRepo.GetCurrentPeriodUsage(ctx, accountID, resource)
	if err != nil {
		return false, 0, core.InternalServerError("failed to get usage")
	}
	
	remaining := limit - usage
	return usage < limit, remaining, nil
}

// GetInvoices returns account invoices
func (s *SubscriptionService) GetInvoices(ctx context.Context, accountID uuid.UUID) ([]subscriptioninterface.Invoice, error) {
	// Get subscription
	subscription, err := s.subscriptionRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, core.NotFound(subscriptionconstants.ErrSubscriptionNotFound)
	}
	
	if subscription.GetStripeCustomerID() == "" {
		return []subscriptioninterface.Invoice{}, nil
	}
	
	// Get invoices from Stripe
	providerInvoices, err := s.paymentProvider.ListInvoices(ctx, subscription.GetStripeCustomerID(), 100)
	if err != nil {
		return nil, core.InternalServerError("failed to get invoices")
	}
	
	// Convert to interface
	invoices := make([]subscriptioninterface.Invoice, len(providerInvoices))
	for i, pi := range providerInvoices {
		invoice := &subscriptionmodel.Invoice{
			ID:              pi.ID,
			AccountID:       accountID,
			SubscriptionID:  subscription.GetID(),
			Amount:          pi.Amount,
			Currency:        pi.Currency,
			Status:          pi.Status,
			PaidAt:          pi.PaidAt,
			DueDate:         pi.DueDate,
			StripeInvoiceID: pi.ID,
			PDF:             pi.PDF,
			CreatedAt:       time.Now(),
		}
		invoices[i] = invoice
	}
	
	return invoices, nil
}

// CreateCheckoutSession creates a checkout session
func (s *SubscriptionService) CreateCheckoutSession(ctx context.Context, accountID uuid.UUID, req subscriptioninterface.CheckoutRequest) (string, error) {
	// Get plan
	plan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		return "", core.NotFound(subscriptionconstants.ErrPlanNotFound)
	}
	
	// Check if already subscribed
	existing, _ := s.subscriptionRepo.GetByAccountID(ctx, accountID)
	
	// Get or create customer ID
	var customerID string
	if existing != nil && existing.GetStripeCustomerID() != "" {
		customerID = existing.GetStripeCustomerID()
	}
	
	// Determine price ID
	var priceID string
	if req.BillingPeriod == subscriptioninterface.BillingPeriodYearly {
		priceID = plan.GetStripePriceYearlyID()
	} else {
		priceID = plan.GetStripePriceMonthlyID()
	}
	
	checkoutReq := subscriptioninterface.CheckoutSessionRequest{
		PriceID:    priceID,
		Quantity:   1,
		SuccessURL: req.SuccessURL,
		CancelURL:  req.CancelURL,
		CustomerID: customerID,
		Metadata: map[string]string{
			"account_id": accountID.String(),
			"plan_id":    plan.GetID().String(),
		},
		AllowPromoCodes: true,
	}
	
	session, err := s.paymentProvider.CreateCheckoutSession(ctx, checkoutReq)
	if err != nil {
		return "", core.InternalServerError("failed to create checkout session")
	}
	
	return session.URL, nil
}

// CreateBillingPortalSession creates a billing portal session
func (s *SubscriptionService) CreateBillingPortalSession(ctx context.Context, accountID uuid.UUID) (string, error) {
	// Get subscription
	subscription, err := s.subscriptionRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return "", core.NotFound(subscriptionconstants.ErrSubscriptionNotFound)
	}
	
	if subscription.GetStripeCustomerID() == "" {
		return "", core.BadRequest("no payment method on file")
	}
	
	return s.paymentProvider.CreateBillingPortalSession(ctx, subscription.GetStripeCustomerID(), s.returnURL)
}

// HandleStripeWebhook handles Stripe webhook events
func (s *SubscriptionService) HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	// Webhook handling would be implemented here
	// This is a placeholder for the actual implementation
	return nil
}

// Helper methods

func (s *SubscriptionService) mapStripeStatus(status string) subscriptioninterface.SubscriptionStatus {
	switch status {
	case "active":
		return subscriptioninterface.StatusActive
	case "trialing":
		return subscriptioninterface.StatusTrialing
	case "past_due":
		return subscriptioninterface.StatusPastDue
	case "canceled":
		return subscriptioninterface.StatusCanceled
	case "paused":
		return subscriptioninterface.StatusPaused
	default:
		return subscriptioninterface.StatusActive
	}
}

func (s *SubscriptionService) checkDowngradeLimits(ctx context.Context, accountID uuid.UUID, currentPlan, newPlan subscriptioninterface.SubscriptionPlan) error {
	currentLimits := currentPlan.GetLimits()
	newLimits := newPlan.GetLimits()
	
	for resource, newLimit := range newLimits {
		currentLimit, exists := currentLimits[resource]
		if !exists || newLimit >= currentLimit {
			continue // Not a downgrade for this resource
		}
		
		// Check current usage
		usage, err := s.usageRepo.GetCurrentPeriodUsage(ctx, accountID, resource)
		if err != nil {
			return core.InternalServerError("failed to check usage")
		}
		
		if usage > newLimit {
			return core.BadRequest(fmt.Sprintf("%s: current usage (%d) exceeds new plan limit (%d)", 
				subscriptionconstants.ErrCannotDowngrade, usage, newLimit))
		}
	}
	
	return nil
}

// Additional types for internal use
type CreateSubscriptionRequest struct {
	AccountID       uuid.UUID
	PlanID          uuid.UUID
	BillingPeriod   subscriptioninterface.BillingPeriod
	PaymentMethodID string
	CustomerEmail   string
}

type UpdateSubscriptionRequest struct {
	NewPlanID     uuid.UUID
	BillingPeriod subscriptioninterface.BillingPeriod
}