package subscriptionmodel

import (
	"time"
	
	"github.com/google/uuid"
)

// Invoice represents a subscription invoice
type Invoice struct {
	ID               string     `json:"id" gorm:"primary_key"`
	AccountID        uuid.UUID  `json:"account_id" gorm:"type:uuid;not null;index"`
	SubscriptionID   uuid.UUID  `json:"subscription_id" gorm:"type:uuid;not null;index"`
	Amount           int64      `json:"amount" gorm:"not null"` // in cents
	Currency         string     `json:"currency" gorm:"not null;default:'usd'"`
	Status           string     `json:"status" gorm:"not null"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	DueDate          time.Time  `json:"due_date" gorm:"not null"`
	StripeInvoiceID  string     `json:"stripe_invoice_id" gorm:"uniqueIndex"`
	PDF              string     `json:"pdf"`
	CreatedAt        time.Time  `json:"created_at"`
}

// GetID returns the invoice ID
func (i *Invoice) GetID() string {
	return i.ID
}

// GetAccountID returns the account ID
func (i *Invoice) GetAccountID() uuid.UUID {
	return i.AccountID
}

// GetSubscriptionID returns the subscription ID
func (i *Invoice) GetSubscriptionID() uuid.UUID {
	return i.SubscriptionID
}

// GetAmount returns the amount in cents
func (i *Invoice) GetAmount() int64 {
	return i.Amount
}

// GetCurrency returns the currency
func (i *Invoice) GetCurrency() string {
	return i.Currency
}

// GetStatus returns the invoice status
func (i *Invoice) GetStatus() string {
	return i.Status
}

// GetPaidAt returns when invoice was paid
func (i *Invoice) GetPaidAt() *time.Time {
	return i.PaidAt
}

// GetDueDate returns the due date
func (i *Invoice) GetDueDate() time.Time {
	return i.DueDate
}

// GetStripeInvoiceID returns the Stripe invoice ID
func (i *Invoice) GetStripeInvoiceID() string {
	return i.StripeInvoiceID
}

// GetPDF returns the PDF URL
func (i *Invoice) GetPDF() string {
	return i.PDF
}

// GetCreatedAt returns creation time
func (i *Invoice) GetCreatedAt() time.Time {
	return i.CreatedAt
}