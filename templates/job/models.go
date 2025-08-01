package job

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type JobModel struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Type         string      `json:"type" gorm:"not null;index"`
	Payload      JSONMap     `json:"payload" gorm:"type:jsonb"`
	Status       JobStatus   `json:"status" gorm:"not null;index;default:'pending'"`
	Priority     JobPriority `json:"priority" gorm:"not null;index;default:1"`
	ScheduledAt  *time.Time  `json:"scheduled_at" gorm:"index"`
	Attempts     int         `json:"attempts" gorm:"default:0"`
	MaxAttempts  int         `json:"max_attempts" gorm:"default:3"`
	Error        string      `json:"error,omitempty"`
	StartedAt    *time.Time  `json:"started_at"`
	CompletedAt  *time.Time  `json:"completed_at"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func (JobModel) TableName() string {
	return "jobs"
}

func (j *JobModel) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	if j.Status == "" {
		j.Status = JobStatusPending
	}
	if j.Priority == 0 {
		j.Priority = JobPriorityNormal
	}
	if j.MaxAttempts == 0 {
		j.MaxAttempts = 3
	}
	return nil
}

func (j *JobModel) GetID() uuid.UUID                       { return j.ID }
func (j *JobModel) GetType() string                        { return j.Type }
func (j *JobModel) GetPayload() map[string]interface{}     { return j.Payload }
func (j *JobModel) GetStatus() JobStatus                   { return j.Status }
func (j *JobModel) GetPriority() JobPriority               { return j.Priority }
func (j *JobModel) GetScheduledAt() *time.Time             { return j.ScheduledAt }
func (j *JobModel) GetAttempts() int                       { return j.Attempts }
func (j *JobModel) GetMaxAttempts() int                    { return j.MaxAttempts }
func (j *JobModel) GetCreatedAt() time.Time                { return j.CreatedAt }
func (j *JobModel) GetUpdatedAt() time.Time                { return j.UpdatedAt }
func (j *JobModel) GetStartedAt() *time.Time               { return j.StartedAt }
func (j *JobModel) GetCompletedAt() *time.Time             { return j.CompletedAt }
func (j *JobModel) GetError() string                       { return j.Error }

type JobResultModel struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	JobID     uuid.UUID `json:"job_id" gorm:"type:uuid;not null;index"`
	Result    JSONMap   `json:"result" gorm:"type:jsonb"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (JobResultModel) TableName() string {
	return "job_results"
}

func (j *JobResultModel) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

func (j *JobResultModel) GetID() uuid.UUID                   { return j.ID }
func (j *JobResultModel) GetJobID() uuid.UUID                { return j.JobID }
func (j *JobResultModel) GetResult() map[string]interface{}  { return j.Result }
func (j *JobResultModel) GetError() string                   { return j.Error }
func (j *JobResultModel) GetCreatedAt() time.Time            { return j.CreatedAt }