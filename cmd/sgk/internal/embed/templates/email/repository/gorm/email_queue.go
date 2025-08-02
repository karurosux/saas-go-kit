package emailgorm

import (
	"context"
	"time"

	"gorm.io/gorm"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

// EmailQueueRepository implements EmailQueue using GORM
type EmailQueueRepository struct {
	db *gorm.DB
}

// NewEmailQueueRepository creates a new email queue repository
func NewEmailQueueRepository(db *gorm.DB) *EmailQueueRepository {
	return &EmailQueueRepository{db: db}
}

// Enqueue adds an email to the queue
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

// Dequeue retrieves pending emails from the queue
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
	
	// Mark messages as sending
	for _, msg := range messages {
		r.db.WithContext(ctx).
			Model(msg).
			Update("status", emailinterface.StatusSending)
	}
	
	return messages, nil
}

// MarkAsSent marks an email as successfully sent
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

// MarkAsFailed marks an email as failed
func (r *EmailQueueRepository) MarkAsFailed(ctx context.Context, id uint, err error) error {
	var message emailinterface.EmailMessage
	if err := r.db.WithContext(ctx).First(&message, id).Error; err != nil {
		return err
	}
	
	message.Attempts++
	message.Error = err.Error()
	
	// If max attempts reached, mark as failed
	if message.Attempts >= message.MaxAttempts {
		message.Status = emailinterface.StatusFailed
	} else {
		message.Status = emailinterface.StatusPending
	}
	
	return r.db.WithContext(ctx).Save(&message).Error
}

// RetryFailed retries failed emails
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

// GetStatus retrieves the status of an email
func (r *EmailQueueRepository) GetStatus(ctx context.Context, id uint) (*emailinterface.EmailMessage, error) {
	var message emailinterface.EmailMessage
	err := r.db.WithContext(ctx).First(&message, id).Error
	return &message, err
}