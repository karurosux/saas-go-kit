package gorm

import (
	"context"
	"errors"
	"time"
	
	"{{.Project.GoModule}}/internal/job/interface"
	"{{.Project.GoModule}}/internal/job/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JobRepository implements job repository using GORM
type JobRepository struct {
	db *gorm.DB
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *gorm.DB) jobinterface.JobRepository {
	return &JobRepository{db: db}
}

// Create creates a new job
func (r *JobRepository) Create(ctx context.Context, job jobinterface.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

// GetByID gets a job by ID
func (r *JobRepository) GetByID(ctx context.Context, id uuid.UUID) (jobinterface.Job, error) {
	var job jobmodel.Job
	err := r.db.WithContext(ctx).First(&job, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found")
		}
		return nil, err
	}
	return &job, nil
}

// GetByStatus gets jobs by status
func (r *JobRepository) GetByStatus(ctx context.Context, status jobinterface.JobStatus, limit int) ([]jobinterface.Job, error) {
	var jobs []jobmodel.Job
	
	query := r.db.WithContext(ctx).Where("status = ?", status)
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Order("priority DESC, scheduled_at ASC").Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	result := make([]jobinterface.Job, len(jobs))
	for i, j := range jobs {
		job := j // Create a copy to avoid pointer issues
		result[i] = &job
	}
	return result, nil
}

// GetPending gets pending jobs
func (r *JobRepository) GetPending(ctx context.Context, limit int) ([]jobinterface.Job, error) {
	var jobs []jobmodel.Job
	
	query := r.db.WithContext(ctx).
		Where("status IN ? AND scheduled_at <= ?", 
			[]jobinterface.JobStatus{jobinterface.StatusPending, jobinterface.StatusRetrying},
			time.Now())
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Order("priority DESC, scheduled_at ASC").Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	result := make([]jobinterface.Job, len(jobs))
	for i, j := range jobs {
		job := j // Create a copy to avoid pointer issues
		result[i] = &job
	}
	return result, nil
}

// GetScheduledBefore gets jobs scheduled before a time
func (r *JobRepository) GetScheduledBefore(ctx context.Context, before time.Time, limit int) ([]jobinterface.Job, error) {
	var jobs []jobmodel.Job
	
	query := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at < ?", jobinterface.StatusPending, before)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Order("priority DESC, scheduled_at ASC").Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	result := make([]jobinterface.Job, len(jobs))
	for i, j := range jobs {
		job := j // Create a copy to avoid pointer issues
		result[i] = &job
	}
	return result, nil
}

// Update updates a job
func (r *JobRepository) Update(ctx context.Context, job jobinterface.Job) error {
	return r.db.WithContext(ctx).Save(job).Error
}

// Delete deletes a job
func (r *JobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&jobmodel.Job{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("job not found")
	}
	return nil
}

// CountByStatus counts jobs by status
func (r *JobRepository) CountByStatus(ctx context.Context, status jobinterface.JobStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&jobmodel.Job{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// GetStale gets stale running jobs
func (r *JobRepository) GetStale(ctx context.Context, staleAfter time.Duration) ([]jobinterface.Job, error) {
	var jobs []jobmodel.Job
	
	staleBefore := time.Now().Add(-staleAfter)
	
	err := r.db.WithContext(ctx).
		Where("status = ? AND started_at < ?", jobinterface.StatusRunning, staleBefore).
		Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	result := make([]jobinterface.Job, len(jobs))
	for i, j := range jobs {
		job := j // Create a copy to avoid pointer issues
		result[i] = &job
	}
	return result, nil
}