package gorm

import (
	"context"
	"errors"
	
	subscriptioninterface "{{.Project.GoModule}}/internal/subscription/interface"
	subscriptionmodel "{{.Project.GoModule}}/internal/subscription/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionPlanRepository struct {
	db *gorm.DB
}

func NewSubscriptionPlanRepository(db *gorm.DB) subscriptioninterface.SubscriptionPlanRepository {
	return &SubscriptionPlanRepository{db: db}
}

func (r *SubscriptionPlanRepository) Create(ctx context.Context, plan subscriptioninterface.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *SubscriptionPlanRepository) GetByID(ctx context.Context, id uuid.UUID) (subscriptioninterface.SubscriptionPlan, error) {
	var plan subscriptionmodel.SubscriptionPlan
	err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subscription plan not found")
		}
		return nil, err
	}
	return &plan, nil
}

func (r *SubscriptionPlanRepository) GetByType(ctx context.Context, planType subscriptioninterface.PlanType) (subscriptioninterface.SubscriptionPlan, error) {
	var plan subscriptionmodel.SubscriptionPlan
	err := r.db.WithContext(ctx).First(&plan, "type = ?", planType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subscription plan not found")
		}
		return nil, err
	}
	return &plan, nil
}

func (r *SubscriptionPlanRepository) GetAll(ctx context.Context, activeOnly bool) ([]subscriptioninterface.SubscriptionPlan, error) {
	var plans []subscriptionmodel.SubscriptionPlan
	
	query := r.db.WithContext(ctx)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	
	err := query.Order("price_monthly ASC").Find(&plans).Error
	if err != nil {
		return nil, err
	}
	
	result := make([]subscriptioninterface.SubscriptionPlan, len(plans))
	for i, p := range plans {
		plan := p
		result[i] = &plan
	}
	return result, nil
}

func (r *SubscriptionPlanRepository) Update(ctx context.Context, plan subscriptioninterface.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *SubscriptionPlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&subscriptionmodel.SubscriptionPlan{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("subscription plan not found")
	}
	return nil
}