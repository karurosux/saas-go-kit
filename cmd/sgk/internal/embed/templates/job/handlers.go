package job

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type handlers struct {
	service JobService
}

func NewHandlers(service JobService) *handlers {
	return &handlers{service: service}
}

type CreateJobRequest struct {
	Type        string                 `json:"type" validate:"required"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    *JobPriority           `json:"priority,omitempty"`
	MaxAttempts *int                   `json:"max_attempts,omitempty"`
	Delay       *string                `json:"delay,omitempty"`       // Duration string like "5m", "1h"
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"` // RFC3339 time
}

type JobResponse struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Status      JobStatus              `json:"status"`
	Priority    JobPriority            `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type JobResultResponse struct {
	ID        uuid.UUID              `json:"id"`
	JobID     uuid.UUID              `json:"job_id"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

func jobToResponse(job Job) *JobResponse {
	return &JobResponse{
		ID:          job.GetID(),
		Type:        job.GetType(),
		Payload:     job.GetPayload(),
		Status:      job.GetStatus(),
		Priority:    job.GetPriority(),
		ScheduledAt: job.GetScheduledAt(),
		Attempts:    job.GetAttempts(),
		MaxAttempts: job.GetMaxAttempts(),
		Error:       job.GetError(),
		StartedAt:   job.GetStartedAt(),
		CompletedAt: job.GetCompletedAt(),
		CreatedAt:   job.GetCreatedAt(),
		UpdatedAt:   job.GetUpdatedAt(),
	}
}

func jobResultToResponse(result JobResult) *JobResultResponse {
	return &JobResultResponse{
		ID:        result.GetID(),
		JobID:     result.GetJobID(),
		Result:    result.GetResult(),
		Error:     result.GetError(),
		CreatedAt: result.GetCreatedAt(),
	}
}

func (h *handlers) CreateJob(c echo.Context) error {
	var req CreateJobRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var options []JobOption
	if req.Priority != nil {
		options = append(options, WithPriority(*req.Priority))
	}
	if req.MaxAttempts != nil {
		options = append(options, WithMaxAttempts(*req.MaxAttempts))
	}

	var job Job
	var err error

	if req.ScheduledAt != nil {
		job, err = h.service.ScheduleJob(c.Request().Context(), req.Type, req.Payload, *req.ScheduledAt, options...)
	} else if req.Delay != nil {
		duration, err := time.ParseDuration(*req.Delay)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid delay format")
		}
		options = append(options, WithDelay(duration))
		job, err = h.service.CreateJob(c.Request().Context(), req.Type, req.Payload, options...)
	} else {
		job, err = h.service.CreateJob(c.Request().Context(), req.Type, req.Payload, options...)
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, jobToResponse(job))
}

func (h *handlers) GetJob(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid job ID")
	}

	job, err := h.service.GetJob(c.Request().Context(), id)
	if err != nil {
		if err == ErrJobNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Job not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, jobToResponse(job))
}

func (h *handlers) GetJobResult(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid job ID")
	}

	result, err := h.service.GetJobResult(c.Request().Context(), id)
	if err != nil {
		if err == ErrJobResultNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Job result not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, jobResultToResponse(result))
}

func (h *handlers) ListJobs(c echo.Context) error {
	status := c.QueryParam("status")
	jobType := c.QueryParam("type")
	limitStr := c.QueryParam("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var jobs []Job
	var err error

	if jobType != "" && status != "" {
		jobs, err = h.service.GetJobsByType(c.Request().Context(), jobType, JobStatus(status), limit)
	} else if status != "" {
		jobs, err = h.service.GetJobsByStatus(c.Request().Context(), JobStatus(status), limit)
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "Either status or type parameter is required")
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	responses := make([]*JobResponse, len(jobs))
	for i, job := range jobs {
		responses[i] = jobToResponse(job)
	}

	return c.JSON(http.StatusOK, responses)
}

func (h *handlers) CancelJob(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid job ID")
	}

	if err := h.service.CancelJob(c.Request().Context(), id); err != nil {
		if err == ErrJobNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Job not found")
		}
		if err == ErrJobAlreadyRunning {
			return echo.NewHTTPError(http.StatusConflict, "Job is already running")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *handlers) RetryJob(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid job ID")
	}

	if err := h.service.RetryJob(c.Request().Context(), id); err != nil {
		if err == ErrJobNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Job not found")
		}
		if err == ErrJobAlreadyRunning {
			return echo.NewHTTPError(http.StatusConflict, "Job is already running")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *handlers) DeleteJob(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid job ID")
	}

	if err := h.service.DeleteJob(c.Request().Context(), id); err != nil {
		if err == ErrJobNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Job not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}