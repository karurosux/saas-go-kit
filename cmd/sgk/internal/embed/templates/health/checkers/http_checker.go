package healthcheckers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
	
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
	healthmodel "{{.Project.GoModule}}/internal/health/model"
)

type HTTPChecker struct {
	name           string
	endpoint       string
	timeout        time.Duration
	expectedStatus int
	critical       bool
	client         *http.Client
}

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

func (c *HTTPChecker) Name() string {
	return c.name
}

func (c *HTTPChecker) Critical() bool {
	return c.critical
}

func (c *HTTPChecker) SetEndpoint(url string) {
	c.endpoint = url
}

func (c *HTTPChecker) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.client.Timeout = timeout
}

func (c *HTTPChecker) SetExpectedStatus(status int) {
	c.expectedStatus = status
}

func (c *HTTPChecker) Check(ctx context.Context) healthinterface.Check {
	start := time.Now()
	check := &healthmodel.Check{
		Name:        c.name,
		LastChecked: time.Now(),
		Metadata:    make(map[string]any),
	}
	
	check.Metadata["endpoint"] = c.endpoint
	check.Metadata["expected_status"] = c.expectedStatus
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Failed to create request: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Request failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	defer resp.Body.Close()
	
	io.Copy(io.Discard, resp.Body)
	
	check.Metadata["status_code"] = resp.StatusCode
	check.Metadata["status_text"] = resp.Status
	check.Metadata["content_length"] = resp.ContentLength
	
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		check.Metadata["content_type"] = contentType
	}
	
	if resp.StatusCode != c.expectedStatus {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Unexpected status code: got %d, expected %d", resp.StatusCode, c.expectedStatus)
	} else {
		check.Status = healthinterface.StatusOK
		check.Message = fmt.Sprintf("Endpoint returned expected status %d", resp.StatusCode)
	}
	
	responseTime := time.Since(start)
	check.Metadata["response_time_ms"] = responseTime.Milliseconds()
	
	if responseTime > 5*time.Second && check.Status == healthinterface.StatusOK {
		check.Status = healthinterface.StatusDegraded
		check.Message += fmt.Sprintf(" (slow response: %s)", responseTime)
	}
	
	check.Duration = responseTime
	return check
}
