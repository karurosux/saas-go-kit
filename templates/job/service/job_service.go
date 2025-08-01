package jobservice

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/job/constants"
	"{{.Project.GoModule}}/internal/job/interface"
	"{{.Project.GoModule}}/internal/job/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// JobService implements job management
type JobService struct {
	jobRepo       jobinterface.JobRepository
	resultRepo    jobinterface.JobResultRepository
	queue         jobinterface.JobQueue
	handlers      map[string]jobinterface.JobHandler
	handlersMutex sync.RWMutex
	maxRetries    int
	jobTimeout    time.Duration
}

// NewJobService creates a new job service
func NewJobService(
	jobRepo jobinterface.JobRepository,
	resultRepo jobinterface.JobResultRepository,
	queue jobinterface.JobQueue,
	maxRetries int,
	jobTimeout time.Duration,
) jobinterface.JobService {
	if maxRetries <= 0 {
		maxRetries = jobconstants.DefaultMaxRetries
	}
	if jobTimeout <= 0 {
		jobTimeout = jobconstants.DefaultJobTimeout
	}
	
	return &JobService{
		jobRepo:    jobRepo,
		resultRepo: resultRepo,
		queue:      queue,
		handlers:   make(map[string]jobinterface.JobHandler),
		maxRetries: maxRetries,
		jobTimeout: jobTimeout,
	}
}

// CreateJob creates a new job
func (s *JobService) CreateJob(ctx context.Context, req jobinterface.CreateJobRequest) (jobinterface.Job, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Check if handler exists for job type
	s.handlersMutex.RLock()
	_, hasHandler := s.handlers[req.GetType()]
	s.handlersMutex.RUnlock()
	
	if !hasHandler {
		return nil, core.BadRequest(jobconstants.ErrNoHandlerRegistered)
	}
	
	// Create job
	payloadData, err := json.Marshal(req.GetPayload())
	if err != nil {
		return nil, core.BadRequest(jobconstants.ErrInvalidPayload)
	}
	
	scheduledAt := time.Now()
	if req.GetScheduledAt() != nil {
		scheduledAt = *req.GetScheduledAt()
	}
	
	job := &jobmodel.Job{
		ID:          uuid.New(),
		Type:        req.GetType(),
		Payload:     datatypes.JSON(payloadData),
		Status:      jobinterface.StatusPending,
		Priority:    req.GetPriority(),
		MaxRetries:  s.maxRetries,
		ScheduledAt: scheduledAt,
		CreatedBy:   req.GetCreatedBy(),
	}
	
	// Save to repository
	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, core.InternalServerError("failed to create job")
	}
	
	// Enqueue if scheduled for now or past
	if scheduledAt.Before(time.Now().Add(time.Minute)) {
		if err := s.queue.Enqueue(ctx, job); err != nil {
			// Log error but don't fail - job is saved and can be picked up later
			fmt.Printf("Failed to enqueue job %s: %v\n", job.ID, err)
		}
	}
	
	return job, nil
}

// GetJob gets a job by ID
func (s *JobService) GetJob(ctx context.Context, jobID uuid.UUID) (jobinterface.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, core.NotFound(jobconstants.ErrJobNotFound)
	}
	return job, nil
}

// CancelJob cancels a pending or running job
func (s *JobService) CancelJob(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return core.NotFound(jobconstants.ErrJobNotFound)
	}
	
	// Check if job can be cancelled
	status := job.GetStatus()
	if status == jobinterface.StatusCompleted {
		return core.BadRequest(jobconstants.ErrJobAlreadyCompleted)
	}
	if status == jobinterface.StatusCancelled {
		return core.BadRequest(jobconstants.ErrJobCancelled)
	}
	
	// Update job status
	job.SetStatus(jobinterface.StatusCancelled)
	now := time.Now()
	job.SetCompletedAt(&now)
	job.SetError("Job cancelled by user")
	
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return core.InternalServerError("failed to cancel job")
	}
	
	// Create result
	result := &jobmodel.JobResult{
		ID:       uuid.New(),
		JobID:    job.GetID(),
		Status:   jobinterface.StatusCancelled,
		Error:    "Job cancelled by user",
		Duration: 0,
	}
	
	s.resultRepo.Create(ctx, result)
	
	return nil
}

// RetryJob retries a failed job
func (s *JobService) RetryJob(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return core.NotFound(jobconstants.ErrJobNotFound)
	}
	
	// Check if job can be retried
	if job.GetStatus() != jobinterface.StatusFailed {
		return core.BadRequest("only failed jobs can be retried")
	}
	
	if job.GetRetries() >= job.GetMaxRetries() {
		return core.BadRequest(jobconstants.ErrMaxRetriesExceeded)
	}
	
	// Reset job for retry
	job.SetStatus(jobinterface.StatusPending)
	job.SetError("")
	job.SetStartedAt(nil)
	job.SetCompletedAt(nil)
	job.IncrementRetries()
	
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return core.InternalServerError("failed to retry job")
	}
	
	// Re-enqueue job
	if err := s.queue.Enqueue(ctx, job); err != nil {
		return core.InternalServerError("failed to enqueue job")
	}
	
	return nil
}

// ProcessNextJob processes the next job from the queue
func (s *JobService) ProcessNextJob(ctx context.Context) error {
	// Dequeue job
	job, err := s.queue.Dequeue(ctx)
	if err != nil {
		return err
	}
	
	return s.processJob(ctx, job)
}

// ProcessJob processes a specific job
func (s *JobService) ProcessJob(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return core.NotFound(jobconstants.ErrJobNotFound)
	}
	
	return s.processJob(ctx, job)
}

// processJob internal method to process a job
func (s *JobService) processJob(ctx context.Context, job jobinterface.Job) error {
	// Check job status
	if job.GetStatus() != jobinterface.StatusPending && job.GetStatus() != jobinterface.StatusRetrying {
		return core.BadRequest("job is not in pending or retrying status")
	}
	
	// Get handler
	s.handlersMutex.RLock()
	handler, exists := s.handlers[job.GetType()]
	s.handlersMutex.RUnlock()
	
	if !exists {
		job.SetStatus(jobinterface.StatusFailed)
		job.SetError(jobconstants.ErrNoHandlerRegistered)
		s.jobRepo.Update(ctx, job)
		return core.InternalServerError(jobconstants.ErrNoHandlerRegistered)
	}
	
	// Start job
	startTime := time.Now()
	job.SetStatus(jobinterface.StatusRunning)
	job.SetStartedAt(&startTime)
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return core.InternalServerError("failed to update job status")
	}
	
	// Create context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, s.jobTimeout)
	defer cancel()
	
	// Add job info to context
	jobCtx = context.WithValue(jobCtx, jobconstants.ContextKeyJobID, job.GetID())
	jobCtx = context.WithValue(jobCtx, jobconstants.ContextKeyJobType, job.GetType())
	
	// Execute job
	var jobErr error
	done := make(chan bool)
	
	go func() {
		jobErr = handler.Handle(jobCtx, job)
		done <- true
	}()
	
	select {
	case <-done:
		// Job completed
	case <-jobCtx.Done():
		// Job timed out
		jobErr = fmt.Errorf(jobconstants.ErrJobTimeout)
	}
	
	// Calculate duration
	duration := time.Since(startTime)
	completedAt := time.Now()
	
	// Update job status
	if jobErr != nil {
		job.SetStatus(jobinterface.StatusFailed)
		job.SetError(jobErr.Error())
		
		// Check if should retry
		if job.GetRetries() < job.GetMaxRetries() {
			job.SetStatus(jobinterface.StatusRetrying)
			// Re-enqueue with delay
			go func() {
				time.Sleep(jobconstants.DefaultRetryDelay)
				s.queue.Enqueue(context.Background(), job)
			}()
		}
	} else {
		job.SetStatus(jobinterface.StatusCompleted)
		job.SetCompletedAt(&completedAt)
	}
	
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return core.InternalServerError("failed to update job")
	}
	
	// Create result
	result := &jobmodel.JobResult{
		ID:       uuid.New(),
		JobID:    job.GetID(),
		Status:   job.GetStatus(),
		Duration: duration.Milliseconds(),
	}
	
	if jobErr != nil {
		result.Error = jobErr.Error()
	}
	
	if err := s.resultRepo.Create(ctx, result); err != nil {
		// Log error but don't fail
		fmt.Printf("Failed to create job result: %v\n", err)
	}
	
	return jobErr
}

// GetJobsByStatus gets jobs by status
func (s *JobService) GetJobsByStatus(ctx context.Context, status jobinterface.JobStatus, limit int) ([]jobinterface.Job, error) {
	return s.jobRepo.GetByStatus(ctx, status, limit)
}

// GetJobResults gets results for a job
func (s *JobService) GetJobResults(ctx context.Context, jobID uuid.UUID) ([]jobinterface.JobResult, error) {
	return s.resultRepo.GetByJobID(ctx, jobID)
}

// GetQueueStatus gets the queue status
func (s *JobService) GetQueueStatus(ctx context.Context) (jobinterface.QueueStatus, error) {
	status := &jobmodel.QueueStatus{}
	
	// Get counts by status
	var err error
	status.Pending, err = s.jobRepo.CountByStatus(ctx, jobinterface.StatusPending)
	if err != nil {
		return nil, core.InternalServerError("failed to get pending count")
	}
	
	status.Running, err = s.jobRepo.CountByStatus(ctx, jobinterface.StatusRunning)
	if err != nil {
		return nil, core.InternalServerError("failed to get running count")
	}
	
	status.Completed, err = s.jobRepo.CountByStatus(ctx, jobinterface.StatusCompleted)
	if err != nil {
		return nil, core.InternalServerError("failed to get completed count")
	}
	
	status.Failed, err = s.jobRepo.CountByStatus(ctx, jobinterface.StatusFailed)
	if err != nil {
		return nil, core.InternalServerError("failed to get failed count")
	}
	
	status.Total = status.Pending + status.Running + status.Completed + status.Failed
	
	// Get oldest pending job
	pendingJobs, err := s.jobRepo.GetPending(ctx, 1)
	if err == nil && len(pendingJobs) > 0 {
		age := time.Since(pendingJobs[0].GetCreatedAt())
		status.OldestPendingAge = &age
	}
	
	return status, nil
}

// RegisterHandler registers a job handler
func (s *JobService) RegisterHandler(jobType string, handler jobinterface.JobHandler) {
	s.handlersMutex.Lock()
	defer s.handlersMutex.Unlock()
	s.handlers[jobType] = handler
}

// UnregisterHandler unregisters a job handler
func (s *JobService) UnregisterHandler(jobType string) {
	s.handlersMutex.Lock()
	defer s.handlersMutex.Unlock()
	delete(s.handlers, jobType)
}