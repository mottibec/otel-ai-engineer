package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/mottibechhofer/otel-ai-engineer/sandbox"
	"github.com/mottibechhofer/otel-ai-engineer/tools"
)

var sandboxManager *sandbox.Manager

// InitializeSandboxTools initializes the sandbox manager
func InitializeSandboxTools() error {
	logger := sandbox.NewSimpleLogger("sandbox")
	var err error
	sandboxManager, err = sandbox.NewManager(logger)
	if err != nil {
		return fmt.Errorf("failed to initialize sandbox manager: %w", err)
	}
	return nil
}

// GetSandboxManager returns the global sandbox manager
func GetSandboxManager() *sandbox.Manager {
	return sandboxManager
}

// GetSandboxTools returns an array of sandbox tool definitions
func GetSandboxTools() []tools.Tool {
	// Ensure sandbox manager is initialized
	if sandboxManager == nil {
		if err := InitializeSandboxTools(); err != nil {
			// Return empty array if initialization fails
			return []tools.Tool{}
		}
	}

	return getSandboxToolsList()
}

// getSandboxToolsList returns the sandbox tools array (internal helper)
func getSandboxToolsList() []tools.Tool {
	return []tools.Tool{
		{
			Name:        "create_sandbox",
			Description: "Create a new OpenTelemetry collector sandbox for testing. A sandbox is an isolated environment with a collector instance that can be used to test configurations before deploying to production.",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name for the sandbox (e.g., 'test-traces-pipeline', 'debug-exporters')",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Optional description of what this sandbox is testing",
					},
					"collector_config": map[string]interface{}{
						"type":        "string",
						"description": "Complete OpenTelemetry collector configuration in YAML format",
					},
					"collector_version": map[string]interface{}{
						"type":        "string",
						"description": "Collector version to use (e.g., '0.110.0', 'latest'). Defaults to 'latest'",
					},
					"generate_traces": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to generate synthetic trace data (default: false)",
					},
					"generate_metrics": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to generate synthetic metrics data (default: false)",
					},
					"generate_logs": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to generate synthetic logs data (default: false)",
					},
					"trace_rate": map[string]interface{}{
						"type":        "number",
						"description": "Traces per second to generate (default: 1)",
					},
					"metric_rate": map[string]interface{}{
						"type":        "number",
						"description": "Metrics per second to generate (default: 1)",
					},
					"log_rate": map[string]interface{}{
						"type":        "number",
						"description": "Logs per second to generate (default: 1)",
					},
				},
				Required: []string{"name", "collector_config"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input CreateSandboxInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return createSandboxHandler(input)
			},
		},
		{
			Name:        "list_sandboxes",
			Description: "List all active sandboxes with their current status",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input ListSandboxesInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return listSandboxesHandler(input)
			},
		},
		{
			Name:        "get_sandbox",
			Description: "Get detailed information about a specific sandbox including its configuration, status, and last validation results",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"sandbox_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the sandbox to retrieve",
					},
				},
				Required: []string{"sandbox_id"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input GetSandboxInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return getSandboxHandler(input)
			},
		},
		{
			Name:        "start_telemetry",
			Description: "Start generating synthetic telemetry data in a sandbox using telemetrygen. This sends traces, metrics, and/or logs to the collector.",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"sandbox_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the sandbox",
					},
					"duration": map[string]interface{}{
						"type":        "number",
						"description": "How long to generate telemetry in seconds (default: 30)",
					},
					"auto_validate": map[string]interface{}{
						"type":        "boolean",
						"description": "Automatically run validation after telemetry generation completes (default: false)",
					},
				},
				Required: []string{"sandbox_id"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input StartTelemetryInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return startTelemetryHandler(input)
			},
		},
		{
			Name:        "validate_sandbox",
			Description: "Validate a sandbox configuration and operation. This checks for configuration issues, pipeline problems, telemetry flow, and provides AI-powered recommendations.",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"sandbox_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the sandbox to validate",
					},
					"collect_logs": map[string]interface{}{
						"type":        "boolean",
						"description": "Include collector logs in validation (default: true)",
					},
					"collect_metrics": map[string]interface{}{
						"type":        "boolean",
						"description": "Include collector metrics in validation (default: true)",
					},
					"ai_analysis": map[string]interface{}{
						"type":        "boolean",
						"description": "Run AI-powered analysis and generate recommendations (default: true)",
					},
				},
				Required: []string{"sandbox_id"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input ValidateSandboxInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return validateSandboxHandler(input)
			},
		},
		{
			Name:        "get_sandbox_logs",
			Description: "Retrieve collector logs from a sandbox. Useful for debugging configuration issues.",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"sandbox_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the sandbox",
					},
					"tail": map[string]interface{}{
						"type":        "number",
						"description": "Number of recent log lines to retrieve (default: 100)",
					},
				},
				Required: []string{"sandbox_id"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input GetSandboxLogsInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return getSandboxLogsHandler(input)
			},
		},
		{
			Name:        "get_sandbox_metrics",
			Description: "Retrieve internal collector metrics from a sandbox. Shows receiver, processor, exporter stats, queue sizes, and resource usage.",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"sandbox_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the sandbox",
					},
				},
				Required: []string{"sandbox_id"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input GetSandboxMetricsInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return getSandboxMetricsHandler(input)
			},
		},
		{
			Name:        "stop_sandbox",
			Description: "Stop a sandbox and all its containers (collector and telemetry generators). The sandbox can be restarted later.",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"sandbox_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the sandbox to stop",
					},
				},
				Required: []string{"sandbox_id"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input StopSandboxInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return stopSandboxHandler(input)
			},
		},
		{
			Name:        "delete_sandbox",
			Description: "Permanently delete a sandbox and all its resources. This cannot be undone.",
			Schema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"sandbox_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the sandbox to delete",
					},
				},
				Required: []string{"sandbox_id"},
			},
			Handler: func(inputJSON json.RawMessage) (interface{}, error) {
				var input DeleteSandboxInput
				if err := json.Unmarshal(inputJSON, &input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal input: %w", err)
				}
				return deleteSandboxHandler(input)
			},
		},
	}
}

// RegisterSandboxTools registers all sandbox-related tools
func RegisterSandboxTools(registry *tools.ToolRegistry) error {
	if err := InitializeSandboxTools(); err != nil {
		return err
	}

	sandboxTools := getSandboxToolsList()
	registry.RegisterTools(sandboxTools)
	return nil
}

// Tool input structures

type CreateSandboxInput struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	CollectorConfig  string `json:"collector_config"`
	CollectorVersion string `json:"collector_version"`
	GenerateTraces   bool   `json:"generate_traces"`
	GenerateMetrics  bool   `json:"generate_metrics"`
	GenerateLogs     bool   `json:"generate_logs"`
	TraceRate        int    `json:"trace_rate"`
	MetricRate       int    `json:"metric_rate"`
	LogRate          int    `json:"log_rate"`
}

type ListSandboxesInput struct{}

type GetSandboxInput struct {
	SandboxID string `json:"sandbox_id"`
}

type StartTelemetryInput struct {
	SandboxID    string `json:"sandbox_id"`
	Duration     int    `json:"duration"`
	AutoValidate bool   `json:"auto_validate"`
}

type ValidateSandboxInput struct {
	SandboxID      string `json:"sandbox_id"`
	CollectLogs    bool   `json:"collect_logs"`
	CollectMetrics bool   `json:"collect_metrics"`
	AIAnalysis     bool   `json:"ai_analysis"`
}

type GetSandboxLogsInput struct {
	SandboxID string `json:"sandbox_id"`
	Tail      int    `json:"tail"`
}

type GetSandboxMetricsInput struct {
	SandboxID string `json:"sandbox_id"`
}

type StopSandboxInput struct {
	SandboxID string `json:"sandbox_id"`
}

type DeleteSandboxInput struct {
	SandboxID string `json:"sandbox_id"`
}

// Tool handlers

func createSandboxHandler(input CreateSandboxInput) (interface{}, error) {
	ctx := context.Background()

	req := sandbox.CreateSandboxRequest{
		Name:             input.Name,
		Description:      input.Description,
		CollectorConfig:  input.CollectorConfig,
		CollectorVersion: input.CollectorVersion,
		TelemetryConfig: sandbox.TelemetryConfig{
			GenerateTraces:  input.GenerateTraces,
			GenerateMetrics: input.GenerateMetrics,
			GenerateLogs:    input.GenerateLogs,
			TraceRate:       input.TraceRate,
			MetricRate:      input.MetricRate,
			LogRate:         input.LogRate,
		},
	}

	sb, err := sandboxManager.CreateSandbox(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"sandbox_id": sb.ID,
		"sandbox":    sb,
		"message":    fmt.Sprintf("Sandbox '%s' created successfully", sb.Name),
	}, nil
}

func listSandboxesHandler(input ListSandboxesInput) (interface{}, error) {
	sandboxes := sandboxManager.ListSandboxes()

	return map[string]interface{}{
		"success":   true,
		"sandboxes": sandboxes,
		"count":     len(sandboxes),
	}, nil
}

func getSandboxHandler(input GetSandboxInput) (interface{}, error) {
	sb, err := sandboxManager.GetSandbox(input.SandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sandbox: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"sandbox": sb,
	}, nil
}

func startTelemetryHandler(input StartTelemetryInput) (interface{}, error) {
	ctx := context.Background()

	duration := time.Duration(input.Duration) * time.Second
	if duration == 0 {
		duration = 30 * time.Second
	}

	req := sandbox.StartSandboxRequest{
		Duration:     duration,
		AutoValidate: input.AutoValidate,
	}

	err := sandboxManager.StartTelemetry(ctx, input.SandboxID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start telemetry: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Telemetry generation started for %v", duration),
	}, nil
}

func validateSandboxHandler(input ValidateSandboxInput) (interface{}, error) {
	ctx := context.Background()

	// Default to true if not specified
	collectLogs := true
	collectMetrics := true
	aiAnalysis := true

	req := sandbox.ValidateSandboxRequest{
		CollectLogs:    collectLogs,
		CollectMetrics: collectMetrics,
		AIAnalysis:     aiAnalysis,
	}

	result, err := sandboxManager.ValidateSandbox(ctx, input.SandboxID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate sandbox: %w", err)
	}

	return map[string]interface{}{
		"success":    true,
		"validation": result,
	}, nil
}

func getSandboxLogsHandler(input GetSandboxLogsInput) (interface{}, error) {
	ctx := context.Background()

	tail := input.Tail
	if tail == 0 {
		tail = 100
	}

	logs, err := sandboxManager.GetCollectorLogs(ctx, input.SandboxID, tail)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"logs":    logs,
		"count":   len(logs),
	}, nil
}

func getSandboxMetricsHandler(input GetSandboxMetricsInput) (interface{}, error) {
	ctx := context.Background()

	metrics, err := sandboxManager.GetCollectorMetrics(ctx, input.SandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"metrics": metrics,
	}, nil
}

func stopSandboxHandler(input StopSandboxInput) (interface{}, error) {
	ctx := context.Background()

	err := sandboxManager.StopSandbox(ctx, input.SandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to stop sandbox: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Sandbox stopped successfully",
	}, nil
}

func deleteSandboxHandler(input DeleteSandboxInput) (interface{}, error) {
	ctx := context.Background()

	err := sandboxManager.DeleteSandbox(ctx, input.SandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete sandbox: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Sandbox deleted successfully",
	}, nil
}
