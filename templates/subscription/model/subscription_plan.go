package subscriptionmodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/subscription/interface"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SubscriptionPlan represents a subscription plan
type SubscriptionPlan struct {
	ID                   uuid.UUID                         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name                 string                            `json:"name" gorm:"not null"`
	Type                 subscriptioninterface.PlanType    `json:"type" gorm:"not null;uniqueIndex"`
	PriceMonthly         int64                             `json:"price_monthly" gorm:"not null"` // in cents
	PriceYearly          int64                             `json:"price_yearly" gorm:"not null"`  // in cents
	Features             datatypes.JSON                    `json:"features" gorm:"type:jsonb"`
	Limits               datatypes.JSON                    `json:"limits" gorm:"type:jsonb"`
	StripeProductID      string                            `json:"stripe_product_id"`
	StripePriceMonthlyID string                            `json:"stripe_price_monthly_id"`
	StripePriceYearlyID  string                            `json:"stripe_price_yearly_id"`
	TrialDays            int                               `json:"trial_days" gorm:"default:0"`
	IsActive             bool                              `json:"is_active" gorm:"default:true"`
	CreatedAt            time.Time                         `json:"created_at"`
	UpdatedAt            time.Time                         `json:"updated_at"`
}

// GetID returns the plan ID
func (p *SubscriptionPlan) GetID() uuid.UUID {
	return p.ID
}

// GetName returns the plan name
func (p *SubscriptionPlan) GetName() string {
	return p.Name
}

// GetType returns the plan type
func (p *SubscriptionPlan) GetType() subscriptioninterface.PlanType {
	return p.Type
}

// GetPriceMonthly returns the monthly price in cents
func (p *SubscriptionPlan) GetPriceMonthly() int64 {
	return p.PriceMonthly
}

// GetPriceYearly returns the yearly price in cents
func (p *SubscriptionPlan) GetPriceYearly() int64 {
	return p.PriceYearly
}

// GetFeatures returns the plan features
func (p *SubscriptionPlan) GetFeatures() map[string]interface{} {
	var features map[string]interface{}
	if p.Features != nil {
		p.Features.Scan(&features)
	}
	return features
}

// GetLimits returns the plan limits
func (p *SubscriptionPlan) GetLimits() map[string]int64 {
	var limits map[string]int64
	if p.Limits != nil {
		p.Limits.Scan(&limits)
	}
	return limits
}

// GetStripeProductID returns the Stripe product ID
func (p *SubscriptionPlan) GetStripeProductID() string {
	return p.StripeProductID
}

// GetStripePriceMonthlyID returns the Stripe monthly price ID
func (p *SubscriptionPlan) GetStripePriceMonthlyID() string {
	return p.StripePriceMonthlyID
}

// GetStripePriceYearlyID returns the Stripe yearly price ID
func (p *SubscriptionPlan) GetStripePriceYearlyID() string {
	return p.StripePriceYearlyID
}

// GetTrialDays returns the trial days
func (p *SubscriptionPlan) GetTrialDays() int {
	return p.TrialDays
}

// GetIsActive returns if the plan is active
func (p *SubscriptionPlan) GetIsActive() bool {
	return p.IsActive
}

// GetCreatedAt returns creation time
func (p *SubscriptionPlan) GetCreatedAt() time.Time {
	return p.CreatedAt
}

// GetUpdatedAt returns last update time
func (p *SubscriptionPlan) GetUpdatedAt() time.Time {
	return p.UpdatedAt
}