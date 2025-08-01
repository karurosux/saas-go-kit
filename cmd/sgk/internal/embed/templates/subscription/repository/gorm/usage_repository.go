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

type UsageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) subscriptioninterface.UsageRepository {
	return &UsageRepository{db: db}
}

func (r *UsageRepository) Create(ctx context.Context, usage subscriptioninterface.Usage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

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

func (r *UsageRepository) GetByAccount(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time) ([]subscriptioninterface.Usage, error) {
	var usageRecords []subscriptionmodel.Usage
	
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND period >= ? AND period <= ?", accountID, startDate, endDate).
		Order("period ASC, resource ASC").
		Find(&usageRecords).Error
	if err != nil {
		return nil, err
	}
	
	result := make([]subscriptioninterface.Usage, len(usageRecords))
	for i, u := range usageRecords {
		usage := u
		result[i] = &usage
	}
	return result, nil
}

func (r *UsageRepository) Update(ctx context.Context, usage subscriptioninterface.Usage) error {
	return r.db.WithContext(ctx).Save(usage).Error
}

func (r *UsageRepository) IncrementUsage(ctx context.Context, accountID uuid.UUID, resource string, amount int64) error {
	period := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var usage subscriptionmodel.Usage
		err := tx.Where("account_id = ? AND resource = ? AND period = ?", accountID, resource, period).
			First(&usage).Error
		
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
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
		
		usage.IncrementQuantity(amount)
		return tx.Save(&usage).Error
	})
}

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