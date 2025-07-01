package health

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
type Check struct {
	Name        string                 `json:"name"`
	Status      Status                 `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Duration    time.Duration          `json:"duration_ms"`
	LastChecked time.Time              `json:"last_checked"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Report represents the overall health report
type Report struct {
	Status      Status              `json:"status"`
	Version     string              `json:"version,omitempty"`
	Timestamp   time.Time           `json:"timestamp"`
	Checks      map[string]Check    `json:"checks"`
	TotalChecks int                 `json:"total_checks"`
	Healthy     int                 `json:"healthy_checks"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Checker interface for implementing health checks
type Checker interface {
	// Name returns the name of the health check
	Name() string
	
	// Check performs the health check
	Check(ctx context.Context) Check
}

// Service provides health check operations
type Service interface {
	// RegisterChecker registers a health checker
	RegisterChecker(checker Checker)
	
	// UnregisterChecker removes a health checker
	UnregisterChecker(name string)
	
	// Check performs a single health check by name
	Check(ctx context.Context, name string) (*Check, error)
	
	// CheckAll performs all registered health checks
	CheckAll(ctx context.Context) *Report
	
	// GetReport returns the last health report
	GetReport() *Report
	
	// IsHealthy returns true if all checks are passing
	IsHealthy() bool
	
	// StartPeriodicChecks starts periodic health checks
	StartPeriodicChecks(ctx context.Context, interval time.Duration)
	
	// StopPeriodicChecks stops periodic health checks
	StopPeriodicChecks()
}