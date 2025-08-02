package gorm

import (
	"context"
	"errors"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmodel "{{.Project.GoModule}}/internal/auth/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) authinterface.AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account authinterface.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (authinterface.Account, error) {
	var account authmodel.Account
	err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) GetByEmail(ctx context.Context, email string) (authinterface.Account, error) {
	var account authmodel.Account
	err := r.db.WithContext(ctx).First(&account, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) GetByPhone(ctx context.Context, phone string) (authinterface.Account, error) {
	var account authmodel.Account
	err := r.db.WithContext(ctx).First(&account, "phone = ?", phone).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) Update(ctx context.Context, account authinterface.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&authmodel.Account{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("account not found")
	}
	return nil
}

func (r *AccountRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authmodel.Account{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *AccountRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authmodel.Account{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *AccountRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authmodel.Account{}).Where("phone = ? AND phone != ''", phone).Count(&count).Error
	return count > 0, err
}