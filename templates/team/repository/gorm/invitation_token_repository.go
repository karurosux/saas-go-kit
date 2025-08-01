package gorm

import (
	"context"
	"errors"
	"time"
	
	"{{.Project.GoModule}}/internal/team/interface"
	"{{.Project.GoModule}}/internal/team/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InvitationTokenRepository implements invitation token repository using GORM
type InvitationTokenRepository struct {
	db *gorm.DB
}

// NewInvitationTokenRepository creates a new invitation token repository
func NewInvitationTokenRepository(db *gorm.DB) teaminterface.InvitationTokenRepository {
	return &InvitationTokenRepository{db: db}
}

// Create creates a new invitation token
func (r *InvitationTokenRepository) Create(ctx context.Context, token teaminterface.InvitationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByToken gets an invitation token by token string
func (r *InvitationTokenRepository) GetByToken(ctx context.Context, token string) (teaminterface.InvitationToken, error) {
	var inviteToken teammodel.InvitationToken
	err := r.db.WithContext(ctx).First(&inviteToken, "token = ?", token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation token not found")
		}
		return nil, err
	}
	return &inviteToken, nil
}

// GetByMemberID gets an invitation token by member ID
func (r *InvitationTokenRepository) GetByMemberID(ctx context.Context, memberID uuid.UUID) (teaminterface.InvitationToken, error) {
	var inviteToken teammodel.InvitationToken
	err := r.db.WithContext(ctx).
		Where("member_id = ? AND used = ?", memberID, false).
		Order("created_at DESC").
		First(&inviteToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation token not found")
		}
		return nil, err
	}
	return &inviteToken, nil
}

// Update updates an invitation token
func (r *InvitationTokenRepository) Update(ctx context.Context, token teaminterface.InvitationToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

// Delete deletes an invitation token
func (r *InvitationTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&teammodel.InvitationToken{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("invitation token not found")
	}
	return nil
}

// DeleteExpired deletes all expired invitation tokens
func (r *InvitationTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&teammodel.InvitationToken{}).Error
}