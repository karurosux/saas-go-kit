package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/team-go"
	"gorm.io/gorm"
)

type InvitationTokenRepository struct {
	db *gorm.DB
}

func NewInvitationTokenRepository(db *gorm.DB) team.InvitationTokenRepository {
	return &InvitationTokenRepository{db: db}
}

func (r *InvitationTokenRepository) Create(ctx context.Context, token *team.InvitationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *InvitationTokenRepository) FindByToken(ctx context.Context, token string) (*team.InvitationToken, error) {
	var invitationToken team.InvitationToken
	err := r.db.WithContext(ctx).
		First(&invitationToken, "token = ?", token).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, team.ErrInvitationTokenNotFound
		}
		return nil, err
	}
	
	return &invitationToken, nil
}

func (r *InvitationTokenRepository) FindByMemberID(ctx context.Context, memberID uuid.UUID) (*team.InvitationToken, error) {
	var invitationToken team.InvitationToken
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND used_at IS NULL", memberID).
		First(&invitationToken).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, team.ErrInvitationTokenNotFound
		}
		return nil, err
	}
	
	return &invitationToken, nil
}

func (r *InvitationTokenRepository) Update(ctx context.Context, token *team.InvitationToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *InvitationTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&team.InvitationToken{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return team.ErrInvitationTokenNotFound
	}
	
	return nil
}

func (r *InvitationTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&team.InvitationToken{}).Error
}