package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/subscription-go"
	"gorm.io/gorm"
)

type SubscriptionPlanRepository struct {
	db *gorm.DB
}

func NewSubscriptionPlanRepository(db *gorm.DB) subscription.SubscriptionPlanRepository {
	return &SubscriptionPlanRepository{db: db}
}

func (r *SubscriptionPlanRepository) FindAll(ctx context.Context) ([]subscription.SubscriptionPlan, error) {
	var plans []subscription.SubscriptionPlan
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND is_visible = ?", true, true).
		Order("price ASC").
		Find(&plans).Error
	
	if err != nil {
		return nil, err
	}
	
	return plans, nil
}

func (r *SubscriptionPlanRepository) FindAllIncludingHidden(ctx context.Context) ([]subscription.SubscriptionPlan, error) {
	var plans []subscription.SubscriptionPlan
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("price ASC").
		Find(&plans).Error
	
	if err != nil {
		return nil, err
	}
	
	return plans, nil
}

func (r *SubscriptionPlanRepository) FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*subscription.SubscriptionPlan, error) {
	var plan subscription.SubscriptionPlan
	query := r.db.WithContext(ctx)
	
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	
	err := query.First(&plan, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, subscription.ErrSubscriptionPlanNotFound
		}
		return nil, err
	}
	
	return &plan, nil
}

func (r *SubscriptionPlanRepository) FindByCode(ctx context.Context, code string) (*subscription.SubscriptionPlan, error) {
	var plan subscription.SubscriptionPlan
	err := r.db.WithContext(ctx).
		First(&plan, "code = ? AND is_active = ?", code, true).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, subscription.ErrSubscriptionPlanNotFound
		}
		return nil, err
	}
	
	return &plan, nil
}

func (r *SubscriptionPlanRepository) Create(ctx context.Context, plan *subscription.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *SubscriptionPlanRepository) Update(ctx context.Context, plan *subscription.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *SubscriptionPlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&subscription.SubscriptionPlan{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return subscription.ErrSubscriptionPlanNotFound
	}
	
	return nil
}