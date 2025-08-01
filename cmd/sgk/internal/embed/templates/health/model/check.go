package healthmodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/health/interface"
)

// Check represents a single health check result
type Check struct {
	Name        string                 `json:"name"`
	Status      healthinterface.Status `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Duration    time.Duration          `json:"duration_ms"`
	LastChecked time.Time              `json:"last_checked"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GetName returns the check name
func (c *Check) GetName() string {
	return c.Name
}

// GetStatus returns the check status
func (c *Check) GetStatus() healthinterface.Status {
	return c.Status
}

// GetMessage returns the check message
func (c *Check) GetMessage() string {
	return c.Message
}

// GetDuration returns the check duration
func (c *Check) GetDuration() time.Duration {
	return c.Duration
}

// GetLastChecked returns when the check was last performed
func (c *Check) GetLastChecked() time.Time {
	return c.LastChecked
}

// GetMetadata returns the check metadata
func (c *Check) GetMetadata() map[string]interface{} {
	return c.Metadata
}