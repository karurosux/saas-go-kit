package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultService implements the health Service interface
type DefaultService struct {
	mu              sync.RWMutex
	checkers        map[string]Checker
	lastReport      *Report
	version         string
	periodicCancel  context.CancelFunc
	periodicRunning bool
}

// NewService creates a new health service
func NewService(version string) Service {
	return &DefaultService{
		checkers:   make(map[string]Checker),
		version:    version,
		lastReport: nil,
	}
}

// RegisterChecker registers a health checker
func (s *DefaultService) RegisterChecker(checker Checker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.checkers[checker.Name()] = checker
}

// UnregisterChecker removes a health checker
func (s *DefaultService) UnregisterChecker(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.checkers, name)
}

// Check performs a single health check by name
func (s *DefaultService) Check(ctx context.Context, name string) (*Check, error) {
	s.mu.RLock()
	checker, exists := s.checkers[name]
	s.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("health check '%s' not found", name)
	}
	
	check := checker.Check(ctx)
	return &check, nil
}

// CheckAll performs all registered health checks
func (s *DefaultService) CheckAll(ctx context.Context) *Report {
	s.mu.RLock()
	checkersCopy := make(map[string]Checker, len(s.checkers))
	for name, checker := range s.checkers {
		checkersCopy[name] = checker
	}
	s.mu.RUnlock()
	
	report := &Report{
		Status:      StatusOK,
		Version:     s.version,
		Timestamp:   time.Now(),
		Checks:      make(map[string]Check),
		TotalChecks: len(checkersCopy),
		Healthy:     0,
		Metadata:    make(map[string]interface{}),
	}
	
	// Run all checks concurrently
	checkChan := make(chan struct {
		name  string
		check Check
	}, len(checkersCopy))
	
	var wg sync.WaitGroup
	for name, checker := range checkersCopy {
		wg.Add(1)
		go func(n string, c Checker) {
			defer wg.Done()
			
			// Create a timeout context for individual checks
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			
			check := c.Check(checkCtx)
			checkChan <- struct {
				name  string
				check Check
			}{name: n, check: check}
		}(name, checker)
	}
	
	// Wait for all checks to complete
	wg.Wait()
	close(checkChan)
	
	// Collect results
	for result := range checkChan {
		report.Checks[result.name] = result.check
		
		if result.check.Status == StatusOK {
			report.Healthy++
		} else if result.check.Status == StatusDegraded && report.Status == StatusOK {
			report.Status = StatusDegraded
		} else if result.check.Status == StatusDown {
			report.Status = StatusDown
		}
	}
	
	// Update last report
	s.mu.Lock()
	s.lastReport = report
	s.mu.Unlock()
	
	return report
}

// GetReport returns the last health report
func (s *DefaultService) GetReport() *Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.lastReport == nil {
		return &Report{
			Status:      StatusOK,
			Version:     s.version,
			Timestamp:   time.Now(),
			Checks:      make(map[string]Check),
			TotalChecks: 0,
			Healthy:     0,
			Metadata:    make(map[string]interface{}),
		}
	}
	
	return s.lastReport
}

// IsHealthy returns true if all checks are passing
func (s *DefaultService) IsHealthy() bool {
	report := s.GetReport()
	return report.Status == StatusOK
}

// StartPeriodicChecks starts periodic health checks
func (s *DefaultService) StartPeriodicChecks(ctx context.Context, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Stop existing periodic checks if running
	if s.periodicRunning && s.periodicCancel != nil {
		s.periodicCancel()
	}
	
	// Create new context with cancel
	periodicCtx, cancel := context.WithCancel(ctx)
	s.periodicCancel = cancel
	s.periodicRunning = true
	
	// Start periodic checker goroutine
	go func() {
		// Run initial check
		s.CheckAll(periodicCtx)
		
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-periodicCtx.Done():
				s.mu.Lock()
				s.periodicRunning = false
				s.mu.Unlock()
				return
			case <-ticker.C:
				s.CheckAll(periodicCtx)
			}
		}
	}()
}

// StopPeriodicChecks stops periodic health checks
func (s *DefaultService) StopPeriodicChecks() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.periodicRunning && s.periodicCancel != nil {
		s.periodicCancel()
		s.periodicRunning = false
	}
}