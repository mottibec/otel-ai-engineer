package storage

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
)

// MemoryStorage is an in-memory implementation of Storage
type MemoryStorage struct {
	runs    map[string]*Run
	events  map[string][]*events.AgentEvent
	emitter events.EventEmitter
	mu      sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		runs:    make(map[string]*Run),
		events:  make(map[string][]*events.AgentEvent),
		emitter: events.NewEmitter(),
	}
}

// CreateRun creates a new run
func (m *MemoryStorage) CreateRun(run *Run) error {
	if run == nil {
		return fmt.Errorf("run cannot be nil")
	}
	if run.ID == "" {
		return fmt.Errorf("run ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.runs[run.ID]; exists {
		return fmt.Errorf("run with ID %s already exists", run.ID)
	}

	m.runs[run.ID] = run
	m.events[run.ID] = make([]*events.AgentEvent, 0)

	return nil
}

// GetRun retrieves a run by ID
func (m *MemoryStorage) GetRun(runID string) (*Run, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	run, exists := m.runs[runID]
	if !exists {
		return nil, fmt.Errorf("run with ID %s not found", runID)
	}

	// Return a copy to prevent external modifications
	runCopy := *run
	return &runCopy, nil
}

// ListRuns retrieves runs with optional filtering
func (m *MemoryStorage) ListRuns(opts RunListOptions) ([]*Run, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all runs
	allRuns := make([]*Run, 0, len(m.runs))
	for _, run := range m.runs {
		// Apply filters
		if opts.Status != nil && run.Status != *opts.Status {
			continue
		}
		if opts.Since != nil && run.StartTime.Before(*opts.Since) {
			continue
		}

		// Create a copy
		runCopy := *run
		allRuns = append(allRuns, &runCopy)
	}

	// Sort by start time (newest first)
	sort.Slice(allRuns, func(i, j int) bool {
		return allRuns[i].StartTime.After(allRuns[j].StartTime)
	})

	// Apply pagination
	if opts.Limit == 0 {
		opts.Limit = 100 // Default limit
	}

	start := opts.Offset
	if start >= len(allRuns) {
		return []*Run{}, nil
	}

	end := start + opts.Limit
	if end > len(allRuns) {
		end = len(allRuns)
	}

	return allRuns[start:end], nil
}

// UpdateRun updates a run
func (m *MemoryStorage) UpdateRun(runID string, update *RunUpdate) error {
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	run, exists := m.runs[runID]
	if !exists {
		return fmt.Errorf("run with ID %s not found", runID)
	}

	run.ApplyUpdate(update)

	return nil
}

// DeleteRun deletes a run and its events
func (m *MemoryStorage) DeleteRun(runID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.runs[runID]; !exists {
		return fmt.Errorf("run with ID %s not found", runID)
	}

	delete(m.runs, runID)
	delete(m.events, runID)

	return nil
}

// AddEvent adds an event to a run
func (m *MemoryStorage) AddEvent(runID string, event *events.AgentEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.runs[runID]; !exists {
		return fmt.Errorf("run with ID %s not found", runID)
	}

	// Check if event already exists (deduplicate by ID)
	exists := false
	for _, existingEvent := range m.events[runID] {
		if existingEvent.ID == event.ID {
			exists = true
			break
		}
	}

	if !exists {
		m.events[runID] = append(m.events[runID], event)
		// Emit event to subscribers (only if this was a new event)
		m.emitter.Emit(event)
	}

	return nil
}

// GetEvents retrieves events for a run
func (m *MemoryStorage) GetEvents(runID string, after *time.Time) ([]*events.AgentEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	runEvents, exists := m.events[runID]
	if !exists {
		return nil, fmt.Errorf("run with ID %s not found", runID)
	}

	// Filter by timestamp if provided
	if after == nil {
		// Return all events
		result := make([]*events.AgentEvent, len(runEvents))
		copy(result, runEvents)
		return result, nil
	}

	// Filter events after the given time
	result := make([]*events.AgentEvent, 0)
	for _, event := range runEvents {
		if event.Timestamp.After(*after) {
			result = append(result, event)
		}
	}

	return result, nil
}

// GetEventCount returns the number of events for a run
func (m *MemoryStorage) GetEventCount(runID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	runEvents, exists := m.events[runID]
	if !exists {
		return 0, fmt.Errorf("run with ID %s not found", runID)
	}

	return len(runEvents), nil
}

// Subscribe creates a subscription to events for a specific run
func (m *MemoryStorage) Subscribe(runID string) (<-chan *events.AgentEvent, func()) {
	return m.emitter.Subscribe(runID)
}

// SubscribeAll creates a subscription to all events
func (m *MemoryStorage) SubscribeAll() (<-chan *events.AgentEvent, func()) {
	return m.emitter.SubscribeAll()
}

// Close closes the storage
func (m *MemoryStorage) Close() error {
	m.emitter.Close()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.runs = nil
	m.events = nil

	return nil
}
