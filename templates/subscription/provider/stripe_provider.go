package provider

import (
	"context"
	"fmt"
	
	"{{.Project.GoModule}}/internal/subscription/interface"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/invoice"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/billingportal/session"
)

// StripeProvider implements payment provider using Stripe
type StripeProvider struct {
	apiKey string
}

// NewStripeProvider creates a new Stripe payment provider
func NewStripeProvider(apiKey string) subscriptioninterface.PaymentProvider {
	stripe.Key = apiKey
	return &StripeProvider{apiKey: apiKey}
}

// CreateCustomer creates a new Stripe customer
func (p *StripeProvider) CreateCustomer(ctx context.Context, email string, metadata map[string]string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}
	
	if metadata != nil {
		params.Metadata = metadata
	}
	
	c, err := customer.New(params)
	if err != nil {
		return "", err
	}
	
	return c.ID, nil
}

// UpdateCustomer updates a Stripe customer
func (p *StripeProvider) UpdateCustomer(ctx context.Context, customerID string, metadata map[string]string) error {
	params := &stripe.CustomerParams{}
	if metadata != nil {
		params.Metadata = metadata
	}
	
	_, err := customer.Update(customerID, params)
	return err
}

// DeleteCustomer deletes a Stripe customer
func (p *StripeProvider) DeleteCustomer(ctx context.Context, customerID string) error {
	_, err := customer.Del(customerID, nil)
	return err
}

// CreateSubscription creates a new Stripe subscription
func (p *StripeProvider) CreateSubscription(ctx context.Context, req subscriptioninterface.CreateSubscriptionRequest) (*subscriptioninterface.ProviderSubscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(req.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(req.PriceID),
			},
		},
	}
	
	if req.TrialDays > 0 {
		params.TrialPeriodDays = stripe.Int64(int64(req.TrialDays))
	}
	
	if req.PaymentMethodID != "" {
		params.DefaultPaymentMethod = stripe.String(req.PaymentMethodID)
	}
	
	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}
	
	s, err := sub.New(params)
	if err != nil {
		return nil, err
	}
	
	return p.mapSubscription(s), nil
}

// UpdateSubscription updates a Stripe subscription
func (p *StripeProvider) UpdateSubscription(ctx context.Context, subscriptionID string, req subscriptioninterface.UpdateSubscriptionRequest) (*subscriptioninterface.ProviderSubscription, error) {
	// Get current subscription
	s, err := sub.Get(subscriptionID, nil)
	if err != nil {
		return nil, err
	}
	
	// Update subscription items
	if req.PriceID != "" && len(s.Items.Data) > 0 {
		_, err = sub.Update(subscriptionID, &stripe.SubscriptionParams{
			Items: []*stripe.SubscriptionItemsParams{
				{
					ID:    stripe.String(s.Items.Data[0].ID),
					Price: stripe.String(req.PriceID),
				},
			},
		})
		if err != nil {
			return nil, err
		}
	}
	
	// Update metadata if provided
	if req.Metadata != nil {
		params := &stripe.SubscriptionParams{}
		params.Metadata = req.Metadata
		s, err = sub.Update(subscriptionID, params)
		if err != nil {
			return nil, err
		}
	}
	
	return p.mapSubscription(s), nil
}

// CancelSubscription cancels a Stripe subscription
func (p *StripeProvider) CancelSubscription(ctx context.Context, subscriptionID string, immediately bool) error {
	params := &stripe.SubscriptionCancelParams{}
	if !immediately {
		params.InvoiceNow = stripe.Bool(true)
		params.Prorate = stripe.Bool(true)
	}
	
	_, err := sub.Cancel(subscriptionID, params)
	return err
}

// ResumeSubscription resumes a canceled subscription
func (p *StripeProvider) ResumeSubscription(ctx context.Context, subscriptionID string) error {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}
	
	_, err := sub.Update(subscriptionID, params)
	return err
}

// AttachPaymentMethod attaches a payment method to a customer
func (p *StripeProvider) AttachPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	
	_, err := paymentmethod.Attach(paymentMethodID, params)
	return err
}

// DetachPaymentMethod detaches a payment method
func (p *StripeProvider) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	return err
}

// SetDefaultPaymentMethod sets the default payment method for a customer
func (p *StripeProvider) SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}
	
	_, err := customer.Update(customerID, params)
	return err
}

// ListPaymentMethods lists payment methods for a customer
func (p *StripeProvider) ListPaymentMethods(ctx context.Context, customerID string) ([]subscriptioninterface.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}
	
	i := paymentmethod.List(params)
	var methods []subscriptioninterface.PaymentMethod
	
	for i.Next() {
		pm := i.PaymentMethod()
		method := subscriptioninterface.PaymentMethod{
			ID:   pm.ID,
			Type: string(pm.Type),
		}
		
		if pm.Card != nil {
			method.Last4 = pm.Card.Last4
			method.Brand = string(pm.Card.Brand)
			method.ExpMonth = int(pm.Card.ExpMonth)
			method.ExpYear = int(pm.Card.ExpYear)
		}
		
		methods = append(methods, method)
	}
	
	if err := i.Err(); err != nil {
		return nil, err
	}
	
	return methods, nil
}

// CreateInvoice creates a new invoice
func (p *StripeProvider) CreateInvoice(ctx context.Context, customerID string, items []subscriptioninterface.InvoiceItem) (*subscriptioninterface.ProviderInvoice, error) {
	params := &stripe.InvoiceParams{
		Customer: stripe.String(customerID),
	}
	
	inv, err := invoice.New(params)
	if err != nil {
		return nil, err
	}
	
	// Add invoice items
	for _, item := range items {
		itemParams := &stripe.InvoiceItemParams{
			Customer:    stripe.String(customerID),
			Invoice:     stripe.String(inv.ID),
			Price:       stripe.String(item.PriceID),
			Quantity:    stripe.Int64(item.Quantity),
			Description: stripe.String(item.Description),
		}
		
		_, err = invoice.New(itemParams)
		if err != nil {
			return nil, err
		}
	}
	
	// Finalize invoice
	inv, err = invoice.FinalizeInvoice(inv.ID, nil)
	if err != nil {
		return nil, err
	}
	
	return p.mapInvoice(inv), nil
}

// GetInvoice gets an invoice by ID
func (p *StripeProvider) GetInvoice(ctx context.Context, invoiceID string) (*subscriptioninterface.ProviderInvoice, error) {
	inv, err := invoice.Get(invoiceID, nil)
	if err != nil {
		return nil, err
	}
	
	return p.mapInvoice(inv), nil
}

// ListInvoices lists invoices for a customer
func (p *StripeProvider) ListInvoices(ctx context.Context, customerID string, limit int) ([]subscriptioninterface.ProviderInvoice, error) {
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(customerID),
	}
	params.Limit = stripe.Int64(int64(limit))
	
	i := invoice.List(params)
	var invoices []subscriptioninterface.ProviderInvoice
	
	for i.Next() {
		inv := i.Invoice()
		invoices = append(invoices, *p.mapInvoice(inv))
	}
	
	if err := i.Err(); err != nil {
		return nil, err
	}
	
	return invoices, nil
}

// CreateCheckoutSession creates a Stripe checkout session
func (p *StripeProvider) CreateCheckoutSession(ctx context.Context, req subscriptioninterface.CheckoutSessionRequest) (*subscriptioninterface.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(req.Quantity),
			},
		},
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
	}
	
	if req.CustomerID != "" {
		params.Customer = stripe.String(req.CustomerID)
	}
	
	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}
	
	if req.AllowPromoCodes {
		params.AllowPromotionCodes = stripe.Bool(true)
	}
	
	s, err := session.New(params)
	if err != nil {
		return nil, err
	}
	
	return &subscriptioninterface.CheckoutSession{
		ID:         s.ID,
		URL:        s.URL,
		CustomerID: s.Customer.ID,
		Status:     string(s.Status),
		ExpiresAt:  time.Unix(s.ExpiresAt, 0),
	}, nil
}

// CreateBillingPortalSession creates a billing portal session
func (p *StripeProvider) CreateBillingPortalSession(ctx context.Context, customerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}
	
	s, err := portalsession.New(params)
	if err != nil {
		return "", err
	}
	
	return s.URL, nil
}

// Helper methods

func (p *StripeProvider) mapSubscription(s *stripe.Subscription) *subscriptioninterface.ProviderSubscription {
	ps := &subscriptioninterface.ProviderSubscription{
		ID:                s.ID,
		CustomerID:        s.Customer.ID,
		Status:            string(s.Status),
		CurrentPeriodStart: time.Unix(s.CurrentPeriodStart, 0),
		CurrentPeriodEnd:   time.Unix(s.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
	}
	
	if s.TrialEnd > 0 {
		trialEnd := time.Unix(s.TrialEnd, 0)
		ps.TrialEnd = &trialEnd
	}
	
	if s.CancelAt > 0 {
		cancelAt := time.Unix(s.CancelAt, 0)
		ps.CancelAt = &cancelAt
	}
	
	if s.DefaultPaymentMethod != nil {
		ps.DefaultPaymentMethod = s.DefaultPaymentMethod.ID
	}
	
	// Map subscription items
	for _, item := range s.Items.Data {
		ps.Items = append(ps.Items, subscriptioninterface.SubscriptionItem{
			ID:       item.ID,
			PriceID:  item.Price.ID,
			Quantity: item.Quantity,
		})
	}
	
	return ps
}

func (p *StripeProvider) mapInvoice(inv *stripe.Invoice) *subscriptioninterface.ProviderInvoice {
	pi := &subscriptioninterface.ProviderInvoice{
		ID:         inv.ID,
		CustomerID: inv.Customer.ID,
		Amount:     inv.Total,
		Currency:   string(inv.Currency),
		Status:     string(inv.Status),
		DueDate:    time.Unix(inv.DueDate, 0),
		PDF:        inv.InvoicePDF,
	}
	
	if inv.StatusTransitions != nil && inv.StatusTransitions.PaidAt > 0 {
		paidAt := time.Unix(inv.StatusTransitions.PaidAt, 0)
		pi.PaidAt = &paidAt
	}
	
	return pi
}