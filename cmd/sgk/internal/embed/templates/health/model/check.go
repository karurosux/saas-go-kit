package healthmodel

import (
	"time"
	
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
)

type Check struct {
	Name        string                 `json:"name"`
	Status      healthinterface.Status `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Duration    time.Duration          `json:"duration_ms"`
	LastChecked time.Time              `json:"last_checked"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

func (c *Check) GetName() string {
	return c.Name
}

func (c *Check) GetStatus() healthinterface.Status {
	return c.Status
}

func (c *Check) GetMessage() string {
	return c.Message
}

func (c *Check) GetDuration() time.Duration {
	return c.Duration
}

func (c *Check) GetLastChecked() time.Time {
	return c.LastChecked
}

func (c *Check) GetMetadata() map[string]any {
	return c.Metadata
}
