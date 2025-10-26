package service

import (
	"context"
	"sync"
)

// ActiveRunManagerImpl implements ActiveRunManager interface
type ActiveRunManagerImpl struct {
	runs map[string]*ActiveRun
	mu   sync.RWMutex
}

// NewActiveRunManager creates a new active run manager
func NewActiveRunManager() *ActiveRunManagerImpl {
	return &ActiveRunManagerImpl{
		runs: make(map[string]*ActiveRun),
	}
}

// Add adds a new active run
func (m *ActiveRunManagerImpl) Add(runID string, ctx context.Context, cancel context.CancelFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.runs[runID] = &ActiveRun{
		RunID:          runID,
		Context:        ctx,
		CancelFunc:     cancel,
		PendingMessage: make(chan string, 10), // Buffered channel
	}
}

// Get retrieves an active run
func (m *ActiveRunManagerImpl) Get(runID string) (*ActiveRun, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	run, exists := m.runs[runID]
	return run, exists
}

// Remove removes an active run
func (m *ActiveRunManagerImpl) Remove(runID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.runs, runID)
}

// CancelAll cancels all active runs (useful for shutdown)
func (m *ActiveRunManagerImpl) CancelAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, run := range m.runs {
		if run.CancelFunc != nil {
			run.CancelFunc()
		}
	}
	m.runs = make(map[string]*ActiveRun)
}
