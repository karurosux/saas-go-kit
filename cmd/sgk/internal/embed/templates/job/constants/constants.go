package jobconstants

import "time"

// Default values
const (
	// DefaultMaxRetries is the default maximum number of retries
	DefaultMaxRetries = 3
	
	// DefaultRetryDelay is the default delay between retries
	DefaultRetryDelay = 5 * time.Minute
	
	// DefaultJobTimeout is the default job execution timeout
	DefaultJobTimeout = 30 * time.Minute
	
	// DefaultStaleJobTimeout is when a running job is considered stale
	DefaultStaleJobTimeout = 1 * time.Hour
	
	// DefaultWorkerConcurrency is the default number of concurrent workers
	DefaultWorkerConcurrency = 5
	
	// DefaultPollInterval is the default interval for polling new jobs
	DefaultPollInterval = 5 * time.Second
	
	// DefaultResultRetention is how long to keep job results
	DefaultResultRetention = 7 * 24 * time.Hour
)

// Job types
const (
	JobTypeEmail          = "email"
	JobTypeWebhook        = "webhook"
	JobTypeDataExport     = "data_export"
	JobTypeDataImport     = "data_import"
	JobTypeReportGenerate = "report_generate"
	JobTypeCleanup        = "cleanup"
	JobTypeNotification   = "notification"
)

// Context keys
const (
	// ContextKeyJobID stores the current job ID in context
	ContextKeyJobID = "job_id"
	
	// ContextKeyJobType stores the job type in context
	ContextKeyJobType = "job_type"
	
	// ContextKeyJobWorker stores the worker ID in context
	ContextKeyJobWorker = "job_worker"
)