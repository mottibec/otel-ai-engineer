package sandbox

import (
	"context"
	"fmt"

	"github.com/mottibechhofer/otel-ai-engineer/sandbox"
	sandboxTools "github.com/mottibechhofer/otel-ai-engineer/tools/sandbox"
)

// SandboxService handles business logic for sandbox management
type SandboxService struct {
	manager *sandbox.Manager
}

// NewSandboxService creates a new sandbox service
func NewSandboxService() *SandboxService {
	return &SandboxService{
		manager: nil, // Will be lazily initialized
	}
}

// ensureManager ensures the sandbox manager is initialized
func (ss *SandboxService) ensureManager() error {
	if ss.manager == nil {
		manager := sandboxTools.GetSandboxManager()
		if manager == nil {
			// Try to initialize it
			if err := sandboxTools.InitializeSandboxTools(); err != nil {
				return fmt.Errorf("failed to initialize sandbox manager: %w", err)
			}
			manager = sandboxTools.GetSandboxManager()
			if manager == nil {
				return fmt.Errorf("sandbox manager not initialized")
			}
		}
		ss.manager = manager
	}
	return nil
}

// ListSandboxes lists all sandboxes
func (ss *SandboxService) ListSandboxes(ctx context.Context) (*ListSandboxesResponse, error) {
	if err := ss.ensureManager(); err != nil {
		return nil, err
	}

	sandboxes := ss.manager.ListSandboxes()
	return &ListSandboxesResponse{
		Sandboxes: sandboxes,
		Count:     len(sandboxes),
	}, nil
}

// GetSandbox retrieves a sandbox by ID
func (ss *SandboxService) GetSandbox(ctx context.Context, sandboxID string) (*sandbox.Sandbox, error) {
	if err := ss.ensureManager(); err != nil {
		return nil, err
	}

	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID cannot be empty")
	}

	return ss.manager.GetSandbox(sandboxID)
}

// CreateSandbox creates a new sandbox
func (ss *SandboxService) CreateSandbox(ctx context.Context, req sandbox.CreateSandboxRequest) (*CreateSandboxResponse, error) {
	if err := ss.ensureManager(); err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if req.CollectorConfig == "" {
		return nil, fmt.Errorf("collector_config is required")
	}

	sb, err := ss.manager.CreateSandbox(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}

	return &CreateSandboxResponse{
		Success:   true,
		SandboxID: sb.ID,
		Sandbox:   sb,
	}, nil
}

// StartTelemetry starts telemetry generation for a sandbox
func (ss *SandboxService) StartTelemetry(ctx context.Context, sandboxID string, req sandbox.StartSandboxRequest) (*StartTelemetryResponse, error) {
	if err := ss.ensureManager(); err != nil {
		return nil, err
	}

	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID cannot be empty")
	}

	if err := ss.manager.StartTelemetry(ctx, sandboxID, req); err != nil {
		return nil, fmt.Errorf("failed to start telemetry: %w", err)
	}

	return &StartTelemetryResponse{
		Success: true,
		Message: "Telemetry generation started",
	}, nil
}

// ValidateSandbox validates a sandbox configuration
func (ss *SandboxService) ValidateSandbox(ctx context.Context, sandboxID string, req sandbox.ValidateSandboxRequest) (*ValidateSandboxResponse, error) {
	if err := ss.ensureManager(); err != nil {
		return nil, err
	}

	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID cannot be empty")
	}

	result, err := ss.manager.ValidateSandbox(ctx, sandboxID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate sandbox: %w", err)
	}

	return &ValidateSandboxResponse{
		Success:    true,
		Validation: result,
	}, nil
}

// GetSandboxLogs retrieves logs for a sandbox collector
func (ss *SandboxService) GetSandboxLogs(ctx context.Context, sandboxID string, tail int) (*GetSandboxLogsResponse, error) {
	if err := ss.ensureManager(); err != nil {
		return nil, err
	}

	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID cannot be empty")
	}

	if tail <= 0 {
		tail = 100 // default
	}

	logEntries, err := ss.manager.GetCollectorLogs(ctx, sandboxID, tail)
	if err != nil {
		return nil, fmt.Errorf("failed to get collector logs: %w", err)
	}

	// Convert LogEntry to strings
	logs := make([]string, 0, len(logEntries))
	for _, entry := range logEntries {
		logs = append(logs, entry.Message)
	}

	return &GetSandboxLogsResponse{
		Success: true,
		Logs:    logs,
		Count:   len(logs),
	}, nil
}

// GetSandboxMetrics retrieves metrics for a sandbox collector
func (ss *SandboxService) GetSandboxMetrics(ctx context.Context, sandboxID string) (*GetSandboxMetricsResponse, error) {
	if err := ss.ensureManager(); err != nil {
		return nil, err
	}

	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID cannot be empty")
	}

	metrics, err := ss.manager.GetCollectorMetrics(ctx, sandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collector metrics: %w", err)
	}

	return &GetSandboxMetricsResponse{
		Success: true,
		Metrics: metrics,
	}, nil
}

// StopSandbox stops a sandbox
func (ss *SandboxService) StopSandbox(ctx context.Context, sandboxID string) (*StopSandboxResponse, error) {
	if err := ss.ensureManager(); err != nil {
		return nil, err
	}

	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID cannot be empty")
	}

	if err := ss.manager.StopSandbox(ctx, sandboxID); err != nil {
		return nil, fmt.Errorf("failed to stop sandbox: %w", err)
	}

	return &StopSandboxResponse{
		Success: true,
		Message: "Sandbox stopped",
	}, nil
}

// DeleteSandbox deletes a sandbox
func (ss *SandboxService) DeleteSandbox(ctx context.Context, sandboxID string) (*DeleteSandboxResponse, error) {
	if err := ss.ensureManager(); err != nil {
		return nil, err
	}

	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID cannot be empty")
	}

	if err := ss.manager.DeleteSandbox(ctx, sandboxID); err != nil {
		return nil, fmt.Errorf("failed to delete sandbox: %w", err)
	}

	return &DeleteSandboxResponse{
		Success: true,
		Message: "Sandbox deleted",
	}, nil
}

