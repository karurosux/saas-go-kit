package gorm

import (
	"context"
	"fmt"
	"time"

	"github.com/karurosux/saas-go-kit/job-go"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type jobStore struct {
	db *gorm.DB
}

func NewJobStore(db *gorm.DB) job.JobStore {
	return &jobStore{db: db}
}

func (s *jobStore) Create(ctx context.Context, j job.Job) error {
	model, ok := j.(*job.JobModel)
	if !ok {
		return fmt.Errorf("invalid job type: expected *job.JobModel")
	}
	return s.db.WithContext(ctx).Create(model).Error
}

func (s *jobStore) GetByID(ctx context.Context, id uuid.UUID) (job.Job, error) {
	var model job.JobModel
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, job.ErrJobNotFound
		}
		return nil, err
	}
	return &model, nil
}

func (s *jobStore) GetPending(ctx context.Context, limit int) ([]job.Job, error) {
	var models []job.JobModel
	err := s.db.WithContext(ctx).
		Where("status = ?", job.JobStatusPending).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	jobs := make([]job.Job, len(models))
	for i := range models {
		jobs[i] = &models[i]
	}
	return jobs, nil
}

func (s *jobStore) GetScheduledReady(ctx context.Context, now time.Time, limit int) ([]job.Job, error) {
	var models []job.JobModel
	err := s.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", job.JobStatusScheduled, now).
		Order("priority DESC, scheduled_at ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	jobs := make([]job.Job, len(models))
	for i := range models {
		jobs[i] = &models[i]
	}
	return jobs, nil
}

func (s *jobStore) GetByStatus(ctx context.Context, status job.JobStatus, limit int) ([]job.Job, error) {
	var models []job.JobModel
	err := s.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	jobs := make([]job.Job, len(models))
	for i := range models {
		jobs[i] = &models[i]
	}
	return jobs, nil
}

func (s *jobStore) GetByType(ctx context.Context, jobType string, status job.JobStatus, limit int) ([]job.Job, error) {
	var models []job.JobModel
	query := s.db.WithContext(ctx).Where("type = ?", jobType)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}

	jobs := make([]job.Job, len(models))
	for i := range models {
		jobs[i] = &models[i]
	}
	return jobs, nil
}

func (s *jobStore) UpdateStatus(ctx context.Context, id uuid.UUID, status job.JobStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
		"error":  errorMsg,
	}
	return s.db.WithContext(ctx).Model(&job.JobModel{}).Where("id = ?", id).Updates(updates).Error
}

func (s *jobStore) UpdateAttempts(ctx context.Context, id uuid.UUID, attempts int) error {
	return s.db.WithContext(ctx).Model(&job.JobModel{}).Where("id = ?", id).Update("attempts", attempts).Error
}

func (s *jobStore) MarkStarted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&job.JobModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     job.JobStatusRunning,
		"started_at": now,
	}).Error
}

func (s *jobStore) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&job.JobModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       job.JobStatusCompleted,
		"completed_at": now,
		"error":        "",
	}).Error
}

func (s *jobStore) MarkFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&job.JobModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       job.JobStatusFailed,
		"completed_at": now,
		"error":        errorMsg,
	}).Error
}

func (s *jobStore) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&job.JobModel{}).Error
}

func (s *jobStore) CleanupOld(ctx context.Context, before time.Time) error {
	return s.db.WithContext(ctx).
		Where("completed_at < ? AND status IN ?", before, []job.JobStatus{job.JobStatusCompleted, job.JobStatusFailed, job.JobStatusCanceled}).
		Delete(&job.JobModel{}).Error
}