package jobconstants

// Error messages
const (
	ErrJobNotFound         = "job not found"
	ErrJobAlreadyRunning   = "job is already running"
	ErrJobAlreadyCompleted = "job has already completed"
	ErrJobCancelled        = "job has been cancelled"
	ErrJobFailed           = "job execution failed"
	ErrInvalidJobType      = "invalid job type"
	ErrNoHandlerRegistered = "no handler registered for job type"
	ErrQueueFull           = "job queue is full"
	ErrWorkerNotRunning    = "worker is not running"
	ErrMaxRetriesExceeded  = "maximum retries exceeded"
	ErrJobTimeout          = "job execution timeout"
	ErrInvalidPriority     = "invalid job priority"
	ErrInvalidPayload      = "invalid job payload"
)