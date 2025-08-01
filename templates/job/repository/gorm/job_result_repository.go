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

// JobResultRepository implements job result repository using GORM
type JobResultRepository struct {
	db *gorm.DB
}

// NewJobResultRepository creates a new job result repository
func NewJobResultRepository(db *gorm.DB) jobinterface.JobResultRepository {
	return &JobResultRepository{db: db}
}

// Create creates a new job result
func (r *JobResultRepository) Create(ctx context.Context, result jobinterface.JobResult) error {
	return r.db.WithContext(ctx).Create(result).Error
}

// GetByID gets a job result by ID
func (r *JobResultRepository) GetByID(ctx context.Context, id uuid.UUID) (jobinterface.JobResult, error) {
	var result jobmodel.JobResult
	err := r.db.WithContext(ctx).First(&result, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job result not found")
		}
		return nil, err
	}
	return &result, nil
}

// GetByJobID gets job results by job ID
func (r *JobResultRepository) GetByJobID(ctx context.Context, jobID uuid.UUID) ([]jobinterface.JobResult, error) {
	var results []jobmodel.JobResult
	
	err := r.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Order("created_at DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	interfaceResults := make([]jobinterface.JobResult, len(results))
	for i, res := range results {
		result := res // Create a copy to avoid pointer issues
		interfaceResults[i] = &result
	}
	return interfaceResults, nil
}

// GetRecent gets recent job results
func (r *JobResultRepository) GetRecent(ctx context.Context, limit int) ([]jobinterface.JobResult, error) {
	var results []jobmodel.JobResult
	
	query := r.db.WithContext(ctx)
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Order("created_at DESC").Find(&results).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	interfaceResults := make([]jobinterface.JobResult, len(results))
	for i, res := range results {
		result := res // Create a copy to avoid pointer issues
		interfaceResults[i] = &result
	}
	return interfaceResults, nil
}

// Delete deletes a job result
func (r *JobResultRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&jobmodel.JobResult{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("job result not found")
	}
	return nil
}

// DeleteOlderThan deletes job results older than a specific time
func (r *JobResultRepository) DeleteOlderThan(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&jobmodel.JobResult{}).Error
}