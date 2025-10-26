package deployers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DockerDeployer handles Docker-based collector deployments
type DockerDeployer struct {
	dockerPath string
}

// NewDockerDeployer creates a new Docker deployer
func NewDockerDeployer() (*DockerDeployer, error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	return &DockerDeployer{
		dockerPath: dockerPath,
	}, nil
}

// GetTargetType returns the target type
func (d *DockerDeployer) GetTargetType() TargetType {
	return TargetDocker
}

// Deploy deploys a collector as a Docker container
func (d *DockerDeployer) Deploy(config DeploymentConfig) (*DeploymentResult, error) {
	// Generate unique collector ID
	collectorID := fmt.Sprintf("%s-%d", config.CollectorName, time.Now().Unix())
	containerName := fmt.Sprintf("otel-collector-%s", collectorID)

	// Create temporary config file
	configDir := "/tmp/otel-configs"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, fmt.Sprintf("%s.yaml", collectorID))
	if err := os.WriteFile(configPath, []byte(config.YAMLConfig), 0644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	// Get deployment parameters
	network := "otel-network" // default
	if net, ok := config.Parameters["network"].(string); ok && net != "" {
		network = net
	}

	image := "otel/opentelemetry-collector-contrib:latest"
	if img, ok := config.Parameters["image"].(string); ok && img != "" {
		image = img
	}

	lawrenceURL := "http://lawrence:4320"
	if url, ok := config.Parameters["lawrence_url"].(string); ok && url != "" {
		lawrenceURL = url
	}

	// Build docker run command
	args := []string{
		"run",
		"-d",
		"--name", containerName,
		"--network", network,
		"-v", fmt.Sprintf("%s:/etc/otelcol/config.yaml:ro", configPath),
		"-e", fmt.Sprintf("OTEL_OPAMP_SERVER=%s", lawrenceURL),
		"-e", fmt.Sprintf("OTEL_AGENT_ID=%s", collectorID),
		image,
		"--config=/etc/otelcol/config.yaml",
	}

	// Execute docker run
	cmd := exec.Command(d.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to start docker container: %s - %w", string(output), err)
	}

	// Wait a moment for container to start
	time.Sleep(2 * time.Second)

	// Check container status
	status, err := d.getContainerStatus(containerName)
	if err != nil {
		// Container might have exited, check logs
		logs, logErr := d.getContainerLogs(containerName)
		if logErr != nil {
			return nil, fmt.Errorf("container failed to start: %w (logs unavailable)", err)
		}
		return nil, fmt.Errorf("container failed to start: %w\nLogs: %s", err, logs)
	}

	return &DeploymentResult{
		Success:     true,
		CollectorID: collectorID,
		TargetType:  string(TargetDocker),
		Status:      status,
		Message:     fmt.Sprintf("Collector deployed in container %s", containerName),
		DeployedAt:  time.Now(),
	}, nil
}

// Stop stops and removes a collector container
func (d *DockerDeployer) Stop(collectorID string, params map[string]interface{}) error {
	containerName := fmt.Sprintf("otel-collector-%s", collectorID)

	// First, try to stop the container
	stopCmd := exec.Command(d.dockerPath, "stop", containerName)
	if err := stopCmd.Run(); err != nil && !strings.Contains(err.Error(), "No such container") {
		// Log but continue to remove
	}

	// Remove the container
	rmCmd := exec.Command(d.dockerPath, "rm", containerName)
	if err := rmCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	// Clean up config file
	configPath := filepath.Join("/tmp/otel-configs", fmt.Sprintf("%s.yaml", collectorID))
	_ = os.Remove(configPath)

	return nil
}

// List lists all running collector containers
func (d *DockerDeployer) List() ([]CollectorInfo, error) {
	// List all otel-collector-* containers
	args := []string{
		"ps",
		"--filter", "name=otel-collector-",
		"--format", "{{.Names}},{{.Status}},{{.CreatedAt}}",
	}

	cmd := exec.Command(d.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var collectors []CollectorInfo
	if len(strings.TrimSpace(string(output))) == 0 {
		return collectors, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}

		containerName := parts[0]
		status := parts[1]

		// Extract collector ID from container name (otel-collector-{id})
		collectorID := strings.TrimPrefix(containerName, "otel-collector-")

		// Try to parse deployment time from created at field
		var deployedAt time.Time
		if len(parts) >= 3 {
			// CreatedAt field from docker ps
			// This is approximate, actual deployment time would need tracking
			deployedAt = time.Now() // Fallback
		}

		collectors = append(collectors, CollectorInfo{
			CollectorID:  collectorID,
			CollectorName: collectorID,
			TargetType:    string(TargetDocker),
			Status:        status,
			DeployedAt:    deployedAt,
			ConfigPath:    fmt.Sprintf("/tmp/otel-configs/%s.yaml", collectorID),
		})
	}

	return collectors, nil
}

// getContainerStatus gets the status of a container
func (d *DockerDeployer) getContainerStatus(containerName string) (string, error) {
	cmd := exec.Command(d.dockerPath, "inspect", "-f", "{{.State.Status}}", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getContainerLogs retrieves logs from a container
func (d *DockerDeployer) getContainerLogs(containerName string) (string, error) {
	cmd := exec.Command(d.dockerPath, "logs", "--tail", "50", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

