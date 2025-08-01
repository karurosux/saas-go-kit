package gorm

import (
	"context"
	"errors"
	"time"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmodel "{{.Project.GoModule}}/internal/auth/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) authinterface.TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(ctx context.Context, token authinterface.Token) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *TokenRepository) GetByToken(ctx context.Context, token string) (authinterface.Token, error) {
	var t authmodel.Token
	err := r.db.WithContext(ctx).First(&t, "token = ?", token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) GetByAccountAndType(ctx context.Context, accountID uuid.UUID, tokenType authinterface.TokenType) ([]authinterface.Token, error) {
	var tokens []authmodel.Token
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND type = ?", accountID, tokenType).
		Order("created_at DESC").
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	
	result := make([]authinterface.Token, len(tokens))
	for i, t := range tokens {
		token := t
		result[i] = &token
	}
	return result, nil
}

func (r *TokenRepository) Update(ctx context.Context, token authinterface.Token) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *TokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&authmodel.Token{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("token not found")
	}
	return nil
}

func (r *TokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&authmodel.Token{}).Error
}

func (r *TokenRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&authmodel.Token{}).
		Where("id = ?", id).
		Update("used", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("token not found")
	}
	return nil
}