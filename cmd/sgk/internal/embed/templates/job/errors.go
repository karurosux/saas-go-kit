package job

import "errors"

var (
	ErrJobNotFound       = errors.New("job not found")
	ErrJobResultNotFound = errors.New("job result not found")
	ErrJobAlreadyRunning = errors.New("job is already running")
	ErrJobCanceled       = errors.New("job was canceled")
	ErrJobFailed         = errors.New("job failed")
	ErrInvalidJobType    = errors.New("invalid job type")
	ErrNoHandlerFound    = errors.New("no handler found for job type")
	ErrQueueEmpty        = errors.New("queue is empty")
	ErrServiceStopped    = errors.New("job service is stopped")
)