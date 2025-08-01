package sseservice

import (
	"context"
	"sync"
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/sse/constants"
	"{{.Project.GoModule}}/internal/sse/interface"
	"{{.Project.GoModule}}/internal/sse/model"
	"github.com/google/uuid"
)

// hub implements the Hub interface
type hub struct {
	clients      map[string]sseinterface.Client
	userClients  map[uuid.UUID][]sseinterface.Client
	register     chan sseinterface.Client
	unregister   chan string
	broadcast    chan sseinterface.BroadcastMessage
	mu           sync.RWMutex
	config       sseinterface.Config
	stats        sseinterface.HubStats
	startTime    time.Time
	shutdown     chan struct{}
	shutdownOnce sync.Once
}

// NewHub creates a new SSE hub
func NewHub(config sseinterface.Config) sseinterface.Hub {
	return &hub{
		clients:     make(map[string]sseinterface.Client),
		userClients: make(map[uuid.UUID][]sseinterface.Client),
		register:    make(chan sseinterface.Client, 100),
		unregister:  make(chan string, 100),
		broadcast:   make(chan sseinterface.BroadcastMessage, 1000),
		config:      config,
		startTime:   time.Now(),
		shutdown:    make(chan struct{}),
		stats: sseinterface.HubStats{
			ClientsPerUser: make(map[uuid.UUID]int),
		},
	}
}

// Run starts the hub's event loop
func (h *hub) Run(ctx context.Context) {
	var heartbeatTicker *time.Ticker
	if h.config.EnableHeartbeat && h.config.HeartbeatInterval > 0 {
		heartbeatTicker = time.NewTicker(h.config.HeartbeatInterval)
		defer heartbeatTicker.Stop()
	}
	
	for {
		select {
		case <-ctx.Done():
			h.shutdownHub()
			return
			
		case <-h.shutdown:
			h.shutdownHub()
			return
			
		case client := <-h.register:
			h.registerClient(client)
			
		case clientID := <-h.unregister:
			h.unregisterClient(clientID)
			
		case message := <-h.broadcast:
			h.broadcastMessage(message)
			
		case <-func() <-chan time.Time {
			if heartbeatTicker != nil {
				return heartbeatTicker.C
			}
			return make(chan time.Time) // Never fires
		}():
			h.sendHeartbeat()
		}
	}
}

// RegisterClient registers a new client
func (h *hub) RegisterClient(client sseinterface.Client) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Check max clients limit
	if h.config.MaxClients > 0 && len(h.clients) >= h.config.MaxClients {
		return core.BadRequest(sseconstants.ErrMaxClientsReached)
	}
	
	// Check max clients per user limit
	userID := client.GetUserID()
	if h.config.MaxClientsPerUser > 0 {
		userClients := h.userClients[userID]
		if len(userClients) >= h.config.MaxClientsPerUser {
			// Disconnect oldest client
			if len(userClients) > 0 {
				oldestClient := userClients[0]
				oldestClient.Disconnect()
				h.unregisterClientLocked(oldestClient.GetID())
			}
		}
	}
	
	// Register via channel
	select {
	case h.register <- client:
		return nil
	default:
		return core.Internal("failed to register client: channel full")
	}
}

// UnregisterClient unregisters a client
func (h *hub) UnregisterClient(clientID string) error {
	select {
	case h.unregister <- clientID:
		return nil
	default:
		return core.Internal("failed to unregister client: channel full")
	}
}

// GetClient retrieves a client by ID
func (h *hub) GetClient(clientID string) (sseinterface.Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	client, exists := h.clients[clientID]
	return client, exists
}

// Broadcast sends a message to the specified targets
func (h *hub) Broadcast(message sseinterface.BroadcastMessage) error {
	select {
	case h.broadcast <- message:
		return nil
	default:
		return core.Internal(sseconstants.ErrBroadcastFailed)
	}
}

// GetStats returns current hub statistics
func (h *hub) GetStats() sseinterface.HubStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Update runtime stats
	h.stats.TotalClients = len(h.clients)
	h.stats.ConnectedUsers = len(h.userClients)
	h.stats.Uptime = time.Since(h.startTime)
	
	// Copy clients per user map
	clientsPerUser := make(map[uuid.UUID]int)
	for userID, clients := range h.userClients {
		clientsPerUser[userID] = len(clients)
	}
	h.stats.ClientsPerUser = clientsPerUser
	
	return h.stats
}

// Shutdown gracefully shuts down the hub
func (h *hub) Shutdown() error {
	h.shutdownOnce.Do(func() {
		close(h.shutdown)
	})
	return nil
}

// Private methods

func (h *hub) registerClient(client sseinterface.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	clientID := client.GetID()
	userID := client.GetUserID()
	
	// Add to clients map
	h.clients[clientID] = client
	
	// Add to user clients map
	h.userClients[userID] = append(h.userClients[userID], client)
	
	// Update stats
	h.stats.ConnectionsOpened++
}

func (h *hub) unregisterClient(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.unregisterClientLocked(clientID)
}

func (h *hub) unregisterClientLocked(clientID string) {
	client, exists := h.clients[clientID]
	if !exists {
		return
	}
	
	userID := client.GetUserID()
	
	// Remove from clients map
	delete(h.clients, clientID)
	
	// Remove from user clients map
	userClients := h.userClients[userID]
	for i, c := range userClients {
		if c.GetID() == clientID {
			h.userClients[userID] = append(userClients[:i], userClients[i+1:]...)
			break
		}
	}
	
	// Clean up empty user entries
	if len(h.userClients[userID]) == 0 {
		delete(h.userClients, userID)
	}
	
	// Update stats
	h.stats.ConnectionsClosed++
}

func (h *hub) broadcastMessage(msg sseinterface.BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Set event metadata if missing
	if msg.Event.ID == "" {
		msg.Event.ID = uuid.New().String()
	}
	if msg.Event.Created.IsZero() {
		msg.Event.Created = time.Now()
	}
	
	var targetClients []sseinterface.Client
	
	// Send to specific clients
	if len(msg.ClientIDs) > 0 {
		for _, clientID := range msg.ClientIDs {
			if client, ok := h.clients[clientID]; ok {
				targetClients = append(targetClients, client)
			}
		}
	} else if len(msg.UserIDs) > 0 {
		// Send to specific users
		for _, userID := range msg.UserIDs {
			if userClients, ok := h.userClients[userID]; ok {
				targetClients = append(targetClients, userClients...)
			}
		}
	} else {
		// Send to all clients
		for _, client := range h.clients {
			targetClients = append(targetClients, client)
		}
	}
	
	// Send event to target clients
	for _, client := range targetClients {
		if client.IsConnected() {
			select {
			case client.GetChannel() <- msg.Event:
				h.stats.EventsSent++
			default:
				// Client channel is full, event dropped
				h.stats.EventsDropped++
			}
		}
	}
}

func (h *hub) sendHeartbeat() {
	heartbeatEvent := ssemodel.NewHeartbeatEvent()
	h.stats.LastHeartbeat = time.Now()
	
	message := sseinterface.BroadcastMessage{
		Event: heartbeatEvent,
	}
	
	h.broadcastMessage(message)
}

func (h *hub) shutdownHub() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Disconnect all clients
	for _, client := range h.clients {
		client.Disconnect()
	}
	
	// Clear maps
	h.clients = make(map[string]sseinterface.Client)
	h.userClients = make(map[uuid.UUID][]sseinterface.Client)
}