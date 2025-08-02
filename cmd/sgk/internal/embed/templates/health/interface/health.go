package healthinterface

import (
	"context"
	"time"
)

type Status string

const (
	StatusOK       Status = "ok"
	StatusDegraded Status = "degraded"
	StatusDown     Status = "down"
)

type Check interface {
	GetName() string
	GetStatus() Status
	GetMessage() string
	GetDuration() time.Duration
	GetLastChecked() time.Time
	GetMetadata() map[string]any
}

type Report interface {
	GetStatus() Status
	GetVersion() string
	GetTimestamp() time.Time
	GetChecks() map[string]Check
	GetTotalChecks() int
	GetHealthyChecks() int
	GetMetadata() map[string]any
}

type Checker interface {
	Name() string
	
	Check(ctx context.Context) Check
	
	Critical() bool
}

type HealthService interface {
	RegisterChecker(checker Checker)
	
	UnregisterChecker(name string)
	
	Check(ctx context.Context, name string) (Check, error)
	
	CheckAll(ctx context.Context) Report
	
	GetReport() Report
	
	IsHealthy() bool
	
	StartPeriodicChecks(ctx context.Context, interval time.Duration)
	
	StopPeriodicChecks()
}

type DatabaseChecker interface {
	Checker
	SetConnectionTimeout(timeout time.Duration)
}

type RedisChecker interface {
	Checker
	SetPingTimeout(timeout time.Duration)
}

type HTTPChecker interface {
	Checker
	SetEndpoint(url string)
	SetTimeout(timeout time.Duration)
	SetExpectedStatus(status int)
}

type DiskSpaceChecker interface {
	Checker
	SetPath(path string)
	SetThreshold(percentage float64)
}

type MemoryChecker interface {
	Checker
	SetThreshold(percentage float64)
}
