package healthcheckers

import (
	"context"
	"fmt"
	"time"
	
	healthconstants "{{.Project.GoModule}}/internal/health/constants"
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
	healthmodel "{{.Project.GoModule}}/internal/health/model"
	"github.com/redis/go-redis/v9"
)

type RedisChecker struct {
	client      *redis.Client
	name        string
	critical    bool
	pingTimeout time.Duration
}

func NewRedisChecker(client *redis.Client, critical bool) healthinterface.RedisChecker {
	return &RedisChecker{
		client:      client,
		name:        healthconstants.CheckerRedis,
		critical:    critical,
		pingTimeout: 5 * time.Second,
	}
}

func (c *RedisChecker) Name() string {
	return c.name
}

func (c *RedisChecker) Critical() bool {
	return c.critical
}

func (c *RedisChecker) SetPingTimeout(timeout time.Duration) {
	c.pingTimeout = timeout
}

func (c *RedisChecker) Check(ctx context.Context) healthinterface.Check {
	start := time.Now()
	check := &healthmodel.Check{
		Name:        c.name,
		LastChecked: time.Now(),
		Metadata:    make(map[string]any),
	}
	
	checkCtx, cancel := context.WithTimeout(ctx, c.pingTimeout)
	defer cancel()
	
	if err := c.client.Ping(checkCtx).Err(); err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Redis ping failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	
	info, err := c.client.Info(checkCtx, "server", "clients", "memory", "stats").Result()
	if err != nil {
		check.Status = healthinterface.StatusDegraded
		check.Message = fmt.Sprintf("Failed to get Redis info: %v", err)
	} else {
		check.Metadata["info"] = parseRedisInfo(info)
		check.Status = healthinterface.StatusOK
		check.Message = "Redis is healthy"
	}
	
	testKey := fmt.Sprintf("health_check_%d", time.Now().UnixNano())
	testValue := "test"
	
	if err := c.client.Set(checkCtx, testKey, testValue, 10*time.Second).Err(); err != nil {
		check.Status = healthinterface.StatusDegraded
		check.Message = fmt.Sprintf("Redis SET operation failed: %v", err)
	} else {
		val, err := c.client.Get(checkCtx, testKey).Result()
		if err != nil || val != testValue {
			check.Status = healthinterface.StatusDegraded
			check.Message = "Redis GET operation failed or returned incorrect value"
		}
		
		c.client.Del(checkCtx, testKey)
	}
	
	poolStats := c.client.PoolStats()
	check.Metadata["pool_hits"] = poolStats.Hits
	check.Metadata["pool_misses"] = poolStats.Misses
	check.Metadata["pool_timeouts"] = poolStats.Timeouts
	check.Metadata["total_conns"] = poolStats.TotalConns
	check.Metadata["idle_conns"] = poolStats.IdleConns
	
	check.Duration = time.Since(start)
	return check
}

func parseRedisInfo(info string) map[string]any {
	result := make(map[string]any)
	
	lines := []string{}
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}
		
		parts := splitOnce(line, ':')
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			
			switch key {
			case "redis_version", "redis_mode", "used_memory_human", 
			     "connected_clients", "total_connections_received",
			     "instantaneous_ops_per_sec":
				result[key] = value
			}
		}
	}
	
	return result
}

func splitOnce(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
