package job

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/job/constants"
	"{{.Project.GoModule}}/internal/job/controller"
	"{{.Project.GoModule}}/internal/job/interface"
	"{{.Project.GoModule}}/internal/job/repository/gorm"
	"{{.Project.GoModule}}/internal/job/service"
	"{{.Project.GoModule}}/internal/job/worker"
	"{{.Project.GoModule}}/internal/job/worker/handlers"
	"github.com/labstack/echo/v4"
	gormdb "gorm.io/gorm"
)

// RegisterModule registers the job module with the container
func RegisterModule(c core.Container) error {
	// Get dependencies from container
	e, ok := c.Get("echo").(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance not found in container")
	}
	
	db, ok := c.Get("db").(*gormdb.DB)
	if !ok {
		return fmt.Errorf("database instance not found in container")
	}
	
	// Run migrations
	if err := gorm.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to run job migrations: %w", err)
	}
	
	// Create repositories
	jobRepo := gorm.NewJobRepository(db)
	resultRepo := gorm.NewJobResultRepository(db)
	
	// Create queue
	queueSize := 10000
	if size := os.Getenv("JOB_QUEUE_SIZE"); size != "" {
		if s, err := strconv.Atoi(size); err == nil && s > 0 {
			queueSize = s
		}
	}
	queue := jobservice.NewMemoryQueue(queueSize)
	
	// Get configuration
	maxRetries := jobconstants.DefaultMaxRetries
	if retries := os.Getenv("JOB_MAX_RETRIES"); retries != "" {
		if r, err := strconv.Atoi(retries); err == nil && r > 0 {
			maxRetries = r
		}
	}
	
	jobTimeout := jobconstants.DefaultJobTimeout
	if timeout := os.Getenv("JOB_TIMEOUT"); timeout != "" {
		if t, err := time.ParseDuration(timeout); err == nil {
			jobTimeout = t
		}
	}
	
	// Create job service
	jobService := jobservice.NewJobService(
		jobRepo,
		resultRepo,
		queue,
		maxRetries,
		jobTimeout,
	)
	
	// Register default handlers
	registerDefaultHandlers(jobService)
	
	// Create controller
	jobController := jobcontroller.NewJobController(jobService)
	
	// Register routes
	jobController.RegisterRoutes(e, "/jobs")
	
	// Create and start worker if enabled
	if os.Getenv("JOB_WORKER_ENABLED") != "false" {
		concurrency := jobconstants.DefaultWorkerConcurrency
		if conc := os.Getenv("JOB_WORKER_CONCURRENCY"); conc != "" {
			if c, err := strconv.Atoi(conc); err == nil && c > 0 {
				concurrency = c
			}
		}
		
		pollInterval := jobconstants.DefaultPollInterval
		if interval := os.Getenv("JOB_POLL_INTERVAL"); interval != "" {
			if i, err := time.ParseDuration(interval); err == nil {
				pollInterval = i
			}
		}
		
		jobWorker := jobworker.NewJobWorker(
			jobService,
			jobRepo,
			concurrency,
			pollInterval,
		)
		
		// Start worker
		if err := jobWorker.Start(context.Background()); err != nil {
			return fmt.Errorf("failed to start job worker: %w", err)
		}
		
		// Register worker in container
		c.Set("job.worker", jobWorker)
		
		// Register shutdown handler
		c.OnShutdown(func() {
			jobWorker.Stop()
		})
	}
	
	// Register components in container for other modules to use
	c.Set("job.service", jobService)
	c.Set("job.jobRepository", jobRepo)
	c.Set("job.resultRepository", resultRepo)
	c.Set("job.queue", queue)
	
	// Start cleanup job for old results
	startCleanupJob(resultRepo)
	
	return nil
}

// registerDefaultHandlers registers built-in job handlers
func registerDefaultHandlers(service jobinterface.JobService) {
	// Register email handler
	service.RegisterHandler(jobconstants.JobTypeEmail, handlers.NewEmailHandler())
	
	// Add more default handlers as needed
}

// startCleanupJob starts a periodic cleanup job for old results
func startCleanupJob(resultRepo jobinterface.JobResultRepository) {
	retention := jobconstants.DefaultResultRetention
	if ret := os.Getenv("JOB_RESULT_RETENTION"); ret != "" {
		if r, err := time.ParseDuration(ret); err == nil {
			retention = r
		}
	}
	
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run daily
		defer ticker.Stop()
		
		for range ticker.C {
			ctx := context.Background()
			before := time.Now().Add(-retention)
			
			if err := resultRepo.DeleteOlderThan(ctx, before); err != nil {
				fmt.Printf("Error cleaning up old job results: %v\n", err)
			}
		}
	}()
}