package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/invoice"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	portalSession "github.com/stripe/stripe-go/v76/billingportal/session"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"
)

type StripeProvider struct {
	secretKey     string
	webhookSecret string
}

func NewStripeProvider() *StripeProvider {
	return &StripeProvider{}
}

func (p *StripeProvider) Initialize(config PaymentConfig) error {
	p.secretKey = config.SecretKey
	p.webhookSecret = config.WebhookSecret
	stripe.Key = p.secretKey
	return nil
}

func (p *StripeProvider) GetProviderName() string {
	return "stripe"
}

func (p *StripeProvider) CreateCheckoutSession(ctx context.Context, options CheckoutOptions) (*CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(options.SuccessURL),
		CancelURL:  stripe.String(options.CancelURL),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(options.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
	}

	if options.CustomerID != "" {
		params.Customer = stripe.String(options.CustomerID)
	}

	if options.TrialPeriodDays > 0 {
		params.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
			TrialPeriodDays: stripe.Int64(int64(options.TrialPeriodDays)),
		}
	}

	if options.AllowPromotionCodes {
		params.AllowPromotionCodes = stripe.Bool(true)
	}

	if len(options.Metadata) > 0 {
		params.Metadata = options.Metadata
	}

	s, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return &CheckoutSession{
		ID:              s.ID,
		URL:             s.URL,
		Status:          string(s.Status),
		CustomerID:      getStringID(s.Customer),
		SubscriptionID:  getStringID(s.Subscription),
		PaymentIntentID: getStringID(s.PaymentIntent),
		AmountTotal:     s.AmountTotal,
		Currency:        string(s.Currency),
		ExpiresAt:       toTime(s.ExpiresAt),
		Metadata:        s.Metadata,
	}, nil
}

func (p *StripeProvider) GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSession, error) {
	s, err := session.Get(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkout session: %w", err)
	}

	return &CheckoutSession{
		ID:              s.ID,
		URL:             s.URL,
		Status:          string(s.Status),
		CustomerID:      getStringID(s.Customer),
		SubscriptionID:  getStringID(s.Subscription),
		PaymentIntentID: getStringID(s.PaymentIntent),
		AmountTotal:     s.AmountTotal,
		Currency:        string(s.Currency),
		ExpiresAt:       toTime(s.ExpiresAt),
		Metadata:        s.Metadata,
	}, nil
}

func (p *StripeProvider) CreateCustomer(ctx context.Context, info CustomerInfo) (*Customer, error) {
	params := &stripe.CustomerParams{
		Email:       stripe.String(info.Email),
		Name:        stripe.String(info.Name),
		Phone:       stripe.String(info.Phone),
		Description: stripe.String(info.Description),
	}

	if len(info.Metadata) > 0 {
		params.Metadata = info.Metadata
	}

	c, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return &Customer{
		ID:               c.ID,
		Email:            c.Email,
		Name:             c.Name,
		Phone:            c.Phone,
		Description:      c.Description,
		DefaultPaymentID: getStringID(c.InvoiceSettings.DefaultPaymentMethod),
		Metadata:         c.Metadata,
		CreatedAt:        toTime(c.Created),
	}, nil
}

func (p *StripeProvider) UpdateCustomer(ctx context.Context, customerID string, updates CustomerInfo) (*Customer, error) {
	params := &stripe.CustomerParams{}

	if updates.Email != "" {
		params.Email = stripe.String(updates.Email)
	}
	if updates.Name != "" {
		params.Name = stripe.String(updates.Name)
	}
	if updates.Phone != "" {
		params.Phone = stripe.String(updates.Phone)
	}
	if updates.Description != "" {
		params.Description = stripe.String(updates.Description)
	}
	if len(updates.Metadata) > 0 {
		params.Metadata = updates.Metadata
	}

	c, err := customer.Update(customerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return &Customer{
		ID:               c.ID,
		Email:            c.Email,
		Name:             c.Name,
		Phone:            c.Phone,
		Description:      c.Description,
		DefaultPaymentID: getStringID(c.InvoiceSettings.DefaultPaymentMethod),
		Metadata:         c.Metadata,
		CreatedAt:        toTime(c.Created),
	}, nil
}

func (p *StripeProvider) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	c, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &Customer{
		ID:               c.ID,
		Email:            c.Email,
		Name:             c.Name,
		Phone:            c.Phone,
		Description:      c.Description,
		DefaultPaymentID: getStringID(c.InvoiceSettings.DefaultPaymentMethod),
		Metadata:         c.Metadata,
		CreatedAt:        toTime(c.Created),
	}, nil
}

func (p *StripeProvider) CreateSubscription(ctx context.Context, options SubscriptionOptions) (*PaymentSubscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(options.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(options.PriceID),
			},
		},
	}

	if options.TrialPeriodDays > 0 {
		params.TrialPeriodDays = stripe.Int64(int64(options.TrialPeriodDays))
	}

	if options.DefaultPaymentID != "" {
		params.DefaultPaymentMethod = stripe.String(options.DefaultPaymentID)
	}

	if len(options.Metadata) > 0 {
		params.Metadata = options.Metadata
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return stripeSubscriptionToPaymentSubscription(sub), nil
}

func (p *StripeProvider) UpdateSubscription(ctx context.Context, subscriptionID string, options UpdateSubscriptionOptions) (*PaymentSubscription, error) {
	params := &stripe.SubscriptionParams{}

	if options.CancelAtPeriodEnd {
		params.CancelAtPeriodEnd = stripe.Bool(true)
	}

	if options.DefaultPaymentID != "" {
		params.DefaultPaymentMethod = stripe.String(options.DefaultPaymentID)
	}

	if options.ProrationBehavior != "" {
		params.ProrationBehavior = stripe.String(options.ProrationBehavior)
	}

	if len(options.Metadata) > 0 {
		params.Metadata = options.Metadata
	}

	if options.PriceID != "" {
		sub, err := subscription.Get(subscriptionID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get subscription: %w", err)
		}

		params.Items = []*stripe.SubscriptionItemsParams{}
		for _, item := range sub.Items.Data {
			params.Items = append(params.Items, &stripe.SubscriptionItemsParams{
				ID:      stripe.String(item.ID),
				Deleted: stripe.Bool(true),
			})
		}
		params.Items = append(params.Items, &stripe.SubscriptionItemsParams{
			Price: stripe.String(options.PriceID),
		})
	}

	sub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return stripeSubscriptionToPaymentSubscription(sub), nil
}

func (p *StripeProvider) CancelSubscription(ctx context.Context, subscriptionID string, immediately bool) error {
	params := &stripe.SubscriptionCancelParams{}
	if immediately {
		params.InvoiceNow = stripe.Bool(true)
		params.Prorate = stripe.Bool(true)
	} else {
		_, err := subscription.Update(subscriptionID, &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		})
		return err
	}

	_, err := subscription.Cancel(subscriptionID, params)
	return err
}

func (p *StripeProvider) GetSubscription(ctx context.Context, subscriptionID string) (*PaymentSubscription, error) {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return stripeSubscriptionToPaymentSubscription(sub), nil
}

func (p *StripeProvider) AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error {
	_, err := paymentmethod.Attach(paymentMethodID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	})
	return err
}

func (p *StripeProvider) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	return err
}

func (p *StripeProvider) ListPaymentMethods(ctx context.Context, customerID string) ([]*PaymentMethod, error) {
	iter := paymentmethod.List(&stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	})

	var methods []*PaymentMethod
	for iter.Next() {
		pm := iter.PaymentMethod()
		method := &PaymentMethod{
			ID:        pm.ID,
			Type:      string(pm.Type),
			CreatedAt: toTime(pm.Created),
		}

		if pm.Card != nil {
			method.Card = &CardDetails{
				Brand:    string(pm.Card.Brand),
				Last4:    pm.Card.Last4,
				ExpMonth: int(pm.Card.ExpMonth),
				ExpYear:  int(pm.Card.ExpYear),
				Country:  pm.Card.Country,
			}
		}

		methods = append(methods, method)
	}

	return methods, iter.Err()
}

func (p *StripeProvider) SetDefaultPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error {
	_, err := customer.Update(customerID, &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	})
	return err
}

func (p *StripeProvider) CreatePortalSession(ctx context.Context, customerID string, returnURL string) (*PortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	ps, err := portalSession.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create portal session: %w", err)
	}

	return &PortalSession{
		ID:        ps.ID,
		URL:       ps.URL,
		ReturnURL: ps.ReturnURL,
		CreatedAt: toTime(ps.Created),
	}, nil
}

func (p *StripeProvider) ConstructWebhookEvent(payload []byte, signature string) (*WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, p.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to construct webhook event: %w", err)
	}

	return &WebhookEvent{
		ID:        event.ID,
		Type:      string(event.Type),
		Data:      event.Data.Raw,
		CreatedAt: toTime(event.Created),
	}, nil
}

func (p *StripeProvider) HandleWebhookEvent(ctx context.Context, event *WebhookEvent) error {
	return nil
}

func (p *StripeProvider) GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	inv, err := invoice.Get(invoiceID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return stripeInvoiceToInvoice(inv), nil
}

func (p *StripeProvider) ListInvoices(ctx context.Context, customerID string, limit int) ([]*Invoice, error) {
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(customerID),
	}
	if limit > 0 {
		params.Limit = stripe.Int64(int64(limit))
	}

	iter := invoice.List(params)
	var invoices []*Invoice
	for iter.Next() {
		invoices = append(invoices, stripeInvoiceToInvoice(iter.Invoice()))
	}

	return invoices, iter.Err()
}

// Helper functions

func toTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func toTimePtr(timestamp int64) *time.Time {
	if timestamp == 0 {
		return nil
	}
	t := toTime(timestamp)
	return &t
}

func getStringID(obj interface{}) string {
	if obj == nil {
		return ""
	}
	switch v := obj.(type) {
	case *stripe.Customer:
		if v == nil {
			return ""
		}
		return v.ID
	case *stripe.Subscription:
		if v == nil {
			return ""
		}
		return v.ID
	case *stripe.PaymentIntent:
		if v == nil {
			return ""
		}
		return v.ID
	case *stripe.PaymentMethod:
		if v == nil {
			return ""
		}
		return v.ID
	default:
		return ""
	}
}

func stripeSubscriptionToPaymentSubscription(sub *stripe.Subscription) *PaymentSubscription {
	ps := &PaymentSubscription{
		ID:                 sub.ID,
		CustomerID:         getStringID(sub.Customer),
		Status:             string(sub.Status),
		CurrentPeriodStart: toTime(sub.CurrentPeriodStart),
		CurrentPeriodEnd:   toTime(sub.CurrentPeriodEnd),
		CancelAt:           toTimePtr(sub.CancelAt),
		CanceledAt:         toTimePtr(sub.CanceledAt),
		EndedAt:            toTimePtr(sub.EndedAt),
		TrialStart:         toTimePtr(sub.TrialStart),
		TrialEnd:           toTimePtr(sub.TrialEnd),
		DefaultPaymentID:   getStringID(sub.DefaultPaymentMethod),
		LatestInvoiceID:    getStringID(sub.LatestInvoice),
		Metadata:           sub.Metadata,
		Items:              []SubscriptionItem{},
	}

	for _, item := range sub.Items.Data {
		ps.Items = append(ps.Items, SubscriptionItem{
			ID:       item.ID,
			PriceID:  item.Price.ID,
			Quantity: item.Quantity,
		})
	}

	return ps
}

func stripeInvoiceToInvoice(inv *stripe.Invoice) *Invoice {
	i := &Invoice{
		ID:               inv.ID,
		Number:           inv.Number,
		CustomerID:       getStringID(inv.Customer),
		SubscriptionID:   getStringID(inv.Subscription),
		Status:           string(inv.Status),
		AmountDue:        inv.AmountDue,
		AmountPaid:       inv.AmountPaid,
		Currency:         string(inv.Currency),
		InvoicePDF:       inv.InvoicePDF,
		HostedInvoiceURL: inv.HostedInvoiceURL,
		CreatedAt:        toTime(inv.Created),
		PaidAt:           toTimePtr(inv.StatusTransitions.PaidAt),
		DueDate:          toTimePtr(inv.DueDate),
		Lines:            []InvoiceLine{},
	}

	for _, line := range inv.Lines.Data {
		i.Lines = append(i.Lines, InvoiceLine{
			Description: line.Description,
			Quantity:    line.Quantity,
			UnitAmount:  line.Price.UnitAmount,
			Amount:      line.Amount,
		})
	}

	return i
}