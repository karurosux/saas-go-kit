package jobmodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/job/interface"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// JobResult represents the result of a job execution
type JobResult struct {
	ID        uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	JobID     uuid.UUID              `json:"job_id" gorm:"type:uuid;not null;index"`
	Status    jobinterface.JobStatus `json:"status" gorm:"not null"`
	Output    datatypes.JSON         `json:"output" gorm:"type:jsonb"`
	Error     string                 `json:"error,omitempty"`
	Duration  int64                  `json:"duration" gorm:"not null"` // in milliseconds
	CreatedAt time.Time              `json:"created_at"`
}

// GetID returns the result ID
func (jr *JobResult) GetID() uuid.UUID {
	return jr.ID
}

// GetJobID returns the job ID
func (jr *JobResult) GetJobID() uuid.UUID {
	return jr.JobID
}

// GetStatus returns the job status
func (jr *JobResult) GetStatus() jobinterface.JobStatus {
	return jr.Status
}

// GetOutput returns the job output
func (jr *JobResult) GetOutput() map[string]interface{} {
	var output map[string]interface{}
	if jr.Output != nil {
		jr.Output.Scan(&output)
	}
	return output
}

// GetError returns the error message
func (jr *JobResult) GetError() string {
	return jr.Error
}

// GetDuration returns the execution duration
func (jr *JobResult) GetDuration() time.Duration {
	return time.Duration(jr.Duration) * time.Millisecond
}

// GetCreatedAt returns creation time
func (jr *JobResult) GetCreatedAt() time.Time {
	return jr.CreatedAt
}