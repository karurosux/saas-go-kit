package subscription

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type paymentService struct {
	provider         PaymentProvider
	subscriptionRepo SubscriptionRepository
	planRepo         SubscriptionPlanRepository
}

func NewPaymentService(
	provider PaymentProvider,
	subscriptionRepo SubscriptionRepository,
	planRepo SubscriptionPlanRepository,
) PaymentService {
	return &paymentService{
		provider:         provider,
		subscriptionRepo: subscriptionRepo,
		planRepo:         planRepo,
	}
}

func (s *paymentService) CreateCheckoutSession(ctx context.Context, req *CreateCheckoutRequest) (*CheckoutSession, error) {
	plan, err := s.planRepo.FindByID(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	if plan.StripePriceID == "" {
		return nil, fmt.Errorf("plan does not have a Stripe price ID configured")
	}

	// Check if user already has a subscription to get customer ID
	subscription, _ := s.subscriptionRepo.FindByAccountID(ctx, req.AccountID)
	var customerID string
	if subscription != nil {
		customerID = subscription.StripeCustomerID
	}

	options := CheckoutOptions{
		CustomerID:          customerID,
		PriceID:             plan.StripePriceID,
		SuccessURL:          req.SuccessURL,
		CancelURL:           req.CancelURL,
		TrialPeriodDays:     plan.TrialDays,
		AllowPromotionCodes: req.AllowPromotionCodes,
		Metadata: map[string]string{
			"account_id": req.AccountID.String(),
			"plan_id":    req.PlanID.String(),
		},
	}

	return s.provider.CreateCheckoutSession(ctx, options)
}

func (s *paymentService) HandleWebhookEvent(ctx context.Context, payload []byte, signature string) error {
	event, err := s.provider.ConstructWebhookEvent(payload, signature)
	if err != nil {
		return fmt.Errorf("failed to construct webhook event: %w", err)
	}

	switch event.Type {
	case WebhookCheckoutCompleted:
		return s.handleCheckoutCompleted(ctx, event)
	case WebhookSubscriptionUpdated:
		return s.handleSubscriptionUpdated(ctx, event)
	case WebhookSubscriptionDeleted:
		return s.handleSubscriptionDeleted(ctx, event)
	case WebhookInvoicePaymentSucceeded:
		return s.handleInvoicePaymentSucceeded(ctx, event)
	case WebhookInvoicePaymentFailed:
		return s.handleInvoicePaymentFailed(ctx, event)
	default:
		// Log unknown webhook type but don't fail
		return nil
	}
}

func (s *paymentService) CreateCustomerPortalSession(ctx context.Context, accountID uuid.UUID, returnURL string) (*PortalSession, error) {
	subscription, err := s.subscriptionRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	if subscription.StripeCustomerID == "" {
		return nil, fmt.Errorf("no Stripe customer ID found for this account")
	}

	return s.provider.CreatePortalSession(ctx, subscription.StripeCustomerID, returnURL)
}

func (s *paymentService) GetPaymentMethods(ctx context.Context, accountID uuid.UUID) ([]*PaymentMethod, error) {
	subscription, err := s.subscriptionRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	if subscription.StripeCustomerID == "" {
		return []*PaymentMethod{}, nil
	}

	return s.provider.ListPaymentMethods(ctx, subscription.StripeCustomerID)
}

func (s *paymentService) GetInvoiceHistory(ctx context.Context, accountID uuid.UUID, limit int) ([]*Invoice, error) {
	subscription, err := s.subscriptionRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	if subscription.StripeCustomerID == "" {
		return []*Invoice{}, nil
	}

	return s.provider.ListInvoices(ctx, subscription.StripeCustomerID, limit)
}

// Webhook handlers

func (s *paymentService) handleCheckoutCompleted(ctx context.Context, event *WebhookEvent) error {
	// Extract session data and update subscription
	// This would need to parse the event data properly
	return nil
}

func (s *paymentService) handleSubscriptionUpdated(ctx context.Context, event *WebhookEvent) error {
	// Update local subscription data based on Stripe subscription changes
	return nil
}

func (s *paymentService) handleSubscriptionDeleted(ctx context.Context, event *WebhookEvent) error {
	// Mark local subscription as cancelled
	return nil
}

func (s *paymentService) handleInvoicePaymentSucceeded(ctx context.Context, event *WebhookEvent) error {
	// Handle successful payment, extend subscription period if needed
	return nil
}

func (s *paymentService) handleInvoicePaymentFailed(ctx context.Context, event *WebhookEvent) error {
	// Handle failed payment, potentially mark subscription as past due
	return nil
}