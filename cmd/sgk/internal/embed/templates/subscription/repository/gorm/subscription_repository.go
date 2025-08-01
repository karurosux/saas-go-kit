package gorm

import (
	"context"
	"errors"
	"time"
	
	subscriptioninterface "{{.Project.GoModule}}/internal/subscription/interface"
	subscriptionmodel "{{.Project.GoModule}}/internal/subscription/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) subscriptioninterface.SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription subscriptioninterface.Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

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

func (r *SubscriptionRepository) Update(ctx context.Context, subscription subscriptioninterface.Subscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

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
	
	result := make([]subscriptioninterface.Subscription, len(subscriptions))
	for i, s := range subscriptions {
		sub := s
		result[i] = &sub
	}
	return result, nil
}

func (r *SubscriptionRepository) GetPastDue(ctx context.Context) ([]subscriptioninterface.Subscription, error) {
	var subscriptions []subscriptionmodel.Subscription
	
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("status = ?", subscriptioninterface.StatusPastDue).
		Find(&subscriptions).Error
	if err != nil {
		return nil, err
	}
	
	result := make([]subscriptioninterface.Subscription, len(subscriptions))
	for i, s := range subscriptions {
		sub := s
		result[i] = &sub
	}
	return result, nil
}