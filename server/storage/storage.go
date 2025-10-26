package storage

import (
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
)

// Storage is the interface for storing and retrieving agent runs and events
type Storage interface {
	// Run management
	CreateRun(run *Run) error
	GetRun(runID string) (*Run, error)
	ListRuns(opts RunListOptions) ([]*Run, error)
	UpdateRun(runID string, update *RunUpdate) error
	DeleteRun(runID string) error

	// Event management
	AddEvent(runID string, event *events.AgentEvent) error
	GetEvents(runID string, after *time.Time) ([]*events.AgentEvent, error)
	GetEventCount(runID string) (int, error)

	// Stream support (for real-time updates)
	Subscribe(runID string) (<-chan *events.AgentEvent, func())
	SubscribeAll() (<-chan *events.AgentEvent, func())

	// Cleanup
	Close() error
}
