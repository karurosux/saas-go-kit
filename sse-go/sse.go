package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Event represents a Server-Sent Event
type Event struct {
	ID      string      `json:"id,omitempty"`
	Type    string      `json:"type,omitempty"`
	Data    interface{} `json:"data"`
	Retry   int         `json:"retry,omitempty"`
	Created time.Time   `json:"created"`
}

// Client represents a connected SSE client
type Client struct {
	ID       string
	UserID   uuid.UUID
	Channel  chan Event
	Context  context.Context
	Cancel   context.CancelFunc
	Metadata map[string]interface{}
}

// Hub manages SSE connections and event broadcasting
type Hub struct {
	clients    map[string]*Client
	userClients map[uuid.UUID][]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
	mu         sync.RWMutex
	config     Config
}

// Config for SSE Hub
type Config struct {
	// BufferSize is the size of each client's event channel
	BufferSize int
	// ClientTimeout is how long to wait before considering a client dead
	ClientTimeout time.Duration
	// HeartbeatInterval is how often to send keepalive messages
	HeartbeatInterval time.Duration
	// MaxClients is the maximum number of concurrent clients (0 = unlimited)
	MaxClients int
	// MaxClientsPerUser is the maximum number of concurrent clients per user (0 = unlimited)
	MaxClientsPerUser int
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	// Target specific user(s) or broadcast to all if empty
	UserIDs []uuid.UUID
	// Target specific client(s)
	ClientIDs []string
	// The event to send
	Event Event
}

// DefaultConfig returns default SSE configuration
func DefaultConfig() Config {
	return Config{
		BufferSize:        10,
		ClientTimeout:     5 * time.Minute,
		HeartbeatInterval: 30 * time.Second,
		MaxClients:        1000,
		MaxClientsPerUser: 5,
	}
}

// NewHub creates a new SSE hub
func NewHub(config Config) *Hub {
	if config.BufferSize <= 0 {
		config.BufferSize = 10
	}
	if config.HeartbeatInterval <= 0 {
		config.HeartbeatInterval = 30 * time.Second
	}

	return &Hub{
		clients:     make(map[string]*Client),
		userClients: make(map[uuid.UUID][]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan BroadcastMessage),
		config:      config,
	}
}

// Run starts the hub's event loop
func (h *Hub) Run(ctx context.Context) {
	heartbeatTicker := time.NewTicker(h.config.HeartbeatInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-heartbeatTicker.C:
			h.sendHeartbeat()
		}
	}
}

// ServeHTTP handles SSE connections
func (h *Hub) ServeHTTP(c echo.Context) error {
	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no") // Disable Nginx buffering

	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		// Try alternative key
		if accountID, ok := c.Get("account_id").(uuid.UUID); ok {
			userID = accountID
		} else {
			return echo.NewHTTPError(401, "Unauthorized")
		}
	}

	// Create client
	ctx, cancel := context.WithCancel(c.Request().Context())
	client := &Client{
		ID:       uuid.New().String(),
		UserID:   userID,
		Channel:  make(chan Event, h.config.BufferSize),
		Context:  ctx,
		Cancel:   cancel,
		Metadata: make(map[string]interface{}),
	}

	// Register client
	h.register <- client

	// Send initial connection event
	client.Channel <- Event{
		Type: "connected",
		Data: map[string]interface{}{
			"client_id": client.ID,
			"timestamp": time.Now(),
		},
	}

	// Cleanup on disconnect
	defer func() {
		h.unregister <- client
		close(client.Channel)
		cancel()
	}()

	// Send events to client
	for {
		select {
		case <-c.Request().Context().Done():
			return nil

		case event, ok := <-client.Channel:
			if !ok {
				return nil
			}

			// Format SSE message
			if event.ID != "" {
				fmt.Fprintf(c.Response(), "id: %s\n", event.ID)
			}
			if event.Type != "" {
				fmt.Fprintf(c.Response(), "event: %s\n", event.Type)
			}
			if event.Retry > 0 {
				fmt.Fprintf(c.Response(), "retry: %d\n", event.Retry)
			}

			// Marshal data to JSON
			data, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Response(), "data: %s\n\n", data)

			// Flush the response
			c.Response().Flush()
		}
	}
}

// SendToUser sends an event to all clients of a specific user
func (h *Hub) SendToUser(userID uuid.UUID, event Event) {
	h.broadcast <- BroadcastMessage{
		UserIDs: []uuid.UUID{userID},
		Event:   event,
	}
}

// SendToUsers sends an event to multiple users
func (h *Hub) SendToUsers(userIDs []uuid.UUID, event Event) {
	h.broadcast <- BroadcastMessage{
		UserIDs: userIDs,
		Event:   event,
	}
}

// SendToClient sends an event to a specific client
func (h *Hub) SendToClient(clientID string, event Event) {
	h.broadcast <- BroadcastMessage{
		ClientIDs: []string{clientID},
		Event:     event,
	}
}

// SendToAll sends an event to all connected clients
func (h *Hub) SendToAll(event Event) {
	h.broadcast <- BroadcastMessage{
		Event: event,
	}
}

// GetConnectedUsers returns a list of currently connected user IDs
func (h *Hub) GetConnectedUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uuid.UUID, 0, len(h.userClients))
	for userID := range h.userClients {
		users = append(users, userID)
	}
	return users
}

// GetUserClientCount returns the number of clients for a specific user
func (h *Hub) GetUserClientCount(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.userClients[userID])
}

// GetTotalClientCount returns the total number of connected clients
func (h *Hub) GetTotalClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}

// Private methods

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check max clients limit
	if h.config.MaxClients > 0 && len(h.clients) >= h.config.MaxClients {
		client.Cancel()
		return
	}

	// Check max clients per user limit
	if h.config.MaxClientsPerUser > 0 {
		userClients := h.userClients[client.UserID]
		if len(userClients) >= h.config.MaxClientsPerUser {
			// Remove oldest client
			oldestClient := userClients[0]
			oldestClient.Cancel()
			h.unregisterClientLocked(oldestClient)
		}
	}

	// Register client
	h.clients[client.ID] = client
	h.userClients[client.UserID] = append(h.userClients[client.UserID], client)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.unregisterClientLocked(client)
}

func (h *Hub) unregisterClientLocked(client *Client) {
	if _, ok := h.clients[client.ID]; !ok {
		return
	}

	delete(h.clients, client.ID)

	// Remove from user clients
	userClients := h.userClients[client.UserID]
	for i, c := range userClients {
		if c.ID == client.ID {
			h.userClients[client.UserID] = append(userClients[:i], userClients[i+1:]...)
			break
		}
	}

	// Clean up empty user entries
	if len(h.userClients[client.UserID]) == 0 {
		delete(h.userClients, client.UserID)
	}
}

func (h *Hub) broadcastMessage(msg BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Set event metadata
	if msg.Event.ID == "" {
		msg.Event.ID = uuid.New().String()
	}
	if msg.Event.Created.IsZero() {
		msg.Event.Created = time.Now()
	}

	// Send to specific clients
	if len(msg.ClientIDs) > 0 {
		for _, clientID := range msg.ClientIDs {
			if client, ok := h.clients[clientID]; ok {
				select {
				case client.Channel <- msg.Event:
				default:
					// Client channel is full, skip
				}
			}
		}
		return
	}

	// Send to specific users
	if len(msg.UserIDs) > 0 {
		for _, userID := range msg.UserIDs {
			for _, client := range h.userClients[userID] {
				select {
				case client.Channel <- msg.Event:
				default:
					// Client channel is full, skip
				}
			}
		}
		return
	}

	// Send to all clients
	for _, client := range h.clients {
		select {
		case client.Channel <- msg.Event:
		default:
			// Client channel is full, skip
		}
	}
}

func (h *Hub) sendHeartbeat() {
	h.broadcastMessage(BroadcastMessage{
		Event: Event{
			Type: "heartbeat",
			Data: map[string]interface{}{
				"timestamp": time.Now(),
			},
		},
	})
}

func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Cancel all client contexts
	for _, client := range h.clients {
		client.Cancel()
	}

	// Clear maps
	h.clients = make(map[string]*Client)
	h.userClients = make(map[uuid.UUID][]*Client)
}