package storage

import (
	"time"

	"github.com/mottibechhofer/otel-ai-engineer/agent/events"
)

// RunStatus represents the status of an agent run
type RunStatus string

const (
	RunStatusRunning   RunStatus = "running"
	RunStatusPaused    RunStatus = "paused"
	RunStatusSuccess   RunStatus = "success"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCancelled RunStatus = "cancelled"
)

// Run represents a single agent execution
type Run struct {
	ID              string               `json:"id"`
	AgentID         string               `json:"agent_id"`
	AgentName       string               `json:"agent_name"`
	Status          RunStatus            `json:"status"`
	Prompt          string               `json:"prompt"`
	Model           string               `json:"model"`
	StartTime       time.Time            `json:"start_time"`
	EndTime         *time.Time           `json:"end_time,omitempty"`
	Duration        string               `json:"duration,omitempty"`
	TotalIterations int                  `json:"total_iterations"`
	TotalToolCalls  int                  `json:"total_tool_calls"`
	TotalTokens     TokenUsage           `json:"total_tokens"`
	Error           string               `json:"error,omitempty"`
	Events          []*events.AgentEvent `json:"-"` // Not serialized, fetched separately
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// RunUpdate contains fields that can be updated on a run
type RunUpdate struct {
	Status          *RunStatus
	EndTime         *time.Time
	ClearEndTime    bool // If true, clears the EndTime field (sets to NULL)
	Duration        *string
	TotalIterations *int
	TotalToolCalls  *int
	TotalTokens     *TokenUsage
	Error           *string
}

// RunListOptions contains options for listing runs
type RunListOptions struct {
	Limit  int
	Offset int
	Status *RunStatus
	Since  *time.Time
}

// ApplyUpdate applies an update to a run
func (r *Run) ApplyUpdate(update *RunUpdate) {
	if update.Status != nil {
		r.Status = *update.Status
	}
	if update.ClearEndTime {
		r.EndTime = nil
	} else if update.EndTime != nil {
		r.EndTime = update.EndTime
	}
	if update.Duration != nil {
		r.Duration = *update.Duration
	}
	if update.TotalIterations != nil {
		r.TotalIterations = *update.TotalIterations
	}
	if update.TotalToolCalls != nil {
		r.TotalToolCalls = *update.TotalToolCalls
	}
	if update.TotalTokens != nil {
		r.TotalTokens = *update.TotalTokens
	}
	if update.Error != nil {
		r.Error = *update.Error
	}
}

// CalculateDuration calculates and sets the duration field
func (r *Run) CalculateDuration() {
	if r.EndTime != nil {
		duration := r.EndTime.Sub(r.StartTime)
		r.Duration = duration.String()
	}
}
