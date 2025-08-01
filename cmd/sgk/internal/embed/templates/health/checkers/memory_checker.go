package healthcheckers

import (
	"context"
	"fmt"
	"runtime"
	"time"
	
	healthconstants "{{.Project.GoModule}}/internal/health/constants"
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
	healthmodel "{{.Project.GoModule}}/internal/health/model"
)

type MemoryChecker struct {
	name      string
	threshold float64
	critical  bool
}

func NewMemoryChecker(threshold float64, critical bool) healthinterface.MemoryChecker {
	if threshold <= 0 || threshold > 100 {
		threshold = healthconstants.DefaultMemoryThreshold
	}
	
	return &MemoryChecker{
		name:      healthconstants.CheckerMemory,
		threshold: threshold,
		critical:  critical,
	}
}

func (c *MemoryChecker) Name() string {
	return c.name
}

func (c *MemoryChecker) Critical() bool {
	return c.critical
}

func (c *MemoryChecker) SetThreshold(percentage float64) {
	if percentage > 0 && percentage <= 100 {
		c.threshold = percentage
	}
}

func (c *MemoryChecker) Check(ctx context.Context) healthinterface.Check {
	start := time.Now()
	check := &healthmodel.Check{
		Name:        c.name,
		LastChecked: time.Now(),
		Metadata:    make(map[string]any),
	}
	
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	totalAlloc := memStats.TotalAlloc
	heapAlloc := memStats.HeapAlloc
	heapSys := memStats.HeapSys
	usedPercent := float64(heapAlloc) / float64(heapSys) * 100
	
	check.Metadata["heap_alloc_bytes"] = heapAlloc
	check.Metadata["heap_alloc_human"] = humanizeBytes(heapAlloc)
	check.Metadata["heap_sys_bytes"] = heapSys
	check.Metadata["heap_sys_human"] = humanizeBytes(heapSys)
	check.Metadata["total_alloc_bytes"] = totalAlloc
	check.Metadata["total_alloc_human"] = humanizeBytes(totalAlloc)
	check.Metadata["sys_bytes"] = memStats.Sys
	check.Metadata["sys_human"] = humanizeBytes(memStats.Sys)
	check.Metadata["num_gc"] = memStats.NumGC
	check.Metadata["gc_cpu_fraction"] = memStats.GCCPUFraction
	check.Metadata["goroutines"] = runtime.NumGoroutine()
	check.Metadata["used_percent"] = fmt.Sprintf("%.2f", usedPercent)
	check.Metadata["threshold_percent"] = c.threshold
	
	if usedPercent >= c.threshold {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Memory usage %.2f%% exceeds threshold %.2f%%", usedPercent, c.threshold)
	} else if usedPercent >= (c.threshold - 10) {
		check.Status = healthinterface.StatusDegraded
		check.Message = fmt.Sprintf("Memory usage %.2f%% approaching threshold %.2f%%", usedPercent, c.threshold)
	} else {
		check.Status = healthinterface.StatusOK
		check.Message = fmt.Sprintf("Memory usage %.2f%% is healthy", usedPercent)
	}
	
	if memStats.NumGC > 0 {
		avgGCPause := time.Duration(memStats.PauseTotalNs / uint64(memStats.NumGC))
		check.Metadata["avg_gc_pause_ns"] = avgGCPause.Nanoseconds()
		check.Metadata["avg_gc_pause_human"] = avgGCPause.String()
		
		if avgGCPause > 100*time.Millisecond && check.Status == healthinterface.StatusOK {
			check.Status = healthinterface.StatusDegraded
			check.Message += fmt.Sprintf(" (high GC pause time: %s)", avgGCPause)
		}
	}
	
	check.Duration = time.Since(start)
	return check
}
