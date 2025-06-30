package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/subscription-go"
	"gorm.io/gorm"
)

type UsageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) subscription.UsageRepository {
	return &UsageRepository{db: db}
}

func (r *UsageRepository) Create(ctx context.Context, usage *subscription.SubscriptionUsage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

func (r *UsageRepository) Update(ctx context.Context, usage *subscription.SubscriptionUsage) error {
	usage.LastUpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(usage).Error
}

func (r *UsageRepository) FindByID(ctx context.Context, id uuid.UUID) (*subscription.SubscriptionUsage, error) {
	var usage subscription.SubscriptionUsage
	err := r.db.WithContext(ctx).
		Preload("Subscription").
		First(&usage, "id = ?", id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, subscription.ErrUsageNotFound
		}
		return nil, err
	}
	
	return &usage, nil
}

func (r *UsageRepository) FindBySubscriptionAndPeriod(ctx context.Context, subscriptionID uuid.UUID, periodStart, periodEnd time.Time) (*subscription.SubscriptionUsage, error) {
	var usage subscription.SubscriptionUsage
	err := r.db.WithContext(ctx).
		Preload("Subscription").
		Where("subscription_id = ? AND period_start = ? AND period_end = ?", 
			subscriptionID, periodStart, periodEnd).
		First(&usage).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, subscription.ErrUsageNotFound
		}
		return nil, err
	}
	
	return &usage, nil
}

func (r *UsageRepository) FindBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]*subscription.SubscriptionUsage, error) {
	var usages []*subscription.SubscriptionUsage
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("period_start DESC").
		Find(&usages).Error
	
	if err != nil {
		return nil, err
	}
	
	return usages, nil
}

func (r *UsageRepository) CreateEvent(ctx context.Context, event *subscription.UsageEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *UsageRepository) FindEventsBySubscription(ctx context.Context, subscriptionID uuid.UUID, limit int) ([]*subscription.UsageEvent, error) {
	var events []*subscription.UsageEvent
	query := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&events).Error
	if err != nil {
		return nil, err
	}
	
	return events, nil
}