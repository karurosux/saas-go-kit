package authservice

import (
	"fmt"
	"sync"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
)

type StrategyRegistry struct {
	strategies map[string]authinterface.AuthStrategy
	mu         sync.RWMutex
}

func NewStrategyRegistry() authinterface.StrategyRegistry {
	return &StrategyRegistry{
		strategies: make(map[string]authinterface.AuthStrategy),
	}
}

func (r *StrategyRegistry) Register(strategy authinterface.AuthStrategy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := strategy.Name()
	if _, exists := r.strategies[name]; exists {
		return fmt.Errorf("strategy %s already registered", name)
	}
	
	r.strategies[name] = strategy
	return nil
}

func (r *StrategyRegistry) Get(name string) (authinterface.AuthStrategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	strategy, exists := r.strategies[name]
	if !exists {
		return nil, fmt.Errorf("strategy %s not found", name)
	}
	
	return strategy, nil
}

func (r *StrategyRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.strategies))
	for name := range r.strategies {
		names = append(names, name)
	}
	
	return names
}