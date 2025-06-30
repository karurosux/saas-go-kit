package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/auth-go"
	"gorm.io/gorm"
)

// GormAccountStore implements auth.AccountStore using GORM
type GormAccountStore struct {
	db *gorm.DB
}

func NewGormAccountStore(db *gorm.DB) *GormAccountStore {
	return &GormAccountStore{db: db}
}

func (s *GormAccountStore) Create(ctx context.Context, account auth.Account) error {
	model := &AccountModel{}
	model.FromAuthAccount(account)
	return s.db.WithContext(ctx).Create(model).Error
}

func (s *GormAccountStore) FindByID(ctx context.Context, id uuid.UUID) (auth.Account, error) {
	var model AccountModel
	if err := s.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return model.ToAuthAccount(), nil
}

func (s *GormAccountStore) FindByEmail(ctx context.Context, email string) (auth.Account, error) {
	var model AccountModel
	if err := s.db.WithContext(ctx).First(&model, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return model.ToAuthAccount(), nil
}

func (s *GormAccountStore) Update(ctx context.Context, account auth.Account) error {
	model := &AccountModel{}
	model.FromAuthAccount(account)
	return s.db.WithContext(ctx).Save(model).Error
}

func (s *GormAccountStore) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&AccountModel{}, "id = ?", id).Error
}

func (s *GormAccountStore) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&AccountModel{}).Count(&count).Error
	return count, err
}

// GormTokenStore implements auth.TokenStore using GORM
type GormTokenStore struct {
	db *gorm.DB
}

func NewGormTokenStore(db *gorm.DB) *GormTokenStore {
	return &GormTokenStore{db: db}
}

func (s *GormTokenStore) Create(ctx context.Context, token auth.VerificationToken) error {
	model := &TokenModel{}
	model.FromAuthToken(token)
	return s.db.WithContext(ctx).Create(model).Error
}

func (s *GormTokenStore) FindByToken(ctx context.Context, token string) (auth.VerificationToken, error) {
	var model TokenModel
	if err := s.db.WithContext(ctx).First(&model, "token = ?", token).Error; err != nil {
		return nil, err
	}
	return model.ToAuthToken(), nil
}

func (s *GormTokenStore) MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&TokenModel{}).
		Where("id = ?", tokenID).
		Update("used_at", now).Error
}

func (s *GormTokenStore) DeleteExpired(ctx context.Context) error {
	return s.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&TokenModel{}).Error
}

func (s *GormTokenStore) DeleteByAccountID(ctx context.Context, accountID uuid.UUID, tokenType auth.TokenType) error {
	return s.db.WithContext(ctx).
		Where("account_id = ? AND type = ?", accountID, tokenType).
		Delete(&TokenModel{}).Error
}