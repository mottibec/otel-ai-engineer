package sandbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager manages sandbox lifecycles
type Manager struct {
	dockerOrchestrator *DockerOrchestrator
	validator          *Validator
	sandboxes          map[string]*Sandbox
	mu                 sync.RWMutex
	logger             Logger
}

// Logger interface for logging
type Logger interface {
	Info(msg string, fields map[string]interface{})
	Error(msg string, err error, fields map[string]interface{})
	Debug(msg string, fields map[string]interface{})
}

// NewManager creates a new sandbox manager
func NewManager(logger Logger) (*Manager, error) {
	orchestrator, err := NewDockerOrchestrator(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker orchestrator: %w", err)
	}

	validator := NewValidator(orchestrator, logger)

	return &Manager{
		dockerOrchestrator: orchestrator,
		validator:          validator,
		sandboxes:          make(map[string]*Sandbox),
		logger:             logger,
	}, nil
}

// CreateSandbox creates a new sandbox environment
func (m *Manager) CreateSandbox(ctx context.Context, req CreateSandboxRequest) (*Sandbox, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate unique ID
	sandboxID := uuid.New().String()

	// Set defaults
	if req.CollectorVersion == "" {
		req.CollectorVersion = "latest"
	}

	if req.TelemetryConfig.OTLPEndpoint == "" {
		req.TelemetryConfig.OTLPEndpoint = "collector:4317"
	}

	if req.TelemetryConfig.OTLPProtocol == "" {
		req.TelemetryConfig.OTLPProtocol = "grpc"
	}

	// Create sandbox object
	sandbox := &Sandbox{
		ID:               sandboxID,
		Name:             req.Name,
		Description:      req.Description,
		CollectorConfig:  req.CollectorConfig,
		CollectorVersion: req.CollectorVersion,
		Status:           SandboxStatusCreating,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		TelemetryConfig:  req.TelemetryConfig,
		Tags:             req.Tags,
		Metadata:         make(map[string]interface{}),
	}

	m.logger.Info("Creating sandbox", map[string]interface{}{
		"sandbox_id": sandboxID,
		"name":       req.Name,
	})

	// Create Docker network for isolation
	networkName, networkID, err := m.dockerOrchestrator.CreateNetwork(ctx, sandboxID)
	if err != nil {
		sandbox.Status = SandboxStatusFailed
		return sandbox, fmt.Errorf("failed to create network: %w", err)
	}

	sandbox.NetworkName = networkName
	sandbox.NetworkID = networkID

	// Deploy collector
	collectorInfo, err := m.dockerOrchestrator.DeployCollector(ctx, DeployCollectorConfig{
		SandboxID:        sandboxID,
		Config:           req.CollectorConfig,
		CollectorVersion: req.CollectorVersion,
		NetworkName:      networkName,
	})
	if err != nil {
		// Cleanup network
		_ = m.dockerOrchestrator.CleanupNetwork(ctx, sandboxID)
		sandbox.Status = SandboxStatusFailed
		return sandbox, fmt.Errorf("failed to deploy collector: %w", err)
	}

	sandbox.CollectorContainerID = collectorInfo.ContainerID
	sandbox.CollectorContainerName = collectorInfo.ContainerName
	sandbox.Status = SandboxStatusRunning
	sandbox.UpdatedAt = time.Now()

	// Store sandbox
	m.sandboxes[sandboxID] = sandbox

	m.logger.Info("Sandbox created successfully", map[string]interface{}{
		"sandbox_id":   sandboxID,
		"collector_id": collectorInfo.ContainerID,
	})

	return sandbox, nil
}

// GetSandbox retrieves a sandbox by ID and refreshes its status
func (m *Manager) GetSandbox(sandboxID string) (*Sandbox, error) {
	m.mu.RLock()
	sandbox, exists := m.sandboxes[sandboxID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("sandbox not found: %s", sandboxID)
	}

	// Refresh status from actual container state
	ctx := context.Background()
	m.refreshSandboxStatus(ctx, sandbox)

	return sandbox, nil
}

// refreshSandboxStatus checks the actual container status and updates the sandbox
func (m *Manager) refreshSandboxStatus(ctx context.Context, sandbox *Sandbox) {
	if sandbox.CollectorContainerID == "" {
		return
	}

	// Get actual container status
	status, err := m.dockerOrchestrator.getContainerStatus(ctx, sandbox.CollectorContainerID)
	if err != nil {
		m.logger.Error("Failed to get container status", err, map[string]interface{}{
			"sandbox_id":   sandbox.ID,
			"container_id": sandbox.CollectorContainerID,
		})
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	oldStatus := sandbox.Status

	// Map Docker status to sandbox status
	switch status {
	case "running":
		if sandbox.Status != SandboxStatusRunning {
			sandbox.Status = SandboxStatusRunning
			sandbox.UpdatedAt = time.Now()
		}
	case "exited", "dead":
		if sandbox.Status != SandboxStatusFailed {
			sandbox.Status = SandboxStatusFailed
			sandbox.UpdatedAt = time.Now()

			// Capture error logs when container fails
			logs, err := m.dockerOrchestrator.GetContainerLogs(ctx, sandbox.CollectorContainerID, 50)
			if err == nil && len(logs) > 0 {
				// Store error logs in metadata
				if sandbox.Metadata == nil {
					sandbox.Metadata = make(map[string]interface{})
				}
				errorLogs := make([]string, 0)
				for _, log := range logs {
					if log.Level == "error" || containsError(log.Message) {
						errorLogs = append(errorLogs, log.Message)
					}
				}
				if len(errorLogs) > 0 {
					sandbox.Metadata["error_logs"] = errorLogs
				}
			}
		}
	case "created", "restarting":
		if sandbox.Status != SandboxStatusCreating {
			sandbox.Status = SandboxStatusCreating
			sandbox.UpdatedAt = time.Now()
		}
	case "paused":
		if sandbox.Status != SandboxStatusStopped {
			sandbox.Status = SandboxStatusStopped
			sandbox.UpdatedAt = time.Now()
		}
	}

	if oldStatus != sandbox.Status {
		m.logger.Info("Sandbox status changed", map[string]interface{}{
			"sandbox_id": sandbox.ID,
			"old_status": oldStatus,
			"new_status": sandbox.Status,
		})
	}
}

// containsError checks if a log message contains error indicators
func containsError(message string) bool {
	errorKeywords := []string{"error:", "Error:", "ERROR:", "failed", "Failed", "FAILED"}
	for _, keyword := range errorKeywords {
		if len(message) > 0 && len(keyword) > 0 {
			// Simple string contains check
			for i := 0; i <= len(message)-len(keyword); i++ {
				if message[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// ListSandboxes returns all sandboxes with refreshed status
func (m *Manager) ListSandboxes() []Sandbox {
	ctx := context.Background()

	m.mu.RLock()
	sandboxList := make([]*Sandbox, 0, len(m.sandboxes))
	for _, sandbox := range m.sandboxes {
		sandboxList = append(sandboxList, sandbox)
	}
	m.mu.RUnlock()

	// Refresh status for each sandbox
	for _, sandbox := range sandboxList {
		m.refreshSandboxStatus(ctx, sandbox)
	}

	// Convert to value slice
	sandboxes := make([]Sandbox, 0, len(sandboxList))
	for _, sandbox := range sandboxList {
		sandboxes = append(sandboxes, *sandbox)
	}

	return sandboxes
}

// StartTelemetry starts generating telemetry in a sandbox
func (m *Manager) StartTelemetry(ctx context.Context, sandboxID string, req StartSandboxRequest) error {
	sandbox, err := m.GetSandbox(sandboxID)
	if err != nil {
		return err
	}

	if sandbox.Status != SandboxStatusRunning {
		return fmt.Errorf("sandbox is not running: %s", sandbox.Status)
	}

	// Override telemetry config if provided
	telemetryConfig := sandbox.TelemetryConfig
	if req.TelemetryConfig != nil {
		telemetryConfig = *req.TelemetryConfig
	}

	// Set duration
	duration := req.Duration
	if duration == 0 {
		// Default to 30 seconds if not specified
		duration = 30 * time.Second
	}

	m.logger.Info("Starting telemetry generation", map[string]interface{}{
		"sandbox_id": sandboxID,
		"duration":   duration,
	})

	// Start telemetrygen containers
	generatorInfo, err := m.dockerOrchestrator.StartTelemetryGeneration(ctx, StartTelemetryConfig{
		SandboxID:       sandboxID,
		NetworkName:     sandbox.NetworkName,
		TelemetryConfig: telemetryConfig,
		Duration:        duration,
	})
	if err != nil {
		return fmt.Errorf("failed to start telemetry generation: %w", err)
	}

	// Store generator info in metadata
	m.mu.Lock()
	sandbox.Metadata["telemetry_generator"] = generatorInfo
	sandbox.UpdatedAt = time.Now()
	m.mu.Unlock()

	// If auto-validate is enabled, wait for duration and then validate
	if req.AutoValidate {
		go func() {
			// Wait for telemetry generation to complete
			time.Sleep(duration + 2*time.Second) // Extra buffer

			m.logger.Info("Auto-validating sandbox", map[string]interface{}{
				"sandbox_id": sandboxID,
			})

			result, err := m.ValidateSandbox(context.Background(), sandboxID, ValidateSandboxRequest{
				CollectLogs:    true,
				CollectMetrics: true,
				AIAnalysis:     true,
			})
			if err != nil {
				m.logger.Error("Auto-validation failed", err, map[string]interface{}{
					"sandbox_id": sandboxID,
				})
				return
			}

			m.mu.Lock()
			sandbox.LastValidation = result
			m.mu.Unlock()

			m.logger.Info("Auto-validation completed", map[string]interface{}{
				"sandbox_id": sandboxID,
				"status":     result.Status,
			})
		}()
	}

	return nil
}

// ValidateSandbox validates a sandbox configuration and operation
func (m *Manager) ValidateSandbox(ctx context.Context, sandboxID string, req ValidateSandboxRequest) (*ValidationResult, error) {
	sandbox, err := m.GetSandbox(sandboxID)
	if err != nil {
		return nil, err
	}

	m.logger.Info("Validating sandbox", map[string]interface{}{
		"sandbox_id": sandboxID,
	})

	// Update sandbox status
	m.mu.Lock()
	sandbox.Status = SandboxStatusValidating
	sandbox.UpdatedAt = time.Now()
	m.mu.Unlock()

	// Run validation
	result, err := m.validator.Validate(ctx, sandbox, req)
	if err != nil {
		m.mu.Lock()
		sandbox.Status = SandboxStatusFailed
		m.mu.Unlock()
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update sandbox with validation result
	m.mu.Lock()
	sandbox.LastValidation = result
	sandbox.Status = SandboxStatusRunning
	sandbox.UpdatedAt = time.Now()
	m.mu.Unlock()

	m.logger.Info("Validation completed", map[string]interface{}{
		"sandbox_id":   sandboxID,
		"status":       result.Status,
		"checks_passed": result.Summary.Passed,
		"checks_failed": result.Summary.Failed,
	})

	return result, nil
}

// StopSandbox stops a sandbox and all its containers
func (m *Manager) StopSandbox(ctx context.Context, sandboxID string) error {
	sandbox, err := m.GetSandbox(sandboxID)
	if err != nil {
		return err
	}

	m.logger.Info("Stopping sandbox", map[string]interface{}{
		"sandbox_id": sandboxID,
	})

	// Stop telemetry generators if running
	if generatorInfo, ok := sandbox.Metadata["telemetry_generator"].(TelemetryGeneratorContainerInfo); ok {
		_ = m.dockerOrchestrator.StopTelemetryGenerators(ctx, generatorInfo)
	}

	// Stop collector
	if err := m.dockerOrchestrator.StopCollector(ctx, sandbox.CollectorContainerID); err != nil {
		m.logger.Error("Failed to stop collector", err, map[string]interface{}{
			"sandbox_id": sandboxID,
		})
	}

	// Cleanup network
	if err := m.dockerOrchestrator.CleanupNetwork(ctx, sandboxID); err != nil {
		m.logger.Error("Failed to cleanup network", err, map[string]interface{}{
			"sandbox_id": sandboxID,
		})
	}

	m.mu.Lock()
	sandbox.Status = SandboxStatusStopped
	sandbox.UpdatedAt = time.Now()
	m.mu.Unlock()

	m.logger.Info("Sandbox stopped", map[string]interface{}{
		"sandbox_id": sandboxID,
	})

	return nil
}

// DeleteSandbox removes a sandbox and all its resources
func (m *Manager) DeleteSandbox(ctx context.Context, sandboxID string) error {
	// Stop sandbox first
	if err := m.StopSandbox(ctx, sandboxID); err != nil {
		m.logger.Error("Failed to stop sandbox during deletion", err, map[string]interface{}{
			"sandbox_id": sandboxID,
		})
	}

	// Remove from memory
	m.mu.Lock()
	delete(m.sandboxes, sandboxID)
	m.mu.Unlock()

	m.logger.Info("Sandbox deleted", map[string]interface{}{
		"sandbox_id": sandboxID,
	})

	return nil
}

// GetCollectorLogs retrieves logs from the collector
func (m *Manager) GetCollectorLogs(ctx context.Context, sandboxID string, tailLines int) ([]LogEntry, error) {
	sandbox, err := m.GetSandbox(sandboxID)
	if err != nil {
		return nil, err
	}

	logs, err := m.dockerOrchestrator.GetContainerLogs(ctx, sandbox.CollectorContainerID, tailLines)
	if err != nil {
		return nil, fmt.Errorf("failed to get collector logs: %w", err)
	}

	return logs, nil
}

// GetCollectorMetrics retrieves internal metrics from the collector
func (m *Manager) GetCollectorMetrics(ctx context.Context, sandboxID string) (*CollectorMetrics, error) {
	sandbox, err := m.GetSandbox(sandboxID)
	if err != nil {
		return nil, err
	}

	metrics, err := m.dockerOrchestrator.GetCollectorMetrics(ctx, sandbox.CollectorContainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collector metrics: %w", err)
	}

	return metrics, nil
}

// Cleanup stops and removes all sandboxes
func (m *Manager) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	sandboxIDs := make([]string, 0, len(m.sandboxes))
	for id := range m.sandboxes {
		sandboxIDs = append(sandboxIDs, id)
	}
	m.mu.Unlock()

	for _, id := range sandboxIDs {
		if err := m.DeleteSandbox(ctx, id); err != nil {
			m.logger.Error("Failed to delete sandbox during cleanup", err, map[string]interface{}{
				"sandbox_id": id,
			})
		}
	}

	return nil
}
