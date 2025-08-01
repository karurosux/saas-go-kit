package ssemodel

import (
	"time"
	
	"{{.Project.GoModule}}/internal/sse/constants"
	"{{.Project.GoModule}}/internal/sse/interface"
	"github.com/google/uuid"
)

// EventBuilder helps create SSE events with a fluent interface
type EventBuilder struct {
	event sseinterface.Event
}

// NewEventBuilder creates a new event builder
func NewEventBuilder() *EventBuilder {
	return &EventBuilder{
		event: sseinterface.Event{
			Created: time.Now(),
		},
	}
}

// WithID sets the event ID
func (b *EventBuilder) WithID(id string) *EventBuilder {
	b.event.ID = id
	return b
}

// WithType sets the event type
func (b *EventBuilder) WithType(eventType string) *EventBuilder {
	b.event.Type = eventType
	return b
}

// WithData sets the event data
func (b *EventBuilder) WithData(data interface{}) *EventBuilder {
	b.event.Data = data
	return b
}

// WithRetry sets the retry interval in milliseconds
func (b *EventBuilder) WithRetry(retry int) *EventBuilder {
	b.event.Retry = retry
	return b
}

// Build returns the constructed event
func (b *EventBuilder) Build() sseinterface.Event {
	if b.event.ID == "" {
		b.event.ID = uuid.New().String()
	}
	return b.event
}

// Predefined event builders for common event types

// NewConnectedEvent creates a connected event
func NewConnectedEvent(clientID string) sseinterface.Event {
	return NewEventBuilder().
		WithType(sseconstants.EventTypeConnected).
		WithData(map[string]interface{}{
			"client_id": clientID,
			"timestamp": time.Now(),
		}).
		Build()
}

// NewDisconnectedEvent creates a disconnected event
func NewDisconnectedEvent(clientID string, reason string) sseinterface.Event {
	return NewEventBuilder().
		WithType(sseconstants.EventTypeDisconnected).
		WithData(map[string]interface{}{
			"client_id": clientID,
			"reason":    reason,
			"timestamp": time.Now(),
		}).
		Build()
}

// NewHeartbeatEvent creates a heartbeat event
func NewHeartbeatEvent() sseinterface.Event {
	return NewEventBuilder().
		WithType(sseconstants.EventTypeHeartbeat).
		WithData(map[string]interface{}{
			"timestamp": time.Now(),
		}).
		Build()
}

// NewErrorEvent creates an error event
func NewErrorEvent(errorMsg string, code int) sseinterface.Event {
	return NewEventBuilder().
		WithType(sseconstants.EventTypeError).
		WithData(map[string]interface{}{
			"error":     errorMsg,
			"code":      code,
			"timestamp": time.Now(),
		}).
		Build()
}

// NewMessageEvent creates a message event
func NewMessageEvent(message string, sender string) sseinterface.Event {
	return NewEventBuilder().
		WithType(sseconstants.EventTypeMessage).
		WithData(map[string]interface{}{
			"message":   message,
			"sender":    sender,
			"timestamp": time.Now(),
		}).
		Build()
}

// NewNotificationEvent creates a notification event
func NewNotificationEvent(title, body string, priority string) sseinterface.Event {
	return NewEventBuilder().
		WithType(sseconstants.EventTypeNotification).
		WithData(map[string]interface{}{
			"title":     title,
			"body":      body,
			"priority":  priority,
			"timestamp": time.Now(),
		}).
		Build()
}

// NewUpdateEvent creates an update event
func NewUpdateEvent(resource string, action string, data interface{}) sseinterface.Event {
	return NewEventBuilder().
		WithType(sseconstants.EventTypeUpdate).
		WithData(map[string]interface{}{
			"resource":  resource,
			"action":    action,
			"data":      data,
			"timestamp": time.Now(),
		}).
		Build()
}

// BroadcastRequest represents a request to broadcast an event via API
type BroadcastRequest struct {
	UserIDs   []uuid.UUID `json:"user_ids,omitempty" validate:"omitempty,dive,uuid"`
	ClientIDs []string    `json:"client_ids,omitempty" validate:"omitempty,dive,uuid"`
	Event     EventRequest `json:"event" validate:"required"`
}

// EventRequest represents an event in API requests
type EventRequest struct {
	ID    string      `json:"id,omitempty"`
	Type  string      `json:"type" validate:"required"`
	Data  interface{} `json:"data" validate:"required"`
	Retry int         `json:"retry,omitempty" validate:"omitempty,min=0"`
}

// ToEvent converts an EventRequest to an Event
func (r *EventRequest) ToEvent() sseinterface.Event {
	event := sseinterface.Event{
		ID:      r.ID,
		Type:    r.Type,
		Data:    r.Data,
		Retry:   r.Retry,
		Created: time.Now(),
	}
	
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	
	return event
}