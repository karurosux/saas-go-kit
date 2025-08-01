package healthmodel

import (
	"time"
	
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
)

type Report struct {
	Status       healthinterface.Status            `json:"status"`
	Version      string                            `json:"version,omitempty"`
	Timestamp    time.Time                         `json:"timestamp"`
	Checks       map[string]healthinterface.Check  `json:"checks"`
	TotalChecks  int                               `json:"total_checks"`
	HealthyChecks int                              `json:"healthy_checks"`
	Metadata     map[string]any            `json:"metadata,omitempty"`
}

func (r *Report) GetStatus() healthinterface.Status {
	return r.Status
}

func (r *Report) GetVersion() string {
	return r.Version
}

func (r *Report) GetTimestamp() time.Time {
	return r.Timestamp
}

func (r *Report) GetChecks() map[string]healthinterface.Check {
	return r.Checks
}

func (r *Report) GetTotalChecks() int {
	return r.TotalChecks
}

func (r *Report) GetHealthyChecks() int {
	return r.HealthyChecks
}

func (r *Report) GetMetadata() map[string]any {
	return r.Metadata
}
