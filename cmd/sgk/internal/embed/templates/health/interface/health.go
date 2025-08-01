package healthinterface

import (
	"context"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusOK       Status = "ok"
	StatusDegraded Status = "degraded"
	StatusDown     Status = "down"
)

// Check represents a single health check result
type Check interface {
	GetName() string
	GetStatus() Status
	GetMessage() string
	GetDuration() time.Duration
	GetLastChecked() time.Time
	GetMetadata() map[string]any
}

// Report represents the overall health report
type Report interface {
	GetStatus() Status
	GetVersion() string
	GetTimestamp() time.Time
	GetChecks() map[string]Check
	GetTotalChecks() int
	GetHealthyChecks() int
	GetMetadata() map[string]any
}

// Checker interface for implementing health checks
type Checker interface {
	// Name returns the name of the health check
	Name() string
	
	// Check performs the health check
	Check(ctx context.Context) Check
	
	// Critical returns true if this check is critical for overall health
	Critical() bool
}

// HealthService provides health check operations
type HealthService interface {
	// RegisterChecker registers a health checker
	RegisterChecker(checker Checker)
	
	// UnregisterChecker removes a health checker
	UnregisterChecker(name string)
	
	// Check performs a single health check by name
	Check(ctx context.Context, name string) (Check, error)
	
	// CheckAll performs all registered health checks
	CheckAll(ctx context.Context) Report
	
	// GetReport returns the last health report
	GetReport() Report
	
	// IsHealthy returns true if all critical checks are passing
	IsHealthy() bool
	
	// StartPeriodicChecks starts periodic health checks
	StartPeriodicChecks(ctx context.Context, interval time.Duration)
	
	// StopPeriodicChecks stops periodic health checks
	StopPeriodicChecks()
}

// DatabaseChecker checks database connectivity
type DatabaseChecker interface {
	Checker
	SetConnectionTimeout(timeout time.Duration)
}

// RedisChecker checks Redis connectivity
type RedisChecker interface {
	Checker
	SetPingTimeout(timeout time.Duration)
}

// HTTPChecker checks HTTP endpoint availability
type HTTPChecker interface {
	Checker
	SetEndpoint(url string)
	SetTimeout(timeout time.Duration)
	SetExpectedStatus(status int)
}

// DiskSpaceChecker checks available disk space
type DiskSpaceChecker interface {
	Checker
	SetPath(path string)
	SetThreshold(percentage float64)
}

// MemoryChecker checks memory usage
type MemoryChecker interface {
	Checker
	SetThreshold(percentage float64)
}
