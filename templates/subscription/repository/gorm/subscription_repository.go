package gorm

import (
	"context"
	"errors"
	"time"
	
	"{{.Project.GoModule}}/internal/subscription/interface"
	"{{.Project.GoModule}}/internal/subscription/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionRepository implements subscription repository using GORM
type SubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *gorm.DB) subscriptioninterface.SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create creates a new subscription
func (r *SubscriptionRepository) Create(ctx context.Context, subscription subscriptioninterface.Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

// GetByID gets a subscription by ID
func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (subscriptioninterface.Subscription, error) {
	var subscription subscriptionmodel.Subscription
	err := r.db.WithContext(ctx).Preload("Plan").First(&subscription, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subscription not found")
		}
		return nil, err
	}
	return &subscription, nil
}

// GetByAccountID gets a subscription by account ID
func (r *SubscriptionRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) (subscriptioninterface.Subscription, error) {
	var subscription subscriptionmodel.Subscription
	err := r.db.WithContext(ctx).Preload("Plan").First(&subscription, "account_id = ?", accountID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subscription not found")
		}
		return nil, err
	}
	return &subscription, nil
}

// GetByStripeSubscriptionID gets a subscription by Stripe subscription ID
func (r *SubscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, stripeSubID string) (subscriptioninterface.Subscription, error) {
	var subscription subscriptionmodel.Subscription
	err := r.db.WithContext(ctx).Preload("Plan").First(&subscription, "stripe_subscription_id = ?", stripeSubID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subscription not found")
		}
		return nil, err
	}
	return &subscription, nil
}

// Update updates a subscription
func (r *SubscriptionRepository) Update(ctx context.Context, subscription subscriptioninterface.Subscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

// Delete deletes a subscription
func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&subscriptionmodel.Subscription{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("subscription not found")
	}
	return nil
}

// GetExpiringTrials gets subscriptions with trials expiring soon
func (r *SubscriptionRepository) GetExpiringTrials(ctx context.Context, daysAhead int) ([]subscriptioninterface.Subscription, error) {
	var subscriptions []subscriptionmodel.Subscription
	
	expiryDate := time.Now().AddDate(0, 0, daysAhead)
	
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("status = ? AND trial_ends_at IS NOT NULL AND trial_ends_at <= ?", 
			subscriptioninterface.StatusTrialing, expiryDate).
		Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	result := make([]subscriptioninterface.Subscription, len(subscriptions))
	for i, s := range subscriptions {
		sub := s // Create a copy to avoid pointer issues
		result[i] = &sub
	}
	return result, nil
}

// GetPastDue gets past due subscriptions
func (r *SubscriptionRepository) GetPastDue(ctx context.Context) ([]subscriptioninterface.Subscription, error) {
	var subscriptions []subscriptionmodel.Subscription
	
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("status = ?", subscriptioninterface.StatusPastDue).
		Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	result := make([]subscriptioninterface.Subscription, len(subscriptions))
	for i, s := range subscriptions {
		sub := s // Create a copy to avoid pointer issues
		result[i] = &sub
	}
	return result, nil
}