package health

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GormDatabaseChecker checks GORM database connectivity
type GormDatabaseChecker struct {
	name string
	db   *gorm.DB
}

// NewDatabaseChecker creates a new GORM database health checker
func NewDatabaseChecker(name string, db *gorm.DB) Checker {
	return &GormDatabaseChecker{
		name: name,
		db:   db,
	}
}

func (c *GormDatabaseChecker) Name() string {
	return c.name
}

func (c *GormDatabaseChecker) Check(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:        c.name,
		Status:      StatusOK,
		LastChecked: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Get underlying SQL database
	sqlDB, err := c.db.DB()
	if err != nil {
		check.Status = StatusDown
		check.Message = fmt.Sprintf("Failed to get database connection: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	// Ping database with context
	if err := sqlDB.PingContext(ctx); err != nil {
		check.Status = StatusDown
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

	// Check connection pool health
	if stats.OpenConnections > 0 && float64(stats.InUse)/float64(stats.MaxOpenConnections) > 0.9 {
		check.Status = StatusDegraded
		check.Message = "Database connection pool is nearly exhausted"
	}

	check.Duration = time.Since(start)
	return check
}