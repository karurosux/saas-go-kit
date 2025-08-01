package job

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCanceled  JobStatus = "canceled"
	JobStatusScheduled JobStatus = "scheduled"
)

type JobPriority int

const (
	JobPriorityLow    JobPriority = 0
	JobPriorityNormal JobPriority = 1
	JobPriorityHigh   JobPriority = 2
	JobPriorityUrgent JobPriority = 3
)

type Job interface {
	GetID() uuid.UUID
	GetType() string
	GetPayload() map[string]interface{}
	GetStatus() JobStatus
	GetPriority() JobPriority
	GetScheduledAt() *time.Time
	GetAttempts() int
	GetMaxAttempts() int
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetStartedAt() *time.Time
	GetCompletedAt() *time.Time
	GetError() string
}

type JobResult interface {
	GetID() uuid.UUID
	GetJobID() uuid.UUID
	GetResult() map[string]interface{}
	GetError() string
	GetCreatedAt() time.Time
}

type JobStore interface {
	Create(ctx context.Context, job Job) error
	GetByID(ctx context.Context, id uuid.UUID) (Job, error)
	GetPending(ctx context.Context, limit int) ([]Job, error)
	GetScheduledReady(ctx context.Context, now time.Time, limit int) ([]Job, error)
	GetByStatus(ctx context.Context, status JobStatus, limit int) ([]Job, error)
	GetByType(ctx context.Context, jobType string, status JobStatus, limit int) ([]Job, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status JobStatus, error string) error
	UpdateAttempts(ctx context.Context, id uuid.UUID, attempts int) error
	MarkStarted(ctx context.Context, id uuid.UUID) error
	MarkCompleted(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, error string) error
	Delete(ctx context.Context, id uuid.UUID) error
	CleanupOld(ctx context.Context, before time.Time) error
}

type JobResultStore interface {
	Create(ctx context.Context, result JobResult) error
	GetByJobID(ctx context.Context, jobID uuid.UUID) (JobResult, error)
	GetByID(ctx context.Context, id uuid.UUID) (JobResult, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CleanupOld(ctx context.Context, before time.Time) error
}

type JobQueue interface {
	Enqueue(ctx context.Context, job Job) error
	Dequeue(ctx context.Context) (Job, error)
	Requeue(ctx context.Context, job Job) error
	Size(ctx context.Context) (int, error)
	Clear(ctx context.Context) error
}

type JobProcessor interface {
	Process(ctx context.Context, job Job) error
}

type JobHandler func(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error)

type JobService interface {
	CreateJob(ctx context.Context, jobType string, payload map[string]interface{}, options ...JobOption) (Job, error)
	ScheduleJob(ctx context.Context, jobType string, payload map[string]interface{}, scheduledAt time.Time, options ...JobOption) (Job, error)
	GetJob(ctx context.Context, id uuid.UUID) (Job, error)
	GetJobResult(ctx context.Context, jobID uuid.UUID) (JobResult, error)
	GetJobsByType(ctx context.Context, jobType string, status JobStatus, limit int) ([]Job, error)
	GetJobsByStatus(ctx context.Context, status JobStatus, limit int) ([]Job, error)
	CancelJob(ctx context.Context, id uuid.UUID) error
	RetryJob(ctx context.Context, id uuid.UUID) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
	RegisterHandler(jobType string, handler JobHandler)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type JobOption func(*JobOptions)

type JobOptions struct {
	Priority    JobPriority
	MaxAttempts int
	Delay       time.Duration
}

func WithPriority(priority JobPriority) JobOption {
	return func(opts *JobOptions) {
		opts.Priority = priority
	}
}

func WithMaxAttempts(attempts int) JobOption {
	return func(opts *JobOptions) {
		opts.MaxAttempts = attempts
	}
}

func WithDelay(delay time.Duration) JobOption {
	return func(opts *JobOptions) {
		opts.Delay = delay
	}
}