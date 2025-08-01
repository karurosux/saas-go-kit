package gorm

import (
	"context"
	"errors"
	
	"{{.Project.GoModule}}/internal/subscription/interface"
	"{{.Project.GoModule}}/internal/subscription/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionPlanRepository implements subscription plan repository using GORM
type SubscriptionPlanRepository struct {
	db *gorm.DB
}

// NewSubscriptionPlanRepository creates a new subscription plan repository
func NewSubscriptionPlanRepository(db *gorm.DB) subscriptioninterface.SubscriptionPlanRepository {
	return &SubscriptionPlanRepository{db: db}
}

// Create creates a new subscription plan
func (r *SubscriptionPlanRepository) Create(ctx context.Context, plan subscriptioninterface.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

// GetByID gets a subscription plan by ID
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

// GetByType gets a subscription plan by type
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

// GetAll gets all subscription plans
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
	
	// Convert to interface slice
	result := make([]subscriptioninterface.SubscriptionPlan, len(plans))
	for i, p := range plans {
		plan := p // Create a copy to avoid pointer issues
		result[i] = &plan
	}
	return result, nil
}

// Update updates a subscription plan
func (r *SubscriptionPlanRepository) Update(ctx context.Context, plan subscriptioninterface.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

// Delete deletes a subscription plan
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