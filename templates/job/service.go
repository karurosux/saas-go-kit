package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type service struct {
	store         JobStore
	resultStore   JobResultStore
	queue         JobQueue
	handlers      map[string]JobHandler
	handlersMu    sync.RWMutex
	workers       int
	pollInterval  time.Duration
	maxRetries    int
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	started       bool
	startedMu     sync.Mutex
}

type ServiceConfig struct {
	Store        JobStore
	ResultStore  JobResultStore
	Queue        JobQueue
	Workers      int
	PollInterval time.Duration
	MaxRetries   int
}

func NewService(config ServiceConfig) JobService {
	if config.Workers <= 0 {
		config.Workers = 5
	}
	if config.PollInterval <= 0 {
		config.PollInterval = 5 * time.Second
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}

	return &service{
		store:        config.Store,
		resultStore:  config.ResultStore,
		queue:        config.Queue,
		handlers:     make(map[string]JobHandler),
		workers:      config.Workers,
		pollInterval: config.PollInterval,
		maxRetries:   config.MaxRetries,
	}
}

func (s *service) RegisterHandler(jobType string, handler JobHandler) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.handlers[jobType] = handler
}

func (s *service) CreateJob(ctx context.Context, jobType string, payload map[string]interface{}, options ...JobOption) (Job, error) {
	opts := &JobOptions{
		Priority:    JobPriorityNormal,
		MaxAttempts: s.maxRetries,
	}
	for _, opt := range options {
		opt(opts)
	}

	var scheduledAt *time.Time
	if opts.Delay > 0 {
		t := time.Now().Add(opts.Delay)
		scheduledAt = &t
	}

	job := &JobModel{
		Type:        jobType,
		Payload:     payload,
		Status:      JobStatusPending,
		Priority:    opts.Priority,
		MaxAttempts: opts.MaxAttempts,
		ScheduledAt: scheduledAt,
	}

	if err := s.store.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	if scheduledAt == nil {
		if err := s.queue.Enqueue(ctx, job); err != nil {
			return nil, fmt.Errorf("failed to enqueue job: %w", err)
		}
	}

	return job, nil
}

func (s *service) ScheduleJob(ctx context.Context, jobType string, payload map[string]interface{}, scheduledAt time.Time, options ...JobOption) (Job, error) {
	opts := &JobOptions{
		Priority:    JobPriorityNormal,
		MaxAttempts: s.maxRetries,
	}
	for _, opt := range options {
		opt(opts)
	}

	job := &JobModel{
		Type:        jobType,
		Payload:     payload,
		Status:      JobStatusScheduled,
		Priority:    opts.Priority,
		MaxAttempts: opts.MaxAttempts,
		ScheduledAt: &scheduledAt,
	}

	if err := s.store.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create scheduled job: %w", err)
	}

	return job, nil
}

func (s *service) GetJob(ctx context.Context, id uuid.UUID) (Job, error) {
	return s.store.GetByID(ctx, id)
}

func (s *service) GetJobResult(ctx context.Context, jobID uuid.UUID) (JobResult, error) {
	return s.resultStore.GetByJobID(ctx, jobID)
}

func (s *service) GetJobsByType(ctx context.Context, jobType string, status JobStatus, limit int) ([]Job, error) {
	return s.store.GetByType(ctx, jobType, status, limit)
}

func (s *service) GetJobsByStatus(ctx context.Context, status JobStatus, limit int) ([]Job, error) {
	return s.store.GetByStatus(ctx, status, limit)
}

func (s *service) CancelJob(ctx context.Context, id uuid.UUID) error {
	job, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if job.GetStatus() == JobStatusRunning {
		return ErrJobAlreadyRunning
	}

	return s.store.UpdateStatus(ctx, id, JobStatusCanceled, "canceled by user")
}

func (s *service) RetryJob(ctx context.Context, id uuid.UUID) error {
	job, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if job.GetStatus() == JobStatusRunning {
		return ErrJobAlreadyRunning
	}

	if err := s.store.UpdateStatus(ctx, id, JobStatusPending, ""); err != nil {
		return err
	}

	if err := s.store.UpdateAttempts(ctx, id, 0); err != nil {
		return err
	}

	return s.queue.Enqueue(ctx, job)
}

func (s *service) DeleteJob(ctx context.Context, id uuid.UUID) error {
	return s.store.Delete(ctx, id)
}

func (s *service) Start(ctx context.Context) error {
	s.startedMu.Lock()
	defer s.startedMu.Unlock()

	if s.started {
		return nil
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.started = true

	// Start worker goroutines
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	// Start scheduler goroutine
	s.wg.Add(1)
	go s.scheduler()

	return nil
}

func (s *service) Stop(ctx context.Context) error {
	s.startedMu.Lock()
	defer s.startedMu.Unlock()

	if !s.started {
		return nil
	}

	s.cancel()
	s.wg.Wait()
	s.started = false

	return nil
}

func (s *service) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			job, err := s.queue.Dequeue(s.ctx)
			if err != nil {
				if err != ErrQueueEmpty {
					// Log error
				}
				time.Sleep(s.pollInterval)
				continue
			}

			if err := s.processJob(s.ctx, job); err != nil {
				// Log error
			}
		}
	}
}

func (s *service) scheduler() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.scheduleJobs()
		}
	}
}

func (s *service) scheduleJobs() {
	jobs, err := s.store.GetScheduledReady(s.ctx, time.Now(), 100)
	if err != nil {
		// Log error
		return
	}

	for _, job := range jobs {
		if err := s.store.UpdateStatus(s.ctx, job.GetID(), JobStatusPending, ""); err != nil {
			// Log error
			continue
		}

		if err := s.queue.Enqueue(s.ctx, job); err != nil {
			// Log error
			continue
		}
	}
}

func (s *service) processJob(ctx context.Context, job Job) error {
	// Mark job as running
	if err := s.store.MarkStarted(ctx, job.GetID()); err != nil {
		return err
	}

	// Get handler
	s.handlersMu.RLock()
	handler, exists := s.handlers[job.GetType()]
	s.handlersMu.RUnlock()

	if !exists {
		return s.store.MarkFailed(ctx, job.GetID(), ErrNoHandlerFound.Error())
	}

	// Execute handler
	result, err := handler(ctx, job.GetPayload())
	if err != nil {
		attempts := job.GetAttempts() + 1
		if err := s.store.UpdateAttempts(ctx, job.GetID(), attempts); err != nil {
			return err
		}

		if attempts < job.GetMaxAttempts() {
			// Requeue for retry
			if err := s.store.UpdateStatus(ctx, job.GetID(), JobStatusPending, err.Error()); err != nil {
				return err
			}
			return s.queue.Enqueue(ctx, job)
		}

		// Max attempts reached, mark as failed
		if err := s.store.MarkFailed(ctx, job.GetID(), err.Error()); err != nil {
			return err
		}

		// Store error result
		jobResult := &JobResultModel{
			JobID: job.GetID(),
			Error: err.Error(),
		}
		return s.resultStore.Create(ctx, jobResult)
	}

	// Mark job as completed
	if err := s.store.MarkCompleted(ctx, job.GetID()); err != nil {
		return err
	}

	// Store successful result
	jobResult := &JobResultModel{
		JobID:  job.GetID(),
		Result: result,
	}
	return s.resultStore.Create(ctx, jobResult)
}