package sseinterface

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SSEService defines the main interface for Server-Sent Events functionality
type SSEService interface {
	// Connection management
	ServeHTTP(c echo.Context) error
	
	// Event broadcasting
	SendToUser(userID uuid.UUID, event Event) error
	SendToUsers(userIDs []uuid.UUID, event Event) error
	SendToClient(clientID string, event Event) error
	SendToAll(event Event) error
	
	// Statistics and monitoring
	GetConnectedUsers() []uuid.UUID
	GetUserClientCount(userID uuid.UUID) int
	GetTotalClientCount() int
	
	// Lifecycle management
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool
}

// Hub manages SSE connections and event broadcasting
type Hub interface {
	// Client management
	RegisterClient(client Client) error
	UnregisterClient(clientID string) error
	GetClient(clientID string) (Client, bool)
	
	// Broadcasting
	Broadcast(message BroadcastMessage) error
	
	// Statistics
	GetStats() HubStats
	
	// Lifecycle
	Run(ctx context.Context)
	Shutdown() error
}

// Client represents a connected SSE client
type Client interface {
	GetID() string
	GetUserID() uuid.UUID
	GetContext() context.Context
	GetChannel() chan Event
	GetMetadata() map[string]interface{}
	SetMetadata(key string, value interface{})
	IsConnected() bool
	Disconnect() error
}

// Event represents a Server-Sent Event
type Event struct {
	ID      string      `json:"id,omitempty"`
	Type    string      `json:"type,omitempty"`
	Data    interface{} `json:"data"`
	Retry   int         `json:"retry,omitempty"`
	Created time.Time   `json:"created"`
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	// Target specific user(s) or broadcast to all if empty
	UserIDs []uuid.UUID `json:"user_ids,omitempty"`
	// Target specific client(s)
	ClientIDs []string `json:"client_ids,omitempty"`
	// The event to send
	Event Event `json:"event"`
}

// Config for SSE Hub
type Config struct {
	// BufferSize is the size of each client's event channel
	BufferSize int `json:"buffer_size"`
	// ClientTimeout is how long to wait before considering a client dead
	ClientTimeout time.Duration `json:"client_timeout"`
	// HeartbeatInterval is how often to send keepalive messages
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	// MaxClients is the maximum number of concurrent clients (0 = unlimited)
	MaxClients int `json:"max_clients"`
	// MaxClientsPerUser is the maximum number of concurrent clients per user (0 = unlimited)
	MaxClientsPerUser int `json:"max_clients_per_user"`
	// EnableHeartbeat enables/disables heartbeat messages
	EnableHeartbeat bool `json:"enable_heartbeat"`
	// EnableMetrics enables/disables metrics collection
	EnableMetrics bool `json:"enable_metrics"`
}

// HubStats represents statistics about the SSE hub
type HubStats struct {
	TotalClients      int                        `json:"total_clients"`
	ConnectedUsers    int                        `json:"connected_users"`
	ClientsPerUser    map[uuid.UUID]int          `json:"clients_per_user"`
	EventsSent        int64                      `json:"events_sent"`
	EventsDropped     int64                      `json:"events_dropped"`
	Uptime            time.Duration              `json:"uptime"`
	LastHeartbeat     time.Time                  `json:"last_heartbeat"`
	ConnectionsOpened int64                      `json:"connections_opened"`
	ConnectionsClosed int64                      `json:"connections_closed"`
}

// EventFilter can filter events before they're sent to clients
type EventFilter interface {
	ShouldSend(event Event, client Client) bool
}

// EventTransformer can transform events before they're sent to clients
type EventTransformer interface {
	Transform(event Event, client Client) Event
}

// EventLogger logs SSE events for debugging/monitoring
type EventLogger interface {
	LogEventSent(event Event, clientID string, userID uuid.UUID)
	LogEventDropped(event Event, clientID string, reason string)
	LogClientConnected(clientID string, userID uuid.UUID)
	LogClientDisconnected(clientID string, userID uuid.UUID, duration time.Duration)
}

// Middleware can be applied to SSE connections
type Middleware interface {
	Process(c echo.Context, next func(c echo.Context) error) error
}