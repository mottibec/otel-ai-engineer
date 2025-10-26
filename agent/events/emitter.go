package events

import (
	"sync"
)

// EventEmitter is the interface for emitting and subscribing to agent events
type EventEmitter interface {
	// Emit sends an event to all subscribers
	Emit(event *AgentEvent)

	// Subscribe creates a subscription to events for a specific run
	// Returns a channel to receive events and a cleanup function
	Subscribe(runID string) (<-chan *AgentEvent, func())

	// SubscribeAll creates a subscription to all events
	// Returns a channel to receive events and a cleanup function
	SubscribeAll() (<-chan *AgentEvent, func())

	// Close shuts down the emitter and all subscriptions
	Close()
}

// Subscriber represents a single subscription
type subscriber struct {
	ch     chan *AgentEvent
	runID  string // empty string means subscribe to all
	closed bool
	mu     sync.Mutex
}

// DefaultEmitter is the default implementation of EventEmitter
type DefaultEmitter struct {
	subscribers []*subscriber
	mu          sync.RWMutex
	closed      bool
}

// NewEmitter creates a new event emitter
func NewEmitter() *DefaultEmitter {
	return &DefaultEmitter{
		subscribers: make([]*subscriber, 0),
	}
}

// Emit sends an event to all matching subscribers
func (e *DefaultEmitter) Emit(event *AgentEvent) {
	if event == nil {
		return
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return
	}

	// Send to all matching subscribers
	for _, sub := range e.subscribers {
		sub.mu.Lock()
		if sub.closed {
			sub.mu.Unlock()
			continue
		}

		// Check if subscriber wants this event
		if sub.runID == "" || sub.runID == event.RunID {
			// Non-blocking send
			select {
			case sub.ch <- event:
			default:
				// Channel is full, skip this subscriber
				// In production, you might want to log this
			}
		}
		sub.mu.Unlock()
	}
}

// Subscribe creates a subscription to events for a specific run
func (e *DefaultEmitter) Subscribe(runID string) (<-chan *AgentEvent, func()) {
	e.mu.Lock()
	defer e.mu.Unlock()

	sub := &subscriber{
		ch:    make(chan *AgentEvent, 100), // Buffered channel
		runID: runID,
	}

	e.subscribers = append(e.subscribers, sub)

	// Return cleanup function
	cleanup := func() {
		sub.mu.Lock()
		if !sub.closed {
			close(sub.ch)
			sub.closed = true
		}
		sub.mu.Unlock()

		// Remove from subscribers list
		e.mu.Lock()
		defer e.mu.Unlock()

		for i, s := range e.subscribers {
			if s == sub {
				e.subscribers = append(e.subscribers[:i], e.subscribers[i+1:]...)
				break
			}
		}
	}

	return sub.ch, cleanup
}

// SubscribeAll creates a subscription to all events
func (e *DefaultEmitter) SubscribeAll() (<-chan *AgentEvent, func()) {
	return e.Subscribe("") // Empty runID means all events
}

// Close shuts down the emitter and all subscriptions
func (e *DefaultEmitter) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return
	}

	e.closed = true

	// Close all subscriber channels
	for _, sub := range e.subscribers {
		sub.mu.Lock()
		if !sub.closed {
			close(sub.ch)
			sub.closed = true
		}
		sub.mu.Unlock()
	}

	e.subscribers = nil
}

// NoOpEmitter is an emitter that does nothing (for when events are disabled)
type NoOpEmitter struct{}

// NewNoOpEmitter creates a new no-op emitter
func NewNoOpEmitter() *NoOpEmitter {
	return &NoOpEmitter{}
}

// Emit does nothing
func (n *NoOpEmitter) Emit(event *AgentEvent) {}

// Subscribe returns a closed channel
func (n *NoOpEmitter) Subscribe(runID string) (<-chan *AgentEvent, func()) {
	ch := make(chan *AgentEvent)
	close(ch)
	return ch, func() {}
}

// SubscribeAll returns a closed channel
func (n *NoOpEmitter) SubscribeAll() (<-chan *AgentEvent, func()) {
	ch := make(chan *AgentEvent)
	close(ch)
	return ch, func() {}
}

// Close does nothing
func (n *NoOpEmitter) Close() {}
