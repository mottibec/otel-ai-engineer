package deployers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	dc "github.com/mottibechhofer/otel-ai-engineer/tools/dockerclient"
)

// DockerDeployer handles Docker-based collector deployments
type DockerDeployer struct {
	dockerPath   string
	dockerClient *dc.Client
}

// NewDockerDeployer creates a new Docker deployer
func NewDockerDeployer() (*DockerDeployer, error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	// Create Docker client for network operations
	dockerClient, err := dc.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerDeployer{
		dockerPath:   dockerPath,
		dockerClient: dockerClient,
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

	// Create config file in a location accessible from host Docker daemon
	// When running inside Docker container:
	// - Write to container mount point (e.g., /otel-configs)
	// - Use host path (from env var) when mounting in docker run
	containerConfigDir := os.Getenv("OTEL_CONFIGS_DIR")
	if containerConfigDir == "" {
		containerConfigDir = "/tmp/otel-configs"
	}

	// Get host path for Docker volume mounting (needed when running inside container)
	hostConfigDir := os.Getenv("OTEL_CONFIGS_HOST_PATH")

	// Create directory in container
	if err := os.MkdirAll(containerConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(containerConfigDir, fmt.Sprintf("%s.yaml", collectorID))
	if err := os.WriteFile(configPath, []byte(config.YAMLConfig), 0644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	// Verify the file exists and is actually a file (not a directory)
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to verify config file exists: %w", err)
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("config path is a directory, not a file: %s", configPath)
	}

	// Determine path to use for Docker volume mount
	var absConfigPath string
	if hostConfigDir != "" {
		// Use host path - this is what Docker daemon on host will see
		absConfigPath = filepath.Join(hostConfigDir, fmt.Sprintf("%s.yaml", collectorID))
		// Ensure it's absolute
		if !filepath.IsAbs(absConfigPath) {
			var err error
			absConfigPath, err = filepath.Abs(absConfigPath)
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path for config file: %w", err)
			}
		}
	} else {
		// Not in containerized environment, use container path as-is
		var err error
		absConfigPath, err = filepath.Abs(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for config file: %w", err)
		}
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

	// Ensure network exists before deploying
	ctx := context.Background()
	if err := d.dockerClient.EnsureNetwork(ctx, network); err != nil {
		return nil, fmt.Errorf("failed to ensure network '%s': %w", network, err)
	}

	// Brief delay to ensure network is fully available for Docker CLI
	time.Sleep(200 * time.Millisecond)

	// Build docker run command
	args := []string{
		"run",
		"-d",
		"--name", containerName,
		"--network", network,
		"-p", "4317", // OTLP gRPC
		"-p", "4318", // OTLP HTTP
		"-v", fmt.Sprintf("%s:/etc/otelcol/config.yaml:ro", absConfigPath),
		"-e", fmt.Sprintf("OTEL_OPAMP_SERVER=%s", lawrenceURL),
		"-e", fmt.Sprintf("OTEL_AGENT_ID=%s", collectorID),
		image,
		"--config=/etc/otelcol/config.yaml",
	}

	// Execute docker run
	cmd := exec.CommandContext(ctx, d.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Provide more detailed error for network issues
		errMsg := string(output)
		if strings.Contains(errMsg, "network") && strings.Contains(errMsg, "not found") {
			return nil, fmt.Errorf("network '%s' not found - this may be a timing issue. Original error: %s - %w. Try ensuring the network exists manually with: docker network create %s", network, errMsg, err, network)
		}
		return nil, fmt.Errorf("failed to start docker container: %s - %w", errMsg, err)
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
	containerConfigDir := os.Getenv("OTEL_CONFIGS_DIR")
	if containerConfigDir == "" {
		containerConfigDir = "/tmp/otel-configs"
	}
	configPath := filepath.Join(containerConfigDir, fmt.Sprintf("%s.yaml", collectorID))
	_ = os.Remove(configPath)

	return nil
}

// List lists all collector containers (including stopped/exited ones)
func (d *DockerDeployer) List() ([]CollectorInfo, error) {
	// List all otel-collector-* containers (including stopped ones with -a)
	args := []string{
		"ps",
		"-a", // Show all containers, including stopped ones
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

		// Get config path using the same logic as deployment
		containerConfigDir := os.Getenv("OTEL_CONFIGS_DIR")
		if containerConfigDir == "" {
			containerConfigDir = "/tmp/otel-configs"
		}
		configPath := filepath.Join(containerConfigDir, fmt.Sprintf("%s.yaml", collectorID))

		collectors = append(collectors, CollectorInfo{
			CollectorID:   collectorID,
			CollectorName: collectorID,
			TargetType:    string(TargetDocker),
			Status:        status,
			DeployedAt:    deployedAt,
			ConfigPath:    configPath,
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
