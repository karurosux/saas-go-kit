package emailgorm

import (
	"context"
	"time"

	"gorm.io/gorm"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

type EmailQueueRepository struct {
	db *gorm.DB
}

func NewEmailQueueRepository(db *gorm.DB) *EmailQueueRepository {
	return &EmailQueueRepository{db: db}
}

func (r *EmailQueueRepository) Enqueue(ctx context.Context, message *emailinterface.EmailMessage) error {
	if message.Status == "" {
		message.Status = emailinterface.StatusPending
	}
	if message.Priority == 0 {
		message.Priority = emailinterface.PriorityNormal
	}
	if message.MaxAttempts == 0 {
		message.MaxAttempts = 3
	}
	
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *EmailQueueRepository) Dequeue(ctx context.Context, limit int) ([]*emailinterface.EmailMessage, error) {
	var messages []*emailinterface.EmailMessage
	
	err := r.db.WithContext(ctx).
		Where("status = ? AND attempts < max_attempts", emailinterface.StatusPending).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&messages).Error
		
	if err != nil {
		return nil, err
	}
	
	for _, msg := range messages {
		r.db.WithContext(ctx).
			Model(msg).
			Update("status", emailinterface.StatusSending)
	}
	
	return messages, nil
}

func (r *EmailQueueRepository) MarkAsSent(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&emailinterface.EmailMessage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  emailinterface.StatusSent,
			"sent_at": &now,
		}).Error
}

func (r *EmailQueueRepository) MarkAsFailed(ctx context.Context, id uint, err error) error {
	var message emailinterface.EmailMessage
	if err := r.db.WithContext(ctx).First(&message, id).Error; err != nil {
		return err
	}
	
	message.Attempts++
	message.Error = err.Error()
	
	if message.Attempts >= message.MaxAttempts {
		message.Status = emailinterface.StatusFailed
	} else {
		message.Status = emailinterface.StatusPending
	}
	
	return r.db.WithContext(ctx).Save(&message).Error
}

func (r *EmailQueueRepository) RetryFailed(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&emailinterface.EmailMessage{}).
		Where("status = ? AND created_at > ?", emailinterface.StatusFailed, time.Now().Add(-24*time.Hour)).
		Updates(map[string]interface{}{
			"status":   emailinterface.StatusPending,
			"attempts": 0,
			"error":    "",
		}).Error
}

func (r *EmailQueueRepository) GetStatus(ctx context.Context, id uint) (*emailinterface.EmailMessage, error) {
	var message emailinterface.EmailMessage
	err := r.db.WithContext(ctx).First(&message, id).Error
	return &message, err
}