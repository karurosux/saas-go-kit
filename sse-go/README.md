# SSE Module

Server-Sent Events (SSE) module for real-time, one-way communication from server to clients. Perfect for notifications, live updates, and activity feeds.

## Features

- **Real-time Communication**: Push events from server to clients instantly
- **Auto-reconnection**: Clients automatically reconnect on connection loss
- **User-based Routing**: Send events to specific users or broadcast to all
- **Multi-client Support**: Users can have multiple concurrent connections
- **Heartbeat**: Automatic keepalive to maintain connections
- **Simple Integration**: Works with existing auth middleware

## Installation

```bash
go get github.com/karurosux/saas-go-kit/sse-go
```

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "github.com/karurosux/saas-go-kit/core-go"
    "github.com/karurosux/saas-go-kit/sse-go"
)

func main() {
    // Create SSE hub with default config
    sseHub := sse.NewHub(sse.DefaultConfig())
    
    // Create SSE module
    sseModule := sse.NewModule(sse.ModuleConfig{
        Hub:         sseHub,
        RoutePrefix: "/api/sse",
        RequireAuth: true,
    })
    
    // Register with your app
    app := core.NewKit(nil, core.KitConfig{})
    app.Register(sseModule)
    app.Mount()
    
    // Keep reference to hub for sending events
    myApp.sseHub = sseHub
}
```

### 2. Send Events

```go
// Send to specific user
sseHub.SendToUser(userID, sse.Event{
    Type: "notification",
    Data: map[string]interface{}{
        "title":   "New follower!",
        "message": "John Doe is now following you",
    },
})

// Send to multiple users
sseHub.SendToUsers([]uuid.UUID{user1, user2}, sse.Event{
    Type: "update",
    Data: map[string]interface{}{
        "action": "dish_posted",
        "dish_id": dishID,
    },
})

// Broadcast to all connected users
sseHub.SendToAll(sse.Event{
    Type: "announcement",
    Data: map[string]interface{}{
        "message": "System maintenance in 10 minutes",
    },
})
```

### 3. Client-side (JavaScript)

```javascript
// Connect to SSE endpoint
const eventSource = new EventSource('/api/sse/stream', {
    withCredentials: true // Include auth cookies
});

// Handle connection
eventSource.addEventListener('connected', (e) => {
    const data = JSON.parse(e.data);
    console.log('Connected with client ID:', data.client_id);
});

// Handle notifications
eventSource.addEventListener('notification', (e) => {
    const notification = JSON.parse(e.data);
    showNotification(notification.title, notification.message);
});

// Handle errors
eventSource.addEventListener('error', (e) => {
    if (eventSource.readyState === EventSource.CLOSED) {
        console.log('Connection closed');
    }
});
```

### 4. Client-side (React Native)

```javascript
import RNEventSource from 'react-native-event-source';

// Connect to SSE
const eventSource = new RNEventSource('https://api.example.com/sse/stream', {
    headers: {
        'Authorization': `Bearer ${authToken}`
    }
});

// Handle events
eventSource.addEventListener('message', (e) => {
    const data = JSON.parse(e.data);
    handleNotification(data);
});
```

## Configuration

```go
type Config struct {
    // BufferSize is the size of each client's event channel
    BufferSize int // Default: 10
    
    // ClientTimeout is how long to wait before considering a client dead
    ClientTimeout time.Duration // Default: 5 minutes
    
    // HeartbeatInterval is how often to send keepalive messages
    HeartbeatInterval time.Duration // Default: 30 seconds
    
    // MaxClients is the maximum number of concurrent clients (0 = unlimited)
    MaxClients int // Default: 1000
    
    // MaxClientsPerUser is the maximum number of concurrent clients per user
    MaxClientsPerUser int // Default: 5
}
```

## Helper Functions

The module provides helper functions for common event types:

```go
// Create a notification event
event := sse.NotificationEvent(
    "review_reminder",
    "Time to review!",
    "Don't forget to review the dish you tried",
    map[string]interface{}{
        "dish_id": dishID,
        "restaurant": "Tony's Pizza",
    },
)

// Create a badge update event
event := sse.BadgeUpdateEvent(5) // User has 5 unread notifications

// Create a new item event (for feeds)
event := sse.NewItemEvent("dish", dishID, dishData)
```

## Event Types

Common event types you might use:

- `connected` - Sent when client connects
- `heartbeat` - Sent periodically to keep connection alive
- `notification` - General notifications
- `badge_update` - Update notification badge count
- `new_item` - New item in feed
- `update` - General data update
- `alert` - Important alerts

## Integration with Other Modules

### With Authentication

The SSE module automatically extracts user ID from the Echo context:

```go
// It looks for these keys in order:
// 1. "user_id" (uuid.UUID)
// 2. "account_id" (uuid.UUID)
```

### With Notifications

Send real-time notifications when events occur:

```go
// In your notification service
func (s *Service) SendNotification(userID uuid.UUID, notification Notification) error {
    // Save to database
    err := s.repo.Save(notification)
    
    // Send real-time event
    s.sseHub.SendToUser(userID, sse.NotificationEvent(
        notification.Type,
        notification.Title,
        notification.Message,
        nil,
    ))
    
    // Update badge count
    count := s.repo.GetUnreadCount(userID)
    s.sseHub.SendToUser(userID, sse.BadgeUpdateEvent(count))
    
    return err
}
```

## API Endpoints

- `GET /sse/stream` - SSE event stream (requires auth)
- `GET /sse/status` - Connection status (public)

## Best Practices

1. **Event Size**: Keep events small (< 1KB) for better performance
2. **Throttling**: Implement throttling for high-frequency updates
3. **Error Handling**: Always handle connection errors on client
4. **Reconnection**: Clients auto-reconnect, but implement exponential backoff
5. **Security**: Always use with authentication middleware
6. **Cleanup**: Events are dropped if client buffer is full

## Security Considerations

- Always use HTTPS in production
- Authenticate users before allowing SSE connections
- Validate user permissions before sending sensitive events
- Don't send sensitive data that shouldn't be cached

## Scaling

For horizontal scaling across multiple servers:

1. Use Redis Pub/Sub to coordinate between servers
2. Use sticky sessions (not recommended) or
3. Use a message queue (recommended)

Example with Redis:
```go
// Subscribe to Redis channel
go func() {
    pubsub := redisClient.Subscribe(ctx, "events")
    for msg := range pubsub.Channel() {
        var event BroadcastMessage
        json.Unmarshal([]byte(msg.Payload), &event)
        hub.broadcast <- event
    }
}()

// Publish events to Redis
func PublishEvent(event BroadcastMessage) {
    data, _ := json.Marshal(event)
    redisClient.Publish(ctx, "events", data)
}
```

## License

This project is licensed under the MIT License.