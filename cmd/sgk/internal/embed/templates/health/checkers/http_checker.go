package healthcheckers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
	
	"{{.Project.GoModule}}/internal/health/constants"
	"{{.Project.GoModule}}/internal/health/interface"
	"{{.Project.GoModule}}/internal/health/model"
)

// HTTPChecker checks HTTP endpoint availability
type HTTPChecker struct {
	name           string
	endpoint       string
	timeout        time.Duration
	expectedStatus int
	critical       bool
	client         *http.Client
}

// NewHTTPChecker creates a new HTTP checker
func NewHTTPChecker(name, endpoint string, critical bool) healthinterface.HTTPChecker {
	return &HTTPChecker{
		name:           name,
		endpoint:       endpoint,
		timeout:        10 * time.Second,
		expectedStatus: http.StatusOK,
		critical:       critical,
		client:         &http.Client{},
	}
}

// Name returns the checker name
func (c *HTTPChecker) Name() string {
	return c.name
}

// Critical returns if this check is critical
func (c *HTTPChecker) Critical() bool {
	return c.critical
}

// SetEndpoint sets the endpoint URL
func (c *HTTPChecker) SetEndpoint(url string) {
	c.endpoint = url
}

// SetTimeout sets the request timeout
func (c *HTTPChecker) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.client.Timeout = timeout
}

// SetExpectedStatus sets the expected HTTP status code
func (c *HTTPChecker) SetExpectedStatus(status int) {
	c.expectedStatus = status
}

// Check performs the HTTP health check
func (c *HTTPChecker) Check(ctx context.Context) healthinterface.Check {
	start := time.Now()
	check := &healthmodel.Check{
		Name:        c.name,
		LastChecked: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	// Store endpoint in metadata
	check.Metadata["endpoint"] = c.endpoint
	check.Metadata["expected_status"] = c.expectedStatus
	
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Failed to create request: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	
	// Perform request
	resp, err := c.client.Do(req)
	if err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Request failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	defer resp.Body.Close()
	
	// Read and discard body to ensure connection can be reused
	io.Copy(io.Discard, resp.Body)
	
	// Store response metadata
	check.Metadata["status_code"] = resp.StatusCode
	check.Metadata["status_text"] = resp.Status
	check.Metadata["content_length"] = resp.ContentLength
	
	// Add response headers if relevant
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		check.Metadata["content_type"] = contentType
	}
	
	// Check status code
	if resp.StatusCode != c.expectedStatus {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Unexpected status code: got %d, expected %d", resp.StatusCode, c.expectedStatus)
	} else {
		check.Status = healthinterface.StatusOK
		check.Message = fmt.Sprintf("Endpoint returned expected status %d", resp.StatusCode)
	}
	
	// Calculate response time
	responseTime := time.Since(start)
	check.Metadata["response_time_ms"] = responseTime.Milliseconds()
	
	// If response is slow, mark as degraded
	if responseTime > 5*time.Second && check.Status == healthinterface.StatusOK {
		check.Status = healthinterface.StatusDegraded
		check.Message += fmt.Sprintf(" (slow response: %s)", responseTime)
	}
	
	check.Duration = responseTime
	return check
}