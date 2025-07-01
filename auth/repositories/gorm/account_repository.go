package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/auth-go"
	"gorm.io/gorm"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) auth.AccountStore {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account auth.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *AccountRepository) FindByID(ctx context.Context, id uuid.UUID) (auth.Account, error) {
	var account auth.DefaultAccount
	err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	
	return &account, nil
}

func (r *AccountRepository) FindByEmail(ctx context.Context, email string) (auth.Account, error) {
	var account auth.DefaultAccount
	err := r.db.WithContext(ctx).First(&account, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	
	return &account, nil
}

func (r *AccountRepository) Update(ctx context.Context, account auth.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&auth.DefaultAccount{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("account not found")
	}
	
	return nil
}

func (r *AccountRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&auth.DefaultAccount{}).Count(&count).Error
	return count, err
}