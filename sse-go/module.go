package sse

import (
	"context"
	"time"

	"github.com/karurosux/saas-go-kit/core-go"
	"github.com/labstack/echo/v4"
)

// Module provides SSE functionality as a SaaS Kit module
type Module struct {
	hub         *Hub
	config      ModuleConfig
	routePrefix string
	name        string
}

// ModuleConfig for SSE module
type ModuleConfig struct {
	Hub          *Hub
	RoutePrefix  string
	RequireAuth  bool
	AuthGroup    string // Optional auth group requirement
}

// NewModule creates a new SSE module
func NewModule(config ModuleConfig) *Module {
	if config.RoutePrefix == "" {
		config.RoutePrefix = "/sse"
	}

	// Create hub if not provided
	if config.Hub == nil {
		config.Hub = NewHub(DefaultConfig())
	}

	// Start the hub in a goroutine
	go config.Hub.Run(context.Background())

	return &Module{
		hub:         config.Hub,
		config:      config,
		routePrefix: config.RoutePrefix,
		name:        "sse",
	}
}

// Mount mounts the SSE routes
func (m *Module) Mount(r *echo.Echo) {
	// Create route group
	group := r.Group(m.routePrefix)

	// Add SSE endpoint
	sseRoute := group.GET("/stream", m.hub.ServeHTTP)
	
	// Set route name for extraction
	sseRoute.Name = "sse.stream"

	// Add status endpoint (public)
	group.GET("/status", m.handleStatus).Name = "sse.status"
}

// GetHub returns the SSE hub for external use
func (m *Module) GetHub() *Hub {
	return m.hub
}

// handleStatus returns SSE connection status
func (m *Module) handleStatus(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"connected_users":  len(m.hub.GetConnectedUsers()),
		"total_clients":    m.hub.GetTotalClientCount(),
		"status":          "operational",
	})
}

// Notification helpers for common notification types

// NotificationEvent creates a notification event
func NotificationEvent(notificationType string, title, message string, data map[string]interface{}) Event {
	payload := map[string]interface{}{
		"type":      notificationType,
		"title":     title,
		"message":   message,
		"timestamp": time.Now(),
	}
	
	// Merge additional data
	for k, v := range data {
		payload[k] = v
	}

	return Event{
		Type: "notification",
		Data: payload,
	}
}

// BadgeUpdateEvent creates a badge update event
func BadgeUpdateEvent(badgeCount int) Event {
	return Event{
		Type: "badge_update",
		Data: map[string]interface{}{
			"count":     badgeCount,
			"timestamp": time.Now(),
		},
	}
}

// NewItemEvent creates a new item event (for feeds)
func NewItemEvent(itemType string, itemID string, item interface{}) Event {
	return Event{
		Type: "new_item",
		Data: map[string]interface{}{
			"item_type": itemType,
			"item_id":   itemID,
			"item":      item,
			"timestamp": time.Now(),
		},
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return m.name
}

// Routes returns the routes to be registered
func (m *Module) Routes() []core.Route {
	// SSE module doesn't use the standard routing mechanism
	// It uses Mount method instead
	return []core.Route{}
}

// Middleware returns global middleware to be applied
func (m *Module) Middleware() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{}
}

// Dependencies returns the names of required modules
func (m *Module) Dependencies() []string {
	return []string{}
}

// Init initializes the module with dependencies
func (m *Module) Init(deps map[string]core.Module) error {
	// SSE module doesn't have dependencies
	return nil
}