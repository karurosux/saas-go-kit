package ssecontroller

import (
	"strconv"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/sse/interface"
	"{{.Project.GoModule}}/internal/sse/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SSEController handles SSE-related HTTP requests
type SSEController struct {
	sseService sseinterface.SSEService
	hub        sseinterface.Hub
}

// NewSSEController creates a new SSE controller
func NewSSEController(sseService sseinterface.SSEService, hub sseinterface.Hub) *SSEController {
	return &SSEController{
		sseService: sseService,
		hub:        hub,
	}
}

// RegisterRoutes registers all SSE-related routes
func (sc *SSEController) RegisterRoutes(e *echo.Echo, basePath string) {
	group := e.Group(basePath)
	
	// SSE connection endpoint
	group.GET("/stream", sc.Stream)
	
	// Broadcasting endpoints
	group.POST("/broadcast", sc.Broadcast)
	group.POST("/broadcast/user/:userID", sc.BroadcastToUser)
	group.POST("/broadcast/users", sc.BroadcastToUsers)
	group.POST("/broadcast/client/:clientID", sc.BroadcastToClient)
	
	// Statistics and monitoring endpoints
	group.GET("/stats", sc.GetStats)
	group.GET("/users", sc.GetConnectedUsers)
	group.GET("/users/:userID/clients", sc.GetUserClients)
	group.GET("/clients/:clientID", sc.GetClient)
	
	// Administrative endpoints
	group.POST("/start", sc.StartService)
	group.POST("/stop", sc.StopService)
	group.GET("/status", sc.GetServiceStatus)
}

// Stream godoc
// @Summary Establish SSE connection
// @Description Establish a Server-Sent Events connection for real-time updates
// @Tags sse
// @Produce text/event-stream
// @Success 200 {string} string "SSE stream"
// @Failure 401 {object} core.ErrorResponse
// @Failure 503 {object} core.ErrorResponse
// @Router /sse/stream [get]
func (sc *SSEController) Stream(c echo.Context) error {
	return sc.sseService.ServeHTTP(c)
}

// Broadcast godoc
// @Summary Broadcast event to all clients
// @Description Broadcast an event to all connected SSE clients
// @Tags sse
// @Accept json
// @Produce json
// @Param request body ssemodel.EventRequest true "Event to broadcast"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /sse/broadcast [post]
func (sc *SSEController) Broadcast(c echo.Context) error {
	var req ssemodel.EventRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	event := req.ToEvent()
	if err := sc.sseService.SendToAll(event); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message":  "Event broadcasted to all clients",
		"event_id": event.ID,
	})
}

// BroadcastToUser godoc
// @Summary Broadcast event to user
// @Description Broadcast an event to all clients of a specific user
// @Tags sse
// @Accept json
// @Produce json
// @Param userID path string true "User ID"
// @Param request body ssemodel.EventRequest true "Event to broadcast"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /sse/broadcast/user/{userID} [post]
func (sc *SSEController) BroadcastToUser(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		return core.Error(c, core.BadRequest("Invalid user ID"))
	}

	var req ssemodel.EventRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	event := req.ToEvent()
	if err := sc.sseService.SendToUser(userID, event); err != nil {
		return core.Error(c, err)
	}

	clientCount := sc.sseService.GetUserClientCount(userID)
	return core.Success(c, map[string]interface{}{
		"message":      "Event broadcasted to user",
		"event_id":     event.ID,
		"user_id":      userID,
		"client_count": clientCount,
	})
}

// BroadcastToUsers godoc
// @Summary Broadcast event to multiple users
// @Description Broadcast an event to all clients of multiple users
// @Tags sse
// @Accept json
// @Produce json
// @Param request body ssemodel.BroadcastRequest true "Broadcast request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /sse/broadcast/users [post]
func (sc *SSEController) BroadcastToUsers(c echo.Context) error {
	var req ssemodel.BroadcastRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	if len(req.UserIDs) == 0 {
		return core.Error(c, core.BadRequest("At least one user ID is required"))
	}

	event := req.Event.ToEvent()
	if err := sc.sseService.SendToUsers(req.UserIDs, event); err != nil {
		return core.Error(c, err)
	}

	// Count total clients for these users
	totalClients := 0
	for _, userID := range req.UserIDs {
		totalClients += sc.sseService.GetUserClientCount(userID)
	}

	return core.Success(c, map[string]interface{}{
		"message":       "Event broadcasted to users",
		"event_id":      event.ID,
		"user_ids":      req.UserIDs,
		"user_count":    len(req.UserIDs),
		"client_count":  totalClients,
	})
}

// BroadcastToClient godoc
// @Summary Broadcast event to specific client
// @Description Broadcast an event to a specific client
// @Tags sse
// @Accept json
// @Produce json
// @Param clientID path string true "Client ID"
// @Param request body ssemodel.EventRequest true "Event to broadcast"
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /sse/broadcast/client/{clientID} [post]
func (sc *SSEController) BroadcastToClient(c echo.Context) error {
	clientID := c.Param("clientID")
	if clientID == "" {
		return core.Error(c, core.BadRequest("Client ID is required"))
	}

	var req ssemodel.EventRequest
	if err := c.Bind(&req); err != nil {
		return core.Error(c, core.BadRequest("Invalid request data"))
	}

	if err := c.Validate(req); err != nil {
		return core.Error(c, core.ValidationError(err))
	}

	event := req.ToEvent()
	if err := sc.sseService.SendToClient(clientID, event); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message":   "Event sent to client",
		"event_id":  event.ID,
		"client_id": clientID,
	})
}

// GetStats godoc
// @Summary Get SSE statistics
// @Description Get current SSE hub statistics and metrics
// @Tags sse
// @Accept json
// @Produce json
// @Success 200 {object} sseinterface.HubStats
// @Failure 500 {object} core.ErrorResponse
// @Router /sse/stats [get]
func (sc *SSEController) GetStats(c echo.Context) error {
	stats := sc.hub.GetStats()
	return core.Success(c, stats)
}

// GetConnectedUsers godoc
// @Summary Get connected users
// @Description Get a list of currently connected user IDs
// @Tags sse
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} core.ErrorResponse
// @Router /sse/users [get]
func (sc *SSEController) GetConnectedUsers(c echo.Context) error {
	users := sc.sseService.GetConnectedUsers()
	
	return core.Success(c, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

// GetUserClients godoc
// @Summary Get user's client count
// @Description Get the number of clients connected for a specific user
// @Tags sse
// @Accept json
// @Produce json
// @Param userID path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /sse/users/{userID}/clients [get]
func (sc *SSEController) GetUserClients(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		return core.Error(c, core.BadRequest("Invalid user ID"))
	}

	clientCount := sc.sseService.GetUserClientCount(userID)
	
	return core.Success(c, map[string]interface{}{
		"user_id":      userID,
		"client_count": clientCount,
	})
}

// GetClient godoc
// @Summary Get client information
// @Description Get information about a specific client
// @Tags sse
// @Accept json
// @Produce json
// @Param clientID path string true "Client ID"
// @Success 200 {object} ssemodel.ClientInfo
// @Failure 400 {object} core.ErrorResponse
// @Failure 404 {object} core.ErrorResponse
// @Router /sse/clients/{clientID} [get]
func (sc *SSEController) GetClient(c echo.Context) error {
	clientID := c.Param("clientID")
	if clientID == "" {
		return core.Error(c, core.BadRequest("Client ID is required"))
	}

	client, exists := sc.hub.GetClient(clientID)
	if !exists {
		return core.Error(c, core.NotFound("Client not found"))
	}

	// Return basic client information
	return core.Success(c, map[string]interface{}{
		"id":        client.GetID(),
		"user_id":   client.GetUserID(),
		"connected": client.IsConnected(),
		"metadata":  client.GetMetadata(),
	})
}

// StartService godoc
// @Summary Start SSE service
// @Description Start the SSE service if it's not running
// @Tags sse
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /sse/start [post]
func (sc *SSEController) StartService(c echo.Context) error {
	if err := sc.sseService.Start(c.Request().Context()); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "SSE service started successfully",
		"status":  "running",
	})
}

// StopService godoc
// @Summary Stop SSE service
// @Description Stop the SSE service if it's running
// @Tags sse
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /sse/stop [post]
func (sc *SSEController) StopService(c echo.Context) error {
	if err := sc.sseService.Stop(); err != nil {
		return core.Error(c, err)
	}

	return core.Success(c, map[string]string{
		"message": "SSE service stopped successfully",
		"status":  "stopped",
	})
}

// GetServiceStatus godoc
// @Summary Get service status
// @Description Get the current status of the SSE service
// @Tags sse
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /sse/status [get]
func (sc *SSEController) GetServiceStatus(c echo.Context) error {
	isRunning := sc.sseService.IsRunning()
	status := "stopped"
	if isRunning {
		status = "running"
	}
	
	stats := sc.hub.GetStats()
	
	return core.Success(c, map[string]interface{}{
		"status":         status,
		"running":        isRunning,
		"total_clients":  stats.TotalClients,
		"connected_users": stats.ConnectedUsers,
		"uptime":         stats.Uptime.String(),
		"events_sent":    stats.EventsSent,
		"events_dropped": stats.EventsDropped,
	})
}