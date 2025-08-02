package healthcheckers

import (
	"context"
	"fmt"
	"syscall"
	"time"
	
	healthconstants "{{.Project.GoModule}}/internal/health/constants"
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
	healthmodel "{{.Project.GoModule}}/internal/health/model"
)

type DiskSpaceChecker struct {
	name      string
	path      string
	threshold float64
	critical  bool
}

func NewDiskSpaceChecker(path string, threshold float64, critical bool) healthinterface.DiskSpaceChecker {
	if threshold <= 0 || threshold > 100 {
		threshold = healthconstants.DefaultDiskSpaceThreshold
	}
	
	return &DiskSpaceChecker{
		name:      healthconstants.CheckerDiskSpace,
		path:      path,
		threshold: threshold,
		critical:  critical,
	}
}

func (c *DiskSpaceChecker) Name() string {
	return c.name
}

func (c *DiskSpaceChecker) Critical() bool {
	return c.critical
}

func (c *DiskSpaceChecker) SetPath(path string) {
	c.path = path
}

func (c *DiskSpaceChecker) SetThreshold(percentage float64) {
	if percentage > 0 && percentage <= 100 {
		c.threshold = percentage
	}
}

func (c *DiskSpaceChecker) Check(ctx context.Context) healthinterface.Check {
	start := time.Now()
	check := &healthmodel.Check{
		Name:        c.name,
		LastChecked: time.Now(),
		Metadata:    make(map[string]any),
	}
	
	var stat syscall.Statfs_t
	if err := syscall.Statfs(c.path, &stat); err != nil {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Failed to get disk stats: %v", err)
		check.Duration = time.Since(start)
		return check
	}
	
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	usedPercent := float64(used) / float64(total) * 100
	
	check.Metadata["path"] = c.path
	check.Metadata["total_bytes"] = total
	check.Metadata["used_bytes"] = used
	check.Metadata["free_bytes"] = free
	check.Metadata["used_percent"] = fmt.Sprintf("%.2f", usedPercent)
	check.Metadata["threshold_percent"] = c.threshold
	
	check.Metadata["total_human"] = humanizeBytes(total)
	check.Metadata["used_human"] = humanizeBytes(used)
	check.Metadata["free_human"] = humanizeBytes(free)
	
	if usedPercent >= c.threshold {
		check.Status = healthinterface.StatusDown
		check.Message = fmt.Sprintf("Disk usage %.2f%% exceeds threshold %.2f%%", usedPercent, c.threshold)
	} else if usedPercent >= (c.threshold - 10) {
		check.Status = healthinterface.StatusDegraded
		check.Message = fmt.Sprintf("Disk usage %.2f%% approaching threshold %.2f%%", usedPercent, c.threshold)
	} else {
		check.Status = healthinterface.StatusOK
		check.Message = fmt.Sprintf("Disk usage %.2f%% is healthy", usedPercent)
	}
	
	check.Duration = time.Since(start)
	return check
}

func humanizeBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
