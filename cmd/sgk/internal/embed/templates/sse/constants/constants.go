package sseconstants

import "time"

// Default configuration values
const (
	DefaultBufferSize        = 10
	DefaultClientTimeout     = 5 * time.Minute
	DefaultHeartbeatInterval = 30 * time.Second
	DefaultMaxClients        = 1000
	DefaultMaxClientsPerUser = 5
)

// SSE event types
const (
	EventTypeConnected    = "connected"
	EventTypeDisconnected = "disconnected"
	EventTypeHeartbeat    = "heartbeat"
	EventTypeError        = "error"
	EventTypeMessage      = "message"
	EventTypeNotification = "notification"
	EventTypeUpdate       = "update"
)

// HTTP headers for SSE
const (
	HeaderContentType     = "text/event-stream"
	HeaderCacheControl    = "no-cache"
	HeaderConnection      = "keep-alive"
	HeaderXAccelBuffering = "no" // Disable Nginx buffering
)

// Environment variable keys
const (
	EnvSSEBufferSize        = "SSE_BUFFER_SIZE"
	EnvSSEClientTimeout     = "SSE_CLIENT_TIMEOUT"
	EnvSSEHeartbeatInterval = "SSE_HEARTBEAT_INTERVAL"
	EnvSSEMaxClients        = "SSE_MAX_CLIENTS"
	EnvSSEMaxClientsPerUser = "SSE_MAX_CLIENTS_PER_USER"
	EnvSSEEnableHeartbeat   = "SSE_ENABLE_HEARTBEAT"
	EnvSSEEnableMetrics     = "SSE_ENABLE_METRICS"
)

// Context keys for SSE middleware
const (
	ContextKeyUserID   = "user_id"
	ContextKeyAccountID = "account_id"
	ContextKeyClientID = "client_id"
	ContextKeySSEHub   = "sse_hub"
)

// Metric names
const (
	MetricConnectionsOpened = "sse_connections_opened_total"
	MetricConnectionsClosed = "sse_connections_closed_total"
	MetricEventsPublished   = "sse_events_published_total"
	MetricEventsDropped     = "sse_events_dropped_total"
	MetricActiveConnections = "sse_active_connections"
	MetricActiveUsers       = "sse_active_users"
)