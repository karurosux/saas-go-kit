package subscriptionmodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/subscription/interface"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Subscription represents an account subscription
type Subscription struct {
	ID                    uuid.UUID                               `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AccountID             uuid.UUID                               `json:"account_id" gorm:"type:uuid;not null;uniqueIndex"`
	PlanID                uuid.UUID                               `json:"plan_id" gorm:"type:uuid;not null"`
	Plan                  *SubscriptionPlan                       `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	Status                subscriptioninterface.SubscriptionStatus `json:"status" gorm:"not null"`
	BillingPeriod         subscriptioninterface.BillingPeriod     `json:"billing_period" gorm:"not null"`
	CurrentPeriodStart    time.Time                               `json:"current_period_start" gorm:"not null"`
	CurrentPeriodEnd      time.Time                               `json:"current_period_end" gorm:"not null"`
	TrialEndsAt           *time.Time                              `json:"trial_ends_at,omitempty"`
	CanceledAt            *time.Time                              `json:"canceled_at,omitempty"`
	CancelAtPeriodEnd     bool                                    `json:"cancel_at_period_end" gorm:"default:false"`
	StripeCustomerID      string                                  `json:"stripe_customer_id" gorm:"index"`
	StripeSubscriptionID  string                                  `json:"stripe_subscription_id" gorm:"uniqueIndex"`
	Metadata              datatypes.JSON                          `json:"metadata" gorm:"type:jsonb"`
	CreatedAt             time.Time                               `json:"created_at"`
	UpdatedAt             time.Time                               `json:"updated_at"`
}

// GetID returns the subscription ID
func (s *Subscription) GetID() uuid.UUID {
	return s.ID
}

// GetAccountID returns the account ID
func (s *Subscription) GetAccountID() uuid.UUID {
	return s.AccountID
}

// GetPlanID returns the plan ID
func (s *Subscription) GetPlanID() uuid.UUID {
	return s.PlanID
}

// GetPlan returns the subscription plan
func (s *Subscription) GetPlan() subscriptioninterface.SubscriptionPlan {
	if s.Plan == nil {
		return nil
	}
	return s.Plan
}

// GetStatus returns the subscription status
func (s *Subscription) GetStatus() subscriptioninterface.SubscriptionStatus {
	return s.Status
}

// GetBillingPeriod returns the billing period
func (s *Subscription) GetBillingPeriod() subscriptioninterface.BillingPeriod {
	return s.BillingPeriod
}

// GetCurrentPeriodStart returns the current period start
func (s *Subscription) GetCurrentPeriodStart() time.Time {
	return s.CurrentPeriodStart
}

// GetCurrentPeriodEnd returns the current period end
func (s *Subscription) GetCurrentPeriodEnd() time.Time {
	return s.CurrentPeriodEnd
}

// GetTrialEndsAt returns when trial ends
func (s *Subscription) GetTrialEndsAt() *time.Time {
	return s.TrialEndsAt
}

// GetCanceledAt returns when subscription was canceled
func (s *Subscription) GetCanceledAt() *time.Time {
	return s.CanceledAt
}

// GetCancelAtPeriodEnd returns if subscription cancels at period end
func (s *Subscription) GetCancelAtPeriodEnd() bool {
	return s.CancelAtPeriodEnd
}

// GetStripeCustomerID returns the Stripe customer ID
func (s *Subscription) GetStripeCustomerID() string {
	return s.StripeCustomerID
}

// GetStripeSubscriptionID returns the Stripe subscription ID
func (s *Subscription) GetStripeSubscriptionID() string {
	return s.StripeSubscriptionID
}

// GetMetadata returns the metadata
func (s *Subscription) GetMetadata() map[string]string {
	var metadata map[string]string
	if s.Metadata != nil {
		s.Metadata.Scan(&metadata)
	}
	return metadata
}

// GetCreatedAt returns creation time
func (s *Subscription) GetCreatedAt() time.Time {
	return s.CreatedAt
}

// GetUpdatedAt returns last update time
func (s *Subscription) GetUpdatedAt() time.Time {
	return s.UpdatedAt
}

// SetStatus sets the subscription status
func (s *Subscription) SetStatus(status subscriptioninterface.SubscriptionStatus) {
	s.Status = status
}

// SetCancelAtPeriodEnd sets if subscription cancels at period end
func (s *Subscription) SetCancelAtPeriodEnd(cancel bool) {
	s.CancelAtPeriodEnd = cancel
}

// SetCanceledAt sets when subscription was canceled
func (s *Subscription) SetCanceledAt(canceledAt *time.Time) {
	s.CanceledAt = canceledAt
}