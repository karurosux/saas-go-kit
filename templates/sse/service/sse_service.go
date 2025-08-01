package sseservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/sse/constants"
	"{{.Project.GoModule}}/internal/sse/interface"
	"{{.Project.GoModule}}/internal/sse/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// sseService implements the SSEService interface
type sseService struct {
	hub     sseinterface.Hub
	config  sseinterface.Config
	running atomic.Bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewSSEService creates a new SSE service
func NewSSEService(hub sseinterface.Hub, config sseinterface.Config) sseinterface.SSEService {
	return &sseService{
		hub:    hub,
		config: config,
	}
}

// Start starts the SSE service
func (s *sseService) Start(ctx context.Context) error {
	if s.running.Load() {
		return core.BadRequest(sseconstants.ErrSSEServiceAlreadyRunning)
	}
	
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running.Store(true)
	
	// Start the hub in a separate goroutine
	go s.hub.Run(s.ctx)
	
	return nil
}

// Stop stops the SSE service
func (s *sseService) Stop() error {
	if !s.running.Load() {
		return core.BadRequest(sseconstants.ErrSSEServiceNotRunning)
	}
	
	s.running.Store(false)
	
	if s.cancel != nil {
		s.cancel()
	}
	
	return s.hub.Shutdown()
}

// IsRunning returns whether the service is running
func (s *sseService) IsRunning() bool {
	return s.running.Load()
}

// ServeHTTP handles SSE connections
func (s *sseService) ServeHTTP(c echo.Context) error {
	if !s.running.Load() {
		return core.Error(c, core.ServiceUnavailable(sseconstants.ErrSSEServiceNotRunning))
	}
	
	// Set SSE headers
	c.Response().Header().Set("Content-Type", sseconstants.HeaderContentType)
	c.Response().Header().Set("Cache-Control", sseconstants.HeaderCacheControl)
	c.Response().Header().Set("Connection", sseconstants.HeaderConnection)
	c.Response().Header().Set("X-Accel-Buffering", sseconstants.HeaderXAccelBuffering)
	
	// Get user ID from context (set by auth middleware)
	userID, err := s.extractUserID(c)
	if err != nil {
		return core.Error(c, core.Unauthorized(sseconstants.ErrUnauthorized))
	}
	
	// Create client
	clientID := uuid.New().String()
	client := ssemodel.NewClient(clientID, userID, s.config.BufferSize)
	
	// Register client with hub
	if err := s.hub.RegisterClient(client); err != nil {
		return core.Error(c, err)
	}
	
	// Send initial connection event
	connectedEvent := ssemodel.NewConnectedEvent(clientID)
	client.GetChannel() <- connectedEvent
	
	// Cleanup on disconnect
	defer func() {
		s.hub.UnregisterClient(clientID)
		client.Disconnect()
	}()
	
	// Send events to client
	return s.streamEvents(c, client)
}

// Event broadcasting methods

func (s *sseService) SendToUser(userID uuid.UUID, event sseinterface.Event) error {
	if !s.running.Load() {
		return core.Internal(sseconstants.ErrSSEServiceNotRunning)
	}
	
	message := sseinterface.BroadcastMessage{
		UserIDs: []uuid.UUID{userID},
		Event:   event,
	}
	
	return s.hub.Broadcast(message)
}

func (s *sseService) SendToUsers(userIDs []uuid.UUID, event sseinterface.Event) error {
	if !s.running.Load() {
		return core.Internal(sseconstants.ErrSSEServiceNotRunning)
	}
	
	message := sseinterface.BroadcastMessage{
		UserIDs: userIDs,
		Event:   event,
	}
	
	return s.hub.Broadcast(message)
}

func (s *sseService) SendToClient(clientID string, event sseinterface.Event) error {
	if !s.running.Load() {
		return core.Internal(sseconstants.ErrSSEServiceNotRunning)
	}
	
	message := sseinterface.BroadcastMessage{
		ClientIDs: []string{clientID},
		Event:     event,
	}
	
	return s.hub.Broadcast(message)
}

func (s *sseService) SendToAll(event sseinterface.Event) error {
	if !s.running.Load() {
		return core.Internal(sseconstants.ErrSSEServiceNotRunning)
	}
	
	message := sseinterface.BroadcastMessage{
		Event: event,
	}
	
	return s.hub.Broadcast(message)
}

// Statistics methods

func (s *sseService) GetConnectedUsers() []uuid.UUID {
	stats := s.hub.GetStats()
	users := make([]uuid.UUID, 0, len(stats.ClientsPerUser))
	for userID := range stats.ClientsPerUser {
		users = append(users, userID)
	}
	return users
}

func (s *sseService) GetUserClientCount(userID uuid.UUID) int {
	stats := s.hub.GetStats()
	return stats.ClientsPerUser[userID]
}

func (s *sseService) GetTotalClientCount() int {
	stats := s.hub.GetStats()
	return stats.TotalClients
}

// Private helper methods

func (s *sseService) extractUserID(c echo.Context) (uuid.UUID, error) {
	// Try user_id first
	if userID := c.Get(sseconstants.ContextKeyUserID); userID != nil {
		if uid, ok := userID.(uuid.UUID); ok {
			return uid, nil
		}
		if uidStr, ok := userID.(string); ok {
			return uuid.Parse(uidStr)
		}
	}
	
	// Try account_id as fallback
	if accountID := c.Get(sseconstants.ContextKeyAccountID); accountID != nil {
		if aid, ok := accountID.(uuid.UUID); ok {
			return aid, nil
		}
		if aidStr, ok := accountID.(string); ok {
			return uuid.Parse(aidStr)
		}
	}
	
	return uuid.Nil, fmt.Errorf(sseconstants.ErrInvalidUserID)
}

func (s *sseService) streamEvents(c echo.Context, client sseinterface.Client) error {
	for {
		select {
		case <-c.Request().Context().Done():
			return nil
			
		case <-client.GetContext().Done():
			return nil
			
		case event, ok := <-client.GetChannel():
			if !ok {
				return nil
			}
			
			if err := s.writeEvent(c, event); err != nil {
				return err
			}
		}
	}
}

func (s *sseService) writeEvent(c echo.Context, event sseinterface.Event) error {
	// Set event metadata if missing
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Created.IsZero() {
		event.Created = time.Now()
	}
	
	// Write SSE format
	if event.ID != "" {
		if _, err := fmt.Fprintf(c.Response(), "id: %s\n", event.ID); err != nil {
			return err
		}
	}
	
	if event.Type != "" {
		if _, err := fmt.Fprintf(c.Response(), "event: %s\n", event.Type); err != nil {
			return err
		}
	}
	
	if event.Retry > 0 {
		if _, err := fmt.Fprintf(c.Response(), "retry: %d\n", event.Retry); err != nil {
			return err
		}
	}
	
	// Marshal data to JSON
	data, err := json.Marshal(event.Data)
	if err != nil {
		return core.Internal(fmt.Sprintf("%s: %v", sseconstants.ErrFailedToMarshalEvent, err))
	}
	
	if _, err := fmt.Fprintf(c.Response(), "data: %s\n\n", data); err != nil {
		return core.Internal(fmt.Sprintf("%s: %v", sseconstants.ErrFailedToWriteEvent, err))
	}
	
	// Flush the response
	if flusher, ok := c.Response().Writer.(http.Flusher); ok {
		flusher.Flush()
	}
	
	return nil
}