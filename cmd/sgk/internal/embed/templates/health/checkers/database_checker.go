package healthcheckers

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	
	healthconstants "{{.Project.GoModule}}/internal/health/constants"
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
	healthmodel "{{.Project.GoModule}}/internal/health/model"
	"gorm.io/gorm"
)

type DatabaseChecker struct {
	db               *gorm.DB
	name             string
	critical         bool
	connectionTimeout time.Duration
}

func NewDatabaseChecker(db *gorm.DB, critical bool) healthinterface.DatabaseChecker {
	return &DatabaseChecker{
		db:                db,
		name:             healthconstants.CheckerDatabase,
		critical:         critical,
		connectionTimeout: 5 * time.Second,
	}
}

func (c *DatabaseChecker) Name() string {
	return c.name
}

func (c *DatabaseChecker) Critical() bool {
	return c.critical
}

func (c *DatabaseChecker) SetConnectionTimeout(timeout time.Duration) {
	c.connectionTimeout = timeout
}

func (c *DatabaseChecker) Check(ctx context.Context) healthinterface.Check {
	start := time.Now()
	check := &healthmodel.Check{
		Name:        c.name,
		LastChecked: time.Now(),
		Metadata:    make(map[string]any),
	}
	
	sqlDB, err := c.db.DB()
	if err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Failed to get database connection: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	
	checkCtx, cancel := context.WithTimeout(ctx, c.connectionTimeout)
	defer cancel()
	
	if err := sqlDB.PingContext(checkCtx); err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Database ping failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	
	stats := sqlDB.Stats()
	check.Metadata["open_connections"] = stats.OpenConnections
	check.Metadata["in_use"] = stats.InUse
	check.Metadata["idle"] = stats.Idle
	check.Metadata["max_open_connections"] = stats.MaxOpenConnections
	
	if stats.OpenConnections == stats.MaxOpenConnections && stats.Idle == 0 {
		check.Status = healthinterface.StatusDegraded
		check.Message = "Database connection pool exhausted"
	} else {
		check.Status = healthinterface.StatusOK
		check.Message = "Database is healthy"
	}
	
	var result int
	if err := c.db.WithContext(checkCtx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Database query failed: %v", err)
	}
	
	check.Duration = time.Since(start)
	return check
}

type DatabaseCheckerWithQuery struct {
	DatabaseChecker
	query string
}

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

func (c *DatabaseCheckerWithQuery) Check(ctx context.Context) healthinterface.Check {
	start := time.Now()
	check := &healthmodel.Check{
		Name:        c.name,
		LastChecked: time.Now(),
		Metadata:    make(map[string]any),
	}
	
	checkCtx, cancel := context.WithTimeout(ctx, c.connectionTimeout)
	defer cancel()
	
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
