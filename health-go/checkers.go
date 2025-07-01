package health

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)


// HTTPChecker checks HTTP endpoint availability
type HTTPChecker struct {
	name   string
	url    string
	client *http.Client
}

// NewHTTPChecker creates a new HTTP health checker
func NewHTTPChecker(name, url string) Checker {
	return &HTTPChecker{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *HTTPChecker) Name() string {
	return c.name
}

func (c *HTTPChecker) Check(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:        c.name,
		Status:      StatusOK,
		LastChecked: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		check.Status = StatusDown
		check.Message = fmt.Sprintf("Failed to create request: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	resp, err := c.client.Do(req)
	if err != nil {
		check.Status = StatusDown
		check.Message = fmt.Sprintf("Request failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	defer resp.Body.Close()

	check.Metadata["status_code"] = resp.StatusCode
	check.Metadata["url"] = c.url

	if resp.StatusCode >= 500 {
		check.Status = StatusDown
		check.Message = fmt.Sprintf("Service returned %d", resp.StatusCode)
	} else if resp.StatusCode >= 400 {
		check.Status = StatusDegraded
		check.Message = fmt.Sprintf("Service returned %d", resp.StatusCode)
	}

	check.Duration = time.Since(start)
	return check
}

// CustomChecker allows for custom health check logic
type CustomChecker struct {
	name     string
	checkFn  func(ctx context.Context) (Status, string, map[string]interface{})
}

// NewCustomChecker creates a new custom health checker
func NewCustomChecker(name string, checkFn func(ctx context.Context) (Status, string, map[string]interface{})) Checker {
	return &CustomChecker{
		name:    name,
		checkFn: checkFn,
	}
}

func (c *CustomChecker) Name() string {
	return c.name
}

func (c *CustomChecker) Check(ctx context.Context) Check {
	start := time.Now()
	
	status, message, metadata := c.checkFn(ctx)
	
	return Check{
		Name:        c.name,
		Status:      status,
		Message:     message,
		Duration:    time.Since(start),
		LastChecked: time.Now(),
		Metadata:    metadata,
	}
}

// SQLDatabaseChecker checks raw SQL database connectivity
type SQLDatabaseChecker struct {
	name string
	db   *sql.DB
}

// NewSQLDatabaseChecker creates a new SQL database health checker
func NewSQLDatabaseChecker(name string, db *sql.DB) Checker {
	return &SQLDatabaseChecker{
		name: name,
		db:   db,
	}
}

func (c *SQLDatabaseChecker) Name() string {
	return c.name
}

func (c *SQLDatabaseChecker) Check(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:        c.name,
		Status:      StatusOK,
		LastChecked: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Ping database
	if err := c.db.PingContext(ctx); err != nil {
		check.Status = StatusDown
		check.Message = fmt.Sprintf("Database ping failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	// Get connection stats
	stats := c.db.Stats()
	check.Metadata["open_connections"] = stats.OpenConnections
	check.Metadata["in_use"] = stats.InUse
	check.Metadata["idle"] = stats.Idle
	check.Metadata["max_open_connections"] = stats.MaxOpenConnections

	// Check connection pool health
	if stats.OpenConnections > 0 && float64(stats.InUse)/float64(stats.MaxOpenConnections) > 0.9 {
		check.Status = StatusDegraded
		check.Message = "Database connection pool is nearly exhausted"
	}

	check.Duration = time.Since(start)
	return check
}