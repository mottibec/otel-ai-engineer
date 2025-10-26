package service

import "github.com/mottibechhofer/otel-ai-engineer/agent/events"

// EventBridgeAdapter adapts the server's EventBridge to the service's interface
type EventBridgeAdapter struct {
	bridge interface {
		GetEmitter() events.EventEmitter
	}
}

// NewEventBridgeAdapter creates a new event bridge adapter
func NewEventBridgeAdapter(bridge interface{ GetEmitter() events.EventEmitter }) *EventBridgeAdapter {
	return &EventBridgeAdapter{bridge: bridge}
}

// GetEmitter returns the event emitter
func (a *EventBridgeAdapter) GetEmitter() events.EventEmitter {
	return a.bridge.GetEmitter()
}
