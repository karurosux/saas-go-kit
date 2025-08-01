package subscriptionmodel

import (
	"time"
	
	subscriptioninterface "{{.Project.GoModule}}/internal/subscription/interface"
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

func (s *Subscription) GetID() uuid.UUID {
	return s.ID
}

func (s *Subscription) GetAccountID() uuid.UUID {
	return s.AccountID
}

func (s *Subscription) GetPlanID() uuid.UUID {
	return s.PlanID
}

func (s *Subscription) GetPlan() subscriptioninterface.SubscriptionPlan {
	if s.Plan == nil {
		return nil
	}
	return s.Plan
}

func (s *Subscription) GetStatus() subscriptioninterface.SubscriptionStatus {
	return s.Status
}

func (s *Subscription) GetBillingPeriod() subscriptioninterface.BillingPeriod {
	return s.BillingPeriod
}

func (s *Subscription) GetCurrentPeriodStart() time.Time {
	return s.CurrentPeriodStart
}

func (s *Subscription) GetCurrentPeriodEnd() time.Time {
	return s.CurrentPeriodEnd
}

func (s *Subscription) GetTrialEndsAt() *time.Time {
	return s.TrialEndsAt
}

func (s *Subscription) GetCanceledAt() *time.Time {
	return s.CanceledAt
}

func (s *Subscription) GetCancelAtPeriodEnd() bool {
	return s.CancelAtPeriodEnd
}

func (s *Subscription) GetStripeCustomerID() string {
	return s.StripeCustomerID
}

func (s *Subscription) GetStripeSubscriptionID() string {
	return s.StripeSubscriptionID
}

func (s *Subscription) GetMetadata() map[string]string {
	var metadata map[string]string
	if s.Metadata != nil {
		s.Metadata.Scan(&metadata)
	}
	return metadata
}

func (s *Subscription) GetCreatedAt() time.Time {
	return s.CreatedAt
}

func (s *Subscription) GetUpdatedAt() time.Time {
	return s.UpdatedAt
}

func (s *Subscription) SetStatus(status subscriptioninterface.SubscriptionStatus) {
	s.Status = status
}

func (s *Subscription) SetCancelAtPeriodEnd(cancel bool) {
	s.CancelAtPeriodEnd = cancel
}

func (s *Subscription) SetCanceledAt(canceledAt *time.Time) {
	s.CanceledAt = canceledAt
}