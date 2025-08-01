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

func (i *Invoice) GetID() string {
	return i.ID
}

func (i *Invoice) GetAccountID() uuid.UUID {
	return i.AccountID
}

func (i *Invoice) GetSubscriptionID() uuid.UUID {
	return i.SubscriptionID
}

func (i *Invoice) GetAmount() int64 {
	return i.Amount
}

func (i *Invoice) GetCurrency() string {
	return i.Currency
}

func (i *Invoice) GetStatus() string {
	return i.Status
}

func (i *Invoice) GetPaidAt() *time.Time {
	return i.PaidAt
}

func (i *Invoice) GetDueDate() time.Time {
	return i.DueDate
}

func (i *Invoice) GetStripeInvoiceID() string {
	return i.StripeInvoiceID
}

func (i *Invoice) GetPDF() string {
	return i.PDF
}

func (i *Invoice) GetCreatedAt() time.Time {
	return i.CreatedAt
}