package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/subscription-go"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) subscription.SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *subscription.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	query := r.db.WithContext(ctx)
	
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	
	err := query.First(&sub, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, subscription.ErrSubscriptionNotFound
		}
		return nil, err
	}
	
	return &sub, nil
}

func (r *SubscriptionRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Preload("Plan.Features").
		First(&sub, "account_id = ?", accountID).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, subscription.ErrSubscriptionNotFound
		}
		return nil, err
	}
	
	return &sub, nil
}

func (r *SubscriptionRepository) FindByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		First(&sub, "stripe_subscription_id = ?", stripeSubscriptionID).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, subscription.ErrSubscriptionNotFound
		}
		return nil, err
	}
	
	return &sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *subscription.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&subscription.Subscription{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return subscription.ErrSubscriptionNotFound
	}
	
	return nil
}