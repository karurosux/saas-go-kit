package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/auth-go"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) auth.TokenStore {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(ctx context.Context, token auth.VerificationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *TokenRepository) FindByToken(ctx context.Context, token string) (auth.VerificationToken, error) {
	var verificationToken auth.DefaultVerificationToken
	err := r.db.WithContext(ctx).First(&verificationToken, "token = ?", token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}
	
	return &verificationToken, nil
}

func (r *TokenRepository) MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&auth.DefaultVerificationToken{}).
		Where("id = ?", tokenID).
		Update("used_at", now)
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("token not found")
	}
	
	return nil
}

func (r *TokenRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&auth.DefaultVerificationToken{}).Error
}

func (r *TokenRepository) DeleteByAccountID(ctx context.Context, accountID uuid.UUID, tokenType auth.TokenType) error {
	return r.db.WithContext(ctx).
		Where("account_id = ? AND type = ?", accountID, tokenType).
		Delete(&auth.DefaultVerificationToken{}).Error
}