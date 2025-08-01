package healthservice

import (
	"context"
	"errors"
	"sync"
	"time"
	
	healthconstants "{{.Project.GoModule}}/internal/health/constants"
	healthinterface "{{.Project.GoModule}}/internal/health/interface"
	healthmodel "{{.Project.GoModule}}/internal/health/model"
)

type HealthService struct {
	checkers       map[string]healthinterface.Checker
	lastReport     healthinterface.Report
	mu             sync.RWMutex
	version        string
	stopChan       chan struct{}
	periodicActive bool
}

func NewHealthService(version string) healthinterface.HealthService {
	return &HealthService{
		checkers: make(map[string]healthinterface.Checker),
		version:  version,
		stopChan: make(chan struct{}),
	}
}

func (s *HealthService) RegisterChecker(checker healthinterface.Checker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkers[checker.Name()] = checker
}

func (s *HealthService) UnregisterChecker(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.checkers, name)
}

func (s *HealthService) Check(ctx context.Context, name string) (healthinterface.Check, error) {
	s.mu.RLock()
	checker, exists := s.checkers[name]
	s.mu.RUnlock()
	
	if !exists {
		return nil, errors.New(healthconstants.ErrCheckerNotFound)
	}
	
	checkCtx, cancel := context.WithTimeout(ctx, healthconstants.DefaultCheckTimeout)
	defer cancel()
	
	checkChan := make(chan healthinterface.Check, 1)
	go func() {
		checkChan <- checker.Check(checkCtx)
	}()
	
	select {
	case check := <-checkChan:
		return check, nil
	case <-checkCtx.Done():
		return &healthmodel.Check{
			Name:        name,
			Status:      healthinterface.StatusDown,
			Message:     healthconstants.ErrCheckTimeout,
			Duration:    healthconstants.DefaultCheckTimeout,
			LastChecked: time.Now(),
		}, nil
	}
}

func (s *HealthService) CheckAll(ctx context.Context) healthinterface.Report {
	s.mu.RLock()
	checkersCopy := make(map[string]healthinterface.Checker, len(s.checkers))
	for k, v := range s.checkers {
		checkersCopy[k] = v
	}
	s.mu.RUnlock()
	
	checks := make(map[string]healthinterface.Check)
	checksChan := make(chan struct {
		name  string
		check healthinterface.Check
	}, len(checkersCopy))
	
	var wg sync.WaitGroup
	for name, checker := range checkersCopy {
		wg.Add(1)
		go func(n string, c healthinterface.Checker) {
			defer wg.Done()
			
			checkCtx, cancel := context.WithTimeout(ctx, healthconstants.DefaultCheckTimeout)
			defer cancel()
			
			start := time.Now()
			checkResult := c.Check(checkCtx)
			
			if check, ok := checkResult.(*healthmodel.Check); ok {
				check.Duration = time.Since(start)
				check.LastChecked = time.Now()
			}
			
			checksChan <- struct {
				name  string
				check healthinterface.Check
			}{name: n, check: checkResult}
		}(name, checker)
	}
	
	go func() {
		wg.Wait()
		close(checksChan)
	}()
	
	healthyCount := 0
	overallStatus := healthinterface.StatusOK
	hasCriticalFailure := false
	
	for result := range checksChan {
		checks[result.name] = result.check
		
		if result.check.GetStatus() == healthinterface.StatusOK {
			healthyCount++
		} else {
			if checker, ok := checkersCopy[result.name]; ok && checker.Critical() {
				hasCriticalFailure = true
				if result.check.GetStatus() == healthinterface.StatusDown {
					overallStatus = healthinterface.StatusDown
				} else if overallStatus != healthinterface.StatusDown {
					overallStatus = healthinterface.StatusDegraded
				}
			} else if !hasCriticalFailure && overallStatus == healthinterface.StatusOK {
				overallStatus = healthinterface.StatusDegraded
			}
		}
	}
	
	report := &healthmodel.Report{
		Status:        overallStatus,
		Version:       s.version,
		Timestamp:     time.Now(),
		Checks:        checks,
		TotalChecks:   len(checks),
		HealthyChecks: healthyCount,
		Metadata: map[string]any{
			"uptime": time.Since(getStartTime()).Seconds(),
		},
	}
	
	s.mu.Lock()
	s.lastReport = report
	s.mu.Unlock()
	
	return report
}

func (s *HealthService) GetReport() healthinterface.Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.lastReport == nil {
		return &healthmodel.Report{
			Status:    healthinterface.StatusOK,
			Version:   s.version,
			Timestamp: time.Now(),
			Checks:    make(map[string]healthinterface.Check),
		}
	}
	
	return s.lastReport
}

func (s *HealthService) IsHealthy() bool {
	report := s.GetReport()
	return report.GetStatus() == healthinterface.StatusOK
}

func (s *HealthService) StartPeriodicChecks(ctx context.Context, interval time.Duration) {
	s.mu.Lock()
	if s.periodicActive {
		s.mu.Unlock()
		return
	}
	s.periodicActive = true
	s.mu.Unlock()
	
	ticker := time.NewTicker(interval)
	
	s.CheckAll(ctx)
	
	go func() {
		for {
			select {
			case <-ticker.C:
				s.CheckAll(ctx)
			case <-s.stopChan:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				s.mu.Lock()
				s.periodicActive = false
				s.mu.Unlock()
				return
			}
		}
	}()
}

func (s *HealthService) StopPeriodicChecks() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.periodicActive {
		close(s.stopChan)
		s.stopChan = make(chan struct{})
		s.periodicActive = false
	}
}

var startTime = time.Now()

func getStartTime() time.Time {
	return startTime
}
