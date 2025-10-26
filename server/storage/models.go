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
	ParentRunID     *string              `json:"parent_run_id,omitempty"`
	SubRunIDs       []string             `json:"sub_run_ids,omitempty"`
	IsHandoff       bool                 `json:"is_handoff"`
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
	ParentRunID     *string
	SubRunIDs       *[]string
	IsHandoff       *bool
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
	if update.ParentRunID != nil {
		r.ParentRunID = update.ParentRunID
	}
	if update.SubRunIDs != nil {
		r.SubRunIDs = *update.SubRunIDs
	}
	if update.IsHandoff != nil {
		r.IsHandoff = *update.IsHandoff
	}
}

// CalculateDuration calculates and sets the duration field
func (r *Run) CalculateDuration() {
	if r.EndTime != nil {
		duration := r.EndTime.Sub(r.StartTime)
		r.Duration = duration.String()
	}
}

// SpanType represents the type of a span
type SpanType string

const (
	SpanTypeTool           SpanType = "tool"
	SpanTypeAPICall        SpanType = "api_call"
	SpanTypeAgentHandoff   SpanType = "agent_handoff"
	SpanTypeIteration      SpanType = "iteration"
	SpanTypeTrace          SpanType = "trace"
)

// Span represents a single span in a trace
type Span struct {
	ID           string                 `json:"id"`
	Type         SpanType               `json:"type"`
	Name         string                 `json:"name"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Duration     string                 `json:"duration,omitempty"`
	DurationMs   int64                  `json:"duration_ms,omitempty"`
	ParentSpanID *string                `json:"parent_span_id,omitempty"`
	Children     []*Span                `json:"children,omitempty"`
	Tags         map[string]interface{} `json:"tags,omitempty"`
	Error        bool                   `json:"error,omitempty"`
	ErrorMsg     string                 `json:"error_msg,omitempty"`
}

// Trace represents a complete trace with a root span
type Trace struct {
	TraceID    string    `json:"trace_id"`
	RootSpan   *Span     `json:"root_span"`
	StartTime  time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Duration   string    `json:"duration,omitempty"`
	DurationMs int64     `json:"duration_ms,omitempty"`
}
