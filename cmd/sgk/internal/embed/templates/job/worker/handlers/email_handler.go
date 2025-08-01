package handlers

import (
	"context"
	"fmt"
	
	"{{.Project.GoModule}}/internal/job/constants"
	"{{.Project.GoModule}}/internal/job/interface"
)

// EmailHandler handles email sending jobs
type EmailHandler struct {
	// In production, this would have an email service dependency
}

// NewEmailHandler creates a new email handler
func NewEmailHandler() jobinterface.JobHandler {
	return &EmailHandler{}
}

// Handle processes an email job
func (h *EmailHandler) Handle(ctx context.Context, job jobinterface.Job) error {
	payload := job.GetPayload()
	
	// Extract email details
	to, ok := payload["to"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'to' field in payload")
	}
	
	subject, ok := payload["subject"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'subject' field in payload")
	}
	
	body, ok := payload["body"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'body' field in payload")
	}
	
	// In production, send actual email here
	fmt.Printf("Sending email to %s with subject: %s\n", to, subject)
	fmt.Printf("Body: %s\n", body)
	
	// Simulate email sending
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Email sent successfully
		return nil
	}
}

// CanHandle checks if this handler can process the job type
func (h *EmailHandler) CanHandle(jobType string) bool {
	return jobType == jobconstants.JobTypeEmail
}