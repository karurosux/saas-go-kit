package gorm

import (
	"context"
	"errors"
	
	"{{.Project.GoModule}}/internal/auth/interface"
	"{{.Project.GoModule}}/internal/auth/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountRepository implements account repository using GORM
type AccountRepository struct {
	db *gorm.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *gorm.DB) authinterface.AccountRepository {
	return &AccountRepository{db: db}
}

// Create creates a new account
func (r *AccountRepository) Create(ctx context.Context, account authinterface.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// GetByID gets an account by ID
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

// GetByEmail gets an account by email
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

// GetByPhone gets an account by phone
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

// Update updates an account
func (r *AccountRepository) Update(ctx context.Context, account authinterface.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// Delete deletes an account
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

// Exists checks if an account exists by ID
func (r *AccountRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authmodel.Account{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail checks if an account exists by email
func (r *AccountRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authmodel.Account{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ExistsByPhone checks if an account exists by phone
func (r *AccountRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&authmodel.Account{}).Where("phone = ? AND phone != ''", phone).Count(&count).Error
	return count > 0, err
}