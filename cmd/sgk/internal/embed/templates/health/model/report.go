package healthmodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/health/interface"
)

// Report represents the overall health report
type Report struct {
	Status       healthinterface.Status            `json:"status"`
	Version      string                            `json:"version,omitempty"`
	Timestamp    time.Time                         `json:"timestamp"`
	Checks       map[string]healthinterface.Check  `json:"checks"`
	TotalChecks  int                               `json:"total_checks"`
	HealthyChecks int                              `json:"healthy_checks"`
	Metadata     map[string]interface{}            `json:"metadata,omitempty"`
}

// GetStatus returns the overall status
func (r *Report) GetStatus() healthinterface.Status {
	return r.Status
}

// GetVersion returns the version
func (r *Report) GetVersion() string {
	return r.Version
}

// GetTimestamp returns the report timestamp
func (r *Report) GetTimestamp() time.Time {
	return r.Timestamp
}

// GetChecks returns all checks
func (r *Report) GetChecks() map[string]healthinterface.Check {
	return r.Checks
}

// GetTotalChecks returns the total number of checks
func (r *Report) GetTotalChecks() int {
	return r.TotalChecks
}

// GetHealthyChecks returns the number of healthy checks
func (r *Report) GetHealthyChecks() int {
	return r.HealthyChecks
}

// GetMetadata returns the report metadata
func (r *Report) GetMetadata() map[string]interface{} {
	return r.Metadata
}