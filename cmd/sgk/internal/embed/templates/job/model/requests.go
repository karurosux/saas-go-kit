package jobmodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/job/constants"
	"{{.Project.GoModule}}/internal/job/interface"
	"github.com/google/uuid"
)

// CreateJobRequest represents a request to create a job
type CreateJobRequest struct {
	Type        string                   `json:"type" validate:"required"`
	Payload     map[string]interface{}   `json:"payload"`
	Priority    jobinterface.JobPriority `json:"priority"`
	ScheduledAt *time.Time               `json:"scheduled_at,omitempty"`
	CreatedBy   uuid.UUID                `json:"created_by"`
}

// GetType returns the job type
func (r *CreateJobRequest) GetType() string {
	return r.Type
}

// GetPayload returns the job payload
func (r *CreateJobRequest) GetPayload() map[string]interface{} {
	if r.Payload == nil {
		return make(map[string]interface{})
	}
	return r.Payload
}

// GetPriority returns the job priority
func (r *CreateJobRequest) GetPriority() jobinterface.JobPriority {
	return r.Priority
}

// GetScheduledAt returns when to schedule the job
func (r *CreateJobRequest) GetScheduledAt() *time.Time {
	return r.ScheduledAt
}

// GetCreatedBy returns who created the job
func (r *CreateJobRequest) GetCreatedBy() uuid.UUID {
	return r.CreatedBy
}

// Validate validates the request
func (r *CreateJobRequest) Validate() error {
	if r.Type == "" {
		return core.BadRequest("job type is required")
	}
	
	if r.Priority < jobinterface.PriorityLow || r.Priority > jobinterface.PriorityUrgent {
		return core.BadRequest(jobconstants.ErrInvalidPriority)
	}
	
	if r.CreatedBy == uuid.Nil {
		return core.BadRequest("created_by is required")
	}
	
	return nil
}

// QueueStatus represents the job queue status
type QueueStatus struct {
	Pending          int64          `json:"pending"`
	Running          int64          `json:"running"`
	Completed        int64          `json:"completed"`
	Failed           int64          `json:"failed"`
	Total            int64          `json:"total"`
	OldestPendingAge *time.Duration `json:"oldest_pending_age,omitempty"`
}

// GetPending returns pending job count
func (qs *QueueStatus) GetPending() int64 {
	return qs.Pending
}

// GetRunning returns running job count
func (qs *QueueStatus) GetRunning() int64 {
	return qs.Running
}

// GetCompleted returns completed job count
func (qs *QueueStatus) GetCompleted() int64 {
	return qs.Completed
}

// GetFailed returns failed job count
func (qs *QueueStatus) GetFailed() int64 {
	return qs.Failed
}

// GetTotal returns total job count
func (qs *QueueStatus) GetTotal() int64 {
	return qs.Total
}

// GetOldestPendingAge returns age of oldest pending job
func (qs *QueueStatus) GetOldestPendingAge() *time.Duration {
	return qs.OldestPendingAge
}