package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DockerOrchestrator manages Docker containers for sandboxes
type DockerOrchestrator struct {
	dockerPath string
	logger     Logger
}

// DeployCollectorConfig holds collector deployment configuration
type DeployCollectorConfig struct {
	SandboxID        string
	Config           string
	CollectorVersion string
	NetworkName      string
}

// CollectorInfo holds information about a deployed collector
type CollectorInfo struct {
	ContainerID   string
	ContainerName string
	Status        string
}

// StartTelemetryConfig holds telemetry generation configuration
type StartTelemetryConfig struct {
	SandboxID       string
	NetworkName     string
	TelemetryConfig TelemetryConfig
	Duration        time.Duration
}

// NewDockerOrchestrator creates a new Docker orchestrator
func NewDockerOrchestrator(logger Logger) (*DockerOrchestrator, error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	return &DockerOrchestrator{
		dockerPath: dockerPath,
		logger:     logger,
	}, nil
}

// CreateNetwork creates an isolated Docker network for a sandbox
func (d *DockerOrchestrator) CreateNetwork(ctx context.Context, sandboxID string) (string, string, error) {
	networkName := fmt.Sprintf("sandbox-%s", sandboxID)

	args := []string{
		"network", "create",
		"--driver", "bridge",
		"--label", fmt.Sprintf("sandbox.id=%s", sandboxID),
		networkName,
	}

	cmd := exec.CommandContext(ctx, d.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to create network: %s - %w", string(output), err)
	}

	networkID := strings.TrimSpace(string(output))

	d.logger.Info("Created sandbox network", map[string]interface{}{
		"sandbox_id":   sandboxID,
		"network_name": networkName,
		"network_id":   networkID,
	})

	return networkName, networkID, nil
}

// DeployCollector deploys an OpenTelemetry collector container
func (d *DockerOrchestrator) DeployCollector(ctx context.Context, config DeployCollectorConfig) (*CollectorInfo, error) {
	containerName := fmt.Sprintf("sandbox-collector-%s", config.SandboxID)

	// Create config file in shared directory
	// Use environment variable for config directory, default to /sandbox-configs
	configDir := os.Getenv("SANDBOX_CONFIGS_DIR")
	if configDir == "" {
		configDir = "/sandbox-configs"
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, fmt.Sprintf("%s.yaml", config.SandboxID))
	if err := os.WriteFile(configPath, []byte(config.Config), 0644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	// Get the host path for volume mounting (for Docker-in-Docker scenarios)
	// When running in Docker, we need the host path for volume mounts
	configHostPath := os.Getenv("SANDBOX_CONFIGS_HOST_PATH")
	if configHostPath == "" {
		// If not set, assume we're not in Docker and use the same path
		configHostPath = configDir
	}
	configHostFilePath := filepath.Join(configHostPath, fmt.Sprintf("%s.yaml", config.SandboxID))

	// Determine collector image
	image := "otel/opentelemetry-collector-contrib"
	if config.CollectorVersion != "" && config.CollectorVersion != "latest" {
		image = fmt.Sprintf("%s:%s", image, config.CollectorVersion)
	} else {
		image = fmt.Sprintf("%s:latest", image)
	}

	// Build docker run command
	args := []string{
		"run",
		"-d",
		"--name", containerName,
		"--network", config.NetworkName,
		"--network-alias", "collector", // Allow other containers to reach it via "collector"
		"-v", fmt.Sprintf("%s:/etc/otelcol-contrib/config.yaml:ro", configHostFilePath),
		"--label", fmt.Sprintf("sandbox.id=%s", config.SandboxID),
		"--label", "sandbox.component=collector",
		// Expose Prometheus metrics endpoint
		"-p", "8888", // Prometheus metrics
		image,
		"--config=/etc/otelcol-contrib/config.yaml",
	}

	cmd := exec.CommandContext(ctx, d.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to start collector container: %s - %w", string(output), err)
	}

	containerID := strings.TrimSpace(string(output))

	// Wait a moment for container to start
	time.Sleep(2 * time.Second)

	// Check container status
	status, err := d.getContainerStatus(ctx, containerID)
	if err != nil {
		// Container might have exited, check logs
		logs, _ := d.GetContainerLogs(ctx, containerID, 50)
		logStr := ""
		for _, log := range logs {
			logStr += log.Message + "\n"
		}
		return nil, fmt.Errorf("container failed to start: %w\nLogs: %s", err, logStr)
	}

	d.logger.Info("Deployed collector", map[string]interface{}{
		"sandbox_id":     config.SandboxID,
		"container_id":   containerID,
		"container_name": containerName,
	})

	return &CollectorInfo{
		ContainerID:   containerID,
		ContainerName: containerName,
		Status:        status,
	}, nil
}

// StartTelemetryGeneration starts telemetrygen containers
func (d *DockerOrchestrator) StartTelemetryGeneration(ctx context.Context, config StartTelemetryConfig) (TelemetryGeneratorContainerInfo, error) {
	var info TelemetryGeneratorContainerInfo

	// Determine OTLP endpoint format
	otlpEndpoint := config.TelemetryConfig.OTLPEndpoint
	if !strings.Contains(otlpEndpoint, ":") {
		if config.TelemetryConfig.OTLPProtocol == "http" {
			otlpEndpoint = otlpEndpoint + ":4318"
		} else {
			otlpEndpoint = otlpEndpoint + ":4317"
		}
	}

	// Start traces generator
	if config.TelemetryConfig.GenerateTraces {
		containerID, err := d.startTracesGenerator(ctx, config.SandboxID, config.NetworkName, otlpEndpoint, config.TelemetryConfig, config.Duration)
		if err != nil {
			return info, fmt.Errorf("failed to start traces generator: %w", err)
		}
		info.TracesContainerID = containerID
	}

	// Start metrics generator
	if config.TelemetryConfig.GenerateMetrics {
		containerID, err := d.startMetricsGenerator(ctx, config.SandboxID, config.NetworkName, otlpEndpoint, config.TelemetryConfig, config.Duration)
		if err != nil {
			return info, fmt.Errorf("failed to start metrics generator: %w", err)
		}
		info.MetricsContainerID = containerID
	}

	// Start logs generator
	if config.TelemetryConfig.GenerateLogs {
		containerID, err := d.startLogsGenerator(ctx, config.SandboxID, config.NetworkName, otlpEndpoint, config.TelemetryConfig, config.Duration)
		if err != nil {
			return info, fmt.Errorf("failed to start logs generator: %w", err)
		}
		info.LogsContainerID = containerID
	}

	d.logger.Info("Started telemetry generation", map[string]interface{}{
		"sandbox_id": config.SandboxID,
		"duration":   config.Duration,
	})

	return info, nil
}

// startTracesGenerator starts a telemetrygen traces container
func (d *DockerOrchestrator) startTracesGenerator(ctx context.Context, sandboxID, networkName, otlpEndpoint string, config TelemetryConfig, duration time.Duration) (string, error) {
	containerName := fmt.Sprintf("sandbox-traces-%s", sandboxID)

	rate := config.TraceRate
	if rate == 0 {
		rate = 1 // Default 1 trace per second
	}

	// Build args for telemetrygen
	args := []string{
		"run",
		"--rm", // Auto-remove after completion
		"--name", containerName,
		"--network", networkName,
		"--label", fmt.Sprintf("sandbox.id=%s", sandboxID),
		"--label", "sandbox.component=telemetrygen",
		"ghcr.io/open-telemetry/opentelemetry-collector-contrib/telemetrygen:latest",
		"traces",
		"--otlp-endpoint", otlpEndpoint,
		"--otlp-insecure",
		"--rate", strconv.Itoa(rate),
		"--duration", duration.String(),
	}

	// Add custom attributes if provided
	for key, value := range config.TraceAttributes {
		args = append(args, "--traces-attribute", fmt.Sprintf("%s=%s", key, value))
	}

	cmd := exec.CommandContext(ctx, d.dockerPath, args...)

	// Start container in background
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start traces generator: %w", err)
	}

	// Get container ID
	containerID, err := d.getContainerIDByName(ctx, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to get traces generator container ID: %w", err)
	}

	return containerID, nil
}

// startMetricsGenerator starts a telemetrygen metrics container
func (d *DockerOrchestrator) startMetricsGenerator(ctx context.Context, sandboxID, networkName, otlpEndpoint string, config TelemetryConfig, duration time.Duration) (string, error) {
	containerName := fmt.Sprintf("sandbox-metrics-%s", sandboxID)

	rate := config.MetricRate
	if rate == 0 {
		rate = 1 // Default 1 metric per second
	}

	args := []string{
		"run",
		"--rm",
		"--name", containerName,
		"--network", networkName,
		"--label", fmt.Sprintf("sandbox.id=%s", sandboxID),
		"--label", "sandbox.component=telemetrygen",
		"ghcr.io/open-telemetry/opentelemetry-collector-contrib/telemetrygen:latest",
		"metrics",
		"--otlp-endpoint", otlpEndpoint,
		"--otlp-insecure",
		"--rate", strconv.Itoa(rate),
		"--duration", duration.String(),
	}

	cmd := exec.CommandContext(ctx, d.dockerPath, args...)

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start metrics generator: %w", err)
	}

	containerID, err := d.getContainerIDByName(ctx, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to get metrics generator container ID: %w", err)
	}

	return containerID, nil
}

// startLogsGenerator starts a telemetrygen logs container
func (d *DockerOrchestrator) startLogsGenerator(ctx context.Context, sandboxID, networkName, otlpEndpoint string, config TelemetryConfig, duration time.Duration) (string, error) {
	containerName := fmt.Sprintf("sandbox-logs-%s", sandboxID)

	rate := config.LogRate
	if rate == 0 {
		rate = 1 // Default 1 log per second
	}

	args := []string{
		"run",
		"--rm",
		"--name", containerName,
		"--network", networkName,
		"--label", fmt.Sprintf("sandbox.id=%s", sandboxID),
		"--label", "sandbox.component=telemetrygen",
		"ghcr.io/open-telemetry/opentelemetry-collector-contrib/telemetrygen:latest",
		"logs",
		"--otlp-endpoint", otlpEndpoint,
		"--otlp-insecure",
		"--rate", strconv.Itoa(rate),
		"--duration", duration.String(),
	}

	cmd := exec.CommandContext(ctx, d.dockerPath, args...)

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start logs generator: %w", err)
	}

	containerID, err := d.getContainerIDByName(ctx, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to get logs generator container ID: %w", err)
	}

	return containerID, nil
}

// GetContainerLogs retrieves logs from a container
func (d *DockerOrchestrator) GetContainerLogs(ctx context.Context, containerID string, tailLines int) ([]LogEntry, error) {
	args := []string{
		"logs",
		"--tail", strconv.Itoa(tailLines),
		"--timestamps",
		containerID,
	}

	cmd := exec.CommandContext(ctx, d.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	return d.parseContainerLogs(string(output)), nil
}

// parseContainerLogs parses Docker log output into structured log entries
func (d *DockerOrchestrator) parseContainerLogs(logOutput string) []LogEntry {
	var logs []LogEntry

	// Docker log format: TIMESTAMP STREAM MESSAGE
	// e.g., 2025-01-29T10:30:45.123456789Z stdout INFO: message here
	timestampRegex := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+(.+)$`)

	lines := strings.Split(logOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := timestampRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			timestamp, _ := time.Parse(time.RFC3339Nano, matches[1])
			message := matches[2]

			// Try to extract log level
			level := "info"
			if strings.Contains(strings.ToLower(message), "error") {
				level = "error"
			} else if strings.Contains(strings.ToLower(message), "warn") {
				level = "warn"
			} else if strings.Contains(strings.ToLower(message), "debug") {
				level = "debug"
			}

			logs = append(logs, LogEntry{
				Timestamp: timestamp,
				Level:     level,
				Message:   message,
			})
		} else {
			// If we can't parse timestamp, just add the line as is
			logs = append(logs, LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   line,
			})
		}
	}

	return logs
}

// GetCollectorMetrics retrieves internal metrics from the collector
func (d *DockerOrchestrator) GetCollectorMetrics(ctx context.Context, containerID string) (*CollectorMetrics, error) {
	// Get container IP
	containerIP, err := d.getContainerIP(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container IP: %w", err)
	}

	// Fetch metrics from Prometheus endpoint
	// Note: This is a simplified version. In production, you'd use a proper Prometheus client
	metricsURL := fmt.Sprintf("http://%s:8888/metrics", containerIP)

	// Use docker exec to curl from within the container
	args := []string{
		"exec",
		containerID,
		"wget",
		"-q",
		"-O", "-",
		"http://localhost:8888/metrics",
	}

	cmd := exec.CommandContext(ctx, d.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Metrics endpoint might not be ready yet
		return &CollectorMetrics{}, nil
	}

	// Parse Prometheus metrics
	metrics := d.parsePrometheusMetrics(string(output))

	d.logger.Debug("Retrieved collector metrics", map[string]interface{}{
		"container_id": containerID,
		"metrics_url":  metricsURL,
	})

	return metrics, nil
}

// parsePrometheusMetrics parses Prometheus format metrics
func (d *DockerOrchestrator) parsePrometheusMetrics(metricsOutput string) *CollectorMetrics {
	metrics := &CollectorMetrics{}

	// This is a simplified parser. A real implementation would use a Prometheus client library
	lines := strings.Split(metricsOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse metric lines: metric_name{labels} value
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		metricName := parts[0]
		valueStr := parts[len(parts)-1]
		value, _ := strconv.ParseFloat(valueStr, 64)

		// Map metrics to our structure
		switch {
		case strings.Contains(metricName, "receiver_accepted_spans"):
			metrics.ReceiverAcceptedSpans = int64(value)
		case strings.Contains(metricName, "receiver_refused_spans"):
			metrics.ReceiverRefusedSpans = int64(value)
		case strings.Contains(metricName, "exporter_sent_spans"):
			metrics.ExporterSentSpans = int64(value)
		case strings.Contains(metricName, "exporter_send_failed_spans"):
			metrics.ExporterFailedSpans = int64(value)
		case strings.Contains(metricName, "queue_size"):
			metrics.QueueSize = int64(value)
		case strings.Contains(metricName, "process_resident_memory_bytes"):
			metrics.MemoryUsageMB = value / 1024 / 1024
		}
	}

	return metrics
}

// StopCollector stops a collector container
func (d *DockerOrchestrator) StopCollector(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, d.dockerPath, "stop", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop collector: %w", err)
	}

	// Remove container
	cmd = exec.CommandContext(ctx, d.dockerPath, "rm", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove collector: %w", err)
	}

	d.logger.Info("Stopped collector", map[string]interface{}{
		"container_id": containerID,
	})

	return nil
}

// StopTelemetryGenerators stops telemetry generator containers
func (d *DockerOrchestrator) StopTelemetryGenerators(ctx context.Context, info TelemetryGeneratorContainerInfo) error {
	if info.TracesContainerID != "" {
		_ = d.stopContainer(ctx, info.TracesContainerID)
	}
	if info.MetricsContainerID != "" {
		_ = d.stopContainer(ctx, info.MetricsContainerID)
	}
	if info.LogsContainerID != "" {
		_ = d.stopContainer(ctx, info.LogsContainerID)
	}

	return nil
}

// CleanupNetwork removes a sandbox network
func (d *DockerOrchestrator) CleanupNetwork(ctx context.Context, sandboxID string) error {
	networkName := fmt.Sprintf("sandbox-%s", sandboxID)

	cmd := exec.CommandContext(ctx, d.dockerPath, "network", "rm", networkName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove network: %w", err)
	}

	d.logger.Info("Removed sandbox network", map[string]interface{}{
		"sandbox_id":   sandboxID,
		"network_name": networkName,
	})

	return nil
}

// Helper methods

func (d *DockerOrchestrator) getContainerStatus(ctx context.Context, containerID string) (string, error) {
	cmd := exec.CommandContext(ctx, d.dockerPath, "inspect", "-f", "{{.State.Status}}", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (d *DockerOrchestrator) getContainerIP(ctx context.Context, containerID string) (string, error) {
	cmd := exec.CommandContext(ctx, d.dockerPath, "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (d *DockerOrchestrator) getContainerIDByName(ctx context.Context, containerName string) (string, error) {
	// Give it a moment to register
	time.Sleep(500 * time.Millisecond)

	cmd := exec.CommandContext(ctx, d.dockerPath, "ps", "-aqf", fmt.Sprintf("name=%s", containerName))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (d *DockerOrchestrator) stopContainer(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, d.dockerPath, "stop", containerID)
	_ = cmd.Run() // Ignore errors, container might already be stopped

	cmd = exec.CommandContext(ctx, d.dockerPath, "rm", containerID)
	return cmd.Run()
}

// InspectContainer returns detailed container information
func (d *DockerOrchestrator) InspectContainer(ctx context.Context, containerID string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, d.dockerPath, "inspect", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse inspect output: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("container not found")
	}

	return result[0], nil
}
