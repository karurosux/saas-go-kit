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

// UsageRepository implements usage repository using GORM
type UsageRepository struct {
	db *gorm.DB
}

// NewUsageRepository creates a new usage repository
func NewUsageRepository(db *gorm.DB) subscriptioninterface.UsageRepository {
	return &UsageRepository{db: db}
}

// Create creates a new usage record
func (r *UsageRepository) Create(ctx context.Context, usage subscriptioninterface.Usage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

// GetByID gets a usage record by ID
func (r *UsageRepository) GetByID(ctx context.Context, id uuid.UUID) (subscriptioninterface.Usage, error) {
	var usage subscriptionmodel.Usage
	err := r.db.WithContext(ctx).First(&usage, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usage record not found")
		}
		return nil, err
	}
	return &usage, nil
}

// GetByAccountAndResource gets usage by account and resource for a period
func (r *UsageRepository) GetByAccountAndResource(ctx context.Context, accountID uuid.UUID, resource string, period time.Time) (subscriptioninterface.Usage, error) {
	var usage subscriptionmodel.Usage
	err := r.db.WithContext(ctx).
		First(&usage, "account_id = ? AND resource = ? AND period = ?", accountID, resource, period).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usage record not found")
		}
		return nil, err
	}
	return &usage, nil
}

// GetByAccount gets all usage records for an account in a date range
func (r *UsageRepository) GetByAccount(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) ([]subscriptioninterface.Usage, error) {
	var usageRecords []subscriptionmodel.Usage
	
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND period >= ? AND period <= ?", accountID, startDate, endDate).
		Order("period ASC, resource ASC").
		Find(&usageRecords).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	result := make([]subscriptioninterface.Usage, len(usageRecords))
	for i, u := range usageRecords {
		usage := u // Create a copy to avoid pointer issues
		result[i] = &usage
	}
	return result, nil
}

// Update updates a usage record
func (r *UsageRepository) Update(ctx context.Context, usage subscriptioninterface.Usage) error {
	return r.db.WithContext(ctx).Save(usage).Error
}

// IncrementUsage increments usage for a resource
func (r *UsageRepository) IncrementUsage(ctx context.Context, accountID uuid.UUID, resource string, amount int64) error {
	period := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var usage subscriptionmodel.Usage
		err := tx.Where("account_id = ? AND resource = ? AND period = ?", accountID, resource, period).
			First(&usage).Error
		
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new usage record
				usage = subscriptionmodel.Usage{
					ID:        uuid.New(),
					AccountID: accountID,
					Resource:  resource,
					Quantity:  amount,
					Period:    period,
				}
				return tx.Create(&usage).Error
			}
			return err
		}
		
		// Increment existing usage
		usage.IncrementQuantity(amount)
		return tx.Save(&usage).Error
	})
}

// GetCurrentPeriodUsage gets current period usage for a resource
func (r *UsageRepository) GetCurrentPeriodUsage(ctx context.Context, accountID uuid.UUID, resource string) (int64, error) {
	period := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	
	var usage subscriptionmodel.Usage
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND resource = ? AND period = ?", accountID, resource, period).
		First(&usage).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	
	return usage.GetQuantity(), nil
}