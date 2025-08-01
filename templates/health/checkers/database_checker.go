package healthcheckers

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	
	"{{.Project.GoModule}}/internal/health/constants"
	"{{.Project.GoModule}}/internal/health/interface"
	"{{.Project.GoModule}}/internal/health/model"
	"gorm.io/gorm"
)

// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
	db               *gorm.DB
	name             string
	critical         bool
	connectionTimeout time.Duration
}

// NewDatabaseChecker creates a new database checker
func NewDatabaseChecker(db *gorm.DB, critical bool) healthinterface.DatabaseChecker {
	return &DatabaseChecker{
		db:                db,
		name:             healthconstants.CheckerDatabase,
		critical:         critical,
		connectionTimeout: 5 * time.Second,
	}
}

// Name returns the checker name
func (c *DatabaseChecker) Name() string {
	return c.name
}

// Critical returns if this check is critical
func (c *DatabaseChecker) Critical() bool {
	return c.critical
}

// SetConnectionTimeout sets the connection timeout
func (c *DatabaseChecker) SetConnectionTimeout(timeout time.Duration) {
	c.connectionTimeout = timeout
}

// Check performs the database health check
func (c *DatabaseChecker) Check(ctx context.Context) healthinterface.Check {
	start := time.Now()
	check := &healthmodel.Check{
		Name:        c.name,
		LastChecked: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	// Get underlying SQL database
	sqlDB, err := c.db.DB()
	if err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Failed to get database connection: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	
	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, c.connectionTimeout)
	defer cancel()
	
	// Ping database
	if err := sqlDB.PingContext(checkCtx); err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Database ping failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	
	// Get connection stats
	stats := sqlDB.Stats()
	check.Metadata["open_connections"] = stats.OpenConnections
	check.Metadata["in_use"] = stats.InUse
	check.Metadata["idle"] = stats.Idle
	check.Metadata["max_open_connections"] = stats.MaxOpenConnections
	
	// Check if connections are exhausted
	if stats.OpenConnections == stats.MaxOpenConnections && stats.Idle == 0 {
		check.Status = healthinterface.StatusDegraded
		check.Message = "Database connection pool exhausted"
	} else {
		check.Status = healthinterface.StatusOK
		check.Message = "Database is healthy"
	}
	
	// Test a simple query
	var result int
	if err := c.db.WithContext(checkCtx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Database query failed: %v", err)
	}
	
	check.Duration = time.Since(start)
	return check
}

// DatabaseCheckerWithQuery allows custom query for health check
type DatabaseCheckerWithQuery struct {
	DatabaseChecker
	query string
}

// NewDatabaseCheckerWithQuery creates a new database checker with custom query
func NewDatabaseCheckerWithQuery(db *gorm.DB, query string, critical bool) healthinterface.DatabaseChecker {
	return &DatabaseCheckerWithQuery{
		DatabaseChecker: DatabaseChecker{
			db:                db,
			name:             healthconstants.CheckerDatabase,
			critical:         critical,
			connectionTimeout: 5 * time.Second,
		},
		query: query,
	}
}

// Check performs the database health check with custom query
func (c *DatabaseCheckerWithQuery) Check(ctx context.Context) healthinterface.Check {
	start := time.Now()
	check := &healthmodel.Check{
		Name:        c.name,
		LastChecked: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, c.connectionTimeout)
	defer cancel()
	
	// Execute custom query
	var result sql.NullString
	if err := c.db.WithContext(checkCtx).Raw(c.query).Scan(&result).Error; err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Health check query failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	
	check.Status = healthinterface.StatusOK
	check.Message = "Database is healthy"
	if result.Valid {
		check.Metadata["query_result"] = result.String
	}
	
	check.Duration = time.Since(start)
	return check
}