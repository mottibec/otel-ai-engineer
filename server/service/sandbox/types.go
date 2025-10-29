package sandbox

import (
	"github.com/mottibechhofer/otel-ai-engineer/sandbox"
)

// ListSandboxesResponse represents the response for listing sandboxes
type ListSandboxesResponse struct {
	Sandboxes []sandbox.Sandbox `json:"sandboxes"`
	Count     int               `json:"count"`
}

// CreateSandboxResponse represents the response for creating a sandbox
type CreateSandboxResponse struct {
	Success   bool             `json:"success"`
	SandboxID string           `json:"sandbox_id"`
	Sandbox   *sandbox.Sandbox `json:"sandbox"`
}

// StartTelemetryResponse represents the response for starting telemetry
type StartTelemetryResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ValidateSandboxResponse represents the response for validating a sandbox
type ValidateSandboxResponse struct {
	Success    bool                      `json:"success"`
	Validation *sandbox.ValidationResult `json:"validation"`
}

// GetSandboxLogsResponse represents the response for getting sandbox logs
type GetSandboxLogsResponse struct {
	Success bool     `json:"success"`
	Logs    []string `json:"logs"`
	Count   int      `json:"count"`
}

// GetSandboxMetricsResponse represents the response for getting sandbox metrics
type GetSandboxMetricsResponse struct {
	Success bool                      `json:"success"`
	Metrics *sandbox.CollectorMetrics `json:"metrics"`
}

// StopSandboxResponse represents the response for stopping a sandbox
type StopSandboxResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeleteSandboxResponse represents the response for deleting a sandbox
type DeleteSandboxResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
