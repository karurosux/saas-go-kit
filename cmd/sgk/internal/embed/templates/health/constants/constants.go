package healthconstants

import "time"

const (
	DefaultCheckTimeout = 5 * time.Second
	
	DefaultPeriodicInterval = 30 * time.Second
	
	DefaultDiskSpaceThreshold = 90.0
	
	DefaultMemoryThreshold = 80.0
)

const (
	CheckerDatabase   = "database"
	CheckerRedis      = "redis"
	CheckerDiskSpace  = "disk_space"
	CheckerMemory     = "memory"
	CheckerHTTP       = "http"
)

const (
	ErrCheckerNotFound = "health checker not found"
	ErrCheckTimeout    = "health check timed out"
	ErrCheckFailed     = "health check failed"
)