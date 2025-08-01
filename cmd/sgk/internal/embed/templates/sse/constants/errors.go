package sseconstants

// Error messages for the SSE module
const (
	ErrSSEServiceNotRunning       = "SSE service is not running"
	ErrSSEServiceAlreadyRunning   = "SSE service is already running"
	ErrClientNotFound             = "client not found"
	ErrClientDisconnected         = "client is disconnected"
	ErrMaxClientsReached          = "maximum number of clients reached"
	ErrMaxClientsPerUserReached   = "maximum number of clients per user reached"
	ErrUnauthorized               = "unauthorized: user authentication required"
	ErrInvalidUserID              = "invalid user ID"
	ErrInvalidClientID            = "invalid client ID"
	ErrInvalidEventType           = "invalid event type"
	ErrEventDataRequired          = "event data is required"
	ErrBroadcastFailed            = "failed to broadcast message"
	ErrHubShutdown                = "hub is shutting down"
	ErrChannelFull                = "client event channel is full"
	ErrContextCanceled            = "context was canceled"
	ErrConnectionClosed           = "connection was closed"
	ErrInvalidConfiguration       = "invalid SSE configuration"
	ErrFailedToMarshalEvent       = "failed to marshal event data"
	ErrFailedToWriteEvent         = "failed to write event to client"
)