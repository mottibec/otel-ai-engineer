package sandbox

import (
	"time"
)

// SandboxStatus represents the current state of a sandbox
type SandboxStatus string

const (
	SandboxStatusCreating  SandboxStatus = "creating"
	SandboxStatusRunning   SandboxStatus = "running"
	SandboxStatusStopped   SandboxStatus = "stopped"
	SandboxStatusFailed    SandboxStatus = "failed"
	SandboxStatusValidating SandboxStatus = "validating"
)

// Sandbox represents an isolated testing environment
type Sandbox struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	CollectorConfig  string                 `json:"collector_config"`   // YAML configuration
	CollectorVersion string                 `json:"collector_version"`  // e.g., "0.110.0", "latest"
	Status           SandboxStatus          `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`

	// Container information
	CollectorContainerID string `json:"collector_container_id,omitempty"`
	CollectorContainerName string `json:"collector_container_name,omitempty"`

	// Network information
	NetworkID   string `json:"network_id,omitempty"`
	NetworkName string `json:"network_name,omitempty"`

	// Test configuration
	TelemetryConfig TelemetryConfig `json:"telemetry_config"`

	// Results
	LastValidation *ValidationResult `json:"last_validation,omitempty"`

	// Metadata
	Tags     map[string]string `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TelemetryConfig defines what kind of telemetry to generate
type TelemetryConfig struct {
	// Telemetrygen configuration
	GenerateTraces  bool              `json:"generate_traces"`
	GenerateMetrics bool              `json:"generate_metrics"`
	GenerateLogs    bool              `json:"generate_logs"`

	// Traces configuration
	TraceRate       int               `json:"trace_rate"`        // Traces per second
	TraceAttributes map[string]string `json:"trace_attributes"`  // Custom attributes
	TraceDuration   time.Duration     `json:"trace_duration"`    // How long to generate

	// Metrics configuration
	MetricRate      int               `json:"metric_rate"`       // Metrics per second
	MetricTypes     []string          `json:"metric_types"`      // e.g., ["counter", "gauge", "histogram"]

	// Logs configuration
	LogRate         int               `json:"log_rate"`          // Logs per second
	LogSeverity     []string          `json:"log_severity"`      // e.g., ["info", "warn", "error"]

	// OTLP endpoint (within sandbox network)
	OTLPEndpoint    string            `json:"otlp_endpoint"`     // e.g., "collector:4317"
	OTLPProtocol    string            `json:"otlp_protocol"`     // "grpc" or "http"
}

// ValidationResult contains the results of validating a sandbox
type ValidationResult struct {
	ID          string              `json:"id"`
	SandboxID   string              `json:"sandbox_id"`
	Status      ValidationStatus    `json:"status"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt time.Time           `json:"completed_at"`
	Duration    time.Duration       `json:"duration"`

	// Validation checks
	Checks      []ValidationCheck   `json:"checks"`
	Summary     ValidationSummary   `json:"summary"`

	// Collected data
	CollectorLogs    []LogEntry      `json:"collector_logs,omitempty"`
	CollectorMetrics CollectorMetrics `json:"collector_metrics,omitempty"`

	// Issues found
	Issues      []ValidationIssue   `json:"issues,omitempty"`

	// AI analysis
	AIAnalysis  string              `json:"ai_analysis,omitempty"`
	Recommendations []string         `json:"recommendations,omitempty"`
}

// ValidationStatus represents validation state
type ValidationStatus string

const (
	ValidationStatusPending   ValidationStatus = "pending"
	ValidationStatusRunning   ValidationStatus = "running"
	ValidationStatusPassed    ValidationStatus = "passed"
	ValidationStatusFailed    ValidationStatus = "failed"
	ValidationStatusPartial   ValidationStatus = "partial"
)

// ValidationCheck represents a single validation check
type ValidationCheck struct {
	Name        string           `json:"name"`
	Category    string           `json:"category"`    // e.g., "pipeline", "receiver", "exporter"
	Status      string           `json:"status"`      // "passed", "failed", "warning"
	Message     string           `json:"message"`
	Details     string           `json:"details,omitempty"`
	Severity    string           `json:"severity"`    // "critical", "high", "medium", "low"
	Timestamp   time.Time        `json:"timestamp"`
}

// ValidationSummary provides overall statistics
type ValidationSummary struct {
	TotalChecks   int `json:"total_checks"`
	Passed        int `json:"passed"`
	Failed        int `json:"failed"`
	Warnings      int `json:"warnings"`
	Critical      int `json:"critical"`

	// Telemetry statistics
	TracesReceived  int64 `json:"traces_received"`
	MetricsReceived int64 `json:"metrics_received"`
	LogsReceived    int64 `json:"logs_received"`

	TracesExported  int64 `json:"traces_exported"`
	MetricsExported int64 `json:"metrics_exported"`
	LogsExported    int64 `json:"logs_exported"`

	// Data loss
	DataLossPercent float64 `json:"data_loss_percent"`
}

// ValidationIssue represents a configuration or operational issue
type ValidationIssue struct {
	Type        string    `json:"type"`        // "configuration", "pipeline", "exporter", "receiver"
	Severity    string    `json:"severity"`    // "critical", "high", "medium", "low"
	Component   string    `json:"component"`   // Which component has the issue
	Message     string    `json:"message"`
	Description string    `json:"description"`
	Suggestion  string    `json:"suggestion,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// LogEntry represents a collector log line
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// CollectorMetrics represents internal collector metrics
type CollectorMetrics struct {
	// Receiver metrics
	ReceiverAcceptedSpans   int64 `json:"receiver_accepted_spans"`
	ReceiverRefusedSpans    int64 `json:"receiver_refused_spans"`
	ReceiverAcceptedMetrics int64 `json:"receiver_accepted_metrics"`
	ReceiverRefusedMetrics  int64 `json:"receiver_refused_metrics"`
	ReceiverAcceptedLogs    int64 `json:"receiver_accepted_logs"`
	ReceiverRefusedLogs     int64 `json:"receiver_refused_logs"`

	// Processor metrics
	ProcessorAcceptedSpans   int64 `json:"processor_accepted_spans"`
	ProcessorRefusedSpans    int64 `json:"processor_refused_spans"`
	ProcessorDroppedSpans    int64 `json:"processor_dropped_spans"`
	ProcessorAcceptedMetrics int64 `json:"processor_accepted_metrics"`
	ProcessorRefusedMetrics  int64 `json:"processor_refused_metrics"`
	ProcessorDroppedMetrics  int64 `json:"processor_dropped_metrics"`

	// Exporter metrics
	ExporterSentSpans      int64 `json:"exporter_sent_spans"`
	ExporterFailedSpans    int64 `json:"exporter_failed_spans"`
	ExporterSentMetrics    int64 `json:"exporter_sent_metrics"`
	ExporterFailedMetrics  int64 `json:"exporter_failed_metrics"`
	ExporterSentLogs       int64 `json:"exporter_sent_logs"`
	ExporterFailedLogs     int64 `json:"exporter_failed_logs"`

	// Queue metrics
	QueueSize              int64 `json:"queue_size"`
	QueueCapacity          int64 `json:"queue_capacity"`

	// Resource metrics
	MemoryUsageMB          float64 `json:"memory_usage_mb"`
	CPUUsagePercent        float64 `json:"cpu_usage_percent"`
}

// SandboxListResponse represents a list of sandboxes
type SandboxListResponse struct {
	Sandboxes []Sandbox `json:"sandboxes"`
	Total     int       `json:"total"`
}

// CreateSandboxRequest represents a request to create a sandbox
type CreateSandboxRequest struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	CollectorConfig  string                 `json:"collector_config"`
	CollectorVersion string                 `json:"collector_version,omitempty"`
	TelemetryConfig  TelemetryConfig        `json:"telemetry_config,omitempty"`
	Tags             map[string]string      `json:"tags,omitempty"`
}

// StartSandboxRequest represents a request to start telemetry generation
type StartSandboxRequest struct {
	Duration         time.Duration   `json:"duration,omitempty"`      // How long to run (0 = run indefinitely)
	TelemetryConfig  *TelemetryConfig `json:"telemetry_config,omitempty"` // Override config
	AutoValidate     bool            `json:"auto_validate"`           // Run validation after completion
}

// ValidateSandboxRequest represents a request to validate a sandbox
type ValidateSandboxRequest struct {
	RunChecks       []string `json:"run_checks,omitempty"`        // Specific checks to run (empty = all)
	CollectLogs     bool     `json:"collect_logs"`                // Include collector logs
	CollectMetrics  bool     `json:"collect_metrics"`             // Include collector metrics
	AIAnalysis      bool     `json:"ai_analysis"`                 // Run AI-powered analysis
}

// TelemetryGeneratorContainerInfo holds info about telemetrygen containers
type TelemetryGeneratorContainerInfo struct {
	TracesContainerID  string `json:"traces_container_id,omitempty"`
	MetricsContainerID string `json:"metrics_container_id,omitempty"`
	LogsContainerID    string `json:"logs_container_id,omitempty"`
}
