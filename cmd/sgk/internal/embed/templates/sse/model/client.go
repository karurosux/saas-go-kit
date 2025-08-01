package ssemodel

import (
	"context"
	"sync"
	"time"
	
	"{{.Project.GoModule}}/internal/sse/interface"
	"github.com/google/uuid"
)

// client implements the Client interface
type client struct {
	id          string
	userID      uuid.UUID
	channel     chan sseinterface.Event
	ctx         context.Context
	cancel      context.CancelFunc
	metadata    map[string]interface{}
	mu          sync.RWMutex
	connected   bool
	connectedAt time.Time
}

// NewClient creates a new SSE client
func NewClient(id string, userID uuid.UUID, bufferSize int) sseinterface.Client {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &client{
		id:          id,
		userID:      userID,
		channel:     make(chan sseinterface.Event, bufferSize),
		ctx:         ctx,
		cancel:      cancel,
		metadata:    make(map[string]interface{}),
		connected:   true,
		connectedAt: time.Now(),
	}
}

func (c *client) GetID() string {
	return c.id
}

func (c *client) GetUserID() uuid.UUID {
	return c.userID
}

func (c *client) GetContext() context.Context {
	return c.ctx
}

func (c *client) GetChannel() chan sseinterface.Event {
	return c.channel
}

func (c *client) GetMetadata() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Return a copy to prevent concurrent access issues
	metadata := make(map[string]interface{})
	for k, v := range c.metadata {
		metadata[k] = v
	}
	return metadata
}

func (c *client) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.metadata[key] = value
}

func (c *client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.connected
}

func (c *client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.connected {
		return nil
	}
	
	c.connected = false
	c.cancel()
	close(c.channel)
	
	return nil
}

// GetConnectedDuration returns how long the client has been connected
func (c *client) GetConnectedDuration() time.Duration {
	return time.Since(c.connectedAt)
}

// SendEvent sends an event to the client (non-blocking)
func (c *client) SendEvent(event sseinterface.Event) bool {
	if !c.IsConnected() {
		return false
	}
	
	select {
	case c.channel <- event:
		return true
	default:
		// Channel is full, event dropped
		return false
	}
}

// ClientInfo represents information about a client for API responses
type ClientInfo struct {
	ID              string                 `json:"id"`
	UserID          uuid.UUID              `json:"user_id"`
	Connected       bool                   `json:"connected"`
	ConnectedAt     time.Time              `json:"connected_at"`
	ConnectedFor    time.Duration          `json:"connected_for"`
	Metadata        map[string]interface{} `json:"metadata"`
	ChannelSize     int                    `json:"channel_size"`
	ChannelCapacity int                    `json:"channel_capacity"`
}

// ToClientInfo converts a client to a ClientInfo struct
func (c *client) ToClientInfo() ClientInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return ClientInfo{
		ID:              c.id,
		UserID:          c.userID,
		Connected:       c.connected,
		ConnectedAt:     c.connectedAt,
		ConnectedFor:    c.GetConnectedDuration(),
		Metadata:        c.GetMetadata(),
		ChannelSize:     len(c.channel),
		ChannelCapacity: cap(c.channel),
	}
}