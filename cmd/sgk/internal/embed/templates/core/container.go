package core

import (
	"fmt"
	"sync"
)

type Container struct {
	services map[string]any
	mu       sync.RWMutex
}

func NewContainer() *Container {
	return &Container{
		services: make(map[string]any),
	}
}

func (c *Container) Set(name string, service any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

func (c *Container) Get(name string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found", name)
	}

	return service, nil
}

func (c *Container) MustGet(name string) any {
	service, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return service
}

func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.services[name]
	return exists
}

