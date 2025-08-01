package jobworker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	
	"{{.Project.GoModule}}/internal/job/constants"
	"{{.Project.GoModule}}/internal/job/interface"
)

// JobWorker implements background job processing
type JobWorker struct {
	service      jobinterface.JobService
	jobRepo      jobinterface.JobRepository
	concurrency  int
	pollInterval time.Duration
	running      atomic.Bool
	stopChan     chan struct{}
	wg           sync.WaitGroup
	workerID     string
}

// NewJobWorker creates a new job worker
func NewJobWorker(
	service jobinterface.JobService,
	jobRepo jobinterface.JobRepository,
	concurrency int,
	pollInterval time.Duration,
) jobinterface.JobWorker {
	if concurrency <= 0 {
		concurrency = jobconstants.DefaultWorkerConcurrency
	}
	if pollInterval <= 0 {
		pollInterval = jobconstants.DefaultPollInterval
	}
	
	return &JobWorker{
		service:      service,
		jobRepo:      jobRepo,
		concurrency:  concurrency,
		pollInterval: pollInterval,
		stopChan:     make(chan struct{}),
		workerID:     generateWorkerID(),
	}
}

// Start begins processing jobs
func (w *JobWorker) Start(ctx context.Context) error {
	if !w.running.CompareAndSwap(false, true) {
		return fmt.Errorf("worker already running")
	}
	
	fmt.Printf("Starting job worker %s with %d workers\n", w.workerID, w.concurrency)
	
	// Start worker goroutines
	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.runWorker(ctx, i)
	}
	
	// Start scheduler for processing scheduled jobs
	w.wg.Add(1)
	go w.runScheduler(ctx)
	
	// Start cleanup worker for stale jobs
	w.wg.Add(1)
	go w.runCleanupWorker(ctx)
	
	return nil
}

// Stop gracefully stops the worker
func (w *JobWorker) Stop() error {
	if !w.running.CompareAndSwap(true, false) {
		return fmt.Errorf("worker not running")
	}
	
	fmt.Printf("Stopping job worker %s\n", w.workerID)
	
	// Signal all workers to stop
	close(w.stopChan)
	
	// Wait for all workers to finish
	w.wg.Wait()
	
	// Reset stop channel for potential restart
	w.stopChan = make(chan struct{})
	
	fmt.Printf("Job worker %s stopped\n", w.workerID)
	
	return nil
}

// IsRunning checks if the worker is running
func (w *JobWorker) IsRunning() bool {
	return w.running.Load()
}

// SetConcurrency sets the number of concurrent workers
func (w *JobWorker) SetConcurrency(concurrency int) {
	if concurrency > 0 {
		w.concurrency = concurrency
	}
}

// runWorker runs a single worker goroutine
func (w *JobWorker) runWorker(ctx context.Context, workerNum int) {
	defer w.wg.Done()
	
	workerCtx := context.WithValue(ctx, jobconstants.ContextKeyJobWorker, fmt.Sprintf("%s-%d", w.workerID, workerNum))
	
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Try to process a job
			if err := w.service.ProcessNextJob(workerCtx); err != nil {
				// No job available or error processing
				// This is normal - just continue
			}
		}
	}
}

// runScheduler processes scheduled jobs
func (w *JobWorker) runScheduler(ctx context.Context) {
	defer w.wg.Done()
	
	ticker := time.NewTicker(time.Minute) // Check every minute
	defer ticker.Stop()
	
	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get jobs scheduled to run
			jobs, err := w.jobRepo.GetScheduledBefore(ctx, time.Now(), 100)
			if err != nil {
				fmt.Printf("Error getting scheduled jobs: %v\n", err)
				continue
			}
			
			// Process each scheduled job
			for _, job := range jobs {
				select {
				case <-w.stopChan:
					return
				case <-ctx.Done():
					return
				default:
					if err := w.service.ProcessJob(ctx, job.GetID()); err != nil {
						fmt.Printf("Error processing scheduled job %s: %v\n", job.GetID(), err)
					}
				}
			}
		}
	}
}

// runCleanupWorker cleans up stale jobs
func (w *JobWorker) runCleanupWorker(ctx context.Context) {
	defer w.wg.Done()
	
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get stale jobs
			staleJobs, err := w.jobRepo.GetStale(ctx, jobconstants.DefaultStaleJobTimeout)
			if err != nil {
				fmt.Printf("Error getting stale jobs: %v\n", err)
				continue
			}
			
			// Mark stale jobs as failed
			for _, job := range staleJobs {
				job.SetStatus(jobinterface.StatusFailed)
				job.SetError("Job timed out - marked as stale")
				now := time.Now()
				job.SetCompletedAt(&now)
				
				if err := w.jobRepo.Update(ctx, job); err != nil {
					fmt.Printf("Error updating stale job %s: %v\n", job.GetID(), err)
				}
			}
		}
	}
}

// generateWorkerID generates a unique worker ID
func generateWorkerID() string {
	return fmt.Sprintf("worker-%d", time.Now().UnixNano())
}