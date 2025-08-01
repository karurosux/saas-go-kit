package healthconstants

import "time"

// Default values
const (
	// DefaultCheckTimeout is the default timeout for health checks
	DefaultCheckTimeout = 5 * time.Second
	
	// DefaultPeriodicInterval is the default interval for periodic checks
	DefaultPeriodicInterval = 30 * time.Second
	
	// DefaultDiskSpaceThreshold is the default disk space threshold (90%)
	DefaultDiskSpaceThreshold = 90.0
	
	// DefaultMemoryThreshold is the default memory threshold (80%)
	DefaultMemoryThreshold = 80.0
)

// Checker names
const (
	CheckerDatabase   = "database"
	CheckerRedis      = "redis"
	CheckerDiskSpace  = "disk_space"
	CheckerMemory     = "memory"
	CheckerHTTP       = "http"
)

// Error messages
const (
	ErrCheckerNotFound = "health checker not found"
	ErrCheckTimeout    = "health check timed out"
	ErrCheckFailed     = "health check failed"
)