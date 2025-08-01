package gorm

import (
	"context"
	"fmt"
	"time"

	"github.com/karurosux/saas-go-kit/job-go"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type jobResultStore struct {
	db *gorm.DB
}

func NewJobResultStore(db *gorm.DB) job.JobResultStore {
	return &jobResultStore{db: db}
}

func (s *jobResultStore) Create(ctx context.Context, result job.JobResult) error {
	model, ok := result.(*job.JobResultModel)
	if !ok {
		return fmt.Errorf("invalid job result type: expected *job.JobResultModel")
	}
	return s.db.WithContext(ctx).Create(model).Error
}

func (s *jobResultStore) GetByJobID(ctx context.Context, jobID uuid.UUID) (job.JobResult, error) {
	var model job.JobResultModel
	err := s.db.WithContext(ctx).Where("job_id = ?", jobID).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, job.ErrJobResultNotFound
		}
		return nil, err
	}
	return &model, nil
}

func (s *jobResultStore) GetByID(ctx context.Context, id uuid.UUID) (job.JobResult, error) {
	var model job.JobResultModel
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, job.ErrJobResultNotFound
		}
		return nil, err
	}
	return &model, nil
}

func (s *jobResultStore) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&job.JobResultModel{}).Error
}

func (s *jobResultStore) CleanupOld(ctx context.Context, before time.Time) error {
	return s.db.WithContext(ctx).Where("created_at < ?", before).Delete(&job.JobResultModel{}).Error
}