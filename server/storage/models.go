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

// PlanStatus represents the status of an observability plan
type PlanStatus string

const (
	PlanStatusDraft     PlanStatus = "draft"
	PlanStatusPending   PlanStatus = "pending"
	PlanStatusExecuting PlanStatus = "executing"
	PlanStatusPartial   PlanStatus = "partial"
	PlanStatusSuccess   PlanStatus = "success"
	PlanStatusFailed    PlanStatus = "failed"
)

// ObservabilityPlan represents a complete observability setup plan
type ObservabilityPlan struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Environment string     `json:"environment"`
	Status      PlanStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	// Components - will be loaded separately
	Services        []*InstrumentedService    `json:"services,omitempty"`
	Infrastructure  []*InfrastructureComponent `json:"infrastructure,omitempty"`
	Pipelines       []*CollectorPipeline       `json:"pipelines,omitempty"`
	Backends        []*Backend                 `json:"backends,omitempty"`
	Dependencies    []*PlanDependency          `json:"dependencies,omitempty"`
}

// InstrumentedService represents a service that needs instrumentation
type InstrumentedService struct {
	ID                  string    `json:"id"`
	PlanID              string    `json:"plan_id"`
	ServiceName         string    `json:"service_name"`
	Language            string    `json:"language"`
	Framework           string    `json:"framework"`
	SDKVersion          string    `json:"sdk_version"`
	ConfigFile          string    `json:"config_file"`
	Status              string    `json:"status"`
	CodeChangesSummary  string    `json:"code_changes_summary"`
	TargetPath          string    `json:"target_path"`
	ExporterEndpoint    string    `json:"exporter_endpoint"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// InfrastructureComponent represents infrastructure to be monitored
type InfrastructureComponent struct {
	ID             string    `json:"id"`
	PlanID         string    `json:"plan_id"`
	ComponentType  string    `json:"component_type"`  // "database", "cache", "queue", "host"
	Name           string    `json:"name"`
	Host           string    `json:"host"`
	ReceiverType   string    `json:"receiver_type"`   // "postgres", "mysql", "redis", "hostmetrics"
	MetricsCollected string  `json:"metrics_collected"` // JSON array or comma-separated
	Status         string    `json:"status"`
	Config         string    `json:"config,omitempty"` // Additional configuration JSON
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CollectorPipeline represents a collector configuration pipeline
type CollectorPipeline struct {
	ID           string    `json:"id"`
	PlanID       string    `json:"plan_id"`
	CollectorID  string    `json:"collector_id"` // Reference to deployed collector
	Name         string    `json:"name"`
	ConfigYAML   string    `json:"config_yaml"`
	Rules        string    `json:"rules"` // JSON object with sampling, filtering, etc.
	Status       string    `json:"status"`
	TargetType   string    `json:"target_type"` // "docker", "kubernetes", "remote"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Backend represents an observability backend
type Backend struct {
	ID            string     `json:"id"`
	PlanID        string     `json:"plan_id"`
	BackendType   string     `json:"backend_type"` // "grafana", "prometheus", "jaeger", "custom"
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	Credentials   string     `json:"credentials"` // Encrypted credentials
	HealthStatus  string     `json:"health_status"` // "healthy", "unhealthy", "unknown"
	LastCheck     *time.Time `json:"last_check,omitempty"`
	DatasourceUID string     `json:"datasource_uid,omitempty"` // For Grafana datasources
	Config        string     `json:"config,omitempty"` // Additional configuration JSON
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PlanDependency represents a relationship between plan components
type PlanDependency struct {
	ID             string `json:"id"`
	PlanID         string `json:"plan_id"`
	SourceID       string `json:"source_id"`
	SourceType     string `json:"source_type"`     // "service", "infrastructure", "pipeline", "backend"
	TargetID       string `json:"target_id"`
	TargetType     string `json:"target_type"`     // Same types
	DependencyType string `json:"dependency_type"` // "data_flow", "depends_on", "used_by"
	CreatedAt      time.Time `json:"created_at"`
}

// PlanUpdate contains fields that can be updated on a plan
type PlanUpdate struct {
	Name        *string
	Description *string
	Environment *string
	Status      *PlanStatus
}
