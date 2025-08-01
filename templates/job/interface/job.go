package jobinterface

import (
	"context"
	"time"
	
	"github.com/google/uuid"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusCancelled JobStatus = "cancelled"
	StatusRetrying  JobStatus = "retrying"
)

// JobPriority represents job priority levels
type JobPriority int

const (
	PriorityLow    JobPriority = 0
	PriorityNormal JobPriority = 1
	PriorityHigh   JobPriority = 2
	PriorityUrgent JobPriority = 3
)

// Job represents a background job
type Job interface {
	GetID() uuid.UUID
	GetType() string
	GetPayload() map[string]interface{}
	GetStatus() JobStatus
	GetPriority() JobPriority
	GetRetries() int
	GetMaxRetries() int
	GetScheduledAt() time.Time
	GetStartedAt() *time.Time
	GetCompletedAt() *time.Time
	GetError() string
	GetCreatedBy() uuid.UUID
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetStatus(status JobStatus)
	SetStartedAt(startedAt *time.Time)
	SetCompletedAt(completedAt *time.Time)
	SetError(error string)
	IncrementRetries()
}

// JobResult represents the result of a job execution
type JobResult interface {
	GetID() uuid.UUID
	GetJobID() uuid.UUID
	GetStatus() JobStatus
	GetOutput() map[string]interface{}
	GetError() string
	GetDuration() time.Duration
	GetCreatedAt() time.Time
}

// JobHandler defines the interface for job handlers
type JobHandler interface {
	// Handle processes a job
	Handle(ctx context.Context, job Job) error
	
	// CanHandle checks if the handler can process the job type
	CanHandle(jobType string) bool
}

// JobQueue defines the interface for job queue operations
type JobQueue interface {
	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, job Job) error
	
	// Dequeue retrieves the next job from the queue
	Dequeue(ctx context.Context) (Job, error)
	
	// DequeueByType retrieves the next job of a specific type
	DequeueByType(ctx context.Context, jobType string) (Job, error)
	
	// Peek looks at the next job without removing it
	Peek(ctx context.Context) (Job, error)
	
	// Size returns the number of jobs in the queue
	Size(ctx context.Context) (int64, error)
	
	// Clear removes all jobs from the queue
	Clear(ctx context.Context) error
}

// JobRepository defines the interface for job persistence
type JobRepository interface {
	Create(ctx context.Context, job Job) error
	GetByID(ctx context.Context, id uuid.UUID) (Job, error)
	GetByStatus(ctx context.Context, status JobStatus, limit int) ([]Job, error)
	GetPending(ctx context.Context, limit int) ([]Job, error)
	GetScheduledBefore(ctx context.Context, before time.Time, limit int) ([]Job, error)
	Update(ctx context.Context, job Job) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByStatus(ctx context.Context, status JobStatus) (int64, error)
	GetStale(ctx context.Context, staleAfter time.Duration) ([]Job, error)
}

// JobResultRepository defines the interface for job result persistence
type JobResultRepository interface {
	Create(ctx context.Context, result JobResult) error
	GetByID(ctx context.Context, id uuid.UUID) (JobResult, error)
	GetByJobID(ctx context.Context, jobID uuid.UUID) ([]JobResult, error)
	GetRecent(ctx context.Context, limit int) ([]JobResult, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteOlderThan(ctx context.Context, before time.Time) error
}

// JobService defines the interface for job management
type JobService interface {
	// Job creation and management
	CreateJob(ctx context.Context, req CreateJobRequest) (Job, error)
	GetJob(ctx context.Context, jobID uuid.UUID) (Job, error)
	CancelJob(ctx context.Context, jobID uuid.UUID) error
	RetryJob(ctx context.Context, jobID uuid.UUID) error
	
	// Job execution
	ProcessNextJob(ctx context.Context) error
	ProcessJob(ctx context.Context, jobID uuid.UUID) error
	
	// Job queries
	GetJobsByStatus(ctx context.Context, status JobStatus, limit int) ([]Job, error)
	GetJobResults(ctx context.Context, jobID uuid.UUID) ([]JobResult, error)
	GetQueueStatus(ctx context.Context) (QueueStatus, error)
	
	// Handler registration
	RegisterHandler(jobType string, handler JobHandler)
	UnregisterHandler(jobType string)
}

// JobWorker defines the interface for background job processing
type JobWorker interface {
	// Start begins processing jobs
	Start(ctx context.Context) error
	
	// Stop gracefully stops the worker
	Stop() error
	
	// IsRunning checks if the worker is running
	IsRunning() bool
	
	// SetConcurrency sets the number of concurrent workers
	SetConcurrency(concurrency int)
}

// Request/Response types

type CreateJobRequest interface {
	GetType() string
	GetPayload() map[string]interface{}
	GetPriority() JobPriority
	GetScheduledAt() *time.Time
	GetCreatedBy() uuid.UUID
	Validate() error
}

type QueueStatus interface {
	GetPending() int64
	GetRunning() int64
	GetCompleted() int64
	GetFailed() int64
	GetTotal() int64
	GetOldestPendingAge() *time.Duration
}