package jobcontroller

import (
	"net/http"
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/job/interface"
	"{{.Project.GoModule}}/internal/job/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JobController handles job management requests
type JobController struct {
	service jobinterface.JobService
}

// NewJobController creates a new job controller
func NewJobController(service jobinterface.JobService) *JobController {
	return &JobController{
		service: service,
	}
}

// RegisterRoutes registers all job-related routes
func (jc *JobController) RegisterRoutes(e *echo.Echo, basePath string) {
	group := e.Group(basePath)
	
	// Job management endpoints
	group.POST("", jc.CreateJob)
	group.GET("/:jobId", jc.GetJob)
	group.DELETE("/:jobId", jc.CancelJob)
	group.POST("/:jobId/retry", jc.RetryJob)
	
	// Job query endpoints
	group.GET("", jc.ListJobs)
	group.GET("/:jobId/results", jc.GetJobResults)
	group.GET("/queue/status", jc.GetQueueStatus)
}

// CreateJobRequest represents the API request to create a job
type CreateJobRequest struct {
	Type        string                   `json:"type" validate:"required"`
	Payload     map[string]interface{}   `json:"payload"`
	Priority    jobinterface.JobPriority `json:"priority"`
	ScheduledAt *time.Time               `json:"scheduled_at,omitempty"`
}

// CreateJob godoc
// @Summary Create a new job
// @Description Create a new background job
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body CreateJobRequest true "Job details"
// @Success 201 {object} jobmodel.Job
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /jobs [post]
func (jc *JobController) CreateJob(c echo.Context) error {
	// Get user ID from context (would come from auth middleware)
	userID := c.Get("user_id")
	if userID == nil {
		return core.Error(c, core.Unauthorized("user not authenticated"))
	}
	
	createdBy, err := uuid.Parse(userID.(string))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid user ID"))
	}
	
	var req CreateJobRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("invalid request body"))
	}
	
	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}
	
	jobReq := &jobmodel.CreateJobRequest{
		Type:        req.Type,
		Payload:     req.Payload,
		Priority:    req.Priority,
		ScheduledAt: req.ScheduledAt,
		CreatedBy:   createdBy,
	}
	
	job, err := jc.service.CreateJob(c.Request().Context(), jobReq)
	if err != nil {
		return core.Error(c, err)
	}
	
	return core.Created(c, job)
}

// GetJob godoc
// @Summary Get job details
// @Description Get details of a specific job
// @Tags jobs
// @Accept json
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} jobmodel.Job
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /jobs/{jobId} [get]
func (jc *JobController) GetJob(c echo.Context) error {
	jobID, err := uuid.Parse(c.Param("jobId"))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid job ID"))
	}
	
	job, err := jc.service.GetJob(c.Request().Context(), jobID)
	if err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, job)
}

// CancelJob godoc
// @Summary Cancel a job
// @Description Cancel a pending or running job
// @Tags jobs
// @Accept json
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /jobs/{jobId} [delete]
func (jc *JobController) CancelJob(c echo.Context) error {
	jobID, err := uuid.Parse(c.Param("jobId"))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid job ID"))
	}
	
	if err := jc.service.CancelJob(c.Request().Context(), jobID); err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Job cancelled successfully",
	})
}

// RetryJob godoc
// @Summary Retry a failed job
// @Description Retry a job that has failed
// @Tags jobs
// @Accept json
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /jobs/{jobId}/retry [post]
func (jc *JobController) RetryJob(c echo.Context) error {
	jobID, err := uuid.Parse(c.Param("jobId"))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid job ID"))
	}
	
	if err := jc.service.RetryJob(c.Request().Context(), jobID); err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, map[string]string{
		"message": "Job retry initiated successfully",
	})
}

// ListJobsRequest represents the request to list jobs
type ListJobsRequest struct {
	Status jobinterface.JobStatus `query:"status"`
	Limit  int                    `query:"limit"`
}

// ListJobs godoc
// @Summary List jobs
// @Description List jobs filtered by status
// @Tags jobs
// @Accept json
// @Produce json
// @Param status query string false "Job status filter"
// @Param limit query int false "Maximum number of jobs to return"
// @Success 200 {array} jobmodel.Job
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /jobs [get]
func (jc *JobController) ListJobs(c echo.Context) error {
	var req ListJobsRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("invalid request parameters"))
	}
	
	// Default limit
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}
	
	// Default to all statuses if not specified
	if req.Status == "" {
		req.Status = jobinterface.StatusPending
	}
	
	jobs, err := jc.service.GetJobsByStatus(c.Request().Context(), req.Status, req.Limit)
	if err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, jobs)
}

// GetJobResults godoc
// @Summary Get job results
// @Description Get execution results for a job
// @Tags jobs
// @Accept json
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {array} jobmodel.JobResult
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /jobs/{jobId}/results [get]
func (jc *JobController) GetJobResults(c echo.Context) error {
	jobID, err := uuid.Parse(c.Param("jobId"))
	if err != nil {
		return core.Error(c, core.BadRequest("invalid job ID"))
	}
	
	results, err := jc.service.GetJobResults(c.Request().Context(), jobID)
	if err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, results)
}

// GetQueueStatus godoc
// @Summary Get queue status
// @Description Get the current job queue status
// @Tags jobs
// @Accept json
// @Produce json
// @Success 200 {object} jobmodel.QueueStatus
// @Failure 500 {object} core.ErrorResponse
// @Router /jobs/queue/status [get]
func (jc *JobController) GetQueueStatus(c echo.Context) error {
	status, err := jc.service.GetQueueStatus(c.Request().Context())
	if err != nil {
		return core.Error(c, err)
	}
	
	return core.Success(c, status)
}