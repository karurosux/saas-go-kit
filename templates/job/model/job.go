package jobmodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/job/interface"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Job represents a background job
type Job struct {
	ID           uuid.UUID               `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Type         string                  `json:"type" gorm:"not null;index"`
	Payload      datatypes.JSON          `json:"payload" gorm:"type:jsonb"`
	Status       jobinterface.JobStatus  `json:"status" gorm:"not null;index"`
	Priority     jobinterface.JobPriority `json:"priority" gorm:"not null;index"`
	Retries      int                     `json:"retries" gorm:"default:0"`
	MaxRetries   int                     `json:"max_retries" gorm:"default:3"`
	ScheduledAt  time.Time               `json:"scheduled_at" gorm:"not null;index"`
	StartedAt    *time.Time              `json:"started_at,omitempty"`
	CompletedAt  *time.Time              `json:"completed_at,omitempty"`
	Error        string                  `json:"error,omitempty"`
	CreatedBy    uuid.UUID               `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

// GetID returns the job ID
func (j *Job) GetID() uuid.UUID {
	return j.ID
}

// GetType returns the job type
func (j *Job) GetType() string {
	return j.Type
}

// GetPayload returns the job payload
func (j *Job) GetPayload() map[string]interface{} {
	var payload map[string]interface{}
	if j.Payload != nil {
		j.Payload.Scan(&payload)
	}
	return payload
}

// GetStatus returns the job status
func (j *Job) GetStatus() jobinterface.JobStatus {
	return j.Status
}

// GetPriority returns the job priority
func (j *Job) GetPriority() jobinterface.JobPriority {
	return j.Priority
}

// GetRetries returns the number of retries
func (j *Job) GetRetries() int {
	return j.Retries
}

// GetMaxRetries returns the maximum number of retries
func (j *Job) GetMaxRetries() int {
	return j.MaxRetries
}

// GetScheduledAt returns when the job is scheduled
func (j *Job) GetScheduledAt() time.Time {
	return j.ScheduledAt
}

// GetStartedAt returns when the job started
func (j *Job) GetStartedAt() *time.Time {
	return j.StartedAt
}

// GetCompletedAt returns when the job completed
func (j *Job) GetCompletedAt() *time.Time {
	return j.CompletedAt
}

// GetError returns the job error
func (j *Job) GetError() string {
	return j.Error
}

// GetCreatedBy returns who created the job
func (j *Job) GetCreatedBy() uuid.UUID {
	return j.CreatedBy
}

// GetCreatedAt returns creation time
func (j *Job) GetCreatedAt() time.Time {
	return j.CreatedAt
}

// GetUpdatedAt returns last update time
func (j *Job) GetUpdatedAt() time.Time {
	return j.UpdatedAt
}

// SetStatus sets the job status
func (j *Job) SetStatus(status jobinterface.JobStatus) {
	j.Status = status
}

// SetStartedAt sets when the job started
func (j *Job) SetStartedAt(startedAt *time.Time) {
	j.StartedAt = startedAt
}

// SetCompletedAt sets when the job completed
func (j *Job) SetCompletedAt(completedAt *time.Time) {
	j.CompletedAt = completedAt
}

// SetError sets the job error
func (j *Job) SetError(error string) {
	j.Error = error
}

// IncrementRetries increments the retry count
func (j *Job) IncrementRetries() {
	j.Retries++
}